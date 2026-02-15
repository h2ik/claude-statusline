package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// BlockProjection displays rate limit utilization from 5-hour and 7-day windows.
type BlockProjection struct {
	renderer *render.Renderer
}

// NewBlockProjection creates a new BlockProjection component.
func NewBlockProjection(r *render.Renderer) *BlockProjection {
	return &BlockProjection{renderer: r}
}

// Name returns the component identifier.
func (c *BlockProjection) Name() string {
	return "block_projection"
}

// Render produces the block projection string with color-coded utilization.
func (c *BlockProjection) Render(in *input.StatusLineInput) string {
	fiveHourPct := in.FiveHour.Utilization * 100.0
	sevenDayPct := in.SevenDay.Utilization * 100.0

	// Graceful degradation when no data (Bedrock case)
	if fiveHourPct == 0 && sevenDayPct == 0 {
		return ""
	}

	var parts []string

	if fiveHourPct > 0 {
		colorFunc := c.getColorForUtilization(fiveHourPct)
		parts = append(parts, colorFunc(fmt.Sprintf("5h: %.0f%%", fiveHourPct)))
	}

	if sevenDayPct > 0 {
		colorFunc := c.getColorForUtilization(sevenDayPct)
		parts = append(parts, colorFunc(fmt.Sprintf("7d: %.0f%%", sevenDayPct)))
	}

	if len(parts) == 0 {
		return ""
	}

	output := "⏳ "
	for i, part := range parts {
		if i > 0 {
			output += " │ "
		}
		output += part
	}

	return output
}

func (c *BlockProjection) getColorForUtilization(pct float64) func(string) string {
	if pct >= 75 {
		return c.renderer.Red
	} else if pct >= 50 {
		return c.renderer.Yellow
	}
	return c.renderer.Green
}
