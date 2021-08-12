package main

import (
    "flag"
    "math/rand"
    "sync"
    "time"

    bbycrgo "github.com/dulli/bbycrgo/pkg"
    log "github.com/sirupsen/logrus"
)

func main() {
    debug := flag.Bool("debug", false, "Enable debug output")
    flag.Parse()
    if *debug {
        log.SetLevel(log.DebugLevel)
    } else {
        log.SetLevel(log.InfoLevel)
    }

    rand.Seed(time.Now().UnixNano())

    // Prepare EventLoops
    events := make(chan bbycrgo.Event)
    progress := &sync.WaitGroup{}

    // Initialize Sensors
    log.Info("Performing sensor setup")
    progress.Add(2)
    go bbycrgo.EventSplitter(events, progress)
    go bbycrgo.SocketServer(events, progress)

    // Initialize Actors
    log.Info("Performing actor setup")
    err := StatusSetup(progress)
    if err != nil {
        log.Fatal(err)
    }
    err = SoundSetup(progress)
    if err != nil {
        log.Fatal(err)
    }
    err = MusicSetup(progress)
    if err != nil {
        log.Fatal(err)
    }
    err = LightsSetup(progress)
    if err != nil {
        log.Fatal(err)
    }

    // Perform finishing actions, like starting the music

    log.Info("Finished setup")
    progress.Wait()
}
