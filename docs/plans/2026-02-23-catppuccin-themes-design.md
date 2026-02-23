# Catppuccin Theme System Design

## Goal

Add all four Catppuccin flavors (Mocha, Latte, Frappe, Macchiato) and a `theme` config field so users can switch palettes. The current hardcoded Mocha palette becomes one of four selectable themes. Existing users see zero change.

## Requirements

1. Support all four canonical Catppuccin flavors
2. Add `theme` field to `[layout]` in TOML config
3. Default to `catppuccin-mocha` when no theme is specified
4. Auto-invert contrast for light themes (Latte uses light text on dark accent backgrounds)
5. Backward-compatible: existing configs without `theme` produce identical output

## Design

### Theme Data Model

New file `internal/render/theme.go` defines a `Theme` struct with named color fields:

```go
type Theme struct {
    Name     string
    Base     lipgloss.Color // Background base (dark for Mocha, light for Latte)
    Overlay0 lipgloss.Color // Dimmed/labels
    Text     lipgloss.Color // Default text
    Green    lipgloss.Color // Clean/good
    Red      lipgloss.Color // Critical
    Yellow   lipgloss.Color // Warning
    Blue     lipgloss.Color // Paths/info
    Mauve    lipgloss.Color // Accent
    Peach    lipgloss.Color // Costs
    Teal     lipgloss.Color // Secondary
}
```

Four package-level instances hold each flavor's hex values. A `ThemeByName(name string) (*Theme, bool)` function provides lookup; unknown names return `(ThemeMocha, false)`.

### Catppuccin Palette Reference

| Role | Mocha | Latte | Frappe | Macchiato |
|------|-------|-------|--------|-----------|
| Base | #1e1e2e | #eff1f5 | #292c3c | #24273a |
| Overlay0 | #6c7086 | #9ca0b0 | #626880 | #6e738d |
| Text | #cdd6f4 | #4c4f69 | #c6d0f5 | #cad3f5 |
| Blue | #89b4fa | #1e66f5 | #8caaee | #8aadf4 |
| Green | #a6e3a1 | #40a02b | #a6d189 | #a6da95 |
| Red | #f38ba8 | #d20f39 | #e78284 | #ed8796 |
| Yellow | #f9e2af | #df8e1d | #e5c890 | #eed49f |
| Peach | #fab387 | #fe640b | #ef9f76 | #f5a97f |
| Mauve | #cba6f7 | #8839ef | #ca9ee6 | #c6a0f6 |
| Teal | #94e2d5 | #179299 | #81c8be | #8bd5ca |

### Renderer Integration

The `Renderer` struct gains a `theme *Theme` field set at construction:

```go
func New(theme *Theme) *Renderer
```

All nine color helper methods (`Dimmed`, `Text`, `Green`, `Red`, `Yellow`, `Blue`, `Mauve`, `Peach`, `Teal`) read from `r.theme` instead of the removed global `Color*` variables.

Components remain untouched. They call `renderer.Blue()` as before and receive the active theme's blue, whatever that may be.

### Segment Category Integration

The hardcoded `segmentCategories` map and `colorDark` global are removed. `SegmentCategoryFor` gains a theme parameter:

```go
func SegmentCategoryFor(componentName string, theme *Theme) SegmentCategory
```

Each semantic group maps to theme colors:

| Group | Background | Foreground |
|-------|-----------|-----------|
| Info | theme.Blue | theme.Base |
| Cost | theme.Peach | theme.Base |
| Metrics | theme.Teal | theme.Base |
| Activity | theme.Green | theme.Base |
| Meta | theme.Mauve | theme.Base |
| Dim | theme.Overlay0 | theme.Text |

Using `theme.Base` as foreground enables auto-invert: dark themes put dark text on bright backgrounds; Latte puts light text on its darker accent backgrounds.

`PowerlineStyle` holds a `*Theme` field via `NewPowerlineStyle(theme *Theme)`.

### Config Integration

Add `Theme` field to `Layout`:

```go
type Layout struct {
    Theme     string       `toml:"theme"`
    Style     string       `toml:"style"`
    IconStyle string       `toml:"icon_style"`
    Padding   int          `toml:"padding"`
    Lines     []LayoutLine `toml:"lines"`
}
```

Example config:

```toml
[layout]
theme = "catppuccin-frappe"
style = "powerline"
icon_style = "nerd-font"
```

Valid values: `catppuccin-mocha`, `catppuccin-latte`, `catppuccin-frappe`, `catppuccin-macchiato`. Empty or missing defaults to `catppuccin-mocha`. Unknown values fall back to Mocha with a stderr warning.

### Wiring in main.go

```go
theme, ok := render.ThemeByName(cfg.Layout.Theme)
if !ok {
    fmt.Fprintf(os.Stderr, "unknown theme %q, using catppuccin-mocha\n", cfg.Layout.Theme)
}
renderer := render.New(theme)
```

## Testing Strategy

**Unit tests:**
- `ThemeByName` returns correct theme for each valid name and `(Mocha, false)` for unknown names
- Renderer color methods use the active theme's colors, not hardcoded Mocha values
- `SegmentCategoryFor` returns the correct theme color pairs for each group across all four themes
- Config loading: `theme` field populates correctly; missing `theme` defaults to `catppuccin-mocha`

**Integration tests:**
- Full pipeline per theme: rendered ANSI output contains that theme's hex values
- Backward compatibility: configs without `theme` produce identical output to current behavior

## Out of Scope

- Custom user-defined color overrides
- Non-Catppuccin themes (e.g., Dracula, Solarized)
- Terminal background auto-detection
- Per-component theme overrides
