package music

import (
	"math/rand"

	"github.com/dulli/bbycrgo/pkg/events"
	log "github.com/sirupsen/logrus"
)

type Playlist interface {
	ListSongs() []string
	Next() Song
	Skip()
	GetPosition() int
	addSong(song Song)
	shuffle()
}

type playlist struct {
	Name  string
	Songs []Song
	Pos   int
}

func (p *playlist) ListSongs() []string {
	lists := make([]string, len(p.Songs))
	for idx := range p.Songs {
		lists[idx] = p.Songs[idx].GetName()
	}
	return lists
}

func (p *playlist) Next() Song {
	p.incPos()
	return p.Songs[p.Pos-1]
}

func (p *playlist) Skip() {
	p.incPos()
	log.WithFields(log.Fields{
		"name": p.Name,
	}).Info("Skipping a song")
}

func (p *playlist) GetPosition() int {
	return p.Pos
}

func (p *playlist) incPos() {
	p.Pos += 1
	if p.Pos >= len(p.Songs) {
		p.Pos = 0
		p.shuffle()
	}
	events.Fire(events.Event{
		Origin: "music",
		Name:   p.Name,
		Type:   "position",
	})
}

func (p *playlist) addSong(song Song) {
	p.Songs = append(p.Songs, song)
}

func (p *playlist) shuffle() {
	rand.Shuffle(len(p.Songs), func(i, j int) {
		p.Songs[i], p.Songs[j] = p.Songs[j], p.Songs[i]
	})
	events.Fire(events.Event{
		Origin: "music",
		Name:   p.Name,
		Type:   "shuffle",
	})
}
