# Catppuccin Theme System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add all four Catppuccin flavors (Mocha, Latte, Frappe, Macchiato) and a `theme` config field so users can switch palettes without changing any component code.

**Architecture:** A `Theme` struct holds named color fields; four package-level instances define each flavor's hex values. The `Renderer` receives a theme at construction and reads from it instead of global color constants. `SegmentCategoryFor` takes a theme parameter and uses `theme.Base` as foreground, enabling automatic contrast inversion for the Latte light theme. The config gains a `theme` string field that defaults to `catppuccin-mocha`.

**Tech Stack:** Go, `github.com/charmbracelet/lipgloss` v1.1.0, `github.com/BurntSushi/toml` v1.6.0

---

## Pre-Flight Check

Run all existing tests before touching anything:

```bash
go test ./... 2>&1
```

Expected: all tests pass. If any fail, stop and fix before proceeding.

---

### Task 1: Theme Data Model

**Files:**
- Create: `internal/render/theme.go`
- Create: `internal/render/theme_test.go`

**Step 1: Write the failing test**

Create `internal/render/theme_test.go`:

```go
package render

import (
	"testing"
)

func TestThemeByName_KnownThemes(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantBlue string
	}{
		{"catppuccin-mocha", "catppuccin-mocha", "#89b4fa"},
		{"catppuccin-latte", "catppuccin-latte", "#1e66f5"},
		{"catppuccin-frappe", "catppuccin-frappe", "#8caaee"},
		{"catppuccin-macchiato", "catppuccin-macchiato", "#8aadf4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme, ok := ThemeByName(tt.name)
			if !ok {
				t.Errorf("ThemeByName(%q) returned ok=false, want true", tt.name)
			}
			if theme.Name != tt.wantName {
				t.Errorf("theme.Name = %q, want %q", theme.Name, tt.wantName)
			}
			if string(theme.Blue) != tt.wantBlue {
				t.Errorf("theme.Blue = %q, want %q", theme.Blue, tt.wantBlue)
			}
		})
	}
}

func TestThemeByName_UnknownReturnsMoncha(t *testing.T) {
	theme, ok := ThemeByName("solarized-dark")
	if ok {
		t.Error("ThemeByName(unknown) returned ok=true, want false")
	}
	if theme.Name != "catppuccin-mocha" {
		t.Errorf("ThemeByName(unknown) returned theme %q, want catppuccin-mocha", theme.Name)
	}
}

func TestThemeByName_EmptyReturnsMoncha(t *testing.T) {
	theme, ok := ThemeByName("")
	if ok {
		t.Error("ThemeByName('') returned ok=true, want false")
	}
	if theme.Name != "catppuccin-mocha" {
		t.Errorf("ThemeByName('') returned theme %q, want catppuccin-mocha", theme.Name)
	}
}

func TestTheme_MochaBaseIsDark(t *testing.T) {
	// Mocha is a dark theme -- Base should be the dark background
	if string(ThemeMocha.Base) != "#1e1e2e" {
		t.Errorf("ThemeMocha.Base = %q, want #1e1e2e", ThemeMocha.Base)
	}
}

func TestTheme_LatteBaseIsLight(t *testing.T) {
	// Latte is a light theme -- Base should be the light background
	if string(ThemeLatte.Base) != "#eff1f5" {
		t.Errorf("ThemeLatte.Base = %q, want #eff1f5", ThemeLatte.Base)
	}
}

func TestTheme_AllFlavorsHaveAllFields(t *testing.T) {
	themes := []*Theme{&ThemeMocha, &ThemeLatte, &ThemeFrappe, &ThemeMacchiato}
	fields := []struct {
		name  string
		value func(*Theme) string
	}{
		{"Base", func(t *Theme) string { return string(t.Base) }},
		{"Overlay0", func(t *Theme) string { return string(t.Overlay0) }},
		{"Text", func(t *Theme) string { return string(t.Text) }},
		{"Green", func(t *Theme) string { return string(t.Green) }},
		{"Red", func(t *Theme) string { return string(t.Red) }},
		{"Yellow", func(t *Theme) string { return string(t.Yellow) }},
		{"Blue", func(t *Theme) string { return string(t.Blue) }},
		{"Mauve", func(t *Theme) string { return string(t.Mauve) }},
		{"Peach", func(t *Theme) string { return string(t.Peach) }},
		{"Teal", func(t *Theme) string { return string(t.Teal) }},
	}
	for _, theme := range themes {
		for _, field := range fields {
			if field.value(theme) == "" {
				t.Errorf("theme %q has empty %s field", theme.Name, field.name)
			}
		}
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/render/ -run TestTheme -v 2>&1
```

Expected: compilation failure — `ThemeByName`, `ThemeMocha`, `ThemeLatte`, etc. not defined.

**Step 3: Create `internal/render/theme.go`**

