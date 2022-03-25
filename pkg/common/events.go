package common

import (
	"os"
	"os/signal"
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
	for ev := range queue {
		for _, listener := range listeners {
			listener(ev)
		}
	}
}

func AwaitSignal() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	EventFire(Event{
		Origin: "System",
		Name:   "SIGINT",
		Type:   "Signal",
	})
}
