package render

import (
	"strings"
	"testing"
)

func TestDefaultStyle_RenderLine_JoinsWithSeparator(t *testing.T) {
	s := NewDefaultStyle(" | ")
	line := LineData{Left: []string{"alpha", "beta", "gamma"}}

	result := s.RenderLine(line, 80)

	if !strings.Contains(result, "alpha") {
		t.Error("expected 'alpha' in result")
	}
	if !strings.Contains(result, " | ") {
		t.Error("expected separator ' | ' in result")
	}
	if strings.Count(result, " | ") != 2 {
		t.Errorf("expected 2 separators, got %d", strings.Count(result, " | "))
	}
}

func TestDefaultStyle_RenderLine_IncludesRight(t *testing.T) {
	s := NewDefaultStyle(" | ")
	line := LineData{
		Left:  []string{"left-item"},
		Right: []string{"right-item"},
	}

	result := s.RenderLine(line, 80)

	if !strings.Contains(result, "left-item") {
		t.Error("expected 'left-item' in result")
	}
	if !strings.Contains(result, "right-item") {
		t.Error("expected 'right-item' appended after left in default style")
	}
}

func TestDefaultStyle_RenderLine_FiltersEmpty(t *testing.T) {
	s := NewDefaultStyle(" | ")
	line := LineData{Left: []string{"a", "", "b"}}

	result := s.RenderLine(line, 80)

	if strings.Contains(result, " |  | ") {
		t.Error("should not have double separator from empty component")
	}
}

func TestDefaultStyle_RenderLine_EmptyLine(t *testing.T) {
	s := NewDefaultStyle(" | ")
	line := LineData{}

	result := s.RenderLine(line, 80)

	if result != "" {
		t.Errorf("expected empty string for empty line, got %q", result)
	}
}

func TestDefaultStyle_ImplementsStyle(t *testing.T) {
	var _ Style = NewDefaultStyle(" | ")
}

func TestDefaultStyle_RenderLine_PreservesOrder(t *testing.T) {
	s := NewDefaultStyle(" | ")
	line := LineData{
		Left:  []string{"first", "second"},
		Right: []string{"third"},
	}

	result := s.RenderLine(line, 80)

	firstIdx := strings.Index(result, "first")
	secondIdx := strings.Index(result, "second")
	thirdIdx := strings.Index(result, "third")

	if firstIdx > secondIdx {
		t.Error("expected 'first' before 'second'")
	}
	if secondIdx > thirdIdx {
		t.Error("expected 'second' before 'third'")
	}
}

func TestDefaultStyle_RenderLine_WhitespaceOnlyFiltered(t *testing.T) {
	s := NewDefaultStyle(" | ")
	line := LineData{Left: []string{"a", "   ", "b"}}

	result := s.RenderLine(line, 80)

	expected := "a | b"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
