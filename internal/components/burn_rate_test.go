package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestBurnRate_Name(t *testing.T) {
	r := render.New()
	c := NewBurnRate(r)

	if c.Name() != "burn_rate" {
		t.Errorf("expected 'burn_rate', got %q", c.Name())
	}
}

func TestBurnRate_Render_ZeroDuration(t *testing.T) {
	r := render.New()
	c := NewBurnRate(r)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:    1.50,
			TotalDurationMS: 0,
		},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for zero duration, got: %s", output)
	}
}

func TestBurnRate_Render_DisplaysRate(t *testing.T) {
	r := render.New()
	c := NewBurnRate(r)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:    1.20,
			TotalDurationMS: 600000, // 10 minutes
		},
	}

	output := c.Render(in)
	if !strings.Contains(output, "\xf0\x9f\x94\xa5") {
		t.Errorf("expected fire emoji in output, got: %s", output)
	}
	if !strings.Contains(output, "$0.12/min") {
		t.Errorf("expected '$0.12/min' for burn rate, got: %s", output)
	}
}

func TestBurnRate_Render_RoundsCorrectly(t *testing.T) {
	r := render.New()
	c := NewBurnRate(r)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:    0.755,
			TotalDurationMS: 180000, // 3 minutes
		},
	}

	output := c.Render(in)
	// 0.755 / 3 = 0.2516666... should round to $0.25/min
	if !strings.Contains(output, "$0.25/min") {
		t.Errorf("expected '$0.25/min' for burn rate, got: %s", output)
	}
}
