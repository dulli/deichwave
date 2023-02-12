//go:build !linux

package hardware

import (
	"github.com/dulli/deichwave/pkg/common"
	"github.com/dulli/deichwave/pkg/rest"
)

type GPIO struct {
}

func (h *GPIO) Setup(cfg common.Config, srv rest.Server) error {
	return ErrDriverNotImplementedForArch
}

func (h *GPIO) Check() error {
	return ErrDriverNotImplementedForArch
}

func (h *GPIO) Close() {
}
