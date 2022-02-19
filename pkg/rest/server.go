package rest

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/dulli/bbycrgo/pkg/common"
	"github.com/dulli/bbycrgo/pkg/lights"
	"github.com/dulli/bbycrgo/pkg/music"
	"github.com/dulli/bbycrgo/pkg/sounds"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/r3labs/sse/v2"

	log "github.com/sirupsen/logrus"
)

// Implements ServerInterface
type Server struct {
	music  music.MusicPlayer
	sounds sounds.SoundPlayer
	lights lights.Renderer
}

func (server Server) Start(m music.MusicPlayer, s sounds.SoundPlayer, l lights.Renderer) error {
	server.music = m
	server.sounds = s
	server.lights = l

	// REST
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	// r.Use(middleware.AllowContentType("application/json")) // TODO add a switch to some functions to also be able to return html snippets for htmx?
	r.Mount("/api/v0", Handler(&server))

	// SSE
	sseServer := sse.New()
	sseServer.AutoReplay = false
	sseServer.AutoStream = true
	common.EventListen(func(ev common.Event) {
		data, _ := json.Marshal(ev)
		sseServer.Publish("events", &sse.Event{
			Data: data,
		})
	})
	r.Group(func(r chi.Router) {
		r.Get("/sse", sseServer.ServeHTTP)
	})

	// Static file host
	fs := http.FileServer(http.Dir("web/public"))
	r.Group(func(r chi.Router) {
		r.Get("/*", fs.ServeHTTP)
	})

	log.WithFields(log.Fields{
		"address": fmt.Sprintf("http://%s:3000/app.html", getLocalIP()),
	}).Info("Started REST API server")
	// TODO make port configurable
	return http.ListenAndServe(":3000", r)
}

func getLocalIP() net.IP {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

// func (s BbycrServer) makeHypermedia(path string, ids []string) []string {
// 	list := make([]string, 0)
// 	for _, id := range ids {
// 		list = append(list, fmt.Sprintf("%s/%s", path, id))
// 	}
// 	return list
// }

// List all endpoints
// (GET /)
func (s Server) GetRoot(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusNotImplemented)
	render.JSON(w, r, "NOK")
}

