package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestBlockProjection_Name(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r, icons.New("emoji"))

	if c.Name() != "block_projection" {
		t.Errorf("expected 'block_projection', got %q", c.Name())
	}
}

func TestBlockProjection_Render_ZeroUtilization(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		FiveHour: input.UsageLimit{Utilization: 0.0},
		SevenDay: input.UsageLimit{Utilization: 0.0},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for zero utilization (Bedrock case), got: %s", output)
	}
}

func TestBlockProjection_Render_LowUtilization(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		FiveHour: input.UsageLimit{Utilization: 0.25},
		SevenDay: input.UsageLimit{Utilization: 0.10},
	}

	output := c.Render(in)
	if !strings.Contains(output, icons.New("emoji").Get(icons.Hourglass)) {
		t.Errorf("expected hourglass icon in output, got: %s", output)
	}
	if !strings.Contains(output, "5h: 25%") {
		t.Errorf("expected '5h: 25%%' in output, got: %s", output)
	}
	if !strings.Contains(output, "7d: 10%") {
		t.Errorf("expected '7d: 10%%' in output, got: %s", output)
	}
}

func TestBlockProjection_Render_HighUtilization(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		FiveHour: input.UsageLimit{Utilization: 0.85},
		SevenDay: input.UsageLimit{Utilization: 0.62},
	}

	output := c.Render(in)
	if !strings.Contains(output, "5h: 85%") {
		t.Errorf("expected '5h: 85%%' in output, got: %s", output)
	}
	if !strings.Contains(output, "7d: 62%") {
		t.Errorf("expected '7d: 62%%' in output, got: %s", output)
	}
}

func TestBlockProjection_Render_OnlyFiveHourData(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		FiveHour: input.UsageLimit{Utilization: 0.45},
		SevenDay: input.UsageLimit{Utilization: 0.0},
	}

	output := c.Render(in)
	if !strings.Contains(output, "5h: 45%") {
		t.Errorf("expected '5h: 45%%' in output, got: %s", output)
	}
}
