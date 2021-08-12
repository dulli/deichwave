package bbycrgo

import (
    "errors"
    "sync"

    log "github.com/sirupsen/logrus"
)

type Event struct {
    Target    string
    Command   string
    Arguments string
    Response  chan string
}
type EventHandler struct {
    Func      func(Event) (string, error)
    Arguments []string
}
type EventHandlerList map[string]EventHandler

var CmdList = make(map[string]map[string][]string)
var ChanList = make(map[string]chan Event)

var InvalidEventData = errors.New("Invalid event data")
var InvalidEventTarget = errors.New("Invalid target endpoint")

func EventSplitter(events chan Event, progress *sync.WaitGroup) {
    defer progress.Done()
    for {
        ev := <-events
        endpoint := ev.Target

        if target, ok := ChanList[endpoint]; ok {
            target <- ev
        } else {
            log.WithFields(log.Fields{
                "endpoint": endpoint,
            }).Error(InvalidEventTarget)
            ev.Response <- InvalidEventTarget.Error()
        }
    }
}

func EventLoop(endpoint string, cmds EventHandlerList, progress *sync.WaitGroup) {
    progress.Add(1)
    defer progress.Done()

    RegisterCmds(endpoint, cmds)
    ChanList[endpoint] = make(chan Event)
    for {
        ev, ok := <-ChanList[endpoint]
        if ok == false {
            log.WithFields(log.Fields{
                "endpoint": endpoint,
            }).Error("EventLoop broke")
            break
        }

        if handler, ok := cmds[ev.Command]; ok {
            resp, err := handler.Func(ev)
            if err != nil {
                ev.Response <- err.Error()
                log.Error(err)
            } else {
                if resp != "" {
                    ev.Response <- resp
                } else {
                    ev.Response <- "OK"
                }
            }
        } else {
            ev.Response <- InvalidEventData.Error()
            log.Error(InvalidEventData)
        }
    }
}

func RegisterCmds(target string, list EventHandlerList) error {
    endpoint := make(map[string][]string)
    for command, ev := range list {
        endpoint[command] = ev.Arguments
    }
    CmdList[target] = endpoint
    return nil
}
