package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all user-configurable settings for LazyPorts.
type Config struct {
	General     General     `toml:"general"`
	Filters     Filters     `toml:"filters"`
	Keybindings Keybindings `toml:"keybindings"`
}

type General struct {
	RefreshInterval int    `toml:"refresh_interval"` // seconds; 0 = manual only
	Theme           string `toml:"theme"`
}

type Filters struct {
	DefaultSort     string `toml:"default_sort"` // port | process | pid
	ShowSystemPorts bool   `toml:"show_system_ports"`
	FocusRange      string `toml:"focus_range"` // e.g. "3000-9000"
}

type Keybindings struct {
	Kill      string `toml:"kill"`
	ForceKill string `toml:"force_kill"`
	Quit      string `toml:"quit"`
}

// Defaults returns the default configuration when no file is present.
func Defaults() Config {
	return Config{
		General: General{
			RefreshInterval: 0,
			Theme:           "catppuccin",
		},
		Filters: Filters{
			DefaultSort:     "port",
			ShowSystemPorts: true,
		},
		Keybindings: Keybindings{
			Kill:      "k",
			ForceKill: "K",
			Quit:      "q",
		},
	}
}

// Load reads ~/.config/lazyports/config.toml. Returns defaults if the file
// does not exist; parse errors are non-fatal (returns defaults with error).
func Load() (Config, error) {
	cfg := Defaults()

	dir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	path := filepath.Join(dir, "lazyports", "config.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err // caller logs the error, uses defaults
	}
	return cfg, nil
}
