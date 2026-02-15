package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestCacheEfficiency_Name(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	if c.Name() != "cache_efficiency" {
		t.Errorf("expected 'cache_efficiency', got %q", c.Name())
	}
}

func TestCacheEfficiency_Render_ZeroTokens(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	in := &input.StatusLineInput{
		CurrentUsage: input.UsageInfo{
			InputTokens:              0,
			CacheReadInputTokens:     0,
			CacheCreationInputTokens: 0,
		},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for zero tokens, got: %s", output)
	}
}

func TestCacheEfficiency_Render_HighEfficiency(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	in := &input.StatusLineInput{
		CurrentUsage: input.UsageInfo{
			InputTokens:              1000,
			CacheReadInputTokens:     7000,
			CacheCreationInputTokens: 2000,
		},
	}

	output := c.Render(in)
	if !strings.Contains(output, "💾") {
		t.Errorf("expected disk emoji in output, got: %s", output)
	}
	// 7000 / 10000 = 70%
	if !strings.Contains(output, "70% cache") {
		t.Errorf("expected '70%% cache' for high efficiency, got: %s", output)
	}
}

func TestCacheEfficiency_Render_LowEfficiency(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	in := &input.StatusLineInput{
		CurrentUsage: input.UsageInfo{
			InputTokens:              7000,
			CacheReadInputTokens:     1000,
			CacheCreationInputTokens: 2000,
		},
	}

	output := c.Render(in)
	// 1000 / 10000 = 10%
	if !strings.Contains(output, "10% cache") {
		t.Errorf("expected '10%% cache' for low efficiency, got: %s", output)
	}
}

func TestCacheEfficiency_Render_MediumEfficiency(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	in := &input.StatusLineInput{
		CurrentUsage: input.UsageInfo{
			InputTokens:              3000,
			CacheReadInputTokens:     5000,
			CacheCreationInputTokens: 2000,
		},
	}

	output := c.Render(in)
	// 5000 / 10000 = 50%
	if !strings.Contains(output, "50% cache") {
		t.Errorf("expected '50%% cache' for medium efficiency, got: %s", output)
	}
}
