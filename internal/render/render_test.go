package render

import (
    "strings"
    "testing"
)

func TestRenderer_RenderLines(t *testing.T) {
    r := New()

    lines := [][]string{
        {"component1", "component2"},
        {"component3"},
    }

    output := r.RenderLines(lines)

    // Should have 2 lines
    lineCount := strings.Count(output, "\n")
    if lineCount != 1 { // 2 lines = 1 newline
        t.Errorf("expected 1 newline, got %d", lineCount)
    }

    // Should contain separator
    if !strings.Contains(output, " │ ") {
        t.Error("expected separator ' │ ' in output")
    }
}

func TestRenderer_RenderLines_FiltersEmpty(t *testing.T) {
    r := New()

    lines := [][]string{
        {"component1", "", "component2"},
        {"", "  ", ""},
        {"component3"},
    }

    output := r.RenderLines(lines)

    // Line with all empty components should be filtered out entirely
    lineCount := strings.Count(output, "\n")
    if lineCount != 1 { // 2 non-empty lines = 1 newline
        t.Errorf("expected 1 newline, got %d", lineCount)
    }

    // Should contain both components from first line
    if !strings.Contains(output, "component1") {
        t.Error("expected 'component1' in output")
    }
    if !strings.Contains(output, "component2") {
        t.Error("expected 'component2' in output")
    }
    if !strings.Contains(output, "component3") {
        t.Error("expected 'component3' in output")
    }
}

func TestRenderer_RenderLines_Empty(t *testing.T) {
    r := New()

    lines := [][]string{}
    output := r.RenderLines(lines)

    if output != "" {
        t.Errorf("expected empty string for empty input, got %q", output)
    }
}

func TestRenderer_StyleHelpers(t *testing.T) {
    r := New()

    // Each style helper should return a non-empty string containing the input
    tests := []struct {
        name   string
        fn     func(string) string
        input  string
    }{
        {"Dimmed", r.Dimmed, "dim text"},
        {"Text", r.Text, "normal text"},
        {"Green", r.Green, "green text"},
        {"Red", r.Red, "red text"},
        {"Yellow", r.Yellow, "yellow text"},
        {"Blue", r.Blue, "blue text"},
        {"Mauve", r.Mauve, "mauve text"},
        {"Peach", r.Peach, "peach text"},
        {"Teal", r.Teal, "teal text"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.fn(tt.input)
            if result == "" {
                t.Errorf("%s returned empty string", tt.name)
            }
            if !strings.Contains(result, tt.input) {
                t.Errorf("%s result %q does not contain input %q", tt.name, result, tt.input)
            }
        })
    }
}
