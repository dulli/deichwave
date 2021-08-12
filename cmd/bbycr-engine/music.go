package main

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	bbycrgo "github.com/dulli/bbycrgo/pkg"
	"github.com/fhs/gompd/mpd"
	shellquote "github.com/kballard/go-shellquote"

	log "github.com/sirupsen/logrus"
)

const (
	ENDPOINT_MUSIC string = "music"

	MPD_NET    string = "unix"
	MPD_ADDR   string = "/run/mpd/socket"
	MPD_PASSWD string = ""
	MPD_SUBSYS string = "player"

	PING_INTERVAL time.Duration = 5

	MUSIC_DIRECTORY  string  = "music/playlists"
	MUSIC_INITIALVOL int     = 50
	MUSIC_SUPPFACTOR float64 = 0.5

	RNG_MAX int = 99
	RNG_MIN int = 25
)

type MusicCmd struct {
	mpd         *MusicMpd
	list        PlaylistList
	rng_level   int
	vol_level   int
	vol_factor  float64
	suppressors int
}
type MusicMpd struct {
	control *mpd.Client
	watcher *mpd.Watcher
}
type Playlist struct {
	name     string
	songs    []string
	position int
}
type PlaylistList []*Playlist

var valid_filetypes = map[string]bool{
	".mp3": true,
	".ogg": true,
}

var InvalidVolumeError = errors.New("Target volume has to be an integer between 0 and 100")
var InvalidRngError = errors.New("Target RNG level has to be an integer between 0 and 100")
var NoValueError = errors.New("Required parameter missing")
var NoPlaylistsError = errors.New("No playlists were found")

func (mcmd *MusicCmd) Queue() error {
	if len(mcmd.list) == 0 {
		return NoPlaylistsError
	}

	playlist_id := 0
	random := rand.Intn(100)
	if random > mcmd.rng_level {
		playlist_id = 1
	}

	playlist := mcmd.list[playlist_id]
	if playlist.position == 0 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(playlist.songs), func(i, j int) {
			playlist.songs[i], playlist.songs[j] = playlist.songs[j], playlist.songs[i]
		})
		log.WithFields(log.Fields{
			"playlist": playlist.name,
		}).Info("Shuffled playlist")
	}
	song := playlist.songs[playlist.position]
	err := mcmd.mpd.control.Add(song)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"song":     filepath.Base(song),
		"playlist": playlist.name,
	}).Debug("Queued song")

	playlist.position = playlist.position + 1
	if playlist.position >= len(playlist.songs) {
		playlist.position = 0
		log.WithFields(log.Fields{
			"playlist": playlist.name,
		}).Debug("Restarting playlist")
	}

	return nil
}

func (mcmd *MusicCmd) Crawl(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	dir := MUSIC_DIRECTORY
	base := filepath.Base(path)
	modpath := strings.Replace(path, MUSIC_DIRECTORY, "", 1)
	dir_parts := strings.Split(modpath, string(os.PathSeparator))
	if len(dir_parts) > 1 {
		dir = dir_parts[1]
	}

	if info.IsDir() {
		// Check if the folder is in the uppermost level, and add it as a playlist
		if base == dir && base != MUSIC_DIRECTORY {
			mcmd.list = append(mcmd.list, &Playlist{name: base})
			log.WithFields(log.Fields{
				"name": base,
			}).Info("Added playlist")
		}
		return nil
	} else {
		ext := filepath.Ext(path)
		if !valid_filetypes[ext] {
			return nil
		}
	}

	idx := -1
	for n := range mcmd.list {
		if mcmd.list[n].name == dir {
			idx = n
			break
		}
	}
	if idx < 0 {
		log.WithFields(log.Fields{
			"playlist": dir,
			"song":     base,
		}).Error("Playlist does not exist")
		return nil
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return nil
	}
	mcmd.list[idx].songs = append(mcmd.list[idx].songs, "file://"+path)
	log.WithFields(log.Fields{
		"playlist": mcmd.list[idx].name,
		"song":     base,
	}).Debug("Added song to playlist")

	return nil
}

