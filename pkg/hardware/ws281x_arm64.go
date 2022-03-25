package hardware

import (
	"image"

	"github.com/dulli/bbycrgo/pkg/common"
	"github.com/dulli/bbycrgo/pkg/lights"
	ws281x "github.com/dulli/go-rpi-ws281x"
	log "github.com/sirupsen/logrus"
)

type LEDws281x struct {
	canvas *ws281x.Canvas
}

func (h *LEDws281x) Setup(l lights.Renderer, cfg common.Config) error {
	config := ws281x.DefaultConfig
	config.Brightness = cfg.Hardware.LEDBrightness
	config.Pin = cfg.Hardware.LEDPin
	config.StripType = ws281x.StripBRG
	colors := lights.ColormapRainbow(256)
	ledCount := l.GetLEDCount()

	rect := image.Rectangle{image.Point{0, 0}, image.Point{ledCount - 1, 0}}
	h.canvas, _ = ws281x.NewCanvas(rect.Max.X+1, 1, &config)
	h.canvas.Initialize()

	l.ReceiveFrame(func(state [][]lights.LEDState) {
		idx := 0
		for _, group := range state {
			for _, led := range group {
				c := colors[led.ColorIndex]
				c.L = led.Brightness / 2
				h.canvas.Set(idx, 0, c.Get())
				idx++
			}
		}
		h.canvas.Render()
	})
	log.WithFields(log.Fields{
		"type":     "led",
		"driver":   "ws281x",
		"platform": "arm64",
	}).Debug("Listening for frames")
	return nil
}

func (h *LEDws281x) Check() error {
	return nil
}

func (h *LEDws281x) Close() {
	h.canvas.Close()
	log.WithFields(log.Fields{
		"type":     "led",
		"driver":   "ws281x",
		"platform": "arm64",
	}).Debug("Stopped listening for frames")
}
