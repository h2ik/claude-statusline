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
	if strings.Contains(output, "us-west-2") {
		t.Errorf("expected region 'us-west-2' to be hidden when show_region=false, got: %s", output)
	}
	// Should still contain the model name (falls back to "Bedrock Model" since aws CLI isn't available)
	if !strings.Contains(output, "Bedrock Model") {
		t.Errorf("expected 'Bedrock Model' in output even with region hidden, got: %s", output)
	}
}

func TestBedrockModel_Render_StripsContextWindowSuffix(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	// Pre-seed the cache for the ARN *without* the suffix
	arn := "arn:aws:bedrock:us-west-2:123456789012:application-inference-profile/abc123"
	_ = c.Set("bedrock:v3:"+arn, []byte("Claude Opus 4\tus-west-2"), 24*time.Hour)

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	// Pass the ARN with a context window suffix
	in := &input.StatusLineInput{
		Model: input.ModelInfo{
			DisplayName: arn + "[1m]",
		},
	}

	output := bm.Render(in)
	if !strings.Contains(output, "Claude Opus 4") {
		t.Errorf("expected 'Claude Opus 4' after stripping suffix, got: %s", output)
	}
}

func TestGetFriendlyName_FromCatalog(t *testing.T) {
	r := render.New(nil)
	cacheDir := t.TempDir()
	c := cache.New(cacheDir)
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	// Pre-seed the cache with a model catalog
	catalog := `[{"id":"anthropic.claude-opus-4-6-v1","name":"Claude Opus 4.6"},{"id":"anthropic.claude-sonnet-4-20250514-v1:0","name":"Claude Sonnet 4"}]`
	_ = c.Set("bedrock:v3:model-catalog", []byte(catalog), 24*time.Hour)

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

	// Totally unknown model — should return the raw ARN
	name := bm.getFriendlyName("arn:aws:bedrock:us-east-2:123456:foundation-model/anthropic.claude-99-turbo-v1:0")
	if name != "arn:aws:bedrock:us-east-2:123456:foundation-model/anthropic.claude-99-turbo-v1:0" {
		t.Errorf("expected raw ARN passthrough, got %q", name)
	}
}

func TestResolveBedrockARN_FallsBackToProfileCatalog(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	// Pre-seed the inference profile catalog cache.
	// This simulates what loadProfileCatalog() would cache from
	// "aws bedrock list-inference-profiles".
	catalog := `[{"arn":"arn:aws:bedrock:us-east-2:123456:application-inference-profile/opaque123","modelArn":"arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-opus-4-6-v1"}]`
	_ = c.Set("bedrock:v3:profile-catalog", []byte(catalog), 24*time.Hour)

	// Also seed the model catalog so getFriendlyName can resolve the model ARN.
	modelCatalog := `[{"id":"anthropic.claude-opus-4-6-v1","name":"Claude Opus 4.6"}]`
	_ = c.Set("bedrock:v3:model-catalog", []byte(modelCatalog), 24*time.Hour)

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	// The AWS CLI call for get-inference-profile will fail (no real AWS in tests),
	// but the profile catalog should resolve the opaque ARN to the model ARN,
	// and then getFriendlyName resolves that to "Claude Opus 4.6".
	name, region := bm.resolveBedrockARN("arn:aws:bedrock:us-east-2:123456:application-inference-profile/opaque123")

	if name != "Claude Opus 4.6" {
		t.Errorf("expected 'Claude Opus 4.6' via profile catalog fallback, got %q", name)
	}
	if region != "us-east-2" {
		t.Errorf("expected region 'us-east-2', got %q", region)
	}
}

func TestResolveBedrockARN_FallsBackToFriendlyNameOnOriginalARN(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	// An inference-profile ARN (not application-) that contains a recognizable model slug.
	// AWS CLI will fail in tests, but the ARN itself contains "claude-opus-4-6"
	// which should match the static fallback in getFriendlyName.
	name, region := bm.resolveBedrockARN("arn:aws:bedrock:us-west-2:123456:inference-profile/us.anthropic.claude-opus-4-6-v1")

	if name != "Claude Opus 4.6" {
		t.Errorf("expected 'Claude Opus 4.6' via static fallback on original ARN, got %q", name)
	}
	if region != "us-west-2" {
		t.Errorf("expected region 'us-west-2', got %q", region)
	}
}

func TestResolveBedrockARN_BedrockModelOnlyWhenNothingMatches(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))

	// Completely opaque ARN, no profile catalog, no model catalog, no static match.
	// This is the only case that should return "Bedrock Model".
	name, region := bm.resolveBedrockARN("arn:aws:bedrock:eu-west-1:999999:application-inference-profile/totallyopaque")

	if name != "Bedrock Model" {
		t.Errorf("expected 'Bedrock Model' as ultimate fallback, got %q", name)
	}
	if region != "eu-west-1" {
		t.Errorf("expected region 'eu-west-1', got %q", region)
	}
}

func TestLoadProfileCatalog_FromCache(t *testing.T) {
	r := render.New(nil)
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	catalog := `[{"arn":"arn:aws:bedrock:us-east-2:123456:application-inference-profile/abc","modelArn":"arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-sonnet-4-6-v1"}]`
	_ = c.Set("bedrock:v3:profile-catalog", []byte(catalog), 24*time.Hour)

	bm := NewBedrockModel(r, c, cfg, nil, icons.New("emoji"))
	profiles := bm.loadProfileCatalog()

	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile entry, got %d", len(profiles))
	}
	if profiles[0].ARN != "arn:aws:bedrock:us-east-2:123456:application-inference-profile/abc" {
		t.Errorf("unexpected ARN: %q", profiles[0].ARN)
	}
	if profiles[0].ModelARN != "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-sonnet-4-6-v1" {
		t.Errorf("unexpected ModelARN: %q", profiles[0].ModelARN)
	}
}
