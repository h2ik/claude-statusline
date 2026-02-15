package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// BurnRate displays the current spending velocity in dollars per minute.
type BurnRate struct {
	renderer *render.Renderer
}

// NewBurnRate creates a new BurnRate component.
func NewBurnRate(r *render.Renderer) *BurnRate {
	return &BurnRate{renderer: r}
}

// Name returns the component identifier.
func (c *BurnRate) Name() string {
	return "burn_rate"
}

// Render produces the burn rate string.
func (c *BurnRate) Render(in *input.StatusLineInput) string {
	if in.Cost.TotalDurationMS == 0 {
		return ""
	}

	minutes := float64(in.Cost.TotalDurationMS) / 60000.0
	ratePerMin := in.Cost.TotalCostUSD / minutes

	return fmt.Sprintf("\xf0\x9f\x94\xa5 %s",
		c.renderer.Peach(fmt.Sprintf("$%.2f/min", ratePerMin)),
	)
}
