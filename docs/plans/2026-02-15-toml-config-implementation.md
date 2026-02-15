# TOML Configuration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add TOML configuration file support for controlling statusline layout and per-component display options.

**Architecture:** New `internal/config/` package with `Load()`, `DefaultConfig()`, and `GetBool()` helpers. `main.go` loads config early and uses it to build component lines. Components that need per-component settings receive `*config.Config` and query it at render time.

**Tech Stack:** Go 1.23.0, `github.com/BurntSushi/toml`, existing `internal/cache` and `internal/components` packages.

---

## Task 1: Add BurntSushi/toml dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add the dependency**

```bash
go get github.com/BurntSushi/toml@latest
```

Expected: Dependency added to `go.mod` and `go.sum`

**Step 2: Verify it builds**

```bash
go build .
```

Expected: Clean build

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "build: Add BurntSushi/toml dependency"
```

---

## Task 2: Create config package with types and tests

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the failing test**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_CreatesDefaultWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}

	// Verify default layout has 3 lines
	if len(cfg.Layout.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(cfg.Layout.Lines))
	}
}

func TestLoad_ParsesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[layout]
lines = [
  ["repo_info"],
  ["model_info", "commits"],
]

[components.bedrock_model]
show_region = false
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Layout.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(cfg.Layout.Lines))
	}
	if len(cfg.Layout.Lines[1]) != 2 {
		t.Errorf("Expected 2 components on line 2, got %d", len(cfg.Layout.Lines[1]))
	}

	// Verify per-component setting
	showRegion := cfg.GetBool("bedrock_model", "show_region", true)
	if showRegion {
		t.Error("Expected show_region=false from config")
	}
}

func TestGetBool_ReturnsFallbackWhenNotSet(t *testing.T) {
	cfg := &Config{
		Components: make(map[string]ComponentConfig),
	}

	val := cfg.GetBool("context_window", "show_tokens", true)
	if !val {
		t.Error("Expected fallback true when key not set")
	}
}

func TestGetBool_ReturnsConfigValue(t *testing.T) {
	falseVal := false
	cfg := &Config{
		Components: map[string]ComponentConfig{
			"context_window": {ShowTokens: &falseVal},
		},
	}

	val := cfg.GetBool("context_window", "show_tokens", true)
	if val {
		t.Error("Expected false from config, got true")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.Layout.Lines) != 3 {
		t.Errorf("Expected 3 lines in default, got %d", len(cfg.Layout.Lines))
	}

	// Verify all current components are present
	allComponents := make(map[string]bool)
	for _, line := range cfg.Layout.Lines {
		for _, comp := range line {
			allComponents[comp] = true
		}
	}

	required := []string{"repo_info", "model_info", "cost_live", "context_window"}
	for _, comp := range required {
		if !allComponents[comp] {
			t.Errorf("Expected %s in default config", comp)
		}
	}

	// Verify default per-component settings
	if cfg.GetBool("bedrock_model", "show_region", false) != true {
		t.Error("Expected show_region default true")
	}
	if cfg.GetBool("context_window", "show_tokens", false) != true {
		t.Error("Expected show_tokens default true")
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/config/ -v
```

Expected: FAIL — `Load`, `DefaultConfig`, `GetBool` not defined

**Step 3: Write minimal implementation**

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the parsed TOML configuration.
type Config struct {
	Layout     LayoutConfig                `toml:"layout"`
	Components map[string]ComponentConfig  `toml:"components"`
}

// LayoutConfig defines which components appear on which lines.
type LayoutConfig struct {
	Lines [][]string `toml:"lines"`
}

// ComponentConfig holds per-component display options.
type ComponentConfig struct {
	ShowRegion *bool `toml:"show_region,omitempty"`
	ShowTokens *bool `toml:"show_tokens,omitempty"`
}

// Load reads the config file, creating it with defaults if missing.
func Load(path string) (*Config, error) {
	// If file doesn't exist, create default
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := writeConfig(path, cfg); err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}
		return cfg, nil
	}

	// Parse existing file
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Ensure Components map is initialized
	if cfg.Components == nil {
		cfg.Components = make(map[string]ComponentConfig)
	}

	return &cfg, nil
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	trueVal := true
	return &Config{
		Layout: LayoutConfig{
			Lines: [][]string{
				{"repo_info"},
				{"bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"},
				{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"},
			},
		},
		Components: map[string]ComponentConfig{
			"bedrock_model": {ShowRegion: &trueVal},
			"context_window": {ShowTokens: &trueVal},
		},
	}
}

