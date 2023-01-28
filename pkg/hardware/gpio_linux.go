package hardware

import (
	"strings"
	"time"

	"github.com/dulli/deichwave/pkg/common"
	log "github.com/sirupsen/logrus"

	"github.com/warthog618/gpiod"
)

type GPIO struct {
	lines []*gpiod.Line
}

// As the PCF8574 doesn't support gpiod's debouncing, we need to implement our own
type debouncer struct {
	last     time.Time
	duration time.Duration
}

func (d *debouncer) active(ev gpiod.LineEvent) bool {
	// TODO: only debounce if a falling edge follows on a rising one
	return time.Since(d.last) < d.duration
}

func (h *GPIO) Setup(cfg common.Config) error {
	for _, setup := range cfg.GPIO {
		// Closure to keep debouncers separated
		err := func() error {
			deb := debouncer{
				last:     time.Now().Add(-time.Second),
				duration: time.Duration(setup.Debounce) * time.Microsecond,
			}
			switch setup.Type {
			case "toggles":
				// Each input has its own pin, config object can contain multiple inputs
				for idx, pin := range setup.Pins {
					actions := strings.Split(setup.Actions[idx], ":")
					err := h.setupToggle(setup.Chip, pin, deb, actions)
					if err != nil {
						return err
					}
				}
			case "rotary":
				// Each input has two to three pins, config object is a single input
				actions := strings.Split(setup.Actions[0], ":")
				if len(setup.Actions) > 1 {
					actions = append(actions, strings.Split(setup.Actions[1], ":")...)
				}
				err := h.setupRotary(setup.Chip, setup.Pins, deb, actions)
				if err != nil {
					return err
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *GPIO) Check() error {
	return nil
}

func (h *GPIO) Close() {
	for _, line := range h.lines {
		line.Close()
	}
}

// Toggles switch between on and off states, so either buttons or switches
func (h *GPIO) setupToggle(chip string, pin int, deb debouncer, actions []string) error {
	log.WithFields(log.Fields{
		"chip":    chip,
		"pin":     pin,
		"actions": actions,
	}).Debug("Setting up toggle")
	line, err := gpiod.RequestLine(chip, pin,
		gpiod.WithPullUp,
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
			if deb.active(ev) {
				log.WithFields(log.Fields{
					"chip":   chip,
					"pin":    pin,
					"driver": "gpio",
				}).Warn("Toggle debounced")
				return
			}
			deb.last = time.Now()

			switch ev.Type {
			case gpiod.LineEventFallingEdge:
				// action[0]
				log.WithFields(log.Fields{
					"chip":   chip,
					"pin":    pin,
					"driver": "gpio",
				}).Info("Button pressed")
			case gpiod.LineEventRisingEdge:
				//action[1]
				log.WithFields(log.Fields{
					"chip":   chip,
					"pin":    pin,
					"driver": "gpio",
				}).Info("Button released")
			}

		}),
		gpiod.WithBothEdges)
	h.lines = append(h.lines, line)
	return err
}

// Rotaries switch between left and right states and have an additional toggle when pressed
func (h *GPIO) setupRotary(chip string, pins []int, deb debouncer, actions []string) error {
	log.WithFields(log.Fields{
		"chip":    chip,
		"pin":     pins,
		"actions": actions,
	}).Debug("Setting up rotary")
	rotflag := false
	line, err := gpiod.RequestLine(chip, pins[0],
		gpiod.WithPullUp,
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
			if deb.active(ev) {
				log.WithFields(log.Fields{
					"chip":   chip,
					"pin":    pins,
					"driver": "gpio",
				}).Warn("Rotary low debounced")
				return
			}
			deb.last = time.Now()

			rotflag = ev.Type == gpiod.LineEventRisingEdge
			log.WithFields(log.Fields{
				"chip":   chip,
				"pin":    pins,
				"flag":   rotflag,
				"driver": "gpio",
			}).Info("Rotary turned low")
		}),
		gpiod.WithBothEdges)
	if err != nil {
		return err
	}
	h.lines = append(h.lines, line)

	line, err = gpiod.RequestLine(chip, pins[1],
		gpiod.WithPullUp,
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
			if deb.active(ev) {
				log.WithFields(log.Fields{
					"chip":   chip,
					"pin":    pins,
					"driver": "gpio",
				}).Warn("Rotary high debounced")
				return
			}
			deb.last = time.Now()

			if ev.Type == gpiod.LineEventFallingEdge {
				rotflag = false
				return
			}
			if rotflag {
				// action[0]
				log.WithFields(log.Fields{
					"chip":   chip,
					"pin":    pins,
					"flag":   rotflag,
					"driver": "gpio",
				}).Info("Rotary turned left")
			} else {
				// action[1]
				log.WithFields(log.Fields{
					"chip":   chip,
					"pin":    pins,
					"flag":   rotflag,
					"driver": "gpio",
				}).Info("Rotary turned right")
			}
		}),
		gpiod.WithBothEdges)
	if err != nil {
		return err
	}
	h.lines = append(h.lines, line)

	// Setup the toggle component
	if len(pins) > 2 {
		err = h.setupToggle(chip, pins[2], deb, actions[2:3])
	}
	return err
}
