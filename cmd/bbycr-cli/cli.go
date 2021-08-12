package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"

    "io"
    "strings"

    "github.com/chzyer/readline"
    bbycrgo "github.com/dulli/bbycrgo/pkg"
    log "github.com/sirupsen/logrus"
)

const (
    PROMPT_GREETING string = `

 888888b.   888888b. Y88b   d88P  .d8888b.  8888888b.  
 888  "88b  888  "88b Y88b d88P  d88P  Y88b 888   Y88b 
 888  .88P  888  .88P  Y88o88P   888    888 888    888 
 8888888K.  8888888K.   Y888P    888        888   d88P 
 888  "Y88b 888  "Y88b   888     888        8888888P"  
 888    888 888    888   888     888    888 888 T88b   
 888   d88P 888   d88P   888     Y88b  d88P 888  T88b  
 8888888P"  8888888P"    888      "Y8888P"  888   T88b 

`
    PROMPT_SYMBOL string = "₿" // ฿
)

func filterInput(r rune) (rune, bool) {
    switch r {
    // block CtrlZ feature
    case readline.CharCtrlZ:
        return r, false
    }
    return r, true
}

func main() {
    debug := flag.Bool("debug", false, "Enable debug output")
    flag.Parse()
    if *debug {
        log.SetLevel(log.DebugLevel)
    } else {
        log.SetLevel(log.InfoLevel)
    }

    fmt.Printf("\033[2J\033[1;1H\033[31m%s\033[0m\n", PROMPT_GREETING)
    defer fmt.Printf("\033[2J\033[1;1Hs")

    log.WithFields(log.Fields{
        "addr": bbycrgo.SOCKET_ADDR,
    }).Debug("Connecting to input socket")
    client, err := bbycrgo.SocketConnect()
    if err != nil {
        log.WithFields(log.Fields{
            "err": err,
        }).Debug("Socket connection failed")
        log.Fatal("Could not connect to the BBYCR, is the engine running?")
        return
    }
    defer client.Close()

    log.Info("Starting BBYCR CLI interface")
    l, err := readline.NewEx(&readline.Config{
        Prompt:      fmt.Sprintf("\033[31m%s\033[0m ", PROMPT_SYMBOL),
        HistoryFile: "/tmp/bbycr-cli.tmp",
        //AutoComplete:    completer, // TODO implement autocompleter
        InterruptPrompt: "^C",
        EOFPrompt:       "exit",

        HistorySearchFold:   true,
        FuncFilterInputRune: filterInput,
    })
    if err != nil {
        panic(err)
    }
    defer l.Close()

    log.SetOutput(l.Stderr())
    for {
        line, err := l.Readline()
        if err == io.EOF {
            break
        }

        // TODO kill prompt if socket closes
        line = strings.TrimSpace(line)
        switch {
        case line == "help":
            // TODO implement help
            log.Warn("Help is not implemented yet")

        case line == "exit":
            goto exit

        case line == "":
            continue

        default:
            resp, err := client.Write(line)
            if err != nil {
                log.Fatal("Could not send anything to the BBYCR, exiting")
                break
            }

            var prettyJSON bytes.Buffer
            err = json.Indent(&prettyJSON, []byte(resp), "", "\t")
            if err != nil && resp != "OK\n" {
                fmt.Print(resp)
            } else {
                fmt.Print(string(prettyJSON.Bytes()))
            }
        }
    }
exit:
}
