# TOML Configuration Design

## Goal

Add a TOML configuration file that controls statusline layout and per-component display options, mirroring the original shell version's configurability.

## Config Path

`~/.claude/statusline/config.toml`

Generated with sensible defaults on first run if the file does not exist.

## Config Format

```toml
# Claude Code Statusline Configuration

[layout]
lines = [
  ["repo_info"],
  ["bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"],
  ["cost_monthly", "cost_weekly", "cost_daily", "cost_live", "contxt_window", "session_mode"],
]

[components.bedrock_model]
show_region = true

[components.context_window]
show_tokens = true
```

### Layout

`lines` is an array of arrays. Each inner array lists the component names for one statusline row. A component absent from all lines is disabled. The order within each line controls render order.

### Per-Component Settings

| Section | Key | Type | Default | Effect |
|---------|-----|------|---------|--------|
| `components.bedrock_model` | `show_region` | bool | `true` | Show AWS region after model name |
| `components.context_window` | `show_tokens` | bool | `true` | Show token counts (e.g., `90K/200K`) alongside percentage |

## Go Package

New package: `internal/config/`

```go
type Config struct {
    Layout     LayoutConfig                `toml:"layout"`
    Components map[string]ComponentConfig  `toml:"components"`
}

type LayoutConfig struct {
    Lines [][]string `toml:"lines"`
}

type ComponentConfig struct {
    ShowRegion *bool `toml:"show_region,omitempty"`
    ShowTokens *bool `toml:"show_tokens,omitempty"`
}
```

### Key Functions

- `Load(path string) (*Config, error)` — read and parse TOML; create default file if missing
- `DefaultConfig() *Config` — return the default configuration
- `(c *Config) GetBool(component, key string, fallback bool) bool` — retrieve a per-component boolean with a fallback

### Dependency

`github.com/BurntSushi/toml` — the standard Go TOML library.

## Integration

1. `main.go` calls `config.Load(configPath)` early.
2. Layout lines come from `cfg.Layout.Lines` instead of the hardcoded slice.
3. Components that support per-component settings receive `*config.Config` and query their options at render time.
4. If the config file is missing, `Load` writes the default and returns it.

## Design Decisions

- **Presence in lines = enabled.** No separate `enabled = true/false` per component. Simpler to reason about.
- **Pointer bools for optional fields.** `*bool` with `omitempty` lets us distinguish "not set" from "set to false" and apply defaults cleanly.
- **Flat ComponentConfig.** All per-component keys live in one struct. New keys require a struct field but no schema changes.
- **BurntSushi/toml over stdlib.** Go's stdlib has no TOML support. BurntSushi/toml is the de facto standard, well-maintained, and minimal.
