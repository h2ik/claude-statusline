package components

import (
	"time"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// TimeDisplay renders the current time in HH:MM format with a clock emoji.
type TimeDisplay struct {
	renderer *render.Renderer
}

// NewTimeDisplay creates a new TimeDisplay component with the given renderer.
func NewTimeDisplay(r *render.Renderer) *TimeDisplay {
	return &TimeDisplay{renderer: r}
}

// Name returns the component identifier used for registry lookup.
func (c *TimeDisplay) Name() string {
	return "time_display"
}

// Render produces the current time string.
func (c *TimeDisplay) Render(in *input.StatusLineInput) string {
	now := time.Now().Format("15:04")
	return "\xf0\x9f\x95\x90 " + c.renderer.Text(now)
}