// GetBool retrieves a boolean setting for a component with a fallback.
func (c *Config) GetBool(component, key string, fallback bool) bool {
	compCfg, ok := c.Components[component]
	if !ok {
		return fallback
	}

	switch key {
	case "show_region":
		if compCfg.ShowRegion != nil {
			return *compCfg.ShowRegion
		}
	case "show_tokens":
		if compCfg.ShowTokens != nil {
			return *compCfg.ShowTokens
		}
	}

	return fallback
}

// writeConfig writes the config to disk as TOML.
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
	f.WriteString("# Claude Code Statusline Configuration\n\n")

	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/config/ -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat(config): Add TOML config loading with defaults"
```

---

## Task 3: Wire config into main.go for layout

**Files:**
- Modify: `main.go`

**Step 1: Add config loading early in main()**

After the infrastructure setup (around line 32), add:

```go
	configPath := filepath.Join(homeDir, ".claude", "statusline", "config.toml")
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
```

Add import: `"github.com/h2ik/claude-statusline/internal/config"`

**Step 2: Replace hardcoded lines with config lines**

Replace the `lines := [][]string{...}` block (around line 57) with:

```go
	// Use config-defined layout
	lines := cfg.Layout.Lines
```

**Step 3: Build and smoke test**

```bash
go build -o claude-statusline .
echo '{"workspace":{"current_dir":"/tmp"},"model":{"display_name":"Claude Opus 4.6"},"session_id":"test","cost":{"total_cost_usd":0.45}}' | ./claude-statusline
```

Expected: Output shows 3 lines with components from default config

**Step 4: Verify config file was created**

```bash
cat ~/.claude/statusline/config.toml
```

Expected: Default TOML config with all components listed

**Step 5: Commit**

```bash
git add main.go
git commit -m "feat: Wire TOML config layout into main"
```

---

## Task 4: Add show_region support to bedrock_model component

**Files:**
- Modify: `internal/components/bedrock_model.go`
- Modify: `internal/components/bedrock_model_test.go`

**Step 1: Write the failing test**

Add to `internal/components/bedrock_model_test.go`:

```go
func TestBedrockModel_Render_HidesRegionWhenConfigured(t *testing.T) {
	r := render.New()
	c := cache.New(t.TempDir())

	falseVal := false
	cfg := &config.Config{
		Components: map[string]config.ComponentConfig{
			"bedrock_model": {ShowRegion: &falseVal},
		},
	}

	comp := NewBedrockModel(r, c, cfg)

	in := &input.StatusLineInput{
		Model: input.ModelInfo{
			DisplayName: "Claude Opus 4.6",
			ID:          "arn:aws:bedrock:us-west-2:123456789012:application-inference-profile/abc123",
		},
	}

	output := comp.Render(in)

	if strings.Contains(output, "us-west-2") {
		t.Errorf("Expected region hidden, but got: %s", output)
	}
	if !strings.Contains(output, "Claude Opus 4.6") {
		t.Errorf("Expected model name, got: %s", output)
	}
}

func TestBedrockModel_Render_ShowsRegionByDefault(t *testing.T) {
	r := render.New()
	c := cache.New(t.TempDir())

	cfg := &config.Config{
		Components: make(map[string]config.ComponentConfig),
	}

	comp := NewBedrockModel(r, c, cfg)

	in := &input.StatusLineInput{
		Model: input.ModelInfo{
			DisplayName: "Claude Opus 4.6",
			ID:          "arn:aws:bedrock:us-west-2:123456789012:application-inference-profile/abc123",
		},
	}

	output := comp.Render(in)

	if !strings.Contains(output, "us-west-2") {
		t.Errorf("Expected region shown by default, got: %s", output)
	}
}
```

Add import: `"github.com/h2ik/claude-statusline/internal/config"`

**Step 2: Run test to verify it fails**

```bash
go test ./internal/components/ -run TestBedrockModel_Render_HidesRegion -v
```

Expected: FAIL — `NewBedrockModel` signature mismatch (needs cfg param)

**Step 3: Update BedrockModel to accept and use config**

In `internal/components/bedrock_model.go`:

Update struct:
```go
type BedrockModel struct {
	renderer *render.Renderer
	cache    *cache.Cache
	config   *config.Config
}
```

Update constructor:
```go
func NewBedrockModel(r *render.Renderer, c *cache.Cache, cfg *config.Config) *BedrockModel {
	return &BedrockModel{renderer: r, cache: c, config: cfg}
}
```

Update Render method (around line 31, after resolving model):
```go
	modelStr := fmt.Sprintf("🧠 %s", resolved)

	// Append region if configured
	if region != "" && c.config.GetBool("bedrock_model", "show_region", true) {
		regionStr := c.renderer.Dimmed(fmt.Sprintf(" (%s)", region))
		return modelStr + regionStr
	}

	return modelStr