func (mcmd *MusicCmd) UpdateVolume() error {
	volume := math.Round(float64(mcmd.vol_level) * mcmd.vol_factor)
	return mcmd.mpd.control.SetVolume(int(volume))
}

func (mcmd *MusicCmd) Parse(ev bbycrgo.Event) (int, error) {
	args, err := shellquote.Split(ev.Arguments)
	if err != nil {
		return 0, err
	}
	if len(args) == 0 {
		return 0, NoValueError
	}
	return strconv.Atoi(args[0])
}

func (mcmd *MusicCmd) Play(ev bbycrgo.Event) (string, error) {
	return "", mcmd.mpd.control.Play(-1)
}

func (mcmd *MusicCmd) Stop(ev bbycrgo.Event) (string, error) {
	return "", mcmd.mpd.control.Stop()
}

func (mcmd *MusicCmd) Next(ev bbycrgo.Event) (string, error) {
	return "", mcmd.mpd.control.Next()
}

func (mcmd *MusicCmd) Suppress(ev bbycrgo.Event) (string, error) {
	mcmd.suppressors = mcmd.suppressors + 1
	if mcmd.suppressors == 1 {
		mcmd.vol_factor = MUSIC_SUPPFACTOR
		return "", mcmd.UpdateVolume()
	}
	return "", nil
}

func (mcmd *MusicCmd) Unsuppress(ev bbycrgo.Event) (string, error) {
	mcmd.suppressors = mcmd.suppressors - 1
	if mcmd.suppressors < 1 {
		mcmd.suppressors = 0
		mcmd.vol_factor = 1
		return "", mcmd.UpdateVolume()
	}
	return "", nil
}

func (mcmd *MusicCmd) SetVolume(ev bbycrgo.Event) (string, error) {
	volume, err := mcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	if volume > 100 || volume < 0 {
		return "", InvalidVolumeError
	}
	mcmd.vol_level = volume
	return "", mcmd.UpdateVolume()
}

func (mcmd *MusicCmd) ChangeVolume(ev bbycrgo.Event) (string, error) {
	volume, err := mcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	mcmd.vol_level = mcmd.vol_level + volume
	if mcmd.vol_level > 100 {
		mcmd.vol_level = 100
	}
	if mcmd.vol_level < 0 {
		mcmd.vol_level = 0
	}
	return "", mcmd.UpdateVolume()
}

func (mcmd *MusicCmd) GetVolume(ev bbycrgo.Event) (string, error) {
	return strconv.Itoa(mcmd.vol_level), nil
}

func (mcmd *MusicCmd) SetRng(ev bbycrgo.Event) (string, error) {
	rng, err := mcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	if rng > 100 || rng < 0 {
		return "", InvalidRngError
	}
	mcmd.rng_level = rng
	return "", nil
}

func (mcmd *MusicCmd) ChangeRng(ev bbycrgo.Event) (string, error) {
	rng, err := mcmd.Parse(ev)
	if err != nil {
		return "", err
	}
	mcmd.rng_level = mcmd.rng_level + rng
	if mcmd.rng_level > RNG_MAX {
		mcmd.rng_level = RNG_MAX
	}
	if mcmd.rng_level < RNG_MIN {
		mcmd.rng_level = RNG_MIN
	}
	return "", nil
}

func (mcmd *MusicCmd) GetRng(ev bbycrgo.Event) (string, error) {
	return strconv.Itoa(mcmd.rng_level), nil
}

func (mcmd *MusicCmd) Info(ev bbycrgo.Event) (string, error) {
	attr, err := mcmd.mpd.control.CurrentSong()
	if err != nil {
		return "", err
	}

	var resp string
	data, err := json.Marshal(attr)
	if err == nil {
		resp = string(data)
	}
	return resp, err
}

