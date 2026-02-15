package components

import (
	"fmt"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CostPeriod renders a rolling cost total for a configurable time window.
// Used for monthly, weekly, and daily cost displays.
type CostPeriod struct {
	renderer *render.Renderer
	history  *cost.History
	name     string
	label    string
	emoji    string
	duration time.Duration
}

func (c *CostPeriod) Name() string {
	return c.name
}

func (c *CostPeriod) Render(in *input.StatusLineInput) string {
	total, _ := c.history.CalculatePeriod(c.duration)

	return fmt.Sprintf("%s %s $%.2f",
		c.emoji,
		c.renderer.Dimmed(c.label),
		total,
	)
}

// NewCostMonthly creates a 30-day rolling cost component.
func NewCostMonthly(r *render.Renderer, h *cost.History) *CostPeriod {
	return &CostPeriod{
		renderer: r,
		history:  h,
		name:     "cost_monthly",
		label:    "30DAY",
		emoji:    "📈",
		duration: 30 * 24 * time.Hour,
	}
}

// NewCostWeekly creates a 7-day rolling cost component.
func NewCostWeekly(r *render.Renderer, h *cost.History) *CostPeriod {
	return &CostPeriod{
		renderer: r,
		history:  h,
		name:     "cost_weekly",
		label:    "7DAY",
		emoji:    "📊",
		duration: 7 * 24 * time.Hour,
	}
}

// NewCostDaily creates a 24-hour rolling cost component.
func NewCostDaily(r *render.Renderer, h *cost.History) *CostPeriod {
	return &CostPeriod{
		renderer: r,
		history:  h,
		name:     "cost_daily",
		label:    "DAY",
		emoji:    "📅",
		duration: 24 * time.Hour,
	}
}