```go
package render

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for a statusline theme.
type Theme struct {
	Name     string
	Base     lipgloss.Color // Background base (dark for dark themes, light for Latte)
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

// Catppuccin flavor instances. Use ThemeByName to look up by config string.
var (
	ThemeMocha = Theme{
		Name:     "catppuccin-mocha",
		Base:     lipgloss.Color("#1e1e2e"),
		Overlay0: lipgloss.Color("#6c7086"),
		Text:     lipgloss.Color("#cdd6f4"),
		Green:    lipgloss.Color("#a6e3a1"),
		Red:      lipgloss.Color("#f38ba8"),
		Yellow:   lipgloss.Color("#f9e2af"),
		Blue:     lipgloss.Color("#89b4fa"),
		Mauve:    lipgloss.Color("#cba6f7"),
		Peach:    lipgloss.Color("#fab387"),
		Teal:     lipgloss.Color("#94e2d5"),
	}

	ThemeLatte = Theme{
		Name:     "catppuccin-latte",
		Base:     lipgloss.Color("#eff1f5"),
		Overlay0: lipgloss.Color("#9ca0b0"),
		Text:     lipgloss.Color("#4c4f69"),
		Green:    lipgloss.Color("#40a02b"),
		Red:      lipgloss.Color("#d20f39"),
		Yellow:   lipgloss.Color("#df8e1d"),
		Blue:     lipgloss.Color("#1e66f5"),
		Mauve:    lipgloss.Color("#8839ef"),
		Peach:    lipgloss.Color("#fe640b"),
		Teal:     lipgloss.Color("#179299"),
	}

	ThemeFrappe = Theme{
		Name:     "catppuccin-frappe",
		Base:     lipgloss.Color("#292c3c"),
		Overlay0: lipgloss.Color("#626880"),
		Text:     lipgloss.Color("#c6d0f5"),
		Green:    lipgloss.Color("#a6d189"),
		Red:      lipgloss.Color("#e78284"),
		Yellow:   lipgloss.Color("#e5c890"),
		Blue:     lipgloss.Color("#8caaee"),
		Mauve:    lipgloss.Color("#ca9ee6"),
		Peach:    lipgloss.Color("#ef9f76"),
		Teal:     lipgloss.Color("#81c8be"),
	}

	ThemeMacchiato = Theme{
		Name:     "catppuccin-macchiato",
		Base:     lipgloss.Color("#24273a"),
		Overlay0: lipgloss.Color("#6e738d"),
		Text:     lipgloss.Color("#cad3f5"),
		Green:    lipgloss.Color("#a6da95"),
		Red:      lipgloss.Color("#ed8796"),
		Yellow:   lipgloss.Color("#eed49f"),
		Blue:     lipgloss.Color("#8aadf4"),
		Mauve:    lipgloss.Color("#c6a0f6"),
		Peach:    lipgloss.Color("#f5a97f"),
		Teal:     lipgloss.Color("#8bd5ca"),
	}
)

// ThemeByName returns the theme for the given name.
// If the name is unknown or empty, it returns ThemeMocha and false.
func ThemeByName(name string) (Theme, bool) {
	switch name {
	case "catppuccin-mocha":
		return ThemeMocha, true
	case "catppuccin-latte":
		return ThemeLatte, true
	case "catppuccin-frappe":
		return ThemeFrappe, true
	case "catppuccin-macchiato":
		return ThemeMacchiato, true
	default:
		return ThemeMocha, false
	}
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/render/ -run TestTheme -v 2>&1
```

Expected: all `TestTheme*` tests PASS.

**Step 5: Commit**

```bash
git add internal/render/theme.go internal/render/theme_test.go
/git-commit
```

Suggested message: `feat(render): Add Theme struct and four Catppuccin flavor definitions`

---

### Task 2: Thread Theme Through Renderer

**Files:**
- Modify: `internal/render/render.go`
- Modify: `internal/render/render_test.go`

**Step 1: Write failing tests for theme-aware renderer**

Add to `internal/render/render_test.go` (append after existing tests):

```go
func TestRenderer_UsesThemeColors(t *testing.T) {
	// Create a renderer with Latte (light theme) -- Blue should be Latte's blue
	r := New(&ThemeLatte)
	result := r.Blue("hello")
	if result == "" {
		t.Fatal("Blue() returned empty string")
	}
	// Latte blue is #1e66f5 -- strip ANSI and verify the color code appears
	if !strings.Contains(result, "hello") {
		t.Errorf("Blue() result does not contain input text: %q", result)
	}
	// Verify Latte blue is different from Mocha blue (they must not be the same renderer)
	rMocha := New(&ThemeMocha)
	mochaBlueStyling := rMocha.Blue("hello")
	if result == mochaBlueStyling {
		t.Error("Latte renderer Blue() produced same output as Mocha renderer Blue() -- theme not applied")
	}
}

func TestRenderer_DefaultThemeIsMocha(t *testing.T) {
	// New(nil) should fall back to ThemeMocha
	r := New(nil)
	if r == nil {
		t.Fatal("New(nil) returned nil")
	}
}
```

