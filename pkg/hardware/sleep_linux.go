package hardware

import (
	"github.com/dulli/deichwave/pkg/common"
)

type SystemSleep struct {
}

func (h *SystemSleep) Setup(cfg common.Config) error {
	return ErrDriverNotImplementedForArch
}

func (h *SystemSleep) Check() error {
	return ErrDriverNotImplementedForArch
}

func (h *SystemSleep) Close() {
}