func (mcmd *MusicCmd) Register() bbycrgo.EventHandlerList {
	cmds := bbycrgo.EventHandlerList{
		"play":          bbycrgo.EventHandler{mcmd.Play, nil},
		"stop":          bbycrgo.EventHandler{mcmd.Stop, nil},
		"skip":          bbycrgo.EventHandler{mcmd.Next, nil},
		"suppress":      bbycrgo.EventHandler{mcmd.Suppress, nil},
		"unsuppress":    bbycrgo.EventHandler{mcmd.Unsuppress, nil},
		"volume":        bbycrgo.EventHandler{mcmd.SetVolume, nil},
		"change-volume": bbycrgo.EventHandler{mcmd.ChangeVolume, nil},
		"get-volume":    bbycrgo.EventHandler{mcmd.GetVolume, nil},
		"info":          bbycrgo.EventHandler{mcmd.Info, nil},
		"rng":           bbycrgo.EventHandler{mcmd.SetRng, nil},
		"change-rng":    bbycrgo.EventHandler{mcmd.ChangeRng, nil},
		"get-rng":       bbycrgo.EventHandler{mcmd.GetRng, nil},
	}
	return cmds
}

func (mcmd *MusicCmd) EventLoop(progress *sync.WaitGroup) {
	defer mcmd.mpd.control.Close()
	defer mcmd.mpd.watcher.Close()
	cmds := mcmd.Register()
	bbycrgo.EventLoop(ENDPOINT_MUSIC, cmds, progress)
}

func (mmpd *MusicMpd) HandleErrors(mcmd *MusicCmd) {
	for err := range mmpd.watcher.Error {
		log.Error(err)
	}
}

func (mmpd *MusicMpd) HandleChanges(mcmd *MusicCmd) {
	for subsystem := range mmpd.watcher.Event {
		if subsystem != "player" {
			log.WithFields(log.Fields{
				"subsystem": subsystem,
			}).Debug("Skipped event")
			continue
		}
		attr, err := mmpd.control.Status()
		if err != nil {
			log.Error(err)
		}
		if _, ok := attr["nextsong"]; !ok {
			mcmd.Queue()
		}
	}
}

func (mmpd *MusicMpd) KeepAlive() {
	ticker := time.NewTicker(PING_INTERVAL * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			mmpd.control.Ping()

		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func (mmpd *MusicMpd) Setup(mcmd *MusicCmd) error {
	client, err := mpd.Dial(MPD_NET, MPD_ADDR)
	if err != nil {
		return err
	}

	err = client.Clear()
	if err != nil {
		return err
	}

	err = client.SetVolume(MUSIC_INITIALVOL)
	if err != nil {
		return err
	}

	err = client.SetVolume(MUSIC_INITIALVOL)
	if err != nil {
		return err
	}

	// Create a watcher that keeps track of errors and events
	watch, err := mpd.NewWatcher(MPD_NET, MPD_ADDR, MPD_PASSWD, MPD_SUBSYS)
	if err != nil {
		return err
	}

	mmpd.control = client
	mmpd.watcher = watch
	go mmpd.KeepAlive()
	go mmpd.HandleErrors(mcmd)
	go mmpd.HandleChanges(mcmd)

	return nil
}

func MusicSetup(progress *sync.WaitGroup) error {
	//err := filepath.Walk(SOUND_DIRECTORY, SoundCrawl)
	mcmd := MusicCmd{
		mpd:        new(MusicMpd),
		list:       make(PlaylistList, 0),
		rng_level:  RNG_MAX,
		vol_level:  MUSIC_INITIALVOL,
		vol_factor: 1.0,
	}
	err := mcmd.mpd.Setup(&mcmd)
	if err != nil {
		return err
	}

	err = filepath.Walk(MUSIC_DIRECTORY, mcmd.Crawl)
	if err != nil {
		return err
	}
	err = mcmd.Queue()
	if err != nil {
		return err
	}
	_, err = mcmd.Play(bbycrgo.Event{})
	if err != nil {
		return err
	}

	go mcmd.EventLoop(progress)

	return nil
}
