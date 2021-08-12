package main

import (
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	bbycrgo "github.com/dulli/bbycrgo/pkg"
	shellquote "github.com/kballard/go-shellquote"
	log "github.com/sirupsen/logrus"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

// TODO implement music suppression

const (
	ENDPOINT_SOUNDS string = "sounds"

	SELECT_SINGLE   int = 0
	SELECT_RANDOM   int = 1
	SELECT_SEQUENCE int = 2

	SOUND_DIRECTORY  string = "sounds/processed"
	SOUND_EXT        string = ".wav"
	SOUND_RANDOMIZER string = ".random"

	RESAMPLE_QUALITY int = 6
)

type SoundCmd struct {
	list  SoundList
	loops LoopList
	rate  beep.SampleRate
}
type Sound struct {
	name     string
	buffers  BufferList
	selector int
	index    int
}
type SoundList map[string]*Sound
type BufferList []*beep.Buffer
type LoopList map[string]*beep.Ctrl

var SoundNotFoundError = errors.New("Sound could not be found")

func (s *Sound) Play() error {
	buffer := s.buffers[s.index]
	streamer := buffer.Streamer(0, buffer.Len())
	speaker.Play(streamer)
	log.WithFields(log.Fields{
		"name":  s.name,
		"index": s.index,
	}).Debug("Played a sound")

	if s.selector == SELECT_SEQUENCE {
		s.index = s.index + 1
		if s.index == len(s.buffers) {
			s.index = 0
		}
	} else if s.selector == SELECT_RANDOM {
		s.index = rand.Intn(len(s.buffers))
	}
	return nil
}

func (s *Sound) Loop(loops LoopList) error {
	if _, ok := loops[s.name]; ok {
		log.Error("Loop already exists")
		return nil
	}

	buffer := s.buffers[s.index]
	streamer := buffer.Streamer(0, buffer.Len())
	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}
	loops[s.name] = ctrl
	speaker.Play(ctrl)

	log.WithFields(log.Fields{
		"name":  s.name,
		"index": s.index,
	}).Debug("Looped a sound")
	return nil
}

func (s *Sound) Unloop(loops LoopList) error {
	if _, ok := loops[s.name]; ok {
		speaker.Lock()
		loops[s.name].Streamer = nil
		speaker.Unlock()
		delete(loops, s.name)

		log.WithFields(log.Fields{
			"name":  s.name,
			"index": s.index,
		}).Debug("Unlooped a sound")
	} else {
		log.Warn("Trying to stop non-existing loop")
	}
	return nil
}

func (scmd *SoundCmd) Load(path string) (*beep.Buffer, error) {
	data, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	streamer, format, err := wav.Decode(data)
	if err != nil {
		return nil, err
	}

	if scmd.rate == 0 {
		log.WithFields(log.Fields{
			"rate": format.SampleRate,
		}).Warn("Sample rate wasn't set, setting to first encountered")
		scmd.rate = format.SampleRate
	}

	buffer := beep.NewBuffer(format)
	if scmd.rate != format.SampleRate {
		resampled := beep.Resample(RESAMPLE_QUALITY, format.SampleRate, scmd.rate, streamer)
		streamer.Close()
		buffer.Append(resampled)

		log.WithFields(log.Fields{
			"file": path,
			"old":  format.SampleRate,
			"new":  scmd.rate,
		}).Debug("Resampled a sound file")
	} else {
		buffer.Append(streamer)
		streamer.Close()

		log.WithFields(log.Fields{
			"file": path,
		}).Debug("Buffered a sound file")
	}
	return buffer, nil
}