**Step 2: Run to verify they fail**

```bash
go test ./internal/render/ -run TestRenderer_UsesThemeColors -v 2>&1
```

Expected: compilation failure — `New()` signature changed.

**Step 3: Update `internal/render/render.go`**

Change the `Renderer` struct to add a `theme` field, update `New` to accept a `*Theme`, and update all color helpers to use `r.theme`:

Replace the entire file content with:

```go
package render

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Renderer handles styling and layout of statusline components.
type Renderer struct {
	separator string
	lg        *lipgloss.Renderer
	style     Style
	theme     Theme
}

// New creates a Renderer with forced TrueColor output.
// Claude Code captures stdout so lipgloss won't auto-detect a TTY;
// we force color output with termenv.WithUnsafe().
// If theme is nil, ThemeMocha is used.
func New(theme *Theme) *Renderer {
	t := ThemeMocha
	if theme != nil {
		t = *theme
	}

	lg := lipgloss.NewRenderer(
		os.Stdout,
		termenv.WithUnsafe(),
		termenv.WithProfile(termenv.TrueColor),
	)

	return &Renderer{
		separator: " │ ",
		lg:        lg,
		style:     NewDefaultStyle(" │ "),
		theme:     t,
	}
}

// SetStyle replaces the active rendering style (e.g. DefaultStyle, PowerlineStyle).
func (r *Renderer) SetStyle(s Style) {
	r.style = s
}

// Theme returns the active theme.
func (r *Renderer) Theme() Theme {
	return r.theme
}

// RenderOutput renders a slice of LineData through the active Style, filtering
// out lines that produce only whitespace. termWidth is passed to the Style so
// powerline-mode can pad/align.
func (r *Renderer) RenderOutput(lines []LineData, termWidth int) string {
	var output []string
	for _, line := range lines {
		rendered := r.style.RenderLine(line, termWidth)
		if strings.TrimSpace(rendered) != "" {
			output = append(output, rendered)
		}
	}
	return strings.Join(output, "\n")
}

// RenderLines joins components per line with separators, filtering out empty components
// and lines that contain only empty components.
// This is a backward-compatible wrapper around RenderOutput.
func (r *Renderer) RenderLines(lines [][]string) string {
	var data []LineData
	for _, line := range lines {
		var nonEmpty []string
		for _, c := range line {
			if strings.TrimSpace(c) != "" {
				nonEmpty = append(nonEmpty, c)
			}
		}
		if len(nonEmpty) > 0 {
			data = append(data, LineData{Left: nonEmpty})
		}
	}
	return r.RenderOutput(data, 80)
}

// Style helpers -- each wraps the input string with the active theme's color.

// Dimmed renders text in Overlay0 (dimmed/label color).
func (r *Renderer) Dimmed(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Overlay0).Render(s)
}

// Text renders text in the default text color.
func (r *Renderer) Text(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Text).Render(s)
}

// Green renders text in green (clean/good status).
func (r *Renderer) Green(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Green).Render(s)
}

// Red renders text in red (critical status).
func (r *Renderer) Red(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Red).Render(s)
}

// Yellow renders text in yellow (warning status).
func (r *Renderer) Yellow(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Yellow).Render(s)
}

// Blue renders text in blue (paths/info).
func (r *Renderer) Blue(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Blue).Render(s)
}

// Mauve renders text in mauve (accent color).
func (r *Renderer) Mauve(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Mauve).Render(s)
}

// Peach renders text in peach (cost-related).
func (r *Renderer) Peach(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Peach).Render(s)
}

// Teal renders text in teal (secondary info).
func (r *Renderer) Teal(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Teal).Render(s)
}
```

**Step 4: Fix all callers of `render.New()` to compile**

The only direct caller is `main.go`. Update line 49 temporarily to `render.New(nil)` so it compiles while you fix the remaining tasks. The final wiring happens in Task 5.

```go
// main.go line 49 -- temporary, will be updated in Task 5
r := render.New(nil)
```

Also update the two test files that call `New()` directly:

In `internal/render/render_test.go`, change all `New()` calls to `New(nil)`.

In `internal/render/integration_test.go`, change all `render.New()` calls to `render.New(nil)`.

In `internal/render/powerline_style_test.go`, change `NewPowerlineStyle(New())` to `NewPowerlineStyle(New(nil))`.

**Step 5: Run all render tests**

```bash
go test ./internal/render/ -v 2>&1
```

Expected: all tests PASS including the new `TestRenderer_UsesThemeColors`.

**Step 6: Run full build check**

```bash
go build ./... 2>&1
```

Expected: no errors.

**Step 7: Commit**

```bash
git add internal/render/render.go internal/render/render_test.go \
        internal/render/integration_test.go internal/render/powerline_style_test.go \
        main.go
/git-commit
```

Suggested message: `feat(render): Thread Theme through Renderer, replace global color constants`

---

