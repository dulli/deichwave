package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dulli/deichwave/pkg/common"
)

var ErrActionNotReady = errors.New("actions can not be performed yet")
var ErrActionFailed = errors.New("action could not be performed")

func (s Server) DoAction(action string) error {
	if s.port == 0 {
		return ErrActionNotReady
	}
	response, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/%s/%s", s.port, s.apiBase, action), "application/json", nil)
	if err == nil && response.StatusCode != 200 {
		err = ErrActionFailed
	}
	return err
}

func (s Server) AddHooks(cfg common.Config) {
	for key, actions := range cfg.Hooks {
		listenHook(s, strings.Split(key, "-"), actions)
	}
}

func listenHook(s Server, target []string, actions []string) {
	common.EventListen(func(ev common.Event) {
		if ev.Origin == target[0] && ev.Type == target[1] {
			for _, action := range actions {
				_ = s.DoAction(action)
			}
		}
	})
}
