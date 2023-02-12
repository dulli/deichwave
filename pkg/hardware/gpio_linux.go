package hardware

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dulli/deichwave/pkg/common"
	"github.com/dulli/deichwave/pkg/rest"
	log "github.com/sirupsen/logrus"

	"github.com/warthog618/gpiod"
)

type GPIO struct {
	lines []*gpiod.Line
	srv   rest.Server
}

// As the PCF8574 doesn't support gpiod's debouncing, we need to implement our own
type debouncer struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer
}
type debounced func(f func())

func (d *debouncer) add(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.after, f)
}

func newDebouncer(duration int) debounced {
	d := &debouncer{after: time.Duration(duration) * time.Microsecond}

	return func(f func()) {
		d.add(f)
	}
}

func (h *GPIO) Setup(cfg common.Config, srv rest.Server) error {
	h.srv = srv
	initialized := 0
	for _, setup := range cfg.GPIO {
		switch setup.Type {
		case "toggles":
			// Each input has its own pin, config object can contain multiple inputs
			for idx, pin := range setup.Pins {
				actions := strings.Split(setup.Actions[idx], ":")
				err := h.setupToggle(setup.Chip, pin, setup.Debounce, actions)
				if err != nil {
					log.WithFields(log.Fields{
						"driver": "gpio",
						"err":    err,
					}).Warn("Failed to setup toggle")
					continue
				}
				initialized += 1
			}
		case "rotary":
			// Each input has two to three pins, config object is a single input
			actions := strings.Split(setup.Actions[0], ":")
			if len(setup.Actions) > 1 {
				actions = append(actions, strings.Split(setup.Actions[1], ":")...)
			}
			err := h.setupRotary(setup.Chip, setup.Pins, setup.Debounce, actions)
			if err != nil {
				log.WithFields(log.Fields{
					"driver": "gpio",
					"err":    err,
				}).Warn("Failed to setup rotary")
				continue
			}
			initialized += 1
		}
	}
	if initialized == 0 {
		return fmt.Errorf("No GPIO input could be initialized")
	}
	return nil
}

func (h *GPIO) Check() error {
	return nil
}

func (h *GPIO) Close() {
	for _, line := range h.lines {
		if line != nil {
			line.Close()
		}
	}
}

// Toggles switch between on and off states, so either buttons or switches
func (h *GPIO) setupToggle(chip string, pin int, debounce int, actions []string) error {
	log.WithFields(log.Fields{
		"chip":    chip,
		"pin":     pin,
		"actions": actions,
	}).Debug("Setting up toggle")
	dh := newDebouncer(debounce)
	dl := newDebouncer(debounce)
	var apiErr error
	line, err := gpiod.RequestLine(chip, pin,
		gpiod.WithPullUp,
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
			switch ev.Type {
			case gpiod.LineEventFallingEdge:
				dh(func() {
					apiErr = h.srv.DoAction(actions[0])
					log.WithFields(log.Fields{
						"chip":   chip,
						"pin":    pin,
						"err":    apiErr,
						"driver": "gpio",
					}).Info("Button pressed")
				})
			case gpiod.LineEventRisingEdge:
				dl(func() {
					if len(actions) < 2 {
						return
					}
					apiErr = h.srv.DoAction(actions[1])
					log.WithFields(log.Fields{
						"chip":   chip,
						"pin":    pin,
						"err":    apiErr,
						"driver": "gpio",
					}).Info("Button released")
				})
			}
		}),
		gpiod.WithBothEdges)
	h.lines = append(h.lines, line)
	return err
}

// Rotaries switch between left and right states and have an additional toggle when pressed
func (h *GPIO) setupRotary(chip string, pins []int, debounce int, actions []string) error {
	dh := newDebouncer(debounce)
	dl := newDebouncer(debounce)
	var apiErr error
	log.WithFields(log.Fields{
		"chip":    chip,
		"pin":     pins,
		"actions": actions,
	}).Debug("Setting up rotary")
	rotflag := false
	line, err := gpiod.RequestLine(chip, pins[0],
		gpiod.WithPullUp,
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
			dl(func() {
				rotflag = ev.Type == gpiod.LineEventRisingEdge
				log.WithFields(log.Fields{
					"chip":   chip,
					"pin":    pins,
					"flag":   rotflag,
					"driver": "gpio",
				}).Info("Rotary turned low")
			})
		}),
		gpiod.WithBothEdges)
	if err != nil {
		return err
	}
	h.lines = append(h.lines, line)

	line, err = gpiod.RequestLine(chip, pins[1],
		gpiod.WithPullUp,
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
			dh(func() {
				if ev.Type == gpiod.LineEventFallingEdge {
					rotflag = false
					return
				}
				if rotflag {
					apiErr = h.srv.DoAction(actions[0])
					log.WithFields(log.Fields{
						"chip":   chip,
						"pin":    pins,
						"flag":   rotflag,
						"err":    apiErr,
						"driver": "gpio",
					}).Info("Rotary turned left")
				} else {
					if len(actions) < 2 {
						return
					}
					apiErr = h.srv.DoAction(actions[1])
					log.WithFields(log.Fields{
						"chip":   chip,
						"pin":    pins,
						"flag":   rotflag,
						"err":    apiErr,
						"driver": "gpio",
					}).Info("Rotary turned right")
				}
			})
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