### Task 3: Theme-Aware Segment Categories

**Files:**
- Modify: `internal/render/segment.go`
- Modify: `internal/render/segment_test.go`
- Modify: `internal/render/powerline_style.go`
- Modify: `internal/render/powerline_style_test.go`

**Step 1: Write failing tests**

Replace `internal/render/segment_test.go` entirely (all existing tests will need updating since `SegmentCategoryFor` signature changes):

```go
package render

import (
	"testing"
)

func TestSegmentCategory_AllComponentsMapped(t *testing.T) {
	known := []string{
		"repo_info", "model_info", "bedrock_model",
		"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "burn_rate",
		"context_window", "cache_efficiency", "block_projection",
		"code_productivity", "commits",
		"version_info", "session_mode",
		"time_display", "submodules",
	}

	for _, name := range known {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat == (SegmentCategory{}) {
			t.Errorf("component %q has no segment category", name)
		}
	}
}

func TestSegmentCategory_UnknownComponentGetsDim(t *testing.T) {
	cat := SegmentCategoryFor("unknown_component", &ThemeMocha)
	if cat.Background != ThemeMocha.Overlay0 {
		t.Errorf("unknown component should use Overlay0 background, got %v", cat.Background)
	}
}

func TestSegmentCategory_InfoGroupIsBlue(t *testing.T) {
	for _, name := range []string{"repo_info", "model_info", "bedrock_model"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Blue {
			t.Errorf("component %q should have blue background, got %v", name, cat.Background)
		}
		// Foreground must be Base (for auto-invert)
		if cat.Foreground != ThemeMocha.Base {
			t.Errorf("component %q should have Base foreground, got %v", name, cat.Foreground)
		}
	}
}

func TestSegmentCategory_CostGroupIsPeach(t *testing.T) {
	for _, name := range []string{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "burn_rate"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Peach {
			t.Errorf("component %q should have peach background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_MetricsGroupIsTeal(t *testing.T) {
	for _, name := range []string{"context_window", "cache_efficiency", "block_projection"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Teal {
			t.Errorf("component %q should have teal background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_ActivityGroupIsGreen(t *testing.T) {
	for _, name := range []string{"code_productivity", "commits"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Green {
			t.Errorf("component %q should have green background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_MetaGroupIsMauve(t *testing.T) {
	for _, name := range []string{"version_info", "session_mode"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Mauve {
			t.Errorf("component %q should have mauve background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_LatteAutoInvert(t *testing.T) {
	// Latte is a light theme. Info segments use Blue background.
	// Foreground should be Base (#eff1f5, light) for proper contrast on Latte's darker blue.
	cat := SegmentCategoryFor("repo_info", &ThemeLatte)
	if cat.Background != ThemeLatte.Blue {
		t.Errorf("Latte repo_info should have Latte Blue background, got %v", cat.Background)
	}
	if cat.Foreground != ThemeLatte.Base {
		t.Errorf("Latte repo_info should have Latte Base foreground, got %v", cat.Foreground)
	}
}

func TestSegmentCategory_CrossThemeDiffers(t *testing.T) {
	// Same component, different themes should produce different backgrounds
	catMocha := SegmentCategoryFor("repo_info", &ThemeMocha)
	catLatte := SegmentCategoryFor("repo_info", &ThemeLatte)
	if catMocha.Background == catLatte.Background {
		t.Error("Mocha and Latte should have different blue backgrounds for repo_info")
	}
}
```

**Step 2: Run to verify they fail**

```bash
go test ./internal/render/ -run TestSegmentCategory -v 2>&1
```

Expected: compilation errors — `SegmentCategoryFor` still has old signature.

**Step 3: Update `internal/render/segment.go`**

Replace the entire file:

```go
package render

// SegmentCategory defines the background and foreground colors for a powerline segment.
type SegmentCategory struct {
	Background string // lipgloss.Color is a string alias
	Foreground string
}

// componentGroup maps a component name to its semantic group.
// Returns "dim" for unknown components.
func componentGroup(name string) string {
	switch name {
	case "repo_info", "model_info", "bedrock_model":
		return "info"
	case "cost_monthly", "cost_weekly", "cost_daily", "cost_live", "burn_rate":
		return "cost"
	case "context_window", "cache_efficiency", "block_projection":
		return "metrics"
	case "code_productivity", "commits":
		return "activity"
	case "version_info", "session_mode":
		return "meta"
	default:
		return "dim"
	}
}

// SegmentCategoryFor returns the SegmentCategory for a given component name,
// using the colors from the provided theme. Unknown components fall back to Dim.
func SegmentCategoryFor(componentName string, theme *Theme) SegmentCategory {
	switch componentGroup(componentName) {
	case "info":
		return SegmentCategory{Background: string(theme.Blue), Foreground: string(theme.Base)}
	case "cost":
		return SegmentCategory{Background: string(theme.Peach), Foreground: string(theme.Base)}
	case "metrics":
		return SegmentCategory{Background: string(theme.Teal), Foreground: string(theme.Base)}
	case "activity":
		return SegmentCategory{Background: string(theme.Green), Foreground: string(theme.Base)}
	case "meta":
		return SegmentCategory{Background: string(theme.Mauve), Foreground: string(theme.Base)}
	default: // dim
		return SegmentCategory{Background: string(theme.Overlay0), Foreground: string(theme.Text)}
	}
}
```

