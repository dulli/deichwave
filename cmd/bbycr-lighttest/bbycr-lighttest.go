package main

import (
	"fmt"
	"image/color"
	"os"
	"time"

	"github.com/dulli/bbycrgo/pkg/common"
	"github.com/dulli/bbycrgo/pkg/lights"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

func main() {
	log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: time.Stamp,
	})

	// Load configuration
	var cfg common.Config
	common.Configure(&cfg)

	// Prepare the lights command module and initialize the led groups
	lightPlayer, err := lights.NewRenderer("light-test", cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Renderer setup incomplete")
	} else {
		log.Info("Renderer setup complete")
	}
	log.Info("Gathering light effect files...")

	// Gather the light effect files
	err = lightPlayer.LoadEffects(cfg.Lights.Path)
	if err != nil {
		log.WithFields(log.Fields{
			"path": cfg.Lights.Path,
			"err":  err,
		}).Fatal("Failed to load the light effect directory, is the path correct?")
	} else {
		log.WithFields(log.Fields{
			"num": lightPlayer.ListEffects(),
		}).Info("Loaded effects")
	}

	colors := lights.ColormapRainbow(256)
	ledCount := lightPlayer.GetLEDCount()
	groupCount := len(lightPlayer.GetGroupCount())
	s := make([]color.Color, ledCount+groupCount)
	for idx := range s {
		s[idx] = color.Black
	}

	preview := app.New()
	window := preview.NewWindow(fmt.Sprintf("LED Strip Preview: %s", os.Args[1]))
	raster := canvas.NewRasterWithPixels(
		func(x, y, w, h int) color.Color {
			factor := w / len(colors)
			if factor < 1 {
				factor = 1
			}
			if y < 5 {
				if x/factor < len(colors) {
					return colors[x/factor].Get()
				}
				return color.White
			}
			if y == 5 {
				return color.White
			}

			factor = w / (ledCount + groupCount)
			if factor < 1 {
				factor = 1
			}
			if x/factor < len(s) {
				return s[x/factor]
			}
			return color.White
		})
	window.SetContent(raster)

	lightPlayer.ReceiveFrame(func(state [][]lights.LEDState) {
		idx := 0
		for _, group := range state {
			for _, led := range group {
				c := colors[led.ColorIndex]
				c.L = led.Brightness / 2
				s[idx] = c.Get()
				idx++
			}
			s[idx] = color.White
			idx++
		}
		raster.Refresh()
	})

	lightPlayer.SetEffect(os.Args[1])
	window.Resize(fyne.NewSize(500, 100))
	window.ShowAndRun()
}
