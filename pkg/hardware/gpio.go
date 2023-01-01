//go:build !linux

package hardware

import (
	"github.com/dulli/deichwave/pkg/common"
)

type GPIO struct {
}

func (h *GPIO) Setup(cfg common.Config) error {
	return ErrDriverNotImplementedForArch
}

func (h *GPIO) Check() error {
	return ErrDriverNotImplementedForArch
}

func (h *GPIO) Close() {
}