// List all playlists
// (GET /music)
func (s Server) GetMusic(w http.ResponseWriter, r *http.Request) {
	data := EntityList{
		Entity: s.music.ListPlaylists(),
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

// Get now playing info
// (GET /music/playing)
func (s Server) GetMusicPlaying(w http.ResponseWriter, r *http.Request) {
	np := s.music.NowPlaying()

	var dataURL string
	if np.Picture.Data != nil {
		dataURL = fmt.Sprintf(
			"data:%s;base64,%s",
			np.Picture.Mime,
			base64.StdEncoding.EncodeToString(np.Picture.Data),
		)
	} else {
		dataURL = ""
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, SongInfo{
		Artist:   np.Artist,
		Title:    np.Title,
		Playlist: np.Playlist,
		Image:    &dataURL,
	})
}

// Get playlist details
// (GET /music/{playlist})
func (s Server) GetMusicPlaylist(w http.ResponseWriter, r *http.Request, playlist Playlist) {
	pl, err := s.music.GetPlaylist(string(playlist))
	if errors.As(err, &music.ErrPlaylistNotFound) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, err.Error())
		return
	}

	data := EntityList{
		Entity: pl.ListSongs(),
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

// Get position in playlist
// (GET /music/{playlist}/position)
func (s Server) GetMusicPlaylistPosition(w http.ResponseWriter, r *http.Request, playlist string) {
	pl, err := s.music.GetPlaylist(string(playlist))
	if errors.As(err, &music.ErrPlaylistNotFound) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, err.Error())
		return
	}

	data := PlaylistPosition{
		Position: pl.GetPosition(),
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

// Get playlist chance
// (GET /music/{playlist}/chance)
func (s Server) GetMusicPlaylistChance(w http.ResponseWriter, r *http.Request, playlist string) {
	pl, err := s.music.GetChance(string(playlist))
	if errors.As(err, &music.ErrPlaylistNotFound) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, err.Error())
		return
	}

	data := PlaylistChance{
		Chance: pl,
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

// Skip the next song in a playlist
// (POST /music/{playlist}/skip)
func (s Server) PostMusicPlaylistSkip(w http.ResponseWriter, r *http.Request, playlist Playlist) {
	pl, err := s.music.GetPlaylist(string(playlist))
	if errors.As(err, &music.ErrPlaylistNotFound) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, err.Error())
		return
	}

	pl.Skip()
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// Start music playback
// (POST /music/play)
func (s Server) PostMusicPlay(w http.ResponseWriter, r *http.Request) {
	s.music.Play()
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// Stop music playback
// (POST /music/pause)
func (s Server) PostMusicPause(w http.ResponseWriter, r *http.Request) {
	s.music.Pause()
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// Stop music playback
// (POST /music/stop)
func (s Server) PostMusicStop(w http.ResponseWriter, r *http.Request) {
	s.music.Stop()
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// Play the next track
// (POST /music/next)
func (s Server) PostMusicNext(w http.ResponseWriter, r *http.Request) {
	s.music.Next()
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// List all sounds
// (GET /sounds)
func (s Server) GetSounds(w http.ResponseWriter, r *http.Request) {
	data := EntityList{
		Entity: s.sounds.ListSounds(),
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

// Get sound details
// (GET /sounds/{sound})
func (s Server) GetSoundsSound(w http.ResponseWriter, r *http.Request, sound Sound) {
	snd, err := s.sounds.GetSound(string(sound))
	if errors.As(err, &sounds.ErrSoundNotFound) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, err.Error())
		return
	}

	data := SoundDetails{
		Name:        snd.GetName(),
		BufferCount: snd.GetBufferCount(),
		Links: SoundActions{
			Play:   fmt.Sprintf("/sounds/%s/play", sound),
			Loop:   fmt.Sprintf("/sounds/%s/loop", sound),
			Unloop: fmt.Sprintf("/sounds/%s/unloop", sound),
		},
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

// Play a sound
// (POST /sounds/{sound}/play)
func (s Server) PostSoundsPlay(w http.ResponseWriter, r *http.Request, sound Sound) {
	snd, err := s.sounds.GetSound(string(sound))
	if errors.As(err, &sounds.ErrSoundNotFound) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, err.Error())
		return
	}

	snd.Play()
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// Loop a sound
// (POST /sounds/{sound}/loop)
func (s Server) PostSoundsLoop(w http.ResponseWriter, r *http.Request, sound Sound) {
	snd, err := s.sounds.GetSound(string(sound))
	if errors.As(err, &sounds.ErrSoundNotFound) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, err.Error())
		return
	}

	snd.Loop()
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// Stop a looped sound
// (POST /sounds/{sound}/unloop)
func (s Server) PostSoundsUnloop(w http.ResponseWriter, r *http.Request, sound Sound) {
	snd, err := s.sounds.GetSound(string(sound))
	if errors.As(err, &sounds.ErrSoundNotFound) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, "NOK")
		return
	}

	snd.Unloop()
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// List all light effects
// (GET /lights)
func (s Server) GetLights(w http.ResponseWriter, r *http.Request) {
	data := EntityList{
		Entity: s.lights.ListEffects(),
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

// Stop all light effects
// (POST /lights/clear)
func (s Server) PostLightsClear(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusNotImplemented)
	render.JSON(w, r, "NOK")
}

// Get light effect details
// (GET /lights/{effect})
func (s Server) GetLightsEffect(w http.ResponseWriter, r *http.Request, effect LightEffect) {
	render.Status(r, http.StatusNotImplemented)
	render.JSON(w, r, "NOK")
}

// Start a light effect
// (POST /lights/{effect}/set)
func (s Server) PostLightsStart(w http.ResponseWriter, r *http.Request, effect LightEffect) {
	err := s.lights.SetEffect(string(effect))
	if errors.As(err, &lights.ErrEffectNotFound) {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, err.Error())
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// Stop a light effect
// (POST /lights/{effect}/stop)
func (s Server) PostLightsStop(w http.ResponseWriter, r *http.Request, effect LightEffect) {
	render.Status(r, http.StatusNotImplemented)
	render.JSON(w, r, "NOK")
}

// Get volume
// (GET /system/volume)
func (s Server) GetSystemVolume(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, AudioLevel{
		Level: common.GetVolume(),
	})
}

// Set Volume
// (POST /system/volume)
func (s Server) PostSystemVolume(w http.ResponseWriter, r *http.Request) {
	var vol PostSystemVolumeJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&vol); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, "NOK")
		return
	}
	common.SetVolume(vol.Level)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}

// Get Intensity
// (GET /system/intensity)
func (s Server) GetSystemIntensity(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, AudioLevel{
		Level: common.GetIntensity(),
	})
}

// Set Intensity
// (POST /system/intensity)
func (s Server) PostSystemIntensity(w http.ResponseWriter, r *http.Request) {
	var intensity PostSystemVolumeJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&intensity); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, "NOK")
		return
	}
	common.SetIntensity(intensity.Level)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "OK")
}
