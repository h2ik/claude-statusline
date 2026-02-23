package render

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Renderer handles styling and layout of statusline components.
type Renderer struct {
	separator string
	lg        *lipgloss.Renderer
	style     Style
	theme     Theme
}

// New creates a Renderer with forced TrueColor output.
// Claude Code captures stdout so lipgloss won't auto-detect a TTY;
// we force color output with termenv.WithUnsafe().
// If theme is nil, ThemeMocha is used.
func New(theme *Theme) *Renderer {
	t := ThemeMocha
	if theme != nil {
		t = *theme
	}

	lg := lipgloss.NewRenderer(
		os.Stdout,
		termenv.WithUnsafe(),
		termenv.WithProfile(termenv.TrueColor),
	)
	lg.SetColorProfile(termenv.TrueColor)

	return &Renderer{
		separator: " │ ",
		lg:        lg,
		style:     NewDefaultStyle(" │ "),
		theme:     t,
	}
}

// SetStyle replaces the active rendering style (e.g. DefaultStyle, PowerlineStyle).
func (r *Renderer) SetStyle(s Style) {
	r.style = s
}

// Theme returns the active theme.
func (r *Renderer) Theme() Theme {
	return r.theme
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

// Style helpers -- each wraps the input string with the active theme's color.

func (r *Renderer) Dimmed(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Overlay0).Render(s)
}

func (r *Renderer) Text(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Text).Render(s)
}

func (r *Renderer) Green(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Green).Render(s)
}

func (r *Renderer) Red(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Red).Render(s)
}

func (r *Renderer) Yellow(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Yellow).Render(s)
}

func (r *Renderer) Blue(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Blue).Render(s)
}

func (r *Renderer) Mauve(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Mauve).Render(s)
}

func (r *Renderer) Peach(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Peach).Render(s)
}

func (r *Renderer) Teal(s string) string {
	return r.lg.NewStyle().Foreground(r.theme.Teal).Render(s)
}
