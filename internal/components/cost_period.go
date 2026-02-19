package components

import (
	"fmt"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CostPeriod renders a rolling cost total by scanning Claude Code's
// native JSONL transcript files for a configurable time window.
type CostPeriod struct {
	renderer *render.Renderer
	scanner  *cost.TranscriptScanner
	name     string
	label    string
	emoji    string
	duration time.Duration
}

func (c *CostPeriod) Name() string {
	return c.name
}

func (c *CostPeriod) Render(in *input.StatusLineInput) string {
	total := c.scanner.CalculatePeriod(c.duration)

	return fmt.Sprintf("%s %s $%.2f",
		c.emoji,
		c.renderer.Dimmed(c.label),
		total,
	)
}

// NewCostMonthly creates a 30-day rolling cost component.
func NewCostMonthly(r *render.Renderer, s *cost.TranscriptScanner) *CostPeriod {
	return &CostPeriod{
		renderer: r,
		scanner:  s,
		name:     "cost_monthly",
		label:    "30DAY",
		emoji:    "📈",
		duration: 30 * 24 * time.Hour,
	}
}

// NewCostWeekly creates a 7-day rolling cost component.
func NewCostWeekly(r *render.Renderer, s *cost.TranscriptScanner) *CostPeriod {
	return &CostPeriod{
		renderer: r,
		scanner:  s,
		name:     "cost_weekly",
		label:    "7DAY",
		emoji:    "📊",
		duration: 7 * 24 * time.Hour,
	}
}

// CostToday renders the cost since midnight local time by scanning
// Claude Code's native JSONL transcript files.
type CostToday struct {
	renderer *render.Renderer
	scanner  *cost.TranscriptScanner
}

func (c *CostToday) Name() string {
	return "cost_daily"
}

func (c *CostToday) Render(in *input.StatusLineInput) string {
	total := c.scanner.CalculateToday()

	return fmt.Sprintf("%s %s $%.2f",
		"📅",
		c.renderer.Dimmed("TODAY"),
		total,
	)
}

// NewCostDaily creates a component showing cost since midnight local time.
func NewCostDaily(r *render.Renderer, s *cost.TranscriptScanner) *CostToday {
	return &CostToday{
		renderer: r,
		scanner:  s,
	}
}
