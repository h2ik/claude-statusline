package components

import (
	"fmt"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CostLive renders the current session's live cost and appends the cost entry
// to the JSONL history for period-based components to read.
type CostLive struct {
	renderer *render.Renderer
	history  *cost.History
	icons    icons.IconSet
}

// NewCostLive creates a new CostLive component.
func NewCostLive(r *render.Renderer, h *cost.History, ic icons.IconSet) *CostLive {
	return &CostLive{renderer: r, history: h, icons: ic}
}

// Name returns the component identifier used for registry lookup.
func (c *CostLive) Name() string {
	return "cost_live"
}

// Render produces the live session cost string and appends to history.
func (c *CostLive) Render(in *input.StatusLineInput) string {
	// Append current session cost to history
	if in.SessionID != "" && in.Cost.TotalCostUSD > 0 {
		entry := cost.Entry{
			SessionID: in.SessionID,
			Cost:      in.Cost.TotalCostUSD,
			Timestamp: time.Now(),
		}
		_ = c.history.Append(entry)
	}

	// Display live session cost
	return fmt.Sprintf("%s %s $%.2f",
		c.icons.Get(icons.Fire),
		c.renderer.Dimmed("LIVE"),
		in.Cost.TotalCostUSD,
	)
}
