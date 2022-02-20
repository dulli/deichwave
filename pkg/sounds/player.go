// Provides a way to play sound effects
package sounds

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dulli/bbycrgo/pkg/common"
	log "github.com/sirupsen/logrus"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

// TODO implement callbacks
// TODO implement concurrent sound buffering

var ErrSoundNotFound = errors.New("sound could not be found")

// SoundPlayer is a collection of playable sound effects
type SoundPlayer interface {
	ListSounds() []string
	GetSound(name string) (Sound, error)
	LoadSounds(root string) error
}

// The player keeps track of the available and looped sounds
type soundPlayer struct {
	Name    string
	list    soundList
	rate    beep.SampleRate
	quality int
	ext     string
	rnd     string
	volume  int
}
type soundList map[string]Sound

// NewPlayer returns a new SoundPlayer object with the given name
// according to the provided config. (see pkg/common/config.go)
func NewPlayer(name string, cfg common.Config) (SoundPlayer, error) {
	player := soundPlayer{
		Name:    name,
		list:    make(soundList),
		rate:    beep.SampleRate(cfg.Audio.Rate),
		quality: cfg.Audio.Quality,
		ext:     cfg.Sounds.Ext,
		rnd:     cfg.Sounds.Randomizer,
		volume:  cfg.Sounds.Volume,
	}
	_, err := common.GetSpeaker(player.rate, cfg.Audio.Volume)
	return &player, err
}

// ListSounds gathers all available sound names and returns them as a slice
// of strings.
func (p *soundPlayer) ListSounds() []string {
	sounds := make([]string, 0)
	for key, sound := range p.list {
		if !sound.getSystem() {
			sounds = append(sounds, key)
		}
	}
	sort.Strings(sounds)
	return sounds
}

// GetSound returns the playable sound object for the given sound name.
func (p *soundPlayer) GetSound(name string) (Sound, error) {
	if val, ok := p.list[name]; ok {
		return val, nil
	}
	return nil, ErrSoundNotFound
}

// LoadSounds recursively crawls the given root directory and loads all
// sounds it can find. If it encounters a top-level directory, all
// audio files in that directory are added to a sound group. Any top-
// level audio files are added as individual sounds.
func (p *soundPlayer) LoadSounds(root string) error {
	root = filepath.Clean(root)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		s, err := visit(root, p.ext, p.rnd, p.rate, p.quality, p.volume, path, d.IsDir())
		if err != nil {
			return err
		}
		if s == nil {
			return nil
		}

		// Now we just have to check whether we have a new sound or one that is part of a group
		// which has already been added, in that case the selector is driven by the first sound added
		// and can only be overwritten using a randomizer
		if _, ok := p.list[s.GetName()]; ok {
			if s.getSelector() == SELECT_RANDOM {
				p.list[s.GetName()].setSelector(SELECT_RANDOM)
				log.WithFields(log.Fields{
					"name": s.GetName(),
				}).Debug("Randomized group")
			} else {
				p.list[s.GetName()].addBuffers(s.getBuffers())
				log.WithFields(log.Fields{
					"name": s.GetName(),
				}).Debug("Added buffer to group")
			}
		} else {
			p.list[s.GetName()] = s
			log.WithFields(log.Fields{
				"name": s.GetName(),
			}).Debug("Added sound")
		}

		return err
	})
	return err
}

// visit is called on all elements encountered while crawling a
// directory with LoadDir
func visit(
	root string,
	ext string,
	randomizer string,
	rate beep.SampleRate,
	quality int,
	volume int,
	currentPath string,
	isDir bool,
) (Sound, error) {
	relpath := strings.Replace(currentPath, root, "", 1)
	parent := strings.Trim(filepath.Dir(relpath), string(os.PathSeparator))
	element := filepath.Base(relpath)

	if element == "." {
		return nil, nil
	}

	if isDir {
		if parent == "" {
			// If the current element is a folder in the uppermost level, create a new sound group
			return NewSound(element, make(bufferList, 0), SELECT_SEQUENCE, volume), nil
		} else {
			// otherwise skip this folder, as we don't expect subfolders to exist
			return nil, fs.SkipDir
		}
	} else if element == randomizer {
		// If the current element is a file named after the magic randomizer string, return a randomizer
		return NewSound(parent, make(bufferList, 0), SELECT_RANDOM, volume), nil
	} else if filepath.Ext(currentPath) == ext {
		// If the current element is a file with the correct file extension, add it as a sound
		// where the name is either the filename or the parent folders name
		if parent != "" {
			element = parent
		} else {
			element = strings.TrimSuffix(element, ext)
		}

		// If we got this far, the file is actually a sound file we want to add so we can buffer it
		buffer, err := loadFile(currentPath, rate, quality)
		if err != nil {
			return nil, err
		}
		return NewSound(element, bufferList{buffer}, SELECT_SINGLE, volume), nil
	}
	return nil, nil
}

func loadFile(
	path string,
	rate beep.SampleRate,
	quality int,
) (*beep.Buffer, error) {
	data, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var streamer beep.StreamSeekCloser
	var format beep.Format
	switch filepath.Ext(path) {
	case ".wav":
		streamer, format, err = wav.Decode(data)
	case ".mp3":
		streamer, format, err = mp3.Decode(data)
	case ".ogg":
		streamer, format, err = vorbis.Decode(data)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"file": path,
			"err":  err,
		}).Error("Could not resample a sound file")
		return nil, err
	}

	buffer := beep.NewBuffer(format)
	if rate != format.SampleRate {
		resampled := beep.Resample(quality, format.SampleRate, rate, streamer)
		streamer.Close()
		buffer.Append(resampled)

		log.WithFields(log.Fields{
			"file": path,
			"is":   format.SampleRate,
			"want": rate,
		}).Debug("Resampled a sound file")
	} else {
		buffer.Append(streamer)
		streamer.Close()

		log.WithFields(log.Fields{
			"file": path,
			"is":   format.SampleRate,
			"want": rate,
		}).Debug("Buffered a sound file")
	}
	data.Close()

	return buffer, nil
}
