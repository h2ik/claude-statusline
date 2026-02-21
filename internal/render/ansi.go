package render

import "regexp"

// ansiEscape matches ANSI escape sequences including color codes, cursor movement, etc.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[mKJHfABCDsuhl]`)

// StripANSI removes all ANSI escape sequences from s, returning plain text.
func StripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// VisualWidth returns the number of visible characters in s, ignoring ANSI escape sequences.
// This is used to calculate padding widths for terminal-width-aware rendering.
// Note: this counts Unicode code points, not display columns. For ASCII-heavy statuslines
// this is accurate enough; emoji/CJK characters may cause slight misalignment.
func VisualWidth(s string) int {
	return len([]rune(StripANSI(s)))
}
