package render_test

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/component"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type mockComp struct {
	name   string
	output string
}

func (m *mockComp) Name() string                           { return m.name }
func (m *mockComp) Render(_ *input.StatusLineInput) string { return m.output }

func TestPowerlinePipeline_EndToEnd(t *testing.T) {
	r := render.New(nil)
	r.SetStyle(render.NewPowerlineStyle(r))

	registry := component.NewRegistry()
	registry.Register(&mockComp{name: "repo_info", output: "~/projects/myrepo (main)"})
	registry.Register(&mockComp{name: "model_info", output: "Claude Opus 4.6"})
	registry.Register(&mockComp{name: "time_display", output: "14:32"})
	registry.Register(&mockComp{name: "cost_daily", output: "TODAY $0.89"})
	registry.Register(&mockComp{name: "context_window", output: "42%"})

	in := &input.StatusLineInput{}
	var lines []render.LineData

	// Line 1: repo_info + model_info on left, time_display on right
	leftNames1, leftContent1 := registry.RenderNamedLine(in, []string{"repo_info", "model_info"})
	rightNames1, rightContent1 := registry.RenderNamedLine(in, []string{"time_display"})
	lines = append(lines, render.LineData{
		Left: leftContent1, LeftNames: leftNames1,
		Right: rightContent1, RightNames: rightNames1,
	})

	// Line 2: cost_daily on left, context_window on right
	leftNames2, leftContent2 := registry.RenderNamedLine(in, []string{"cost_daily"})
	rightNames2, rightContent2 := registry.RenderNamedLine(in, []string{"context_window"})
	lines = append(lines, render.LineData{
		Left: leftContent2, LeftNames: leftNames2,
		Right: rightContent2, RightNames: rightNames2,
	})

	output := r.RenderOutput(lines, 120)

	if output == "" {
		t.Fatal("expected non-empty output")
	}

	// Should have 2 lines
	outputLines := strings.Split(output, "\n")
	if len(outputLines) < 2 {
		t.Errorf("expected at least 2 lines, got %d", len(outputLines))
	}

	// Should contain powerline arrows (both forward and reverse)
	hasForward := strings.Contains(output, "\ue0b0")
	hasReverse := strings.Contains(output, "\ue0b2")
	if !hasForward {
		t.Error("expected forward powerline arrow in output")
	}
	if !hasReverse {
		t.Error("expected reverse powerline arrow in output")
	}

	// Verify plain text content is present (ANSI stripped and re-styled, but text should be there)
	stripped := render.StripANSI(output)
	if !strings.Contains(stripped, "~/projects/myrepo (main)") {
		t.Error("expected repo_info text in output")
	}
	if !strings.Contains(stripped, "Claude Opus 4.6") {
		t.Error("expected model_info text in output")
	}
	if !strings.Contains(stripped, "14:32") {
		t.Error("expected time_display text in output")
	}
}

func TestDefaultPipeline_EndToEnd(t *testing.T) {
	r := render.New(nil)
	// DefaultStyle is already set in New(), no SetStyle needed

	registry := component.NewRegistry()
	registry.Register(&mockComp{name: "repo_info", output: "~/projects/myrepo"})
	registry.Register(&mockComp{name: "model_info", output: "Claude Opus"})

	in := &input.StatusLineInput{}

	leftNames, leftContent := registry.RenderNamedLine(in, []string{"repo_info", "model_info"})
	lines := []render.LineData{
		{Left: leftContent, LeftNames: leftNames},
	}

	output := r.RenderOutput(lines, 80)

	if !strings.Contains(output, "~/projects/myrepo") {
		t.Error("expected repo_info content in output")
	}
	if !strings.Contains(output, " \u2502 ") {
		t.Error("expected default separator in output")
	}
	// No powerline arrows in default style
	if strings.Contains(output, "\ue0b0") {
		t.Error("unexpected powerline arrow in default style output")
	}
}

func TestPowerlinePipeline_SameCategoryMerge(t *testing.T) {
	r := render.New(nil)
	r.SetStyle(render.NewPowerlineStyle(r))

	registry := component.NewRegistry()
	registry.Register(&mockComp{name: "cost_daily", output: "TODAY $0.89"})
	registry.Register(&mockComp{name: "cost_live", output: "LIVE $2.47"})
	registry.Register(&mockComp{name: "burn_rate", output: "$0.12/min"})

	in := &input.StatusLineInput{}

	// All three are Cost category -- should merge into one segment
	leftNames, leftContent := registry.RenderNamedLine(in, []string{"cost_daily", "cost_live", "burn_rate"})
	lines := []render.LineData{
		{Left: leftContent, LeftNames: leftNames},
	}

	output := r.RenderOutput(lines, 80)

	// Only 1 trailing arrow, no inter-segment arrows (all same category)
	arrowCount := strings.Count(output, "\ue0b0")
	if arrowCount != 1 {
		t.Errorf("expected exactly 1 trailing arrow for all-same-category line, got %d", arrowCount)
	}
}

func TestPowerlinePipeline_EmptyComponentsFiltered(t *testing.T) {
	r := render.New(nil)
	r.SetStyle(render.NewPowerlineStyle(r))

	registry := component.NewRegistry()
	registry.Register(&mockComp{name: "repo_info", output: "~/repo"})
	registry.Register(&mockComp{name: "bedrock_model", output: ""}) // empty
	registry.Register(&mockComp{name: "model_info", output: "Claude"})

	in := &input.StatusLineInput{}

	leftNames, leftContent := registry.RenderNamedLine(in, []string{"repo_info", "bedrock_model", "model_info"})
	lines := []render.LineData{
		{Left: leftContent, LeftNames: leftNames},
	}

	output := r.RenderOutput(lines, 80)

	// bedrock_model should be filtered out by RenderNamedLine (empty output)
	// repo_info and model_info are both Info (blue) -- should merge
	if output == "" {
		t.Fatal("expected non-empty output")
	}
}
