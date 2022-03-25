package shell

import (
	"errors"
	"os/exec"

	"github.com/dulli/bbycrgo/pkg/common"
)

var ErrCommandNotFound = errors.New("command could not be found")

type ShellExecutor interface {
	Run(string) (string, error)
}

type shellExec struct {
	commands map[string][]string
}

func NewExecutor(name string, cfg common.Config) (ShellExecutor, error) {
	exec := shellExec{
		commands: cfg.Shell,
	}
	return &exec, nil
}

func (s *shellExec) Run(cmd_name string) (string, error) {
	if cmd_list, ok := s.commands[cmd_name]; ok {
		var cmd *exec.Cmd
		if len(cmd_list) > 1 {
			cmd = exec.Command(cmd_list[0], cmd_list[1:]...)
		} else {
			cmd = exec.Command(cmd_list[0])
		}
		stdout, err := cmd.Output()
		return string(stdout), err
	} else {
		return "", ErrCommandNotFound
	}
}
