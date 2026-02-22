package components

import (
	"fmt"
	"strings"

	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// ModelInfo renders the current Claude model name with an emoji indicator
// based on the model family (Opus, Sonnet, Haiku, or generic).
type ModelInfo struct {
	renderer *render.Renderer
	icons    icons.IconSet
}

// NewModelInfo creates a new ModelInfo component with the given renderer.
func NewModelInfo(r *render.Renderer, ic icons.IconSet) *ModelInfo {
	return &ModelInfo{renderer: r, icons: ic}
}

// Name returns the component identifier used for registry lookup.
func (c *ModelInfo) Name() string {
	return "model_info"
}

// Render produces the model info string from the given input.
// Returns empty for Bedrock ARN models (handled by BedrockModel component).
func (c *ModelInfo) Render(in *input.StatusLineInput) string {
	name := in.Model.DisplayName

	// Skip if this is a Bedrock ARN — let bedrock_model handle it
	if strings.Contains(name, "arn:") {
		return ""
	}

	if name == "" {
		name = "Claude"
	}

	icon := c.getIcon(name)

	return fmt.Sprintf("%s %s", icon, c.renderer.Teal(name))
}

// getIcon returns an icon based on the model family name.
func (c *ModelInfo) getIcon(name string) string {
	lower := strings.ToLower(name)

	switch {
	case strings.Contains(lower, "opus"):
		return c.icons.Get(icons.Brain)
	case strings.Contains(lower, "haiku"):
		return c.icons.Get(icons.Lightning)
	case strings.Contains(lower, "sonnet"):
		return c.icons.Get(icons.Music)
	default:
		return c.icons.Get(icons.Robot)
	}
}
