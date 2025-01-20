//go:generate go run github.com/tc-hib/go-winres make --in "winres.json" --arch "amd64" --product-version=git-tag --file-version=git-tag

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/dulli/deichwave/pkg/common"
	"github.com/dulli/deichwave/pkg/hardware"
	"github.com/dulli/deichwave/pkg/lights"
	"github.com/dulli/deichwave/pkg/music"
	"github.com/dulli/deichwave/pkg/rest"
	"github.com/dulli/deichwave/pkg/shell"
	"github.com/dulli/deichwave/pkg/sounds"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "NONE"
	date    = "UNKNOWN"
	builtBy = "UNKNOWN"
)

func main() {
	log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: time.Stamp,
	})

	// TODO refactor this into a loop calling unified interfaces

	// Load configuration
	var cfg common.Config
	common.Configure(&cfg)

	// Keep track of meta information
	cfg.Meta.Version = version
	cfg.Meta.Build = fmt.Sprintf("commit %s, built at %s by %s", commit, date, builtBy)
	fmt.Printf("%s (%s)\n", cfg.Meta.Name, cfg.Meta.Version)

	log.WithFields(log.Fields{
		"Version": cfg.Meta.Version,
		"Info":    cfg.Meta.Build,
	}).Debug("Build")

	// Prevent hibernation
	driverSleep, err := hardware.GetSystemDriver("sleep")
	if err != nil {
		log.WithFields(log.Fields{
			"driver": "sleep",
			"err":    err,
		}).Error("Failed to load system driver")
	} else {
		err = driverSleep.Setup(cfg)
		if err != nil {
			log.WithFields(log.Fields{
				"driver": "sleep",
				"err":    err,
			}).Error("Failed to setup system driver")
		} else {
			defer driverSleep.Close()
			log.WithFields(log.Fields{
				"driver": "sleep",
			}).Info("Initialized system driver")
		}
	}

	// Prepare the profile switcher
	profileSwitcher, err := common.NewProfilSwitcher(&cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Profile setup incomplete")
	} else {
		log.Info("Profile setup complete")
	}

	// Prepare the lights command module and initialize the led groups
	lightPlayer, err := lights.NewRenderer("light-test", cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Renderer setup incomplete")
	} else {
		log.Info("Renderer setup complete")
	}
	log.Info("Gathering light effect files...")

	// Gather the light effect files
	err = lightPlayer.LoadEffects(cfg.Lights.Path)
	if err != nil {
		log.WithFields(log.Fields{
			"path": cfg.Lights.Path,
			"err":  err,
		}).Fatal("Failed to load the light effect directory, is the path correct?")
	} else {
		log.WithFields(log.Fields{
			"num": lightPlayer.ListEffects(),
		}).Info("Loaded light effects")
	}

	driverLED, err := hardware.GetLEDDriver("ws281x")
	if err != nil {
		log.WithFields(log.Fields{
			"driver": "ws281x",
			"err":    err,
		}).Error("Failed to load LED driver")
	} else {
		err = driverLED.Setup(lightPlayer, cfg)
		if err != nil {
			log.WithFields(log.Fields{
				"driver": "ws281x",
				"err":    err,
			}).Error("Failed to setup LED driver")
		} else {
			defer driverLED.Close()
			log.WithFields(log.Fields{
				"driver": "ws281x",
			}).Info("Initialized LED driver")
		}
	}
	err = lightPlayer.SetEffect("Rainbow")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Could not set initial light effect")
	}

	// Prepare the music command module and initialize the speaker
	musicPlayer, err := music.NewPlayer("music-rest", &cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Player setup incomplete")
	} else {
		log.Info("Player setup complete")
	}
	log.Info("Gathering music files...")

	// Gather the music files
	err = musicPlayer.LoadPlaylists(cfg.Music.Path)
	if err != nil {
		log.WithFields(log.Fields{
			"path": cfg.Music.Path,
			"err":  err,
		}).Fatal("Failed to load the music directory, is the path correct?")
	} else {
		log.WithFields(log.Fields{
			"num": musicPlayer.ListPlaylists(),
		}).Info("Loaded playlists")
	}
	musicPlayer.Play()

	// Prepare the sound command module and initialize the speaker
	soundPlayer, err := sounds.NewPlayer("sounds-rest", cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Player setup incomplete")
	} else {
		log.Info("Player setup complete")
	}
	log.Info("Gathering sound files...")

	// Gather the sound files
	err = soundPlayer.LoadSounds(cfg.Sounds.Path)
	if err != nil {
		log.WithFields(log.Fields{
			"path": cfg.Sounds.Path,
			"err":  err,
		}).Fatal("Failed to load the sound directory, is the path correct?")
	} else {
		log.WithFields(log.Fields{
			"num": len(soundPlayer.ListSounds()),
		}).Info("Loaded sounds")
	}

	// Prepare shell command execution
	shellExec, err := shell.NewExecutor("shell-rest", cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Shell setup incomplete")
	} else {
		log.Info("Shell setup complete")
	}

	api := rest.Server{}
	srv := api.Start(cfg, musicPlayer, soundPlayer, lightPlayer, shellExec, profileSwitcher)
	api.AddHooks(cfg)

	// Start main loop
	go common.EventLoop()
	time.Sleep(time.Second)
	common.EventFire(common.Event{
		Origin: "app",
		Type:   "started",
		Name:   "Deichwave REST Server",
	})

	// GPIO Inputs
	driverGPIO, err := hardware.GetInputDriver("gpio")
	if err != nil {
		log.WithFields(log.Fields{
			"driver": "gpio",
			"err":    err,
		}).Error("Failed to load input driver")
	} else {
		err = driverGPIO.Setup(cfg, api)
		if err != nil {
			log.WithFields(log.Fields{
				"driver": "gpio",
				"err":    err,
			}).Error("Failed to setup input driver")
		} else {
			defer driverGPIO.Close()
			log.WithFields(log.Fields{
				"driver": "gpio",
			}).Info("Initialized input driver")
		}
	}

	// Wait until exit is requested
	sig := common.AwaitSignal()
	log.WithFields(log.Fields{
		"signal": sig,
	}).Warn("Received Signal")

	// Perform clean up
	common.EventFire(common.Event{
		Origin: "app",
		Type:   "stopped",
		Name:   "Deichwave REST Server",
	})

	log.Debug("Sleeping for a few seconds to allow for graceful shutdown")
	time.Sleep(1 * time.Second)
	api.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Unclean shutdown")
	}
	log.Info("Closing")
}
