package main

import (
	"errors"
	"math"
	"net"
	"strconv"
	"sync"

	bbycrgo "github.com/dulli/bbycrgo/pkg"
	shellquote "github.com/kballard/go-shellquote"
)

const (
	ENDPOINT_LIGHTS string = "lights"

	BRIGHTNESS_MAX   uint8 = 255
	BRIGHTNESS_MIN   uint8 = 0
	BRIGHTNESS_START uint8 = 255
)

type LightCmd struct {
	effect     []uint8
	brightness uint8
	intensity  uint8
}

var InvalidIntensityError = errors.New("Target intensity has to be an integer between 0 and 255")
var InvalidBrightnessError = errors.New("Target brightness has to be an integer between 0 and 255")
var NoLightsArgError = errors.New("Required parameter missing")

func (lcmd *LightCmd) Update() error {
	payload := []byte{lcmd.effect[len(lcmd.effect)-1], lcmd.brightness, lcmd.intensity}
	conn, err := net.Dial(bbycrgo.SOCKET_TYPE, bbycrgo.LIGHT_ADDR)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(payload)
	return err
}

func (lcmd *LightCmd) Parse(ev bbycrgo.Event) (int, error) {
	args, err := shellquote.Split(ev.Arguments)
	if err != nil {
		return 0, err
	}
	if len(args) == 0 {
		return 0, NoLightsArgError
	}
	if val, ok := bbycrgo.LightEffects[args[0]]; ok {
		return int(val), nil
	}
	val, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (lcmd *LightCmd) Start(ev bbycrgo.Event) (string, error) {
	value, err := lcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	lcmd.effect = append(lcmd.effect, uint8(value))
	return "", lcmd.Update()
}

func (lcmd *LightCmd) Stop(ev bbycrgo.Event) (string, error) {
	value, err := lcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	new_effect := make([]byte, 1)
	for effect_idx := range lcmd.effect {
		if lcmd.effect[effect_idx] != uint8(value) {
			new_effect = append(new_effect, lcmd.effect[effect_idx])
		}
	}
	if len(new_effect) > 0 {
		new_effect = new_effect[1:]
	}
	lcmd.effect = new_effect
	return "", lcmd.Update()
}

func (lcmd *LightCmd) SetIntensity(ev bbycrgo.Event) (string, error) {
	intensity, err := lcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	if intensity > 100 || intensity < 0 {
		return "", InvalidIntensityError
	}
	lcmd.intensity = uint8(intensity)
	return "", lcmd.Update()
}

func (lcmd *LightCmd) ChangeIntensity(ev bbycrgo.Event) (string, error) {
	intensity, err := lcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	new_intensity := int(lcmd.intensity) + intensity
	if new_intensity > RNG_MAX {
		new_intensity = RNG_MAX
	}
	if new_intensity < RNG_MIN {
		new_intensity = RNG_MIN
	}
	lcmd.intensity = uint8(new_intensity)
	return "", lcmd.Update()
}

func (lcmd *LightCmd) GetIntensity(ev bbycrgo.Event) (string, error) {
	return strconv.Itoa(int(lcmd.intensity)), nil
}

func (lcmd *LightCmd) SetBrightness(ev bbycrgo.Event) (string, error) {
	brightness, err := lcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	if brightness > 0xff || brightness < 0 {
		return "", InvalidBrightnessError
	}
	lcmd.brightness = uint8(brightness)
	return "", lcmd.Update()
}

func (lcmd *LightCmd) ChangeBrightness(ev bbycrgo.Event) (string, error) {
	brightness, err := lcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	new_brightness := uint8(int(lcmd.brightness) + brightness)
	if new_brightness > BRIGHTNESS_MAX {
		new_brightness = BRIGHTNESS_MAX
	}
	if new_brightness < BRIGHTNESS_MIN {
		new_brightness = BRIGHTNESS_MIN
	}
	lcmd.brightness = new_brightness
	return "", lcmd.Update()
}

func (lcmd *LightCmd) GetBrightness(ev bbycrgo.Event) (string, error) {
	return strconv.Itoa(int(lcmd.brightness)), nil
}

func (lcmd *LightCmd) Register() bbycrgo.EventHandlerList {
	cmds := bbycrgo.EventHandlerList{
		"start":             bbycrgo.EventHandler{lcmd.Start, nil},
		"stop":              bbycrgo.EventHandler{lcmd.Stop, nil},
		"intensity":         bbycrgo.EventHandler{lcmd.SetIntensity, nil},
		"change-intensity":  bbycrgo.EventHandler{lcmd.ChangeIntensity, nil},
		"get-intensity":     bbycrgo.EventHandler{lcmd.GetIntensity, nil},
		"brightness":        bbycrgo.EventHandler{lcmd.SetBrightness, nil},
		"change-brightness": bbycrgo.EventHandler{lcmd.ChangeBrightness, nil},
		"get-brightness":    bbycrgo.EventHandler{lcmd.GetBrightness, nil},
	}
	return cmds
}

func (lcmd *LightCmd) EventLoop(progress *sync.WaitGroup) {
	cmds := lcmd.Register()
	bbycrgo.EventLoop(ENDPOINT_LIGHTS, cmds, progress)
}

func LightsSetup(progress *sync.WaitGroup) error {
	lcmd := LightCmd{
		effect:     make([]uint8, 1),
		brightness: BRIGHTNESS_START,
		intensity:  uint8(math.Round(float64(RNG_MAX) / 100 * 0xFF)),
	}

	go lcmd.EventLoop(progress)

	return lcmd.Update()
}
