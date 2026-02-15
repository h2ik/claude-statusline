package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CodeProductivity displays code output metrics (velocity and/or cost per line).
type CodeProductivity struct {
	renderer *render.Renderer
	config   *config.Config
}

// NewCodeProductivity creates a new CodeProductivity component.
func NewCodeProductivity(r *render.Renderer, cfg *config.Config) *CodeProductivity {
	return &CodeProductivity{renderer: r, config: cfg}
}

// Name returns the component identifier.
func (c *CodeProductivity) Name() string {
	return "code_productivity"
}

// Render produces the code productivity string with configurable sub-metrics.
func (c *CodeProductivity) Render(in *input.StatusLineInput) string {
	totalLines := in.Cost.TotalLinesAdded + in.Cost.TotalLinesRemoved

	if totalLines == 0 {
		return ""
	}

	showVelocity := c.config.GetBool("code_productivity", "show_velocity", true)
	showCostPerLine := c.config.GetBool("code_productivity", "show_cost_per_line", true)

	var parts []string

	// Lines per minute
	if showVelocity && in.Cost.TotalDurationMS > 0 {
		minutes := float64(in.Cost.TotalDurationMS) / 60000.0
		linesPerMin := float64(totalLines) / minutes
		parts = append(parts, fmt.Sprintf("%.0f lines/min", linesPerMin))
	}

	// Cost per line
	if showCostPerLine {
		costPerLine := in.Cost.TotalCostUSD / float64(totalLines)
		parts = append(parts, fmt.Sprintf("$%.2f/line", costPerLine))
	}

	if len(parts) == 0 {
		return ""
	}

	output := "✏️ "
	for i, part := range parts {
		if i > 0 {
			output += " │ "
		}
		output += c.renderer.Text(part)
	}

	return output
}
