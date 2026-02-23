package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_CreatesDefaultWhenMissing(t *testing.T) {
	// Use a temp directory so we don't pollute the real filesystem
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "subdir", "config.toml")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// File should have been created
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Fatal("expected config file to be created, but it does not exist")
	}

	// Should have 4 lines in layout
	if len(cfg.Layout.Lines) != 4 {
		t.Fatalf("expected 4 layout lines, got %d", len(cfg.Layout.Lines))
	}

	// Style should default to "default"
	if cfg.Layout.Style != "default" {
		t.Errorf("expected style 'default', got %q", cfg.Layout.Style)
	}

	// Verify line 1
	if len(cfg.Layout.Lines[0].Left) != 1 || cfg.Layout.Lines[0].Left[0] != "repo_info" {
		t.Errorf("line 1: expected Left=[repo_info], got %v", cfg.Layout.Lines[0].Left)
	}

	// Verify line 2
	expectedLine2 := []string{"bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"}
	if len(cfg.Layout.Lines[1].Left) != len(expectedLine2) {
		t.Fatalf("line 2: expected %d components, got %d", len(expectedLine2), len(cfg.Layout.Lines[1].Left))
	}
	for i, comp := range expectedLine2 {
		if cfg.Layout.Lines[1].Left[i] != comp {
			t.Errorf("line 2[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[1].Left[i])
		}
	}

	// Verify line 3
	expectedLine3 := []string{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"}
	if len(cfg.Layout.Lines[2].Left) != len(expectedLine3) {
		t.Fatalf("line 3: expected %d components, got %d", len(expectedLine3), len(cfg.Layout.Lines[2].Left))
	}
	for i, comp := range expectedLine3 {
		if cfg.Layout.Lines[2].Left[i] != comp {
			t.Errorf("line 3[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[2].Left[i])
		}
	}

	// Verify line 4
	expectedLine4 := []string{"burn_rate", "cache_efficiency", "block_projection", "code_productivity"}
	if len(cfg.Layout.Lines[3].Left) != len(expectedLine4) {
		t.Fatalf("line 4: expected %d components, got %d", len(expectedLine4), len(cfg.Layout.Lines[3].Left))
	}
	for i, comp := range expectedLine4 {
		if cfg.Layout.Lines[3].Left[i] != comp {
			t.Errorf("line 4[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[3].Left[i])
		}
	}
}

func TestLoad_ParsesExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Write a custom TOML using the old flat format -- tests backward compat
	content := `
[layout]
lines = [
  ["repo_info"],
  ["bedrock_model", "time_display"]
]

[components.bedrock_model]
show_region = false
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Should have 2 lines (auto-migrated from old format)
	if len(cfg.Layout.Lines) != 2 {
		t.Fatalf("expected 2 layout lines, got %d", len(cfg.Layout.Lines))
	}

	// Style should default to "default" after migration
	if cfg.Layout.Style != "default" {
		t.Errorf("expected style 'default', got %q", cfg.Layout.Style)
	}

	// Verify line 2 (migrated: all components in Left, Right empty)
	if len(cfg.Layout.Lines[1].Left) != 2 {
		t.Fatalf("line 2: expected 2 components, got %d", len(cfg.Layout.Lines[1].Left))
	}
	if cfg.Layout.Lines[1].Left[0] != "bedrock_model" || cfg.Layout.Lines[1].Left[1] != "time_display" {
		t.Errorf("line 2: expected Left=[bedrock_model, time_display], got %v", cfg.Layout.Lines[1].Left)
	}
	if len(cfg.Layout.Lines[1].Right) != 0 {
		t.Errorf("line 2: expected empty Right after migration, got %v", cfg.Layout.Lines[1].Right)
	}

	// Verify show_region is set to false
	comp, ok := cfg.Components["bedrock_model"]
	if !ok {
		t.Fatal("expected bedrock_model in Components map")
	}
	if comp.ShowRegion == nil {
		t.Fatal("expected ShowRegion to be set, got nil")
	}
	if *comp.ShowRegion != false {
		t.Errorf("expected ShowRegion=false, got %v", *comp.ShowRegion)
	}
}

func TestGetBool_ReturnsFallbackWhenNotSet(t *testing.T) {
	cfg := &Config{
		Components: map[string]ComponentConfig{},
	}

	// Component not in map at all - should return fallback
	result := cfg.GetBool("bedrock_model", "show_region", true)
	if result != true {
		t.Errorf("expected fallback true, got %v", result)
	}

	result = cfg.GetBool("bedrock_model", "show_region", false)
	if result != false {
		t.Errorf("expected fallback false, got %v", result)
	}

	// Component in map but field is nil
	cfg.Components["context_window"] = ComponentConfig{}
	result = cfg.GetBool("context_window", "show_tokens", true)
	if result != true {
		t.Errorf("expected fallback true for nil pointer, got %v", result)
	}
}

func TestGetBool_ReturnsConfigValue(t *testing.T) {
	falseVal := false
	trueVal := true

	cfg := &Config{
		Components: map[string]ComponentConfig{
			"context_window": {
				ShowTokens: &falseVal,
			},
			"bedrock_model": {
				ShowRegion: &trueVal,
			},
		},
	}

	// ShowTokens is explicitly false - should return false despite fallback being true
	result := cfg.GetBool("context_window", "show_tokens", true)
	if result != false {
		t.Errorf("expected configured false, got %v", result)
	}

	// ShowRegion is explicitly true - should return true despite fallback being false
	result = cfg.GetBool("bedrock_model", "show_region", false)
	if result != true {
		t.Errorf("expected configured true, got %v", result)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Should have 4 lines
	if len(cfg.Layout.Lines) != 4 {
		t.Fatalf("expected 4 layout lines, got %d", len(cfg.Layout.Lines))
	}

	// Style should be "default"
	if cfg.Layout.Style != "default" {
		t.Errorf("expected style 'default', got %q", cfg.Layout.Style)
	}

	// Verify line 1
	if len(cfg.Layout.Lines[0].Left) != 1 || cfg.Layout.Lines[0].Left[0] != "repo_info" {
		t.Errorf("line 1: expected Left=[repo_info], got %v", cfg.Layout.Lines[0].Left)
	}

	// Verify line 2
	expectedLine2 := []string{"bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"}
	if len(cfg.Layout.Lines[1].Left) != len(expectedLine2) {
		t.Fatalf("line 2: expected %d components, got %d", len(expectedLine2), len(cfg.Layout.Lines[1].Left))
	}
	for i, comp := range expectedLine2 {
		if cfg.Layout.Lines[1].Left[i] != comp {
			t.Errorf("line 2[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[1].Left[i])
		}
	}

	// Verify line 3
	expectedLine3 := []string{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"}
	if len(cfg.Layout.Lines[2].Left) != len(expectedLine3) {
		t.Fatalf("line 3: expected %d components, got %d", len(expectedLine3), len(cfg.Layout.Lines[2].Left))
	}
	for i, comp := range expectedLine3 {
		if cfg.Layout.Lines[2].Left[i] != comp {
			t.Errorf("line 3[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[2].Left[i])
		}
	}

	// Verify line 4
	expectedLine4 := []string{"burn_rate", "cache_efficiency", "block_projection", "code_productivity"}
	if len(cfg.Layout.Lines[3].Left) != len(expectedLine4) {
		t.Fatalf("line 4: expected %d components, got %d", len(expectedLine4), len(cfg.Layout.Lines[3].Left))
	}
	for i, comp := range expectedLine4 {
		if cfg.Layout.Lines[3].Left[i] != comp {
			t.Errorf("line 4[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[3].Left[i])
		}
	}

	// Verify default components exist
	bmComp, ok := cfg.Components["bedrock_model"]
	if !ok {
		t.Fatal("expected bedrock_model in default Components")
	}
	if bmComp.ShowRegion == nil {
		t.Fatal("expected bedrock_model.ShowRegion to be set")
	}
	if *bmComp.ShowRegion != true {
		t.Errorf("expected bedrock_model.ShowRegion=true, got %v", *bmComp.ShowRegion)
	}

	cwComp, ok := cfg.Components["context_window"]
	if !ok {
		t.Fatal("expected context_window in default Components")
	}
	if cwComp.ShowTokens == nil {
		t.Fatal("expected context_window.ShowTokens to be set")
	}
	if *cwComp.ShowTokens != true {
		t.Errorf("expected context_window.ShowTokens=true, got %v", *cwComp.ShowTokens)
	}

	// Verify code_productivity defaults
	cpComp, ok := cfg.Components["code_productivity"]
	if !ok {
		t.Fatal("expected code_productivity in default Components")
	}
	if cpComp.ShowVelocity == nil {
		t.Fatal("expected code_productivity.ShowVelocity to be set")
	}
	if *cpComp.ShowVelocity != true {
		t.Errorf("expected code_productivity.ShowVelocity=true, got %v", *cpComp.ShowVelocity)
	}
	if cpComp.ShowCostPerLine == nil {
		t.Fatal("expected code_productivity.ShowCostPerLine to be set")
	}
	if *cpComp.ShowCostPerLine != true {
		t.Errorf("expected code_productivity.ShowCostPerLine=true, got %v", *cpComp.ShowCostPerLine)
	}
}

func TestLoad_NewFormat_LeftRight(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[layout]
style = "powerline"

[[layout.lines]]
left  = ["repo_info"]
right = ["time_display"]

[[layout.lines]]
left  = ["model_info", "commits"]
right = ["context_window"]
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Layout.Style != "powerline" {
		t.Errorf("expected style 'powerline', got %q", cfg.Layout.Style)
	}
	if len(cfg.Layout.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(cfg.Layout.Lines))
	}
	if cfg.Layout.Lines[0].Left[0] != "repo_info" {
		t.Errorf("unexpected Left[0]: %v", cfg.Layout.Lines[0].Left)
	}
	if cfg.Layout.Lines[0].Right[0] != "time_display" {
		t.Errorf("unexpected Right[0]: %v", cfg.Layout.Lines[0].Right)
	}
}

func TestLoad_OldFormat_AutoMigrates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[layout]
lines = [
  ["repo_info"],
  ["model_info", "commits"],
]
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Layout.Style != "default" {
		t.Errorf("expected style 'default', got %q", cfg.Layout.Style)
	}
	if len(cfg.Layout.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(cfg.Layout.Lines))
	}
	if cfg.Layout.Lines[0].Left[0] != "repo_info" {
		t.Errorf("unexpected Left: %v", cfg.Layout.Lines[0].Left)
	}
	if len(cfg.Layout.Lines[0].Right) != 0 {
		t.Errorf("expected empty Right, got %v", cfg.Layout.Lines[0].Right)
	}
}

func TestDefaultConfig_HasLines(t *testing.T) {
	cfg := DefaultConfig()
	if len(cfg.Layout.Lines) == 0 {
		t.Error("expected lines")
	}
	if cfg.Layout.Style != "default" {
		t.Errorf("expected 'default' style, got %q", cfg.Layout.Style)
	}
}

func TestDefaultPowerlineConfig_HasLines(t *testing.T) {
	cfg := DefaultPowerlineConfig()
	if len(cfg.Layout.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(cfg.Layout.Lines))
	}
	if cfg.Layout.Style != "powerline" {
		t.Errorf("expected 'powerline', got %q", cfg.Layout.Style)
	}
	if len(cfg.Layout.Lines[0].Left) == 0 {
		t.Error("expected Left components")
	}
	if len(cfg.Layout.Lines[0].Right) == 0 {
		t.Error("expected Right components")
	}
}

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
