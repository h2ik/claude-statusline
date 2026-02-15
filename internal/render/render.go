package render

import (
    "strings"

    "github.com/charmbracelet/lipgloss"
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
}

// New creates a Renderer with default settings.
func New() *Renderer {
    return &Renderer{
        separator: " │ ",
    }
}

// RenderLines joins components per line with separators, filtering out empty components
// and lines that contain only empty components.
func (r *Renderer) RenderLines(lines [][]string) string {
    var output []string

    for _, components := range lines {
        var nonEmpty []string
        for _, c := range components {
            if strings.TrimSpace(c) != "" {
                nonEmpty = append(nonEmpty, c)
            }
        }

        if len(nonEmpty) > 0 {
            line := strings.Join(nonEmpty, r.separator)
            output = append(output, line)
        }
    }

    return strings.Join(output, "\n")
}

// Style helpers -- each wraps the input string with a Catppuccin Mocha color.

// Dimmed renders text in Overlay0 (dimmed/label color).
func (r *Renderer) Dimmed(s string) string {
    return lipgloss.NewStyle().Foreground(ColorOverlay0).Render(s)
}

// Text renders text in the default text color.
func (r *Renderer) Text(s string) string {
    return lipgloss.NewStyle().Foreground(ColorText).Render(s)
}

// Green renders text in green (clean/good status).
func (r *Renderer) Green(s string) string {
    return lipgloss.NewStyle().Foreground(ColorGreen).Render(s)
}

// Red renders text in red (critical status).
func (r *Renderer) Red(s string) string {
    return lipgloss.NewStyle().Foreground(ColorRed).Render(s)
}

// Yellow renders text in yellow (warning status).
func (r *Renderer) Yellow(s string) string {
    return lipgloss.NewStyle().Foreground(ColorYellow).Render(s)
}

// Blue renders text in blue (paths/info).
func (r *Renderer) Blue(s string) string {
    return lipgloss.NewStyle().Foreground(ColorBlue).Render(s)
}

// Mauve renders text in mauve (accent color).
func (r *Renderer) Mauve(s string) string {
    return lipgloss.NewStyle().Foreground(ColorMauve).Render(s)
}

// Peach renders text in peach (cost-related).
func (r *Renderer) Peach(s string) string {
    return lipgloss.NewStyle().Foreground(ColorPeach).Render(s)
}

// Teal renders text in teal (secondary info).
func (r *Renderer) Teal(s string) string {
    return lipgloss.NewStyle().Foreground(ColorTeal).Render(s)
}
