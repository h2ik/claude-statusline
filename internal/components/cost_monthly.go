package components

import (
	"fmt"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CostMonthly renders the rolling 30-day cost total from the JSONL history.
type CostMonthly struct {
	renderer *render.Renderer
	history  *cost.History
}

// NewCostMonthly creates a new CostMonthly component.
func NewCostMonthly(r *render.Renderer, h *cost.History) *CostMonthly {
	return &CostMonthly{renderer: r, history: h}
}

// Name returns the component identifier used for registry lookup.
func (c *CostMonthly) Name() string {
	return "cost_monthly"
}

// Render produces the 30-day cost summary string.
func (c *CostMonthly) Render(in *input.StatusLineInput) string {
	total, _ := c.history.CalculatePeriod(30 * 24 * time.Hour)

	return fmt.Sprintf("\xf0\x9f\x93\x88 %s $%.2f",
		c.renderer.Dimmed("30DAY"),
		total,
	)
}
