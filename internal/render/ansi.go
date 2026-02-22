package render

import (
	"regexp"

	"github.com/mattn/go-runewidth"
)

// ansiEscape matches ANSI escape sequences including color codes, cursor movement, etc.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[mKJHfABCDsuhl]`)

// StripANSI removes all ANSI escape sequences from s, returning plain text.
func StripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// VisualWidth returns the number of display columns s occupies, ignoring ANSI escape sequences.
// This correctly handles double-width characters (emoji, CJK) and zero-width joiners.
func VisualWidth(s string) int {
	return runewidth.StringWidth(StripANSI(s))
}
