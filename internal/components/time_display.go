package components

import (
	"time"

	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// TimeDisplay renders the current time in HH:MM format with a clock emoji.
type TimeDisplay struct {
	renderer *render.Renderer
	icons    icons.IconSet
}

// NewTimeDisplay creates a new TimeDisplay component with the given renderer.
func NewTimeDisplay(r *render.Renderer, ic icons.IconSet) *TimeDisplay {
	return &TimeDisplay{renderer: r, icons: ic}
}

// Name returns the component identifier used for registry lookup.
func (c *TimeDisplay) Name() string {
	return "time_display"
}

// Render produces the current time string.
func (c *TimeDisplay) Render(in *input.StatusLineInput) string {
	now := time.Now().Format("15:04")
	return c.icons.Get(icons.Clock) + " " + c.renderer.Text(now)
}
