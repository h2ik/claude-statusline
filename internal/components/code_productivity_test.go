package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestCodeProductivity_Name(t *testing.T) {
	r := render.New()
	cfg := config.DefaultConfig()
	c := NewCodeProductivity(r, cfg)

	if c.Name() != "code_productivity" {
		t.Errorf("expected 'code_productivity', got %q", c.Name())
	}
}

func TestCodeProductivity_Render_NoLinesChanged(t *testing.T) {
	r := render.New()
	cfg := config.DefaultConfig()
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      1.50,
			TotalDurationMS:   600000,
			TotalLinesAdded:   0,
			TotalLinesRemoved: 0,
		},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for no lines changed, got: %s", output)
	}
}

func TestCodeProductivity_Render_BothMetrics(t *testing.T) {
	r := render.New()
	cfg := config.DefaultConfig()
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      0.60,
			TotalDurationMS:   300000, // 5 minutes
			TotalLinesAdded:   80,
			TotalLinesRemoved: 20,
		},
	}

	output := c.Render(in)
	if !strings.Contains(output, "✏️") {
		t.Errorf("expected pencil emoji in output, got: %s", output)
	}
	// 100 lines / 5 min = 20 lines/min
	if !strings.Contains(output, "20 lines/min") {
		t.Errorf("expected '20 lines/min' in output, got: %s", output)
	}
	// $0.60 / 100 lines = $0.01/line (note: 0.006 rounds to $0.01)
	if !strings.Contains(output, "$0.01/line") {
		t.Errorf("expected '$0.01/line' in output, got: %s", output)
	}
}

func TestCodeProductivity_Render_VelocityOnly(t *testing.T) {
	r := render.New()
	showVelocity := true
	showCostPerLine := false
	cfg := &config.Config{
		Components: map[string]config.ComponentConfig{
			"code_productivity": {
				ShowVelocity:    &showVelocity,
				ShowCostPerLine: &showCostPerLine,
			},
		},
	}
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      0.30,
			TotalDurationMS:   120000, // 2 minutes
			TotalLinesAdded:   50,
			TotalLinesRemoved: 10,
		},
	}

	output := c.Render(in)
	// 60 lines / 2 min = 30 lines/min
	if !strings.Contains(output, "30 lines/min") {
		t.Errorf("expected '30 lines/min' in output, got: %s", output)
	}
	if strings.Contains(output, "$") {
		t.Errorf("expected no cost in output when disabled, got: %s", output)
	}
}

func TestCodeProductivity_Render_CostPerLineOnly(t *testing.T) {
	r := render.New()
	showVelocity := false
	showCostPerLine := true
	cfg := &config.Config{
		Components: map[string]config.ComponentConfig{
			"code_productivity": {
				ShowVelocity:    &showVelocity,
				ShowCostPerLine: &showCostPerLine,
			},
		},
	}
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      1.00,
			TotalDurationMS:   300000,
			TotalLinesAdded:   100,
			TotalLinesRemoved: 0,
		},
	}

	output := c.Render(in)
	// $1.00 / 100 lines = $0.01/line
	if !strings.Contains(output, "$0.01/line") {
		t.Errorf("expected '$0.01/line' in output, got: %s", output)
	}
	if strings.Contains(output, "lines/min") {
		t.Errorf("expected no velocity in output when disabled, got: %s", output)
	}
}

func TestCodeProductivity_Render_ZeroDuration(t *testing.T) {
	r := render.New()
	cfg := config.DefaultConfig()
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      0.50,
			TotalDurationMS:   0,
			TotalLinesAdded:   100,
			TotalLinesRemoved: 50,
		},
	}

	output := c.Render(in)
	// Should show cost per line only (velocity requires duration)
	if !strings.Contains(output, "$0.00/line") {
		t.Errorf("expected cost per line in output, got: %s", output)
	}
	if strings.Contains(output, "lines/min") {
		t.Errorf("expected no velocity when duration is zero, got: %s", output)
	}
}
