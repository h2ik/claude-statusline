package render

import "testing"

func TestStripANSI_PlainString(t *testing.T) {
	input := "hello world"
	result := StripANSI(input)
	if result != input {
		t.Errorf("plain string should pass through unchanged, got %q", result)
	}
}

func TestStripANSI_RemovesColorCodes(t *testing.T) {
	input := "\x1b[38;2;163;227;161mclean\x1b[0m"
	result := StripANSI(input)
	if result != "clean" {
		t.Errorf("expected 'clean', got %q", result)
	}
}

func TestStripANSI_RemovesBackgroundCodes(t *testing.T) {
	input := "\x1b[48;2;137;180;250mrepo_info\x1b[0m"
	result := StripANSI(input)
	if result != "repo_info" {
		t.Errorf("expected 'repo_info', got %q", result)
	}
}

func TestStripANSI_EmptyString(t *testing.T) {
	result := StripANSI("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestStripANSI_MultipleSequences(t *testing.T) {
	input := "\x1b[1m\x1b[38;2;100;200;100mgreen bold\x1b[0m normal \x1b[31mred\x1b[0m"
	result := StripANSI(input)
	if result != "green bold normal red" {
		t.Errorf("expected 'green bold normal red', got %q", result)
	}
}

func TestVisualWidth_PlainString(t *testing.T) {
	if got := VisualWidth("hello"); got != 5 {
		t.Errorf("expected width 5, got %d", got)
	}
}

func TestVisualWidth_ANSIColored(t *testing.T) {
	colored := "\x1b[38;2;163;227;161mhello\x1b[0m"
	if got := VisualWidth(colored); got != 5 {
		t.Errorf("expected visual width 5 (ignoring ANSI), got %d", got)
	}
}

func TestVisualWidth_EmptyString(t *testing.T) {
	if got := VisualWidth(""); got != 0 {
		t.Errorf("expected width 0, got %d", got)
	}
}
