package main

import (
    "encoding/json"
    "errors"
    "sync"

    bbycrgo "github.com/dulli/bbycrgo/pkg"
    shellquote "github.com/kballard/go-shellquote"
)

const (
    ENDPOINT_STATUS string = "status"
)

var QueryRequestInvalid = errors.New("Invalid query request")

func StatusSetup(progress *sync.WaitGroup) error {
    go StatusEventLoop(progress)

    return nil
}

func StatusCmdParse(ev bbycrgo.Event) (string, error) {
    var request string

    args, err := shellquote.Split(ev.Arguments)
    if err != nil {
        ev.Response <- err.Error()
        return request, err
    }
    return args[0], nil
}

func StatusQuery(ev bbycrgo.Event) (string, error) {
    request, err := StatusCmdParse(ev)
    if err != nil {
        return "", err
    }

    var data []byte
    var resp string
    switch request {
    case "cmds":
        data, err = json.Marshal(bbycrgo.CmdList)

    default:
        err = QueryRequestInvalid
    }
    if err == nil {
        resp = string(data)
    }
    return resp, err
}

func StatusGetCmds() bbycrgo.EventHandlerList {
    query_keys := []string{"cmds"}
    cmds := bbycrgo.EventHandlerList{
        "query": bbycrgo.EventHandler{StatusQuery, query_keys},
    }
    return cmds
}

func StatusEventLoop(progress *sync.WaitGroup) {
    cmds := StatusGetCmds()
    bbycrgo.EventLoop(ENDPOINT_STATUS, cmds, progress)
}
