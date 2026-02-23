package components

import (
	"regexp"
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// ============================================================
// ModelInfo tests
// ============================================================

func TestModelInfo_Name(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	if c.Name() != "model_info" {
		t.Errorf("expected 'model_info', got %q", c.Name())
	}
}

func TestModelInfo_Render_OpusEmoji(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "Claude Opus 4"},
	}

	output := c.Render(in)
	if !strings.Contains(output, icons.New("emoji").Get(icons.Brain)) {
		t.Errorf("expected brain icon for Opus, got: %s", output)
	}
	if !strings.Contains(output, "Claude Opus 4") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_SonnetEmoji(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "Claude Sonnet 4.5"},
	}

	output := c.Render(in)
	if !strings.Contains(output, icons.New("emoji").Get(icons.Music)) {
		t.Errorf("expected music icon for Sonnet, got: %s", output)
	}
	if !strings.Contains(output, "Claude Sonnet 4.5") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_HaikuEmoji(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "Claude 3.5 Haiku"},
	}

	output := c.Render(in)
	if !strings.Contains(output, icons.New("emoji").Get(icons.Lightning)) {
		t.Errorf("expected lightning icon for Haiku, got: %s", output)
	}
	if !strings.Contains(output, "Claude 3.5 Haiku") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_UnknownModel(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "SomeUnknownModel"},
	}

	output := c.Render(in)
	if !strings.Contains(output, icons.New("emoji").Get(icons.Robot)) {
		t.Errorf("expected robot icon for unknown model, got: %s", output)
	}
	if !strings.Contains(output, "SomeUnknownModel") {
		t.Errorf("expected model name in output, got: %s", output)
	}
}

func TestModelInfo_Render_EmptyDisplayName(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: ""},
	}

	output := c.Render(in)
	if !strings.Contains(output, "Claude") {
		t.Errorf("expected default 'Claude' name when empty, got: %s", output)
	}
}

func TestModelInfo_Render_BedrockARN_ReturnsEmpty(t *testing.T) {
	r := render.New(nil)
	c := NewModelInfo(r, icons.New("emoji"))

	in := &input.StatusLineInput{
		Model: input.ModelInfo{DisplayName: "arn:aws:bedrock:us-east-2:123456:application-inference-profile/abc123"},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for Bedrock ARN, got: %q", output)
	}
}

// ============================================================
// Commits tests
// ============================================================

func TestCommits_Name(t *testing.T) {
	r := render.New(nil)
	c := NewCommits(r, icons.New("emoji"))

	if c.Name() != "commits" {
		t.Errorf("expected 'commits', got %q", c.Name())
	}
}

func TestCommits_Render_NonGitDir(t *testing.T) {
	r := render.New(nil)
	c := NewCommits(r, icons.New("emoji"))

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
	r := render.New(nil)
	c := NewSubmodules(r, icons.New("emoji"))

	if c.Name() != "submodules" {
		t.Errorf("expected 'submodules', got %q", c.Name())
	}
}

func TestSubmodules_Render_NonGitDir(t *testing.T) {
	r := render.New(nil)
	c := NewSubmodules(r, icons.New("emoji"))

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
	r := render.New(nil)
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
	r := render.New(nil)
	c := NewTimeDisplay(r, icons.New("emoji"))

	if c.Name() != "time_display" {
		t.Errorf("expected 'time_display', got %q", c.Name())
	}
}

func TestTimeDisplay_Render_ContainsTimePattern(t *testing.T) {
	r := render.New(nil)
	c := NewTimeDisplay(r, icons.New("emoji"))

	in := &input.StatusLineInput{}

	output := c.Render(in)

	// Should contain a time in HH:MM format somewhere in the output
	matched, _ := regexp.MatchString(`\d{2}:\d{2}`, output)
	if !matched {
		t.Errorf("expected HH:MM time pattern in output, got: %s", output)
	}
}

func TestTimeDisplay_Render_ContainsClockEmoji(t *testing.T) {
	r := render.New(nil)
	c := NewTimeDisplay(r, icons.New("emoji"))

	in := &input.StatusLineInput{}

	output := c.Render(in)

	if !strings.Contains(output, icons.New("emoji").Get(icons.Clock)) {
		t.Errorf("expected clock icon in output, got: %s", output)
	}
}
