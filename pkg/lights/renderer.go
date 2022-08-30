package lights

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/dulli/deichwave/pkg/common"

	log "github.com/sirupsen/logrus"
)

var ErrEffectNotFound = errors.New("light effect could not be found")

type Renderer interface {
	ListEffects() []string
	LoadEffects(root string) error
	GetLEDCount() int
	GetGroupCount() map[string]int
	SetEffect(name string) error
	StopEffect(name string) error
	ReceiveFrame(func([][]LEDState))
}

type scriptRenderer struct {
	Name      string
	ext       string
	current   string
	tick      int
	leds      *tengo.Array
	effects   map[string]*tengo.Compiled
	info      map[string]effectInfo
	nextFrame chan bool
	timer     *time.Timer
	callbacks []func([][]LEDState)
	state     [][]LEDState
	tcount    int
	gcount    map[string]int
	history   []string
}
type effectInfo struct {
	frameTime time.Duration
	maxTick   int
}

type LEDState struct {
	ColorIndex int64
	Brightness float64
}

func NewRenderer(name string, cfg common.Config) (Renderer, error) {
	renderer := scriptRenderer{
		Name:      name,
		ext:       cfg.Lights.Ext,
		tick:      0,
		effects:   make(map[string]*tengo.Compiled),
		info:      make(map[string]effectInfo),
		nextFrame: make(chan bool),
		callbacks: make([]func([][]LEDState), 0),
		state:     make([][]LEDState, len(cfg.LEDs)),
		gcount:    make(map[string]int),
	}
	totalCount := 0
	groups := make([]tengo.Object, len(cfg.LEDs))
	for name, group := range cfg.LEDs {
		totalCount += group.Count
		renderer.gcount[name] = group.Count
		b := make([]tengo.Object, group.Count)
		c := make([]tengo.Object, group.Count)
		for idx := range b {
			b[idx] = &tengo.Float{Value: 0}
			c[idx] = &tengo.Int{Value: 0}
		}

		groups[group.Order-1] = &tengo.Map{
			Value: map[string]tengo.Object{
				"name":       &tengo.String{Value: name},
				"count":      &tengo.Int{Value: int64(group.Count)},
				"brightness": &tengo.Array{Value: b},
				"color":      &tengo.Array{Value: c},
			},
		}
	}

	renderer.tcount = totalCount

	renderer.leds = &tengo.Array{Value: groups}
	go renderer.run()
	return &renderer, nil
}

func (r *scriptRenderer) GetLEDCount() int {
	return r.tcount
}

func (r *scriptRenderer) GetGroupCount() map[string]int {
	return r.gcount
}

// ListEffects gathers all available effect names and returns them as a slice
// of strings.
func (r *scriptRenderer) ListEffects() []string {
	idx := 0
	effects := make([]string, len(r.info))
	for key := range r.info {
		effects[idx] = key
		idx++
	}
	sort.Strings(effects)
	return effects
}

func (r *scriptRenderer) LoadEffects(root string) error {
	root = filepath.Clean(root)

	// Add the directories as playlists
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if path == root {
			return nil
		}

		file := filepath.Base(path)
		ext := filepath.Ext(file)
		if ext != r.ext {
			return nil
		}

		effect := strings.TrimSuffix(file, ext)
		r.compile(root, effect)
		log.WithFields(log.Fields{
			"name": effect,
		}).Debug("Added effect")
		return nil
	})
	return err
}

func (r *scriptRenderer) compile(root string, effect string) {
	exec := fmt.Sprintf(`
		effect := import("%s")
		info_maxtick := effect.info.maxtick
		info_frametime := effect.info.frametime
		leds = effect.frame(leds, tick)
	`, effect)
	script := tengo.NewScript([]byte(exec))
	err := script.SetImportDir(root)
	if err != nil {
		panic(err)
	}
	script.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	script.EnableFileImport(true)

	_ = script.Add("leds", r.leds)
	_ = script.Add("tick", r.tick)

	compiled, err := script.Run()
	if err != nil {
		panic(err)
	}
	r.effects[effect] = compiled
	r.info[effect] = effectInfo{
		maxTick:   compiled.Get("info_maxtick").Int(),
		frameTime: time.Duration(float64(time.Second) * compiled.Get("info_frametime").Float()),
	}
}

func (r *scriptRenderer) SetEffect(name string) error {
	if _, ok := r.info[name]; !ok {
		return ErrEffectNotFound
	}
	r.history = append(r.history, name)
	log.WithFields(log.Fields{
		"name": name,
	}).Info("Setting light effect")
	r.Next(name)
	return nil
}

func (r *scriptRenderer) StopEffect(name string) error {
	// Find the last occurrence of the given effect in the history
	effect_idx := -1
	for i, v := range r.history {
		if v == name {
			effect_idx = i
		}
	}
	if effect_idx == -1 {
		return ErrEffectNotFound
	}

	is_last := effect_idx == len(r.history)-1
	if is_last {
		r.history = r.history[:len(r.history)-1]
		if len(r.history) > 0 {
			r.Next(r.history[len(r.history)-1])
		}
	} else {
		r.history = append(r.history[:effect_idx], r.history[effect_idx+1:]...)
	}
	log.WithFields(log.Fields{
		"name": name,
	}).Info("Stopped light effect")
	return nil
}

func (r *scriptRenderer) ReceiveFrame(cb func([][]LEDState)) {
	r.callbacks = append(r.callbacks, cb)
}

func (r *scriptRenderer) Next(effect string) {
	if effect == "" {
		effect = r.current
	}
	if effect == "" {
		return
	}

	// Schedule next frame request if necessary (i.e. frameTime is > 0 and another
	// frame wasn't already manually requested)
	if r.timer != nil {
		r.timer.Stop()
	}
	if r.info[effect].frameTime > 0 {
		r.timer = time.AfterFunc(r.info[effect].frameTime, func() {
			r.nextFrame <- true
		})
	}
	r.updateFrame(effect)

	// Decode the tengo variables and feed them into the registered callbacks
	for gidx, group := range r.leds.Value {
		if len(r.state[gidx]) == 0 {
			countObj, _ := group.IndexGet(&tengo.String{Value: "count"})
			count := countObj.(*tengo.Int)
			r.state[gidx] = make([]LEDState, count.Value)
		}
		for lidx := range r.state[gidx] {
			tidx := &tengo.Int{Value: int64(lidx)}

			cObj, _ := group.IndexGet(&tengo.String{Value: "color"})
			c, _ := cObj.IndexGet(tidx)
			r.state[gidx][lidx].ColorIndex = c.(*tengo.Int).Value

			bObj, _ := group.IndexGet(&tengo.String{Value: "brightness"})
			b, _ := bObj.IndexGet(tidx)
			r.state[gidx][lidx].Brightness = b.(*tengo.Float).Value
		}
	}
	for _, cb := range r.callbacks {
		cb(r.state)
	}
}

func (r *scriptRenderer) updateFrame(effect string) {
	if effect != r.current {
		r.tick = 0
		r.current = effect
	} else if r.info[effect].maxTick > 0 && r.tick < r.info[effect].maxTick-1 {
		r.tick++
	} else {
		r.tick = 0
	}

	compiled := r.effects[effect]
	_ = compiled.Set("tick", r.tick)
	err := compiled.Run()
	if err != nil {
		panic(err)
	}
}

func (r *scriptRenderer) run() {
	for range r.nextFrame {
		r.Next("")
	}
}
