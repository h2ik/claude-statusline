package render

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Catppuccin Mocha colors (hardcoded)
var (
	ColorOverlay0 = lipgloss.Color("#6c7086") // Dimmed/labels
	ColorText     = lipgloss.Color("#cdd6f4") // Text/values
	ColorGreen    = lipgloss.Color("#a6e3a1") // Clean/good
	ColorRed      = lipgloss.Color("#f38ba8") // Critical
	ColorYellow   = lipgloss.Color("#f9e2af") // Warning
	ColorBlue     = lipgloss.Color("#89b4fa") // Paths/info
	ColorMauve    = lipgloss.Color("#cba6f7") // Accent
	ColorPeach    = lipgloss.Color("#fab387") // Costs
	ColorTeal     = lipgloss.Color("#94e2d5") // Secondary
)

// Renderer handles styling and layout of statusline components.
type Renderer struct {
	separator string
	lg        *lipgloss.Renderer
	style     Style
}

// New creates a Renderer with forced TrueColor output.
// Claude Code captures stdout so lipgloss won't auto-detect a TTY;
// we force color output with termenv.WithUnsafe().
func New() *Renderer {
	lg := lipgloss.NewRenderer(
		os.Stdout,
		termenv.WithUnsafe(),
		termenv.WithProfile(termenv.TrueColor),
	)

	return &Renderer{
		separator: " │ ",
		lg:        lg,
		style:     NewDefaultStyle(" │ "),
	}
}

// SetStyle replaces the active rendering style (e.g. DefaultStyle, PowerlineStyle).
func (r *Renderer) SetStyle(s Style) {
	r.style = s
}

// RenderOutput renders a slice of LineData through the active Style, filtering
// out lines that produce only whitespace. termWidth is passed to the Style so
// powerline-mode can pad/align.
func (r *Renderer) RenderOutput(lines []LineData, termWidth int) string {
	var output []string
	for _, line := range lines {
		rendered := r.style.RenderLine(line, termWidth)
		if strings.TrimSpace(rendered) != "" {
			output = append(output, rendered)
		}
	}
	return strings.Join(output, "\n")
}

// RenderLines joins components per line with separators, filtering out empty components
// and lines that contain only empty components.
// This is a backward-compatible wrapper around RenderOutput.
func (r *Renderer) RenderLines(lines [][]string) string {
	var data []LineData
	for _, line := range lines {
		var nonEmpty []string
		for _, c := range line {
			if strings.TrimSpace(c) != "" {
				nonEmpty = append(nonEmpty, c)
			}
		}
		if len(nonEmpty) > 0 {
			data = append(data, LineData{Left: nonEmpty})
		}
	}
	return r.RenderOutput(data, 80)
}

// Style helpers -- each wraps the input string with a Catppuccin Mocha color.

// Dimmed renders text in Overlay0 (dimmed/label color).
func (r *Renderer) Dimmed(s string) string {
	return r.lg.NewStyle().Foreground(ColorOverlay0).Render(s)
}

// Text renders text in the default text color.
func (r *Renderer) Text(s string) string {
	return r.lg.NewStyle().Foreground(ColorText).Render(s)
}

// Green renders text in green (clean/good status).
func (r *Renderer) Green(s string) string {
	return r.lg.NewStyle().Foreground(ColorGreen).Render(s)
}

// Red renders text in red (critical status).
func (r *Renderer) Red(s string) string {
	return r.lg.NewStyle().Foreground(ColorRed).Render(s)
}

// Yellow renders text in yellow (warning status).
func (r *Renderer) Yellow(s string) string {
	return r.lg.NewStyle().Foreground(ColorYellow).Render(s)
}

// Blue renders text in blue (paths/info).
func (r *Renderer) Blue(s string) string {
	return r.lg.NewStyle().Foreground(ColorBlue).Render(s)
}

// Mauve renders text in mauve (accent color).
func (r *Renderer) Mauve(s string) string {
	return r.lg.NewStyle().Foreground(ColorMauve).Render(s)
}

// Peach renders text in peach (cost-related).
func (r *Renderer) Peach(s string) string {
	return r.lg.NewStyle().Foreground(ColorPeach).Render(s)
}

// Teal renders text in teal (secondary info).
func (r *Renderer) Teal(s string) string {
	return r.lg.NewStyle().Foreground(ColorTeal).Render(s)
}
