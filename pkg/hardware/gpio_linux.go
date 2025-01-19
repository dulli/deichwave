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
	var (
		apiErr            error
		mu                sync.Mutex
		enabled           bool = true
		lastReceivedType  gpiod.LineEventType
		lastTriggeredType gpiod.LineEventType
	)
	debounceTime := time.Duration(debounce) * time.Microsecond

	press := func() {
		apiErr = h.srv.DoAction(actions[0])
		log.WithFields(log.Fields{
			"chip":   chip,
			"pin":    pin,
			"err":    apiErr,
			"driver": "gpio",
		}).Debug("Toggle on")
	}
	release := func() {
		if len(actions) < 2 {
			return
		}
		apiErr = h.srv.DoAction(actions[1])
		log.WithFields(log.Fields{
			"chip":   chip,
			"pin":    pin,
			"err":    apiErr,
			"driver": "gpio",
		}).Debug("Toggle off")
	}

	// Handler for (active) interrupts and expiring debounce timers
	update := func(t *gpiod.LineEventType) {
		if *t == lastTriggeredType {
			return
		}
		lastTriggeredType = *t
		if *t == gpiod.LineEventRisingEdge {
			press()
		} else {
			release()
		}
	}

	// Handler for hardware events
	handler := func(ev gpiod.LineEvent) {
		mu.Lock()
		defer mu.Unlock()
		lastReceivedType = ev.Type
		if !enabled {
			return
		}

		// Start a timer to debounce by disabling the interrupt for some time
		enabled = false
		time.AfterFunc(debounceTime, func() {
			mu.Lock()
			defer mu.Unlock()
			enabled = true
			go update(&lastReceivedType)
		})

		// Handle this event
		go update(&ev.Type)
	}
	line, err := gpiod.RequestLine(chip, pin, gpiod.WithPullUp,
		gpiod.WithEventHandler(handler), gpiod.WithBothEdges)
	h.lines = append(h.lines, line)
	return err
}

// Rotaries switch between left and right states and have an additional toggle when pressed
func (h *GPIO) setupRotary(chip string, pins []int, debounce int, actions []string) error {
	debounceTime := time.Duration(debounce) * time.Microsecond
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
			rotflag = ev.Type == gpiod.LineEventRisingEdge
		}),
		gpiod.WithBothEdges)
	if err != nil {
		return err
	}
	h.lines = append(h.lines, line)

	line, err = gpiod.RequestLine(chip, pins[1],
		gpiod.WithPullUp,
		gpiod.WithDebounce(debounceTime),
		gpiod.WithEventHandler(func(ev gpiod.LineEvent) {
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
				}).Debug("Rotary turned left")
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
				}).Debug("Rotary turned right")
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
