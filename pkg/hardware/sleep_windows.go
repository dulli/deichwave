package hardware

import (
	"syscall"

	"github.com/dulli/deichwave/pkg/common"
)

// Execution States
const (
	EsDisplayRequired = 0x00000002
	EsSystemRequired  = 0x00000001
	EsContinuous      = 0x80000000
)

type SystemSleep struct {
	proc *syscall.LazyProc
}

func (h *SystemSleep) Setup(cfg common.Config) error {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	h.proc = kernel32.NewProc("SetThreadExecutionState")
	_, _, err := h.proc.Call(uintptr(EsContinuous | EsSystemRequired | EsDisplayRequired))
	if syserr, ok := err.(syscall.Errno); ok {
		if syserr == 0 {
			return nil
		}
	}
	return err
}

func (h *SystemSleep) Check() error {
	return nil
}

func (h *SystemSleep) Close() {
	_, _, _ = h.proc.Call(uintptr(EsContinuous))
}
