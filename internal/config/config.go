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
	Style     string       `toml:"style"`
	IconStyle string       `toml:"icon_style"`
	Padding   int          `toml:"padding"`
	Lines     []LayoutLine `toml:"lines"`
}

// LayoutLine describes the left and right component groups for a single
// statusline row.
type LayoutLine struct {
	Left  []string `toml:"left"`
	Right []string `toml:"right"`
}

// ComponentConfig holds per-component configuration options.
// Pointer bools distinguish "not set" from "set to false".
type ComponentConfig struct {
	ShowRegion      *bool `toml:"show_region,omitempty"`
	ShowTokens      *bool `toml:"show_tokens,omitempty"`
	ShowVelocity    *bool `toml:"show_velocity,omitempty"`
	ShowCostPerLine *bool `toml:"show_cost_per_line,omitempty"`
}

// legacyLayout mirrors the old flat lines format ([][]string) so we can detect
// and auto-migrate configs written before left/right support was added.
type legacyLayout struct {
	Lines [][]string `toml:"lines"`
}

// legacyConfig is the full config shape using the old layout format.
type legacyConfig struct {
	Layout     legacyLayout               `toml:"layout"`
	Components map[string]ComponentConfig `toml:"components"`
}

// Load reads a TOML configuration file from the given path.
// If the file does not exist, it creates one with default values and returns
// the default configuration.
//
// For backward compatibility, Load supports both the new left/right layout
// format and the old flat lines format. Old-format configs are auto-migrated:
// all components go to Left, Right stays empty, Style defaults to "default".
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

	// First pass: try unmarshaling into the new Config struct (left/right format).
	var cfg Config
	newFmtErr := toml.Unmarshal(data, &cfg)

	if newFmtErr == nil && len(cfg.Layout.Lines) > 0 {
		// New format parsed successfully.
		if cfg.Components == nil {
			cfg.Components = make(map[string]ComponentConfig)
		}
		if cfg.Layout.Style == "" {
			cfg.Layout.Style = "default"
		}
		if cfg.Layout.Padding == 0 {
			cfg.Layout.Padding = 5
		}
		return &cfg, nil
	}

	// Second pass: try the legacy flat [][]string format.
	var legacy legacyConfig
	if err := toml.Unmarshal(data, &legacy); err != nil {
		// Neither format worked. Return the original new-format error if we
		// had one, otherwise the legacy error.
		if newFmtErr != nil {
			return nil, newFmtErr
		}
		return nil, err
	}

	if len(legacy.Layout.Lines) > 0 {
		// Migrate: all components go to Left, Right stays empty.
		cfg.Layout.Style = "default"
		cfg.Layout.Lines = make([]LayoutLine, len(legacy.Layout.Lines))
		for i, line := range legacy.Layout.Lines {
			cfg.Layout.Lines[i] = LayoutLine{
				Left:  line,
				Right: nil,
			}
		}
		if legacy.Components != nil {
			cfg.Components = legacy.Components
		} else {
			cfg.Components = make(map[string]ComponentConfig)
		}
		return &cfg, nil
	}

	// Neither format had lines -- return what we have with a default style.
	if cfg.Components == nil {
		cfg.Components = make(map[string]ComponentConfig)
	}
	if cfg.Layout.Style == "" {
		cfg.Layout.Style = "default"
	}
	return &cfg, nil
}

// DefaultConfig returns a Config with the default layout and component settings.
// The default layout places all components on the Left side with an empty Right.
func DefaultConfig() *Config {
	showRegion := true
	showTokens := true
	showVelocity := true
	showCostPerLine := true

	return &Config{
		Layout: Layout{
			Style:     "default",
			IconStyle: "emoji",
			Padding:   5,
			Lines: []LayoutLine{
				{Left: []string{"repo_info"}},
				{Left: []string{"bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"}},
				{Left: []string{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"}},
				{Left: []string{"burn_rate", "cache_efficiency", "block_projection", "code_productivity"}},
			},
		},
		Components: map[string]ComponentConfig{
			"bedrock_model": {
				ShowRegion: &showRegion,
			},
			"context_window": {
				ShowTokens: &showTokens,
			},
			"code_productivity": {
				ShowVelocity:    &showVelocity,
				ShowCostPerLine: &showCostPerLine,
			},
		},
	}
}

// DefaultPowerlineConfig returns a Config with a powerline-style layout that
// uses both Left and Right component groups per line.
func DefaultPowerlineConfig() *Config {
	showRegion := true
	showTokens := true
	showVelocity := true
	showCostPerLine := true

	return &Config{
		Layout: Layout{
			Style:     "powerline",
			IconStyle: "nerd-font",
			Padding:   5,
			Lines: []LayoutLine{
				{
					Left:  []string{"repo_info", "bedrock_model", "model_info"},
					Right: []string{"commits", "submodules", "version_info", "time_display"},
				},
				{
					Left:  []string{"cost_monthly", "cost_weekly", "cost_daily", "cost_live"},
					Right: []string{"context_window", "session_mode", "burn_rate", "cache_efficiency", "block_projection", "code_productivity"},
				},
			},
		},
		Components: map[string]ComponentConfig{
			"bedrock_model": {
				ShowRegion: &showRegion,
			},
			"context_window": {
				ShowTokens: &showTokens,
			},
			"code_productivity": {
				ShowVelocity:    &showVelocity,
				ShowCostPerLine: &showCostPerLine,
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
	case "show_velocity":
		if comp.ShowVelocity != nil {
			return *comp.ShowVelocity
		}
	case "show_cost_per_line":
		if comp.ShowCostPerLine != nil {
			return *comp.ShowCostPerLine
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
	defer func() { _ = f.Close() }()

	// Write header comment
	if _, err := f.WriteString("# Claude Code Statusline Configuration\n\n"); err != nil {
		return err
	}

	return toml.NewEncoder(f).Encode(cfg)
}
