package common

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

var ErrProfileNotFound = errors.New("light effect could not be found")

type ProfileSwitcher struct {
	names []string
	paths map[string]string
	cfg   *Config
}

func NewProfilSwitcher(cfg *Config) (ProfileSwitcher, error) {
	// Gather profiles
	// TODO make configurable
	var p ProfileSwitcher
	cfgDir := filepath.Dir(cfg.File)
	cfgExt := filepath.Ext(cfg.File)
	cfgSub := filepath.Join(cfgDir, "profiles")
	p.cfg = cfg
	p.paths = make(map[string]string)
	if _, err := os.Stat(cfgSub); !os.IsNotExist(err) {
		err = filepath.WalkDir(cfgSub, func(path string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if filepath.Ext(d.Name()) == cfgExt {
				basename := filepath.Base(path)
				name := strings.TrimSuffix(basename, filepath.Ext(basename))
				p.names = append(p.names, name)
				p.paths[name] = path
			}
			return nil
		})
		if err != nil {
			return p, err
		}
	}
	sort.Strings(p.names)
	if len(p.paths) > 0 {
		log.WithFields(log.Fields{
			"profiles": p.names,
		}).Info("Found profiles")
	}
	return p, nil
}

func (p *ProfileSwitcher) ListProfiles() []string {
	return p.names
}

func (p *ProfileSwitcher) SetProfile(name string) error {
	if _, ok := p.paths[name]; !ok {
		return ErrProfileNotFound
	}
	err := cleanenv.ReadConfig(p.paths[name], p.cfg)
	EventFire(Event{
		Origin: "config",
		Type:   "changed",
	})
	log.WithFields(log.Fields{
		"profile": name,
	}).Info("Profile changed")
	return err
}

func ConfigChangeListener(handler func()) {
	EventListen(func(ev Event) {
		if ev.Origin != "config" && ev.Type != "changed" {
			return
		}
		handler()
	})
}
