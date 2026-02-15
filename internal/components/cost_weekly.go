package components

import (
	"fmt"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CostWeekly renders the rolling 7-day cost total from the JSONL history.
type CostWeekly struct {
	renderer *render.Renderer
	history  *cost.History
}

// NewCostWeekly creates a new CostWeekly component.
func NewCostWeekly(r *render.Renderer, h *cost.History) *CostWeekly {
	return &CostWeekly{renderer: r, history: h}
}

// Name returns the component identifier used for registry lookup.
func (c *CostWeekly) Name() string {
	return "cost_weekly"
}

// Render produces the 7-day cost summary string.
func (c *CostWeekly) Render(in *input.StatusLineInput) string {
	total, _ := c.history.CalculatePeriod(7 * 24 * time.Hour)

	return fmt.Sprintf("\xf0\x9f\x93\x8a %s $%.2f",
		c.renderer.Dimmed("7DAY"),
		total,
	)
}
