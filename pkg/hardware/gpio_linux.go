package hardware

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dulli/deichwave/pkg/common"
	"github.com/dulli/deichwave/pkg/rest"
	log "github.com/sirupsen/logrus"

	"github.com/warthog618/go-gpiocdev"
)

type GPIO struct {
	lines []*gpiocdev.Line
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
		lastReceivedType  gpiocdev.LineEventType
		lastTriggeredType gpiocdev.LineEventType
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
	update := func(t *gpiocdev.LineEventType) {
		if *t == lastTriggeredType {
			return
		}
		lastTriggeredType = *t
		if *t == gpiocdev.LineEventRisingEdge {
			press()
		} else {
			release()
		}
	}

	// Handler for hardware events
	handler := func(ev gpiocdev.LineEvent) {
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
	line, err := gpiocdev.RequestLine(chip, pin, gpiocdev.WithPullUp,
		gpiocdev.WithEventHandler(handler), gpiocdev.WithBothEdges)
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
	line, err := gpiocdev.RequestLine(chip, pins[0],
		gpiocdev.WithPullUp,
		gpiocdev.WithEventHandler(func(ev gpiocdev.LineEvent) {
			rotflag = ev.Type == gpiocdev.LineEventRisingEdge
		}),
		gpiocdev.WithBothEdges)
	if err != nil {
		return err
	}
	h.lines = append(h.lines, line)

	line, err = gpiocdev.RequestLine(chip, pins[1],
		gpiocdev.WithPullUp,
		gpiocdev.WithDebounce(debounceTime),
		gpiocdev.WithEventHandler(func(ev gpiocdev.LineEvent) {
			if ev.Type == gpiocdev.LineEventFallingEdge {
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
		gpiocdev.WithBothEdges)
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
