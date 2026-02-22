package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CacheEfficiency displays the cache hit ratio as a percentage.
type CacheEfficiency struct {
	renderer *render.Renderer
	icons    icons.IconSet
}

// NewCacheEfficiency creates a new CacheEfficiency component.
func NewCacheEfficiency(r *render.Renderer, ic icons.IconSet) *CacheEfficiency {
	return &CacheEfficiency{renderer: r, icons: ic}
}

// Name returns the component identifier.
func (c *CacheEfficiency) Name() string {
	return "cache_efficiency"
}

// Render produces the cache efficiency string with color coding.
func (c *CacheEfficiency) Render(in *input.StatusLineInput) string {
	usage := in.CurrentUsage
	totalTokens := usage.InputTokens + usage.CacheReadInputTokens + usage.CacheCreationInputTokens

	if totalTokens == 0 {
		return ""
	}

	percentage := float64(usage.CacheReadInputTokens) / float64(totalTokens) * 100.0

	// Color based on efficiency
	var colorFunc func(string) string
	if percentage >= 70 {
		colorFunc = c.renderer.Green
	} else if percentage >= 40 {
		colorFunc = c.renderer.Yellow
	} else {
		colorFunc = c.renderer.Red
	}

	return fmt.Sprintf("%s %s",
		c.icons.Get(icons.FloppyDisk),
		colorFunc(fmt.Sprintf("%.0f%% cache", percentage)),
	)
}