```

Add import: `"github.com/h2ik/claude-statusline/internal/config"`

**Step 4: Update all existing tests**

In `internal/components/bedrock_model_test.go`, update all calls to `NewBedrockModel` to pass `&config.Config{Components: make(map[string]config.ComponentConfig)}` as the third argument.

**Step 5: Run tests to verify they pass**

```bash
go test ./internal/components/ -run TestBedrockModel -v
```

Expected: PASS

**Step 6: Update main.go to pass config**

In `main.go`, change:
```go
	registry.Register(components.NewBedrockModel(r, c))
```

to:
```go
	registry.Register(components.NewBedrockModel(r, c, cfg))
```

**Step 7: Build and verify**

```bash
go build -o claude-statusline .
```

Expected: Clean build

**Step 8: Commit**

```bash
git add internal/components/bedrock_model.go internal/components/bedrock_model_test.go main.go
git commit -m "feat(components): Add show_region config to bedrock_model"
```

---

## Task 5: Add show_tokens support to context_window component

**Files:**
- Modify: `internal/components/context_window.go`
- Modify: `internal/components/context_session_test.go`

**Step 1: Write the failing test**

Add to `internal/components/context_session_test.go`:

```go
func TestContextWindow_Render_HidesTokensWhenConfigured(t *testing.T) {
	r := render.New()

	falseVal := false
	cfg := &config.Config{
		Components: map[string]config.ComponentConfig{
			"context_window": {ShowTokens: &falseVal},
		},
	}

	c := NewContextWindow(r, cfg)

	in := &input.StatusLineInput{
		ContextWindow: input.ContextWindowInfo{
			UsedPercent:   45,
			UsedTokens:    90000,
			MaxTokens:     200000,
		},
	}

	output := c.Render(in)

	if strings.Contains(output, "90K") || strings.Contains(output, "200K") {
		t.Errorf("Expected tokens hidden, got: %s", output)
	}
	if !strings.Contains(output, "45%") {
		t.Errorf("Expected percentage shown, got: %s", output)
	}
}

func TestContextWindow_Render_ShowsTokensByDefault(t *testing.T) {
	r := render.New()

	cfg := &config.Config{
		Components: make(map[string]config.ComponentConfig),
	}

	c := NewContextWindow(r, cfg)

	in := &input.StatusLineInput{
		ContextWindow: input.ContextWindowInfo{
			UsedPercent:   45,
			UsedTokens:    90000,
			MaxTokens:     200000,
		},
	}

	output := c.Render(in)

	if !strings.Contains(output, "90K") {
		t.Errorf("Expected tokens shown by default, got: %s", output)
	}
}
```

Add import: `"github.com/h2ik/claude-statusline/internal/config"`

**Step 2: Run test to verify it fails**

```bash
go test ./internal/components/ -run TestContextWindow_Render_HidesTokens -v
```

Expected: FAIL — `NewContextWindow` signature mismatch

**Step 3: Update ContextWindow to accept and use config**

In `internal/components/context_window.go`:

Update struct:
```go
type ContextWindow struct {
	renderer *render.Renderer
	config   *config.Config
}
```

Update constructor:
```go
func NewContextWindow(r *render.Renderer, cfg *config.Config) *ContextWindow {
	return &ContextWindow{renderer: r, config: cfg}
}
```

Update Render method (around line 43, the percentage display section):

```go
	// Format percentage
	percentStr := fmt.Sprintf("%d%%", in.ContextWindow.UsedPercent)

	// Add token counts if configured
	if c.config.GetBool("context_window", "show_tokens", true) {
		usedK := in.ContextWindow.UsedTokens / 1000
		maxK := in.ContextWindow.MaxTokens / 1000
		percentStr = fmt.Sprintf("%s (%dK/%dK)", percentStr, usedK, maxK)
	}

	return fmt.Sprintf("%s %s", emoji, colorFn(percentStr))
