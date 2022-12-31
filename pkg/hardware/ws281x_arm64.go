package hardware

import (
	"image"

	"github.com/dulli/deichwave/pkg/common"
	"github.com/dulli/deichwave/pkg/lights"
	ws281x "github.com/dulli/go-rpi-ws281x"
	log "github.com/sirupsen/logrus"
)

type LEDws281x struct {
	canvas *ws281x.Canvas
}

func (h *LEDws281x) Setup(l lights.Renderer, cfg common.Config) error {
	var err error
	config := ws281x.DefaultConfig
	config.Brightness = int(float64(cfg.Hardware.LEDBrightness) * 0.01 * 255)
	config.Pin = cfg.Hardware.LEDPin
	config.StripType = ws281x.StripBRG
	colors := lights.ColormapRainbow(256)
	ledCount := l.GetLEDCount()

	rect := image.Rectangle{image.Point{0, 0}, image.Point{ledCount - 1, 0}}
	h.canvas, err = ws281x.NewCanvas(rect.Max.X+1, 1, &config)
	if err != nil {
		return err
	}
	err = h.canvas.Initialize()
	if err != nil {
		return err
	}

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
