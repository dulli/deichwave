package common

import (
	"errors"
	"math"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

var ErrSpeakerContextReused = errors.New("the speaker was already initialized, the existing context is reused")
var initialized int
var mixer *beep.Mixer
var volumeLevel int
var volumeStream *effects.Volume
var intensityLevel int

func GetSpeaker(rate beep.SampleRate, buffersize int, volume int) (int, error) {
	if initialized == 0 {
		err := speaker.Init(rate, buffersize)
		if err != nil {
			return 0, err
		}

		initialized = rate.N(time.Second)
		return initialized, nil
	}
	mixer = &beep.Mixer{}
	volumeStream = &effects.Volume{
		Streamer: mixer,
		Base:     2,
		Volume:   1,
		Silent:   true,
	}
	SetIntensity(0)
	SetVolume(volume)
	speaker.Play(volumeStream)
	return initialized, ErrSpeakerContextReused
}

func Play(streamers ...beep.Streamer) {
	speaker.Lock()
	mixer.Add(streamers...)
	speaker.Unlock()
}

func SetVolume(volume int) {
	if volume > 100 {
		volume = 100
	} else if volume < 0 {
		volume = 0
	}
	volumeLevel = volume
	if volume == 0 {
		volumeStream.Silent = true
	} else {
		volumeStream.Silent = false
		volumeStream.Volume = math.Log2(float64(volume) / 100)
	}
	EventFire(Event{
		Origin: "audio",
		Type:   "volume",
	})
}

func ChangeVolume(delta int) {
	SetVolume(volumeLevel + delta)
}

func GetVolume() int {
	return volumeLevel
}

func SetIntensity(intensity int) {
	if intensity > 100 {
		intensity = 100
	} else if intensity < 0 {
		intensity = 0
	}
	intensityLevel = intensity
	EventFire(Event{
		Origin: "audio",
		Type:   "intensity",
	})
}

func ChangeIntensity(delta int) {
	SetIntensity(intensityLevel + delta)
}

func GetIntensity() int {
	return intensityLevel
}
