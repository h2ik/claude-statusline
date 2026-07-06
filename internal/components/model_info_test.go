package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestModelInfo_Render_FableEmoji(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "Claude Fable 5"},
	}

	output := c.Render(in)
	if !strings.Contains(output, icons.New("emoji").Get(icons.Sparkles)) {
		t.Errorf("expected sparkles icon for Fable, got: %s", output)
	}
	if !strings.Contains(output, "Claude Fable 5") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_MythosEmoji(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "Claude Mythos 5"},
	}

	output := c.Render(in)
	if !strings.Contains(output, icons.New("emoji").Get(icons.Sparkles)) {
		t.Errorf("expected sparkles icon for Mythos, got: %s", output)
	}
	if !strings.Contains(output, "Claude Mythos 5") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_MythosPreviewEmoji(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "Claude Mythos Preview"},
	}

	output := c.Render(in)
	if !strings.Contains(output, icons.New("emoji").Get(icons.Sparkles)) {
		t.Errorf("expected sparkles icon for Mythos Preview, got: %s", output)
	}
}
