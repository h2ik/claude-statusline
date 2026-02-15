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

	// Should have 3 lines in layout
	if len(cfg.Layout.Lines) != 3 {
		t.Fatalf("expected 3 layout lines, got %d", len(cfg.Layout.Lines))
	}

	// Verify line 1
	if len(cfg.Layout.Lines[0]) != 1 || cfg.Layout.Lines[0][0] != "repo_info" {
		t.Errorf("line 1: expected [repo_info], got %v", cfg.Layout.Lines[0])
	}

	// Verify line 2
	expectedLine2 := []string{"bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"}
	if len(cfg.Layout.Lines[1]) != len(expectedLine2) {
		t.Fatalf("line 2: expected %d components, got %d", len(expectedLine2), len(cfg.Layout.Lines[1]))
	}
	for i, comp := range expectedLine2 {
		if cfg.Layout.Lines[1][i] != comp {
			t.Errorf("line 2[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[1][i])
		}
	}

	// Verify line 3
	expectedLine3 := []string{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"}
	if len(cfg.Layout.Lines[2]) != len(expectedLine3) {
		t.Fatalf("line 3: expected %d components, got %d", len(expectedLine3), len(cfg.Layout.Lines[2]))
	}
	for i, comp := range expectedLine3 {
		if cfg.Layout.Lines[2][i] != comp {
			t.Errorf("line 3[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[2][i])
		}
	}
}

func TestLoad_ParsesExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Write a custom TOML with 2 lines and show_region=false
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

	// Should have 2 lines
	if len(cfg.Layout.Lines) != 2 {
		t.Fatalf("expected 2 layout lines, got %d", len(cfg.Layout.Lines))
	}

	// Verify line 2
	if len(cfg.Layout.Lines[1]) != 2 {
		t.Fatalf("line 2: expected 2 components, got %d", len(cfg.Layout.Lines[1]))
	}
	if cfg.Layout.Lines[1][0] != "bedrock_model" || cfg.Layout.Lines[1][1] != "time_display" {
		t.Errorf("line 2: expected [bedrock_model, time_display], got %v", cfg.Layout.Lines[1])
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

	// Should have 3 lines
	if len(cfg.Layout.Lines) != 3 {
		t.Fatalf("expected 3 layout lines, got %d", len(cfg.Layout.Lines))
	}

	// Verify line 1
	if len(cfg.Layout.Lines[0]) != 1 || cfg.Layout.Lines[0][0] != "repo_info" {
		t.Errorf("line 1: expected [repo_info], got %v", cfg.Layout.Lines[0])
	}

	// Verify line 2
	expectedLine2 := []string{"bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"}
	if len(cfg.Layout.Lines[1]) != len(expectedLine2) {
		t.Fatalf("line 2: expected %d components, got %d", len(expectedLine2), len(cfg.Layout.Lines[1]))
	}
	for i, comp := range expectedLine2 {
		if cfg.Layout.Lines[1][i] != comp {
			t.Errorf("line 2[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[1][i])
		}
	}

	// Verify line 3
	expectedLine3 := []string{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"}
	if len(cfg.Layout.Lines[2]) != len(expectedLine3) {
		t.Fatalf("line 3: expected %d components, got %d", len(expectedLine3), len(cfg.Layout.Lines[2]))
	}
	for i, comp := range expectedLine3 {
		if cfg.Layout.Lines[2][i] != comp {
			t.Errorf("line 3[%d]: expected %q, got %q", i, comp, cfg.Layout.Lines[2][i])
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
}
