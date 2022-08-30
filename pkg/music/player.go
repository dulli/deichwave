package music

import (
	"errors"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dulli/deichwave/pkg/common"
	log "github.com/sirupsen/logrus"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	"github.com/bogem/id3v2/v2"
)

var ErrPlaylistNotFound = errors.New("playlist could not be found")

type MusicPlayer interface {
	ListPlaylists() []string
	GetPlaylist(name string) (Playlist, error)
	LoadPlaylists(root string) error
	Play()
	Pause()
	Stop()
	Next()
	NowPlaying() SongInfo
	GetChance(name string) (int, error)
}

// The player keeps track of the available playlists
type musicPlayer struct {
	Name            string
	list            map[string]Playlist
	rate            beep.SampleRate
	quality         int
	ext             string
	nextSong        chan bool
	keys            []string
	chances         []int
	chancesMin      []int
	chancesMax      []int
	volume          int
	currentPlaylist string
	control         *beep.Ctrl
	nowPlaying      SongInfo
}

func NewPlayer(name string, cfg common.Config) (MusicPlayer, error) {
	player := musicPlayer{
		Name:       name,
		list:       make(map[string]Playlist),
		rate:       beep.SampleRate(cfg.Audio.Rate),
		quality:    cfg.Audio.Quality,
		ext:        cfg.Music.Ext,
		nextSong:   make(chan bool),
		chancesMin: cfg.Music.StartRNG,
		chancesMax: cfg.Music.EndRNG,
		volume:     cfg.Music.Volume,
		control:    nil,
	}
	_, err := common.GetSpeaker(player.rate, cfg.Audio.Volume)
	rand.Seed(time.Now().UnixNano())
	go player.run()
	return &player, err
}

// Gather all available playlists
func (p *musicPlayer) ListPlaylists() []string {
	return p.keys
}

func (p *musicPlayer) GetPlaylist(name string) (Playlist, error) {
	if val, ok := p.list[name]; ok {
		return val, nil
	}
	return nil, ErrPlaylistNotFound
}

func (p *musicPlayer) LoadPlaylists(root string) error {
	root = filepath.Clean(root)

	// Add the directories as playlists
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path == root {
			return nil
		}

		directory := filepath.Base(path)
		p.list[directory] = &playlist{Name: directory, Songs: make([]Song, 0), Pos: 0}
		log.WithFields(log.Fields{
			"list": directory,
		}).Debug("Added a playlist")
		return nil
	})

	// Add the directoriy contents as songs to the playlists
	for key := range p.list {
		p.keys = append(p.keys, key)
		directory := filepath.Join(root, key)
		err = filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) != p.ext {
				return nil
			}

			fname := filepath.Base(path)
			song := NewSong(fname[:strings.LastIndexByte(fname, '.')], path)
			p.list[key].addSong(song)

			log.WithFields(log.Fields{
				"list": key,
				"file": path,
			}).Debug("Added a music file")
			return nil
		})
		p.list[key].shuffle()
	}
	sort.Strings(p.keys)
	return err
}

func (p *musicPlayer) Next() {
	// use rng to determine next playlist
	maxrng := 0
	p.updateChances()
	for _, c := range p.chances {
		maxrng += c
	}
	random := rand.Intn(maxrng)
	log.WithFields(log.Fields{
		"chances": p.chances,
		"maxrng":  maxrng,
		"random":  random,
	}).Debug("RNG Result")

	playlistIndex := 0
	level := 0
	for _, chance := range p.chances {
		level += chance
		if random <= level {
			break
		}
		playlistIndex += 1
	}
	p.currentPlaylist = p.keys[playlistIndex]
	pl := p.list[p.currentPlaylist]
	song := pl.Next()
	go p.play(song)
}

func (p *musicPlayer) play(s Song) {
	data, err := os.Open(s.getPath())
	if err != nil {
		log.WithFields(log.Fields{
			"song": s.GetName(),
			"err":  err,
		}).Error("Song not found")
		return
	}

	streamer, format, err := mp3.Decode(data)
	if err != nil {
		log.WithFields(log.Fields{
			"song": s.GetName(),
			"err":  err,
		}).Error("Could not decode song")
		return
	}

	var volstreamer beep.Streamer = streamer
	if p.rate != format.SampleRate {
		volstreamer = beep.Resample(4, format.SampleRate, p.rate, streamer)

		log.WithFields(log.Fields{
			"song": s.GetName(),
			"is":   format.SampleRate,
			"want": p.rate,
		}).Debug("Resampling song")
	}

	volume := &effects.Volume{
		Streamer: volstreamer,
		Base:     2,
		Volume:   math.Log2(float64(p.volume) / 100),
		Silent:   false,
	}
	if p.control != nil {
		speaker.Lock()
		p.control.Streamer = nil
		p.control = nil
		speaker.Unlock()
	}
	p.control = &beep.Ctrl{Streamer: beep.Seq(volume, beep.Callback(func() {
		streamer.Close()
		log.WithFields(log.Fields{
			"name": s.GetName(),
		}).Debug("Finished a song")

		p.nextSong <- true
	})), Paused: false}
	common.Play(p.control)

	// Load id3 tags of currently playing song
	tag, err := id3v2.Open(s.getPath(), id3v2.Options{Parse: true})
	if err != nil {
		log.Error("Error while opening mp3 file: ", err)
		return
	}
	defer tag.Close()

	p.nowPlaying = SongInfo{
		Artist:   tag.Artist(),
		Title:    tag.Title(),
		Playlist: p.currentPlaylist,
	}
	pictures := tag.GetFrames(tag.CommonID("Attached picture"))
	pic, ok := pictures[0].(id3v2.PictureFrame)
	if ok {
		p.nowPlaying.Picture = SongPicture{
			Data: pic.Picture,
			Mime: pic.MimeType,
		}
	} else {
		p.nowPlaying.Picture = SongPicture{}
	}

	common.EventFire(common.Event{
		Origin: "music",
		Type:   "playing",
	})

	log.WithFields(log.Fields{
		"name": s.GetName(),
	}).Info("Playing a song")
}

func (p *musicPlayer) Play() {
	if p.control != nil {
		speaker.Lock()
		p.control.Paused = false
		speaker.Unlock()
	} else {
		p.Next()
	}

	log.Debug("Started music playback")
}

func (p *musicPlayer) Pause() {
	if p.control != nil {
		speaker.Lock()
		p.control.Paused = true
		speaker.Unlock()
	}

	log.Debug("Paused music playback")
}

func (p *musicPlayer) Stop() {
	if p.control != nil {
		speaker.Lock()
		p.control.Streamer = nil
		p.control = nil
		speaker.Unlock()
	}

	log.Debug("Stopped music playback")
}

func (p *musicPlayer) NowPlaying() SongInfo {
	return p.nowPlaying
}

func (p *musicPlayer) updateChances() {
	intensity := common.GetIntensity()
	newChances := make([]int, len(p.chancesMin))
	for i := range p.chancesMin {
		y0 := float64(p.chancesMin[i])
		y1 := float64(p.chancesMax[i])
		x := float64(intensity)
		newChances[i] = int(y0 + math.Round(x*(y1-y0)/100))
	}
	p.chances = newChances
}

func (p *musicPlayer) GetChance(name string) (int, error) {
	p.updateChances()
	for i, v := range p.keys {
		if v == name {
			return p.chances[i], nil
		}
	}
	return -1, ErrPlaylistNotFound
}

func (p *musicPlayer) run() {
	for range p.nextSong {
		p.Next()
	}
}