```

Add import: `"github.com/h2ik/claude-statusline/internal/config"`

**Step 4: Update all existing tests**

In `internal/components/context_session_test.go`, update all calls to `NewContextWindow` to pass `&config.Config{Components: make(map[string]config.ComponentConfig)}` as the second argument.

**Step 5: Run tests to verify they pass**

```bash
go test ./internal/components/ -run TestContextWindow -v
```

Expected: PASS

**Step 6: Update main.go to pass config**

In `main.go`, change:
```go
	registry.Register(components.NewContextWindow(r))
```

to:
```go
	registry.Register(components.NewContextWindow(r, cfg))
```

**Step 7: Build and verify**

```bash
go build -o claude-statusline .
```

Expected: Clean build

**Step 8: Commit**

```bash
git add internal/components/context_window.go internal/components/context_session_test.go main.go
git commit -m "feat(components): Add show_tokens config to context_window"
```

---

## Task 6: Final verification and documentation update

**Files:**
- Modify: `docs/ARCHITECTURE.md`

**Step 1: Run all tests**

```bash
go test ./... -v -count=1
```

Expected: ALL PASS

**Step 2: Build and smoke test with custom config**

```bash
go build -o claude-statusline .

# Test with default config
echo '{"workspace":{"current_dir":"/tmp"},"model":{"display_name":"Claude Opus 4.6","id":"arn:aws:bedrock:us-west-2:123456789012:application-inference-profile/abc123"},"session_id":"test","cost":{"total_cost_usd":0.45},"context_window":{"used_percent":45,"used_tokens":90000,"max_tokens":200000}}' | ./claude-statusline
```

Expected: Shows region and token counts

**Step 3: Test with show_region=false**

Edit `~/.claude/statusline/config.toml`:
```toml
[components.bedrock_model]
show_region = false
```

Run statusline again:

Expected: Region hidden

**Step 4: Test with show_tokens=false**

Edit `~/.claude/statusline/config.toml`:
```toml
[components.context_window]
show_tokens = false
```

Run statusline again:

Expected: Token counts hidden, percentage still shown

**Step 5: Test with custom layout**

Edit `~/.claude/statusline/config.toml`:
```toml
[layout]
lines = [
  ["repo_info"],
  ["model_info", "commits"],
]
```

Run statusline again:

Expected: Only 2 lines, only specified components

**Step 6: Update ARCHITECTURE.md**

Add section after "Caching" section:

```markdown
## Configuration

TOML config at `~/.claude/statusline/config.toml`:
- Layout control: which components appear on which lines
- Per-component display toggles: `show_region`, `show_tokens`
- Auto-generated with defaults on first run
- `github.com/BurntSushi/toml` for parsing

Component registry builds lines from `cfg.Layout.Lines`. Components query their settings via `cfg.GetBool()`.
```

**Step 7: Commit**

```bash
git add docs/ARCHITECTURE.md
git commit -m "docs: Document TOML configuration system"
```

---

## Summary of Changes

| File | Action | Purpose |
|------|--------|---------|
| `go.mod`, `go.sum` | MODIFY | Add BurntSushi/toml dependency |
| `internal/config/config.go` | CREATE | Config types, Load, DefaultConfig, GetBool |
| `internal/config/config_test.go` | CREATE | Tests for config loading and defaults |
| `main.go` | MODIFY | Load config, use cfg.Layout.Lines, pass cfg to components |
| `internal/components/bedrock_model.go` | MODIFY | Accept *config.Config, use show_region setting |
| `internal/components/bedrock_model_test.go` | MODIFY | Update constructor calls, add show_region tests |
| `internal/components/context_window.go` | MODIFY | Accept *config.Config, use show_tokens setting |
| `internal/components/context_session_test.go` | MODIFY | Update constructor calls, add show_tokens tests |
| `docs/ARCHITECTURE.md` | MODIFY | Document configuration system |

## Verification

1. `go test ./... -v` — all tests pass
2. `go build -o claude-statusline .` — builds clean
3. Smoke test with JSON input — 3 lines rendered
4. Config file created at `~/.claude/statusline/config.toml`
5. Edit config, verify changes apply (region hidden, tokens hidden, custom layout)
6. All components respect their config settings
