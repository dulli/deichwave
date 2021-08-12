package main

import (
    "net"

    "flag"
    "math"
    "time"

    bbycrgo "github.com/dulli/bbycrgo/pkg"
    ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
    log "github.com/sirupsen/logrus"
)

const (
    EFFECT_FPS = 30

    LED_LEVEL = 0xFF
)

type LedGroup struct {
    count uint8
    left  bool
    right bool
    front bool
    rear  bool
}

var effect_state = []byte{bbycrgo.LightEffects["kitt"], 0xff, 100}
var effect_groups []LedGroup
var effect_offset uint8
var led_id uint8
var tick uint8
var decider uint64

func main() {
    debug := flag.Bool("debug", false, "Enable debug output")
    flag.Parse()
    if *debug {
        log.SetLevel(log.DebugLevel)
    } else {
        log.SetLevel(log.InfoLevel)
    }

    effect_groups = []LedGroup{
        LedGroup{count: 2, front: true},
        LedGroup{count: 4, front: true},
        LedGroup{count: 4, rear: true},
        LedGroup{count: 5, rear: true, right: true},
        LedGroup{count: 5, rear: true},
        LedGroup{count: 5, rear: true, left: true},
        LedGroup{count: 4, rear: true},
        LedGroup{count: 4, front: true},
        LedGroup{count: 5, front: true, left: true},
        LedGroup{count: 5, front: true},
        LedGroup{count: 5, front: true, right: true},
    }

    for _, group := range effect_groups {
        effect_offset = effect_offset + group.count
    }

    opt := ws2811.DefaultOptions
    opt.Channels[0].Brightness = LED_LEVEL
    opt.Channels[0].LedCount = int(effect_offset)
    opt.Channels[0].StripeType = ws2811.WS2811StripBRG
    dev, err := ws2811.MakeWS2811(&opt)
    if err != nil {
        panic(err)
    }
    err = dev.Init()
    if err != nil {
        panic(err)
    }
    defer dev.Fini()

    go EffectServer()
    log.WithFields(log.Fields{
        "fps": EFFECT_FPS,
    }).Info("Waiting for light changes")

    last_brightness := effect_state[1]
    ticker := time.NewTicker((1000 / EFFECT_FPS) * time.Millisecond)
    for _ = range ticker.C {
        tick = tick + 1
        effect_offset = 0

        if last_brightness != effect_state[1] {
            dev.SetBrightness(0, int(effect_state[1]))
            last_brightness = effect_state[1]
            log.Debug("Changed brightness")
        }
        for _, group := range effect_groups {
            for led_idx := uint8(0); led_idx < group.count; led_idx++ {
                led_id = led_idx + effect_offset

                switch effect_state[0] {
                case bbycrgo.LightEffects["red"]:
                    dev.Leds(0)[led_id] = 0xff0000

                case bbycrgo.LightEffects["green"]:
                    dev.Leds(0)[led_id] = 0x00ff00

                case bbycrgo.LightEffects["blue"]:
                    dev.Leds(0)[led_id] = 0x0000ff

                case bbycrgo.LightEffects["blaulicht"]:
                    decider = uint64(math.Floor(float64(tick) / (0.5 * float64(EFFECT_FPS))))
                    if decider%2 == 0 {
                        dev.Leds(0)[led_id] = 0x0000ff
                    } else {
                        dev.Leds(0)[led_id] = 0xff0000
                    }

                case bbycrgo.LightEffects["bierpong"]:
                    if group.front {
                        dev.Leds(0)[led_id] = 0x0000ff
                    } else if group.rear {
                        dev.Leds(0)[led_id] = 0xff0000
                    }

                case bbycrgo.LightEffects["flash"]:
                    dev.Leds(0)[led_id] = 0xffffff * uint32(tick%2)

                case bbycrgo.LightEffects["kitt"]:
                    decider = uint64(math.Floor(float64(tick) / 0.1 * float64(EFFECT_FPS)))
                    if decider%uint64(group.count) == uint64(led_idx) {
                        dev.Leds(0)[led_id] = 0xff0000
                    } else {
                        dev.Leds(0)[led_id] = 0x000000
                    }

                default:
                    dev.Leds(0)[led_id] = uint32(0xf8e968*(1-float64(effect_state[2])/100) + 0x7fe5fb*float64(effect_state[2])/100)
                }
            }
            effect_offset = effect_offset + group.count
        }
        err = dev.Render()
        if err != nil {
            log.Error(err)
        }
    }

}

func EffectServer() {
    log.WithFields(log.Fields{
        "type": bbycrgo.SOCKET_TYPE,
        "addr": bbycrgo.LIGHT_ADDR,
    }).Info("Listening on Socket")
    sock, err := net.Listen(bbycrgo.SOCKET_TYPE, bbycrgo.LIGHT_ADDR)
    if err != nil {
        log.WithFields(log.Fields{
            "err": err,
        }).Error("Couldn't open socket")
        return
    }
    defer sock.Close()

    for {
        conn, err := sock.Accept()
        if err != nil {
            log.WithFields(log.Fields{
                "err": err,
            }).Warn("Couldn't accept socket connection")
            return
        }
        _, err = conn.Read(effect_state)
        if err != nil {
            log.WithFields(log.Fields{
                "err": err,
            }).Warn("Couldn't read from socket")
            return
        }
        log.WithFields(log.Fields{
            "effect": effect_state,
        }).Info("Updated state")
        conn.Close()
    }
}