> Note: `SegmentCategory` now uses `string` instead of `lipgloss.Color` to avoid the import cycle. `lipgloss.Color` is itself a `type Color string` alias, so passing `string(theme.Blue)` is lossless. Convert back when calling lipgloss: `lipgloss.Color(seg.Background)`.

**Step 4: Update `internal/render/powerline_style.go`**

The `PowerlineStyle` needs a theme to pass to `SegmentCategoryFor`, and must convert `string` back to `lipgloss.Color` when calling lipgloss. Replace the entire file:

```go
package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	arrowRight = "\ue0b0" // filled right-pointing arrow
	arrowLeft  = "\ue0b2" // filled left-pointing arrow
)

// PowerlineStyle renders statusline components as colored background segments
// separated by Nerd Font powerline arrow characters. Adjacent components in the
// same semantic category are merged into a single segment.
type PowerlineStyle struct {
	lg    *lipgloss.Renderer
	theme Theme
}

// NewPowerlineStyle creates a PowerlineStyle renderer from an existing Renderer.
func NewPowerlineStyle(r *Renderer) *PowerlineStyle {
	return &PowerlineStyle{lg: r.lg, theme: r.theme}
}

// segment groups one or more component outputs that share the same SegmentCategory.
type segment struct {
	category SegmentCategory
	parts    []string
}

// buildSegments groups consecutive components by SegmentCategory, merging adjacent
// same-category components into a single segment. Empty components are skipped.
func buildSegments(names, contents []string, theme *Theme) []segment {
	var segments []segment
	for i, content := range contents {
		stripped := StripANSI(content)
		if strings.TrimSpace(stripped) == "" {
			continue
		}
		name := ""
		if i < len(names) {
			name = names[i]
		}
		cat := SegmentCategoryFor(name, theme)
		if len(segments) > 0 && segments[len(segments)-1].category == cat {
			segments[len(segments)-1].parts = append(segments[len(segments)-1].parts, stripped)
		} else {
			segments = append(segments, segment{category: cat, parts: []string{stripped}})
		}
	}
	return segments
}

// renderSegmentText renders the text content of a segment with background and
// foreground colors from its category, with horizontal padding.
func (s *PowerlineStyle) renderSegmentText(seg segment) string {
	text := strings.Join(seg.parts, " \u2502 ")
	return s.lg.NewStyle().
		Background(lipgloss.Color(seg.category.Background)).
		Foreground(lipgloss.Color(seg.category.Foreground)).
		Padding(0, 1).
		Render(text)
}

// renderLeftSegments renders left-aligned segments with forward-pointing arrows
// between different-category segments and a trailing arrow after the last segment.
func (s *PowerlineStyle) renderLeftSegments(segments []segment) string {
	if len(segments) == 0 {
		return ""
	}
	var parts []string
	for i, seg := range segments {
		parts = append(parts, s.renderSegmentText(seg))
		if i < len(segments)-1 {
			arrow := s.lg.NewStyle().
				Foreground(lipgloss.Color(seg.category.Background)).
				Background(lipgloss.Color(segments[i+1].category.Background)).
				Render(arrowRight)
			parts = append(parts, arrow)
		}
	}
	last := segments[len(segments)-1]
	trailingArrow := s.lg.NewStyle().
		Foreground(lipgloss.Color(last.category.Background)).
		Render(arrowRight)
	parts = append(parts, trailingArrow)
	return strings.Join(parts, "")
}

// renderRightSegments renders right-aligned segments with a leading reverse arrow
// before the first segment and reverse arrows between different-category segments.
func (s *PowerlineStyle) renderRightSegments(segments []segment) string {
	if len(segments) == 0 {
		return ""
	}
	var parts []string
	first := segments[0]
	leadingArrow := s.lg.NewStyle().
		Foreground(lipgloss.Color(first.category.Background)).
		Render(arrowLeft)
	parts = append(parts, leadingArrow)
	for i, seg := range segments {
		parts = append(parts, s.renderSegmentText(seg))
		if i < len(segments)-1 {
			arrow := s.lg.NewStyle().
				Foreground(lipgloss.Color(segments[i+1].category.Background)).
				Background(lipgloss.Color(seg.category.Background)).
				Render(arrowLeft)
			parts = append(parts, arrow)
		}
	}
	return strings.Join(parts, "")
}

// RenderLine renders a complete statusline from LineData, with left-aligned segments
// on the left, right-aligned segments on the right, and space-padding in between
// to fill the terminal width.
func (s *PowerlineStyle) RenderLine(line LineData, termWidth int) string {
	leftSegs := buildSegments(line.LeftNames, line.Left, &s.theme)
	rightSegs := buildSegments(line.RightNames, line.Right, &s.theme)
	if len(leftSegs) == 0 && len(rightSegs) == 0 {
		return ""
	}
	leftStr := s.renderLeftSegments(leftSegs)
	rightStr := s.renderRightSegments(rightSegs)
	leftWidth := VisualWidth(leftStr)
	rightWidth := VisualWidth(rightStr)
	padding := termWidth - leftWidth - rightWidth
	if padding < 0 {
		padding = 0
	}
	return leftStr + strings.Repeat(" ", padding) + rightStr
}
```

