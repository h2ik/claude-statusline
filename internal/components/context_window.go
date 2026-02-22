package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// ContextWindow displays the context window usage percentage with color-coded
// thresholds: green (<50%), yellow (50-74%), red (75-89%), and red + warning (90%+).
type ContextWindow struct {
	renderer *render.Renderer
	config   *config.Config
	icons    icons.IconSet
}

// NewContextWindow creates a new ContextWindow component.
func NewContextWindow(r *render.Renderer, cfg *config.Config, ic icons.IconSet) *ContextWindow {
	return &ContextWindow{renderer: r, config: cfg, icons: ic}
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
		warning = " " + c.icons.Get(icons.Warning)
	} else if pct >= 75 {
		colorFunc = c.renderer.Red
	} else if pct >= 50 {
		colorFunc = c.renderer.Yellow
	} else {
		colorFunc = c.renderer.Green
	}

	// Format with tokens if available and configured
	tokens := ""
	showTokens := c.config.GetBool("context_window", "show_tokens", true)
	if showTokens && in.ContextWindow.ContextWindowSize > 0 {
		used := float64(pct) / 100.0 * float64(in.ContextWindow.ContextWindowSize)
		tokens = fmt.Sprintf(" (%.0fK/%dK)",
			used/1000.0,
			in.ContextWindow.ContextWindowSize/1000,
		)
	}

	return fmt.Sprintf("%s %s%s%s",
		c.icons.Get(icons.Brain),
		colorFunc(fmt.Sprintf("%d%%", pct)),
		colorFunc(tokens),
		warning,
	)
}
