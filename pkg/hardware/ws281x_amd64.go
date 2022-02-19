package hardware

import (
	"github.com/dulli/bbycrgo/pkg/common"
	"github.com/dulli/bbycrgo/pkg/lights"
)

type LEDws281x struct {
}

func (h *LEDws281x) Setup(l lights.Renderer, cfg common.Config) error {
	return ErrDriverNotImplementedForArch
}

func (h *LEDws281x) Check() error {
	return ErrDriverNotImplementedForArch
}

func (h *LEDws281x) Close() {
}
