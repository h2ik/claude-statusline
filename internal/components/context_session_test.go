package components

import (
    "strings"
    "testing"

    "github.com/h2ik/claude-statusline/internal/input"
    "github.com/h2ik/claude-statusline/internal/render"
)

// ============================================================
// ContextWindow tests
// ============================================================

func TestContextWindow_Name(t *testing.T) {
    r := render.New()
    c := NewContextWindow(r)

    if c.Name() != "context_window" {
        t.Errorf("expected 'context_window', got %q", c.Name())
    }
}

func TestContextWindow_Render_EmptyWhenZeroPercent(t *testing.T) {
    r := render.New()
    c := NewContextWindow(r)

    in := &input.StatusLineInput{
        ContextWindow: input.ContextWindow{
            UsedPercentage: 0,
        },
    }

    output := c.Render(in)
    if output != "" {
        t.Errorf("expected empty string when percentage is 0, got: %q", output)
    }
}

func TestContextWindow_Render_GreenZone(t *testing.T) {
    r := render.New()
    c := NewContextWindow(r)

    in := &input.StatusLineInput{
        ContextWindow: input.ContextWindow{
            UsedPercentage: 25,
        },
    }

    output := c.Render(in)
    if !strings.Contains(output, "25%") {
        t.Errorf("expected '25%%' in output, got: %s", output)
    }
    if !strings.Contains(output, "\xf0\x9f\xa7\xa0") { // brain emoji
        t.Errorf("expected brain emoji in output, got: %s", output)
    }
}

func TestContextWindow_Render_YellowZone(t *testing.T) {
    r := render.New()
    c := NewContextWindow(r)

    in := &input.StatusLineInput{
        ContextWindow: input.ContextWindow{
            UsedPercentage: 55,
        },
    }

    output := c.Render(in)
    if !strings.Contains(output, "55%") {
        t.Errorf("expected '55%%' in output, got: %s", output)
    }
}

func TestContextWindow_Render_RedZone(t *testing.T) {
    r := render.New()
    c := NewContextWindow(r)

    in := &input.StatusLineInput{
        ContextWindow: input.ContextWindow{
            UsedPercentage: 80,
        },
    }

    output := c.Render(in)
    if !strings.Contains(output, "80%") {
        t.Errorf("expected '80%%' in output, got: %s", output)
    }
}

func TestContextWindow_Render_WarningAt95(t *testing.T) {
    r := render.New()
    c := NewContextWindow(r)

    in := &input.StatusLineInput{
        ContextWindow: input.ContextWindow{
            UsedPercentage: 95,
        },
    }

    output := c.Render(in)
    if !strings.Contains(output, "95%") {
        t.Errorf("expected '95%%' in output, got: %s", output)
    }
    if !strings.Contains(output, "\xe2\x9a\xa0\xef\xb8\x8f") { // warning emoji
        t.Errorf("expected warning emoji at 95%%, got: %s", output)
    }
}

func TestContextWindow_Render_WithTokenCounts(t *testing.T) {
    r := render.New()
    c := NewContextWindow(r)

    in := &input.StatusLineInput{
        ContextWindow: input.ContextWindow{
            UsedPercentage:    45,
            ContextWindowSize: 200000,
        },
    }

    output := c.Render(in)
    if !strings.Contains(output, "45%") {
        t.Errorf("expected '45%%' in output, got: %s", output)
    }
    // 45% of 200000 = 90000 tokens = 90K
    if !strings.Contains(output, "90K/200K") {
        t.Errorf("expected '90K/200K' token count in output, got: %s", output)
    }
}

// ============================================================
// SessionMode tests
// ============================================================

func TestSessionMode_Name(t *testing.T) {
    r := render.New()
    c := NewSessionMode(r)

    if c.Name() != "session_mode" {
        t.Errorf("expected 'session_mode', got %q", c.Name())
    }
}

func TestSessionMode_Render_EmptyWhenNoStyle(t *testing.T) {
    r := render.New()
    c := NewSessionMode(r)

    in := &input.StatusLineInput{
        OutputStyle: input.OutputStyle{
            Name: "",
        },
    }

    output := c.Render(in)
    if output != "" {
        t.Errorf("expected empty string when style is empty, got: %q", output)
    }
}

func TestSessionMode_Render_EmptyWhenDefault(t *testing.T) {
    r := render.New()
    c := NewSessionMode(r)

    in := &input.StatusLineInput{
        OutputStyle: input.OutputStyle{
            Name: "default",
        },
    }

    output := c.Render(in)
    if output != "" {
        t.Errorf("expected empty string when style is 'default', got: %q", output)
    }
}

func TestSessionMode_Render_ExplanatoryStyle(t *testing.T) {
    r := render.New()
    c := NewSessionMode(r)

    in := &input.StatusLineInput{
        OutputStyle: input.OutputStyle{
            Name: "explanatory",
        },
    }

    output := c.Render(in)
    if !strings.Contains(output, "\xf0\x9f\x93\x9a") { // book emoji
        t.Errorf("expected book emoji for 'explanatory' style, got: %s", output)
    }
    if !strings.Contains(output, "Style:") {
        t.Errorf("expected 'Style:' label in output, got: %s", output)
    }
    if !strings.Contains(output, "explanatory") {
        t.Errorf("expected 'explanatory' in output, got: %s", output)
    }
}

func TestSessionMode_Render_UnknownStyle(t *testing.T) {
    r := render.New()
    c := NewSessionMode(r)

    in := &input.StatusLineInput{
        OutputStyle: input.OutputStyle{
            Name: "some-custom-style",
        },
    }

    output := c.Render(in)
    if !strings.Contains(output, "\xe2\x9c\xa8") { // sparkles emoji
        t.Errorf("expected sparkles emoji for unknown style, got: %s", output)
    }
    if !strings.Contains(output, "some-custom-style") {
        t.Errorf("expected 'some-custom-style' in output, got: %s", output)
    }
}
