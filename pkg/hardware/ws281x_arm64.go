package hardware

import (
	"image"

	"github.com/dulli/bbycrgo/pkg/common"
	"github.com/dulli/bbycrgo/pkg/lights"
	ws281x "github.com/dulli/go-rpi-ws281x"
)

type LEDws281x struct {
	canvas *ws281x.Canvas
}

func (h *LEDws281x) Setup(l lights.Renderer, cfg common.Config) error {
	config := ws281x.DefaultConfig
	config.Brightness = cfg.Hardware.LEDBrightness
	config.Pin = cfg.Hardware.LEDPin
	colors := lights.ColormapRainbow(256)
	ledCount := l.GetLEDCount()

	rect := image.Rectangle{image.Point{0, 0}, image.Point{ledCount, 0}}
	h.canvas, _ = ws281x.NewCanvas(rect.Max.X, rect.Max.Y, &ws281x.DefaultConfig)
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
	return nil
}

func (h *LEDws281x) Check() error {
	return nil
}

func (h *LEDws281x) Close() {
	h.canvas.Close()
}
