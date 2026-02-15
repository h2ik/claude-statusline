package components

import (
	"regexp"
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// ============================================================
// ModelInfo tests
// ============================================================

func TestModelInfo_Name(t *testing.T) {
	r := render.New()
	c := NewModelInfo(r)

	if c.Name() != "model_info" {
		t.Errorf("expected 'model_info', got %q", c.Name())
	}
}

func TestModelInfo_Render_OpusEmoji(t *testing.T) {
	r := render.New()
	c := NewModelInfo(r)

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "Claude Opus 4"},
	}

	output := c.Render(in)
	if !strings.Contains(output, "\xf0\x9f\xa7\xa0") { // brain emoji
		t.Errorf("expected brain emoji for Opus, got: %s", output)
	}
	if !strings.Contains(output, "Claude Opus 4") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_SonnetEmoji(t *testing.T) {
	r := render.New()
	c := NewModelInfo(r)

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "Claude Sonnet 4.5"},
	}

	output := c.Render(in)
	if !strings.Contains(output, "\xf0\x9f\x8e\xb5") { // musical note emoji
		t.Errorf("expected musical note emoji for Sonnet, got: %s", output)
	}
	if !strings.Contains(output, "Claude Sonnet 4.5") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_HaikuEmoji(t *testing.T) {
	r := render.New()
	c := NewModelInfo(r)

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "Claude 3.5 Haiku"},
	}

	output := c.Render(in)
	if !strings.Contains(output, "\xe2\x9a\xa1") { // lightning emoji
		t.Errorf("expected lightning emoji for Haiku, got: %s", output)
	}
	if !strings.Contains(output, "Claude 3.5 Haiku") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_UnknownModel(t *testing.T) {
	r := render.New()
	c := NewModelInfo(r)

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "SomeUnknownModel"},
	}

	output := c.Render(in)
	if !strings.Contains(output, "\xf0\x9f\xa4\x96") { // robot emoji
		t.Errorf("expected robot emoji for unknown model, got: %s", output)
	}
	if !strings.Contains(output, "SomeUnknownModel") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_EmptyDisplayName(t *testing.T) {
	r := render.New()
	c := NewModelInfo(r)

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: ""},
	}

	output := c.Render(in)
	if !strings.Contains(output, "Claude") {
		t.Errorf("expected default 'Claude' name when empty, got: %s", output)
	}
}

// ============================================================
// Commits tests
// ============================================================

func TestCommits_Name(t *testing.T) {
	r := render.New()
	c := NewCommits(r)

	if c.Name() != "commits" {
		t.Errorf("expected 'commits', got %q", c.Name())
	}
}

func TestCommits_Render_NonGitDir(t *testing.T) {
	r := render.New()
	c := NewCommits(r)

	in := &input.StatusLineInput{
		Workspace: input.Workspace{
			CurrentDir: "/tmp/definitely-not-a-git-repo-12345",
		},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for non-git directory, got: %q", output)
	}
}

// ============================================================
// Submodules tests
// ============================================================

func TestSubmodules_Name(t *testing.T) {
	r := render.New()
	c := NewSubmodules(r)

	if c.Name() != "submodules" {
		t.Errorf("expected 'submodules', got %q", c.Name())
	}
}

func TestSubmodules_Render_NonGitDir(t *testing.T) {
	r := render.New()
	c := NewSubmodules(r)

	in := &input.StatusLineInput{
		Workspace: input.Workspace{
			CurrentDir: "/tmp/definitely-not-a-git-repo-12345",
		},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for non-git directory, got: %q", output)
	}
}

// ============================================================
// VersionInfo tests
// ============================================================

func TestVersionInfo_Name(t *testing.T) {
	r := render.New()
	tmpDir := t.TempDir()
	c := NewVersionInfo(r, cache.New(tmpDir))

	if c.Name() != "version_info" {
		t.Errorf("expected 'version_info', got %q", c.Name())
	}
}

// ============================================================
// TimeDisplay tests
// ============================================================

func TestTimeDisplay_Name(t *testing.T) {
	r := render.New()
	c := NewTimeDisplay(r)

	if c.Name() != "time_display" {
		t.Errorf("expected 'time_display', got %q", c.Name())
	}
}

func TestTimeDisplay_Render_ContainsTimePattern(t *testing.T) {
	r := render.New()
	c := NewTimeDisplay(r)

	in := &input.StatusLineInput{}

	output := c.Render(in)

	// Should contain a time in HH:MM format somewhere in the output
	matched, _ := regexp.MatchString(`\d{2}:\d{2}`, output)
	if !matched {
		t.Errorf("expected HH:MM time pattern in output, got: %s", output)
	}
}

func TestTimeDisplay_Render_ContainsClockEmoji(t *testing.T) {
	r := render.New()
	c := NewTimeDisplay(r)

	in := &input.StatusLineInput{}

	output := c.Render(in)

	if !strings.Contains(output, "\xf0\x9f\x95\x90") { // clock emoji
		t.Errorf("expected clock emoji in output, got: %s", output)
	}
}

