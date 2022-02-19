package common

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
