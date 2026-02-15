package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestBedrockModel_Name(t *testing.T) {
	r := render.New()
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil)

	if bm.Name() != "bedrock_model" {
		t.Errorf("expected 'bedrock_model', got %q", bm.Name())
	}
}

func TestBedrockModel_Render_EmptyForNonARN(t *testing.T) {
	r := render.New()
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil)

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
	r := render.New()
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil)

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
	r := render.New()
	c := cache.New(t.TempDir())

	falseVal := false
	cfg := &config.Config{
		Components: map[string]config.ComponentConfig{
			"bedrock_model": {ShowRegion: &falseVal},
		},
	}

	bm := NewBedrockModel(r, c, cfg, nil)

	in := &input.StatusLineInput{
		Model: input.ModelInfo{
			DisplayName: "arn:aws:bedrock:us-west-2:123456789012:application-inference-profile/abc123",
		},
	}

	output := bm.Render(in)
	if strings.Contains(output, "us-west-2") {
		t.Errorf("expected region 'us-west-2' to be hidden when show_region=false, got: %s", output)
	}
	// Should still contain the model name (falls back to "Bedrock Model" since aws CLI isn't available)
	if !strings.Contains(output, "Bedrock Model") {
		t.Errorf("expected 'Bedrock Model' in output even with region hidden, got: %s", output)
	}
}
