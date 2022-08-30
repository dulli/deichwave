package sounds

import (
	"math"
	"math/rand"
	"strings"

	"github.com/dulli/deichwave/pkg/common"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	log "github.com/sirupsen/logrus"
)

const (
	SELECT_SINGLE int = iota
	SELECT_RANDOM
	SELECT_SEQUENCE
)

// Sound is a playable sound effect
type Sound interface {
	Play()
	Loop()
	Unloop()
	GetName() string
	getSystem() bool
	getSelector() int
	setSelector(selector int)
	getBuffers() bufferList
	addBuffers(buffers bufferList)
	GetBufferCount() int
}

// A single sound can consist of multiple files, the selector determines which is played
type sound struct {
	Name     string
	system   bool
	Buffers  bufferList
	Selector int
	index    int
	loop     *beep.Ctrl
	volume   int
}
type bufferList []*beep.Buffer

// NewSound returns a new playable sound object with a given name,
// a list of buffers and the method used to select one of the buffers
// when the song is played.
func NewSound(name string, buffers bufferList, selector int, volume int) Sound {
	isSystem := false
	if strings.HasSuffix(name, ".system") {
		isSystem = true
		name = strings.TrimSuffix(name, ".system")
	}
	return &sound{Name: name, system: isSystem, Buffers: buffers, Selector: selector, volume: volume}
}

// Play starts playback of the next buffer that is to be played according
// to the selector method attached to the ssound.
func (s *sound) Play() {
	buffer := s.Buffers[s.index]
	streamer := buffer.Streamer(0, buffer.Len())
	volume := &effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   math.Log2(float64(s.volume) / 100),
		Silent:   false,
	}
	common.Play(volume)
	log.WithFields(log.Fields{
		"name":  s.Name,
		"index": s.index,
	}).Info("Played a sound")

	if s.Selector == SELECT_SEQUENCE {
		s.index = s.index + 1
		if s.index == len(s.Buffers) {
			s.index = 0
		}
	} else if s.Selector == SELECT_RANDOM {
		s.index = rand.Intn(len(s.Buffers))
	}
}

// Loop starts and indefinitely loops the next buffer.
func (s *sound) Loop() {
	if s.loop == nil {
		buffer := s.Buffers[s.index]
		streamer := buffer.Streamer(0, buffer.Len())
		s.loop = &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}
		volume := &effects.Volume{
			Streamer: s.loop,
			Base:     2,
			Volume:   math.Log2(float64(s.volume) / 100),
			Silent:   false,
		}
		common.Play(volume)

		log.WithFields(log.Fields{
			"name":  s.Name,
			"index": s.index,
		}).Debug("Looped a sound")
	}
}

// Unloop stops the currently looped buffer.
func (s *sound) Unloop() {
	if s.loop != nil {
		speaker.Lock()
		s.loop.Streamer = nil
		s.loop = nil
		speaker.Unlock()

		log.WithFields(log.Fields{
			"name":  s.Name,
			"index": s.index,
		}).Debug("Unlooped a sound")
	}
}

func (s *sound) GetName() string {
	return s.Name
}

func (s *sound) getSystem() bool {
	return s.system
}

func (s *sound) getSelector() int {
	return s.Selector
}

func (s *sound) setSelector(selector int) {
	s.Selector = selector
}

func (s *sound) getBuffers() bufferList {
	return s.Buffers
}

func (s *sound) addBuffers(buffers bufferList) {
	s.Buffers = append(s.Buffers, buffers...)
}

func (s *sound) GetBufferCount() int {
	return len(s.Buffers)
}
