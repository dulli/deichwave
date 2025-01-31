package common

import (
	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Meta struct {
		Version string
		Build   string
		Name    string `env:"NAME" env-default:"Deichwave"`
	} `env-prefix:"META_"`
	File  string `env:"CONFIG" env-default:"config/default.toml"`
	Debug bool   `env:"DEBUG" env-default:"false"`
	Audio struct {
		Rate    int `toml:"rate" env:"RATE" env-default:"44100"`
		Buffer  int `toml:"buffer" env:"BUFFER" env-default:"5000"`
		Quality int `toml:"quality" env:"QUALITY" env-default:"6"`
		Volume  int `toml:"volume" env:"VOLUME" env-default:"10"`
	} `toml:"audio" env-prefix:"AUDIO_"`
	Sounds struct {
		Path       string `toml:"path" env:"DIR" env-default:"data/sounds/effects"`
		Ext        string `toml:"ext" env:"EXT" env-default:".ogg"`
		Randomizer string `toml:"randomizer" env:"RND" env-default:".random"`
		Volume     int    `toml:"volume" env:"VOLUME" env-default:"100"`
	} `toml:"sounds" env-prefix:"SOUNDS_"`
	Music struct {
		Path     string `toml:"path" env:"DIR" env-default:"data/music/playlists"`
		Ext      string `toml:"ext" env:"EXT" env-default:".ogg"`
		Volume   int    `toml:"volume" env:"VOLUME" env-default:"50"`
		StartRNG []int  `toml:"startrng" env:"STARTRNG" env-default:"95,5"`
		EndRNG   []int  `toml:"endrng" env:"ENDRNG" env-default:"30,70"`
	} `toml:"music" env-prefix:"MUSIC_"`
	Lights struct {
		Path string `toml:"path" env:"DIR" env-default:"data/lights/effects"`
		Ext  string `toml:"ext" env:"EXT" env-default:".tengo"`
	} `toml:"lights" env-prefix:"LIGHTS_"`
	Shell map[string]map[string][]string `toml:"shell"`
	Hooks map[string][]string            `toml:"hooks"`
	REST  struct {
		Port      int    `toml:"port" env:"PORT" env-default:"3000"`
		HTTPSPort int    `toml:"https_port" env:"HTTPS_PORT" env-default:"3443"`
		TLSCA     string `toml:"tls_ca" env:"TLS_CA" env-default:""`
		TLSCrt    string `toml:"tls_crt" env:"TLS_CRT" env-default:""`
		TLSKey    string `toml:"tls_key" env:"TLS_KEY" env-default:""`
	} `toml:"rest" env-prefix:"REST_"`
	WebIO struct {
		Buttons  map[string][]string `toml:"buttons"`
		Switches map[string][]string `toml:"switches"`
	} `toml:"webio"`
	LEDs map[string]struct {
		Order int `toml:"order"`
		Count int `toml:"count"`
	} `toml:"leds"`
	GPIO map[string]struct {
		Chip     string   `toml:"chip"`
		Pins     []int    `toml:"pins"`
		Type     string   `toml:"type"`
		Debounce int      `toml:"debounce"`
		Actions  []string `toml:"actions"`
	} `toml:"gpio"`
	Hardware struct {
		LEDBrightness int `toml:"led-brightness" env:"LED_BRIGHTNESS" env-default:"100"`
		LEDPin        int `toml:"led-pin" env:"LED_PIN" env-default:"18"`
	} `toml:"hardware" env-prefix:"HW_"`
}

func Configure(cfg *Config) {
	// Load configuration, if no config file exists, fall back to env only
	var err error
	if err := cleanenv.ReadConfig(".env", cfg); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("No .env file was parsed")
	} else {
		log.Debug("Using .env file")
	}
	if err = cleanenv.ReadEnv(cfg); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Could not initialize configuration loader")
	}
	if err := cleanenv.ReadConfig(cfg.File, cfg); err != nil {
		log.WithFields(log.Fields{
			"cfg": cfg.File,
			"err": err,
		}).Debug("Could not create configuration from file")
	}

	// Initialize the command
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.WithFields(log.Fields{
		"cfg": cfg,
	}).Debug("Configuration values")
	log.Info("Configuration completed")
}
