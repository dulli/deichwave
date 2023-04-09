package hardware

import (
	"errors"

	"github.com/dulli/deichwave/pkg/common"
	"github.com/dulli/deichwave/pkg/lights"
	"github.com/dulli/deichwave/pkg/rest"
)

var ErrDriverNotImplementedForArch = errors.New("driver is not implemented for this platform")
var ErrDriverNameNotFound = errors.New("no driver with this identifier is available")

// TODO Unify driver interfaces
type DriverSystem interface {
	Setup(common.Config) error
	Check() error
	Close()
}

func GetSystemDriver(name string) (DriverSystem, error) {
	var d DriverSystem

	err := ErrDriverNameNotFound
	switch name {
	case "sleep":
		d = &SystemSleep{}
		err = d.Check()
	}
	return d, err
}

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

type DriverInput interface {
	Setup(common.Config, rest.Server) error
	Check() error
	Close()
}

func GetInputDriver(name string) (DriverInput, error) {
	var d DriverInput

	err := ErrDriverNameNotFound
	switch name {
	case "gpio":
		d = &GPIO{}
		err = d.Check()
	}
	return d, err
}