**Step 5: Fix `buildSegments` calls in powerline_style_test.go**

In `internal/render/powerline_style_test.go`, the two direct `buildSegments` calls need a theme argument. Update them:

```go
// TestBuildSegments_MergesSameCategory
segs := buildSegments(names, contents, &ThemeMocha)

// TestBuildSegments_SkipsEmpty
segs := buildSegments(names, contents, &ThemeMocha)
```

**Step 6: Run all render tests**

```bash
go test ./internal/render/ -v 2>&1
```

Expected: all tests PASS. If `SegmentCategory` comparison via `==` fails (struct now uses `string` fields instead of `lipgloss.Color`), the struct equality in `buildSegments` still works since string comparison is valid.

**Step 7: Full build check**

```bash
go build ./... 2>&1
```

Expected: no errors.

**Step 8: Commit**

```bash
git add internal/render/segment.go internal/render/segment_test.go \
        internal/render/powerline_style.go internal/render/powerline_style_test.go
/git-commit
```

Suggested message: `feat(render): Make segment categories theme-aware with auto-invert contrast`

---

### Task 4: Config Theme Field

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Step 1: Write failing tests**

Add to `internal/config/config_test.go`:

```go
func TestLoad_ThemeFieldPopulates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[layout]
theme = "catppuccin-frappe"
style = "powerline"

[[layout.lines]]
left  = ["repo_info"]
right = ["time_display"]
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Layout.Theme != "catppuccin-frappe" {
		t.Errorf("expected theme 'catppuccin-frappe', got %q", cfg.Layout.Theme)
	}
}

func TestLoad_MissingThemeDefaultsToMocha(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[layout]
style = "default"

[[layout.lines]]
left = ["repo_info"]
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Layout.Theme != "catppuccin-mocha" {
		t.Errorf("expected theme 'catppuccin-mocha' when field missing, got %q", cfg.Layout.Theme)
	}
}

func TestDefaultConfig_HasMochaTheme(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Layout.Theme != "catppuccin-mocha" {
		t.Errorf("DefaultConfig should have theme 'catppuccin-mocha', got %q", cfg.Layout.Theme)
	}
}

func TestDefaultPowerlineConfig_HasMochaTheme(t *testing.T) {
	cfg := DefaultPowerlineConfig()
	if cfg.Layout.Theme != "catppuccin-mocha" {
		t.Errorf("DefaultPowerlineConfig should have theme 'catppuccin-mocha', got %q", cfg.Layout.Theme)
	}
}
```

**Step 2: Run to verify they fail**

```bash
go test ./internal/config/ -run TestLoad_Theme -run TestDefaultConfig_HasMochaTheme -run TestDefaultPowerlineConfig_HasMochaTheme -v 2>&1
```

Expected: FAIL — `cfg.Layout.Theme` field does not exist yet.

**Step 3: Add `Theme` field to `Layout` struct in `internal/config/config.go`**

Find the `Layout` struct and add the `Theme` field as the first field:

```go
type Layout struct {
	Theme     string       `toml:"theme"`
	Style     string       `toml:"style"`
	IconStyle string       `toml:"icon_style"`
	Padding   int          `toml:"padding"`
	Lines     []LayoutLine `toml:"lines"`
}
```

**Step 4: Add default theme in `Load()`**

In the `Load()` function, add a default for `Theme` alongside the existing default for `Style`. There are three places where defaults are applied. Add `if cfg.Layout.Theme == "" { cfg.Layout.Theme = "catppuccin-mocha" }` in each:

1. After new format parsed successfully (around line 86):
```go
if cfg.Layout.Style == "" {
    cfg.Layout.Style = "default"
}
if cfg.Layout.Theme == "" {
    cfg.Layout.Theme = "catppuccin-mocha"
}
```

2. After legacy migration (around line 108):
```go
cfg.Layout.Style = "default"
cfg.Layout.Theme = "catppuccin-mocha"
```

3. At the bottom fallback (around line 128):
```go
if cfg.Layout.Style == "" {
    cfg.Layout.Style = "default"
}
if cfg.Layout.Theme == "" {
    cfg.Layout.Theme = "catppuccin-mocha"
}
```

**Step 5: Add `Theme` to `DefaultConfig()` and `DefaultPowerlineConfig()`**

