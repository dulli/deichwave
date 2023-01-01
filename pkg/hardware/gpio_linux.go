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

func (h *GPIO) Setup(cfg common.Config) error {
	for _, setup := range cfg.GPIO {
		switch setup.Type {
		case "toggles":
			// Each input has its own pin, config object can contain multiple inputs
			for idx, pin := range setup.Pins {
				actions := strings.Split(setup.Actions[idx], ":")
				err := h.setupToggle(setup.Chip, pin, setup.Debounce, actions)
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
			err := h.setupRotary(setup.Chip, setup.Pins, setup.Debounce, actions)
			if err != nil {
				return err
			}
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
func (h *GPIO) setupToggle(chip string, pin int, debounce int, actions []string) error {
	log.WithFields(log.Fields{
		"chip":    chip,
		"pin":     pin,
		"actions": actions,
	}).Debug("Setting up toggle")
	line, err := gpiod.RequestLine(chip, pin,
		gpiod.WithPullUp,
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
			log.WithFields(log.Fields{
				"ev":     ev,
				"driver": "gpio",
			}).Info("Button pressed")
		}),
		gpiod.WithBothEdges)
	h.lines = append(h.lines, line)
	return err
}

// Rotaries switch between left and right states and have an additional toggle when pressed
func (h *GPIO) setupRotary(chip string, pins []int, debounce int, actions []string) error {
	log.WithFields(log.Fields{
		"chip":    chip,
		"pin":     pins,
		"actions": actions,
	}).Debug("Setting up rotary")
	rotflag := false
	line, err := gpiod.RequestLine(chip, pins[0],
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
			rotflag = ev.Type == gpiod.LineEventRisingEdge
			log.WithFields(log.Fields{
				"ev":     ev,
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
		gpiod.WithDebounce(time.Microsecond*time.Duration(debounce)),
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
			if ev.Type == gpiod.LineEventFallingEdge {
				rotflag = false
				return
			}
			if rotflag {
				log.WithFields(log.Fields{
					"ev":     ev,
					"flag":   rotflag,
					"driver": "gpio",
				}).Info("Rotary turned left")
			} else {
				log.WithFields(log.Fields{
					"ev":     ev,
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
		err = h.setupToggle(chip, pins[2], debounce, actions[2:3])
	}
	return err
}