func (scmd *SoundCmd) Crawl(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	var sound_name string
	dir := SOUND_DIRECTORY
	base := filepath.Base(path)
	modpath := strings.Replace(path, SOUND_DIRECTORY, "", 1)
	dir_parts := strings.Split(modpath, string(os.PathSeparator))
	if len(dir_parts) > 1 {
		dir = dir_parts[1]
	}

	if info.IsDir() {
		// Check if the folder is in the uppermost level, and add it as a sound
		if base == dir && base != SOUND_DIRECTORY {
			scmd.list[base] = &Sound{base, make(BufferList, 0), SELECT_SEQUENCE, 0}
			log.WithFields(log.Fields{
				"name": base,
			}).Info("Added sound group")
		}
		return nil
	} else {
		ext := filepath.Ext(path)

		if dir == base {
			sound_name = strings.TrimSuffix(base, ext)
		} else {
			sound_name = dir
		}

		if base == SOUND_RANDOMIZER {
			if _, ok := scmd.list[sound_name]; ok {
				scmd.list[sound_name].selector = SELECT_RANDOM
				scmd.list[sound_name].index = 0 // TODO randomize first index
			} else {
				log.WithFields(log.Fields{
					"path": path,
				}).Warn("Trying to randomize a non-existing sound")
				return nil
			}
		}

		if ext != SOUND_EXT {
			return nil
		}
	}

	// If we got this far, the file is actually a sound file we want to add so
	// we can buffer it
	buffer, err := scmd.Load(path)
	if err != nil {
		return err
	}

	// Now we just have to check whether its a single sound or in a group, for
	// the latter case, a named sound should already exists due to the folder
	// handling above
	if _, ok := scmd.list[sound_name]; ok {
		scmd.list[sound_name].buffers = append(scmd.list[sound_name].buffers, buffer)
	} else {
		sound_buffers := BufferList{buffer}
		scmd.list[sound_name] = &Sound{sound_name, sound_buffers, SELECT_SINGLE, 0}
		log.WithFields(log.Fields{
			"name": sound_name,
		}).Info("Added sound")
	}
	return nil
}

func (scmd *SoundCmd) Parse(ev bbycrgo.Event) (*Sound, error) {
	var sound *Sound

	args, err := shellquote.Split(ev.Arguments)
	if err != nil {
		return sound, err
	}
	name := args[0]
	if sound, ok := scmd.list[name]; ok {
		return sound, nil
	}
	return sound, SoundNotFoundError
}

func (scmd *SoundCmd) Play(ev bbycrgo.Event) (string, error) {
	sound, err := scmd.Parse(ev)
	if err != nil {
		return "", err
	}
	return "", sound.Play()
}

func (scmd *SoundCmd) Loop(ev bbycrgo.Event) (string, error) {
	sound, err := scmd.Parse(ev)
	if err != nil {
		return "", err
	}
	return "", sound.Loop(scmd.loops)
}

func (scmd *SoundCmd) Unloop(ev bbycrgo.Event) (string, error) {
	sound, err := scmd.Parse(ev)
	if err != nil {
		return "", err
	}
	return "", sound.Unloop(scmd.loops)
}

func (scmd *SoundCmd) Register() bbycrgo.EventHandlerList {
	sound_idx := 0
	sound_names := make([]string, len(scmd.list))
	for key := range scmd.list {
		sound_names[sound_idx] = key
		sound_idx++
	}
	cmds := bbycrgo.EventHandlerList{
		"play":   bbycrgo.EventHandler{scmd.Play, sound_names},
		"loop":   bbycrgo.EventHandler{scmd.Loop, sound_names},
		"unloop": bbycrgo.EventHandler{scmd.Unloop, sound_names},
	}
	return cmds
}

func (scmd *SoundCmd) EventLoop(progress *sync.WaitGroup) {
	cmds := scmd.Register()
	bbycrgo.EventLoop(ENDPOINT_SOUNDS, cmds, progress)
}

func SoundSetup(progress *sync.WaitGroup) error {
	// Prepare the sound command module, gather the sound files and initialize the speaker
	scmd := SoundCmd{
		list:  make(SoundList),
		loops: make(LoopList),
	}
	err := filepath.Walk(SOUND_DIRECTORY, scmd.Crawl)
	if err != nil {
		return err
	}
	err = speaker.Init(scmd.rate, scmd.rate.N(time.Second/10))
	if err != nil {
		return err
	}

	go scmd.EventLoop(progress)

	return nil
}
