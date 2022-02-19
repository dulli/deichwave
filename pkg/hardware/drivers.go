package hardware

import (
	"errors"

	"github.com/dulli/bbycrgo/pkg/common"
	"github.com/dulli/bbycrgo/pkg/lights"
)

var ErrDriverNotImplementedForArch = errors.New("driver is not implemented for this platform")
var ErrDriverNameNotFound = errors.New("no driver with this identifier is available")

type DriverLED interface {
	Setup(lights.Renderer, common.Config) error
	Check() error
	Close()
}

func GetLEDDriver(name string) (DriverLED, error) {
	var d DriverLED

	err := ErrDriverNameNotFound
	switch name {
	case "ws281x":
		d = &LEDws281x{}
		err = d.Check()
	}
	return d, err
}
