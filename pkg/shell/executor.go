package shell

import (
	"errors"
	"os/exec"
	"runtime"

	"github.com/dulli/deichwave/pkg/common"
	log "github.com/sirupsen/logrus"
)

var ErrCommandNotFound = errors.New("command could not be found")

type ShellExecutor interface {
	Run(string) (string, error)
}

type shellExec struct {
	commands map[string][]string
}

func NewExecutor(name string, cfg common.Config) (ShellExecutor, error) {
	var cmds map[string][]string
	cmds, _ = cfg.Shell[runtime.GOOS]
	exec := shellExec{
		commands: cmds,
	}
	return &exec, nil
}

func (s *shellExec) Run(cmd_name string) (string, error) {
	log.WithFields(log.Fields{
		"name": cmd_name,
	}).Info("Running a shell command")

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
