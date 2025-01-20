package common

import (
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

type Event struct {
	Origin string `json:"origin"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

var ready bool
var queue chan Event
var listeners []func(Event)

func EventFire(ev Event) {
	log.WithFields(log.Fields{
		"event": ev,
		"ready": ready,
	}).Debug("Event fired")
	if ready {
		queue <- ev
	}
}

func EventListen(listener func(Event)) {
	listeners = append(listeners, listener)
}

func EventLoop() {
	queue = make(chan Event)
	ready = true
	log.Debug("Eventloop started")

	for ev := range queue {
		log.WithFields(log.Fields{
			"event": ev,
		}).Debug("Event received")

		for _, listener := range listeners {
			listener(ev)
		}
	}
}

func AwaitSignal() os.Signal {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	sig := <-sigchan

	EventFire(Event{
		Origin: "System",
		Name:   "SIGINT",
		Type:   "Signal",
	})
	return sig
}