In `DefaultConfig()`, add to the `Layout` literal:
```go
Layout: Layout{
    Theme:     "catppuccin-mocha",
    Style:     "default",
    // ... rest unchanged
```

In `DefaultPowerlineConfig()`, add similarly:
```go
Layout: Layout{
    Theme:     "catppuccin-mocha",
    Style:     "powerline",
    // ... rest unchanged
```

**Step 6: Run config tests**

```bash
go test ./internal/config/ -v 2>&1
```

Expected: all tests PASS.

**Step 7: Run full test suite**

```bash
go test ./... 2>&1
```

Expected: all tests PASS.

**Step 8: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
/git-commit
```

Suggested message: `feat(config): Add theme field to Layout, default catppuccin-mocha`

---

### Task 5: Wire Theme Through main.go

**Files:**
- Modify: `main.go`

**Step 1: No test needed** (main.go is wiring, not logic; integration test in Task 6 covers this)

**Step 2: Update `main.go`**

Replace the temporary `render.New(nil)` (line 49) and the style selection block with proper theme wiring. The changes are:

1. Load theme from config (after config is loaded, around line 55):
```go
// Load theme from config
theme, ok := render.ThemeByName(cfg.Layout.Theme)
if !ok {
    fmt.Fprintf(os.Stderr, "unknown theme %q, using catppuccin-mocha\n", cfg.Layout.Theme)
}
```

2. Pass theme to `render.New`:
```go
r := render.New(&theme)
```

The full changed section (lines 49–100 approximately) should look like this after editing:

```go
// Load configuration first (needed for theme)
configPath := filepath.Join(homeDir, ".claude", "statusline", "config.toml")
cfg, err := config.Load(configPath)
if err != nil {
    fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
    os.Exit(1)
}

// Load theme from config
theme, ok := render.ThemeByName(cfg.Layout.Theme)
if !ok {
    fmt.Fprintf(os.Stderr, "unknown theme %q, using catppuccin-mocha\n", cfg.Layout.Theme)
}

// Initialize infrastructure
c := cache.New(cacheDir)
_ = c.Prune(30 * 24 * time.Hour)
r := render.New(&theme)
h := cost.NewHistory(filepath.Join(costDir, "history.jsonl"))
scanner := cost.NewTranscriptScanner(projectsDir, c)
```

Note: Move the config loading *before* `render.New()` so the theme is available. The `ic` (icon set) creation stays after config load. Remove the duplicate config loading that was previously later in the function.

**Step 3: Build to verify no errors**

```bash
go build ./... 2>&1
```

Expected: no errors.

**Step 4: Run full test suite**

```bash
go test ./... 2>&1
```

Expected: all tests PASS.

**Step 5: Commit**

```bash
git add main.go
/git-commit
```

Suggested message: `feat: Wire theme from config through renderer in main`

---

### Task 6: Integration Tests for Theme Switching

**Files:**
- Modify: `internal/render/integration_test.go`

**Step 1: Add theme integration tests**

Append to `internal/render/integration_test.go`:

```go
func TestPowerlinePipeline_LatteTheme(t *testing.T) {
	// Latte is a light theme -- segment colors must differ from Mocha
	rLatte := render.New(&render.ThemeLatte)
	rLatte.SetStyle(render.NewPowerlineStyle(rLatte))

	rMocha := render.New(&render.ThemeMocha)
	rMocha.SetStyle(render.NewPowerlineStyle(rMocha))

	registry := component.NewRegistry()
	registry.Register(&mockComp{name: "repo_info", output: "~/projects"})
	registry.Register(&mockComp{name: "cost_daily", output: "$0.89"})

	in := &input.StatusLineInput{}
	leftNames, leftContent := registry.RenderNamedLine(in, []string{"repo_info", "cost_daily"})
	lines := []render.LineData{
		{Left: leftContent, LeftNames: leftNames},
	}

	latteOut := rLatte.RenderOutput(lines, 80)
	mochaOut := rMocha.RenderOutput(lines, 80)

	if latteOut == "" {
		t.Fatal("Latte theme produced empty output")
	}
	// The two themes should produce different ANSI color codes
	if latteOut == mochaOut {
		t.Error("Latte and Mocha produced identical output -- theme colors not applied")
	}
	// Both should still contain the plain text content
	latteStripped := render.StripANSI(latteOut)
	if !strings.Contains(latteStripped, "~/projects") {
		t.Error("Latte output missing repo_info content")
	}
}

func TestDefaultPipeline_LatteTheme(t *testing.T) {
	// Default style with Latte theme -- component foreground colors must differ
	rLatte := render.New(&render.ThemeLatte)
	rMocha := render.New(&render.ThemeMocha)

	latteBlue := rLatte.Blue("hello")
	mochaBlue := rMocha.Blue("hello")

	if latteBlue == mochaBlue {
		t.Error("Latte and Mocha Blue() produced identical styled output -- theme not applied")
	}
}

