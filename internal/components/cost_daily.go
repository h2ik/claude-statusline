package components

import (
	"fmt"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CostDaily renders the rolling 24-hour cost total from the JSONL history.
type CostDaily struct {
	renderer *render.Renderer
	history  *cost.History
}

// NewCostDaily creates a new CostDaily component.
func NewCostDaily(r *render.Renderer, h *cost.History) *CostDaily {
	return &CostDaily{renderer: r, history: h}
}

// Name returns the component identifier used for registry lookup.
func (c *CostDaily) Name() string {
	return "cost_daily"
}

// Render produces the 24-hour cost summary string.
func (c *CostDaily) Render(in *input.StatusLineInput) string {
	total, _ := c.history.CalculatePeriod(24 * time.Hour)

	return fmt.Sprintf("\xf0\x9f\x93\x85 %s $%.2f",
		c.renderer.Dimmed("DAY"),
		total,
	)
}
