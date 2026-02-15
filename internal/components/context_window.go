package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// ContextWindow displays the context window usage percentage with color-coded
// thresholds: green (<50%), yellow (50-74%), red (75-89%), and red + warning (90%+).
type ContextWindow struct {
	renderer *render.Renderer
}

// NewContextWindow creates a new ContextWindow component.
func NewContextWindow(r *render.Renderer) *ContextWindow {
	return &ContextWindow{renderer: r}
}

// Name returns the component identifier.
func (c *ContextWindow) Name() string {
	return "context_window"
}

// Render produces the context window usage string, colored by severity.
func (c *ContextWindow) Render(in *input.StatusLineInput) string {
	pct := in.ContextWindow.UsedPercentage

	if pct == 0 {
		return ""
	}

	// Color based on percentage
	var colorFunc func(string) string
	warning := ""

	if pct >= 90 {
		colorFunc = c.renderer.Red
		warning = " \xe2\x9a\xa0\xef\xb8\x8f"
	} else if pct >= 75 {
		colorFunc = c.renderer.Red
	} else if pct >= 50 {
		colorFunc = c.renderer.Yellow
	} else {
		colorFunc = c.renderer.Green
	}

	// Format with tokens if available
	tokens := ""
	if in.ContextWindow.ContextWindowSize > 0 {
		used := float64(pct) / 100.0 * float64(in.ContextWindow.ContextWindowSize)
		tokens = fmt.Sprintf(" (%.0fK/%dK)",
			used/1000.0,
			in.ContextWindow.ContextWindowSize/1000,
		)
	}

	return fmt.Sprintf("\xf0\x9f\xa7\xa0 %s%s%s",
		colorFunc(fmt.Sprintf("%d%%", pct)),
		colorFunc(tokens),
		warning,
	)
}
