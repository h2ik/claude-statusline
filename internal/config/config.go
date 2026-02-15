package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the statusline configuration.
type Config struct {
	Layout     Layout                     `toml:"layout"`
	Components map[string]ComponentConfig `toml:"components"`
}

// Layout defines which components appear on each line.
type Layout struct {
	Lines [][]string `toml:"lines"`
}

// ComponentConfig holds per-component configuration options.
// Pointer bools distinguish "not set" from "set to false".
type ComponentConfig struct {
	ShowRegion *bool `toml:"show_region,omitempty"`
	ShowTokens *bool `toml:"show_tokens,omitempty"`
}

// Load reads a TOML configuration file from the given path.
// If the file does not exist, it creates one with default values and returns
// the default configuration.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}

		// File doesn't exist - create it with defaults
		cfg := DefaultConfig()
		if err := writeConfig(path, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Initialize Components map if nil after parsing
	if cfg.Components == nil {
		cfg.Components = make(map[string]ComponentConfig)
	}

	return &cfg, nil
}

// DefaultConfig returns a Config with the default layout and component settings.
func DefaultConfig() *Config {
	showRegion := true
	showTokens := true

	return &Config{
		Layout: Layout{
			Lines: [][]string{
				{"repo_info"},
				{"bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"},
				{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"},
			},
		},
		Components: map[string]ComponentConfig{
			"bedrock_model": {
				ShowRegion: &showRegion,
			},
			"context_window": {
				ShowTokens: &showTokens,
			},
		},
	}
}

// GetBool retrieves a boolean value from the ComponentConfig for the given
// component and key name. Returns fallback if the component or key is not set.
func (c *Config) GetBool(component, key string, fallback bool) bool {
	comp, ok := c.Components[component]
	if !ok {
		return fallback
	}

	switch key {
	case "show_region":
		if comp.ShowRegion != nil {
			return *comp.ShowRegion
		}
	case "show_tokens":
		if comp.ShowTokens != nil {
			return *comp.ShowTokens
		}
	}

	return fallback
}

// writeConfig writes the configuration to the given path as TOML, creating
// parent directories as needed.
func writeConfig(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header comment
	if _, err := f.WriteString("# Claude Code Statusline Configuration\n\n"); err != nil {
		return err
	}

	return toml.NewEncoder(f).Encode(cfg)
}