func TestPipeline_BackwardCompat_NoThemeProducesMocha(t *testing.T) {
	// nil theme fallback should produce same output as explicit Mocha
	rNil := render.New(nil)
	rMocha := render.New(&render.ThemeMocha)

	nilBlue := rNil.Blue("hello")
	mochaBlue := rMocha.Blue("hello")

	if nilBlue != mochaBlue {
		t.Errorf("nil theme Blue() = %q, Mocha Blue() = %q -- should be identical", nilBlue, mochaBlue)
	}
}

func TestAllFourThemes_ProduceOutput(t *testing.T) {
	themes := []render.Theme{
		render.ThemeMocha,
		render.ThemeLatte,
		render.ThemeFrappe,
		render.ThemeMacchiato,
	}

	registry := component.NewRegistry()
	registry.Register(&mockComp{name: "repo_info", output: "~/projects"})
	registry.Register(&mockComp{name: "time_display", output: "14:32"})
	in := &input.StatusLineInput{}

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			th := theme // capture
			r := render.New(&th)
			r.SetStyle(render.NewPowerlineStyle(r))

			leftNames, leftContent := registry.RenderNamedLine(in, []string{"repo_info"})
			rightNames, rightContent := registry.RenderNamedLine(in, []string{"time_display"})
			lines := []render.LineData{
				{Left: leftContent, LeftNames: leftNames, Right: rightContent, RightNames: rightNames},
			}

			output := r.RenderOutput(lines, 120)
			if output == "" {
				t.Errorf("theme %q produced empty output", theme.Name)
			}
			stripped := render.StripANSI(output)
			if !strings.Contains(stripped, "~/projects") {
				t.Errorf("theme %q output missing repo_info", theme.Name)
			}
		})
	}
}
```

**Step 2: Run integration tests**

```bash
go test ./internal/render/ -run TestPipeline -run TestAllFour -run TestDefaultPipeline_Latte -v 2>&1
```

Expected: all new integration tests PASS.

**Step 3: Run full test suite one final time**

```bash
go test ./... 2>&1
```

Expected: all tests PASS, zero failures.

**Step 4: Commit**

```bash
git add internal/render/integration_test.go
/git-commit
```

Suggested message: `test(render): Add integration tests for all four Catppuccin theme flavors`

---

### Task 7: Update README

**Files:**
- Modify: `README.md`

**Step 1: Find the config documentation section**

```bash
grep -n "icon_style\|style.*=\|config" README.md | head -30 2>&1
```

**Step 2: Add theme documentation**

Find the section that documents `[layout]` config options (near `icon_style` and `style` entries). Add a `theme` entry documenting valid values and the default.

The addition should look something like:

```markdown
| `theme` | `catppuccin-mocha` | Color palette. Valid values: `catppuccin-mocha`, `catppuccin-latte`, `catppuccin-frappe`, `catppuccin-macchiato` |
```

Or if the docs are prose-style, add a paragraph explaining the theme option near the `style` and `icon_style` docs.

**Step 3: Verify the change looks right**

```bash
grep -A5 -B5 "theme" README.md 2>&1
```

**Step 4: Commit**

```bash
git add README.md
/git-commit
```

Suggested message: `docs: Document theme config option for Catppuccin flavor selection`

---

## Final Verification

Run the complete test suite and build one last time:

```bash
go test ./... -count=1 2>&1 && go build ./... 2>&1 && echo "ALL GOOD"
```

Expected: `ALL GOOD`

---

## Summary of Changes

| File | Change |
|------|--------|
| `internal/render/theme.go` | **NEW** — `Theme` struct, four flavor instances, `ThemeByName()` |
| `internal/render/theme_test.go` | **NEW** — theme lookup and field coverage tests |
| `internal/render/render.go` | Updated — `New(*Theme)`, color helpers read `r.theme`, removed global `Color*` vars |
| `internal/render/segment.go` | Updated — `SegmentCategoryFor(name, theme)`, removed `colorDark` global, `SegmentCategory` uses `string` fields |
| `internal/render/powerline_style.go` | Updated — `PowerlineStyle` holds `theme Theme`, `buildSegments` takes `*Theme` |
| `internal/config/config.go` | Updated — `Layout.Theme` field, defaults to `catppuccin-mocha` in all load paths |
| `main.go` | Updated — load theme from config, pass `&theme` to `render.New()` |
| `internal/render/render_test.go` | Updated — `New(nil)` calls, added theme-aware tests |
| `internal/render/segment_test.go` | Updated — `SegmentCategoryFor` calls pass `&ThemeMocha`, added Latte/cross-theme tests |
| `internal/render/powerline_style_test.go` | Updated — `buildSegments` calls pass `&ThemeMocha` |
| `internal/render/integration_test.go` | Updated — `render.New(nil)` calls, added four-flavor integration tests |
| `internal/config/config_test.go` | Updated — theme field tests |
| `README.md` | Updated — `theme` config option documented |
