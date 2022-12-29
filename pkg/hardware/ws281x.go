//go:build !arm64

package hardware

import (
	"github.com/dulli/deichwave/pkg/common"
	"github.com/dulli/deichwave/pkg/lights"
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
