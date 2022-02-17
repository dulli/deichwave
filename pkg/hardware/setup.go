package hardware

import (
	"github.com/dulli/bbycrgo/pkg/common"
	"github.com/dulli/bbycrgo/pkg/lights"
)

type DriverLED interface {
	Setup(lights.Renderer, common.Config)
	Close()
}
