package components

import (
	"strings"
	"testing"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestBedrockModel_Name(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	if bm.Name() != "bedrock_model" {
		t.Errorf("expected 'bedrock_model', got %q", bm.Name())
	}
}

func TestBedrockModel_Render_EmptyForNonARN(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{
			DisplayName: "claude-sonnet-4",
		},
	}

	output := bm.Render(in)
	if output != "" {
		t.Errorf("expected empty string for non-ARN model, got: %q", output)
	}
}

func TestBedrockModel_Render_ShowsRegionByDefault(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{
			DisplayName: "arn:aws:bedrock:us-west-2:123456789012:application-inference-profile/abc123",
		},
	}

	output := bm.Render(in)
	if !strings.Contains(output, "us-west-2") {
		t.Errorf("expected region 'us-west-2' in output by default, got: %s", output)
	}
}

func TestBedrockModel_Render_HidesRegionWhenConfigured(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())

	falseVal := false
	cfg := &config.Config{
		Components: map[string]config.ComponentConfig{
			"bedrock_model": {ShowRegion: &falseVal},
		},
	}

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{
			DisplayName: "arn:aws:bedrock:us-west-2:123456789012:application-inference-profile/abc123",
		},
	}

	output := bm.Render(in)
	// The region may appear inside the raw ARN fallback name, but it should NOT
	// appear as a parenthetical suffix like "(us-west-2)" when show_region=false.
	if strings.Contains(output, "(us-west-2)") {
		t.Errorf("expected region suffix '(us-west-2)' to be hidden when show_region=false, got: %s", output)
	}
	if output == "" {
		t.Error("expected non-empty output even with region hidden")
	}
}

func TestBedrockModel_Render_FallsBackToFriendlyNameFromARN(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	// ARN with a recognizable model pattern in the resource path — should resolve
	// to friendly name even without AWS CLI access
	in := &input.StatusLineInput{
		Model: input.ModelInfo{
			DisplayName: "arn:aws:bedrock:us-east-2:123456:inference-profile/us.anthropic.claude-opus-4-v1:0",
		},
	}

	output := bm.Render(in)
	if !strings.Contains(output, "Claude Opus 4") {
		t.Errorf("expected 'Claude Opus 4' from ARN fallback, got: %s", output)
	}
}

func TestGetFriendlyName_FromCatalog(t *testing.T) {
	r := render.New(nil)
	cacheDir := t.TempDir()
	c := cache.New(cacheDir)
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	// Pre-seed the cache with a model catalog
	catalog := `[{"id":"anthropic.claude-opus-4-6-v1","name":"Claude Opus 4.6"},{"id":"anthropic.claude-sonnet-4-20250514-v1:0","name":"Claude Sonnet 4"}]`
	_ = c.Set("bedrock:model-catalog", []byte(catalog), 24*time.Hour)

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	// Should match via catalog
	name := bm.getFriendlyName("arn:aws:bedrock:us-east-2:123456:foundation-model/anthropic.claude-opus-4-6-v1")
	if name != "Claude Opus 4.6" {
		t.Errorf("expected 'Claude Opus 4.6', got %q", name)
	}
}

func TestGetFriendlyName_FallbackToHardcoded(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	// No catalog in cache — should fall back to hardcoded map
	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	name := bm.getFriendlyName("arn:aws:bedrock:us-east-2:123456:foundation-model/anthropic.claude-sonnet-4-20250514-v1:0")
	if name != "Claude Sonnet 4" {
		t.Errorf("expected 'Claude Sonnet 4' from hardcoded fallback, got %q", name)
	}
}

func TestGetFriendlyName_RawARNFallback(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	// Totally unknown Bedrock model — should return generic "Bedrock Model" label
	name := bm.getFriendlyName("arn:aws:bedrock:us-east-2:123456:foundation-model/anthropic.claude-99-turbo-v1:0")
	if name != "Bedrock Model" {
		t.Errorf("expected 'Bedrock Model' fallback, got %q", name)
	}
}
