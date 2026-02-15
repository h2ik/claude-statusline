package components

import (
	"fmt"
	"strings"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// ModelInfo renders the current Claude model name with an emoji indicator
// based on the model family (Opus, Sonnet, Haiku, or generic).
type ModelInfo struct {
	renderer *render.Renderer
}

// NewModelInfo creates a new ModelInfo component with the given renderer.
func NewModelInfo(r *render.Renderer) *ModelInfo {
	return &ModelInfo{renderer: r}
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

	emoji := c.getEmoji(name)

	return fmt.Sprintf("%s %s", emoji, c.renderer.Teal(name))
}

// getEmoji returns an emoji based on the model family name.
func (c *ModelInfo) getEmoji(name string) string {
	lower := strings.ToLower(name)

	switch {
	case strings.Contains(lower, "opus"):
		return "\xf0\x9f\xa7\xa0" // brain
	case strings.Contains(lower, "haiku"):
		return "\xe2\x9a\xa1" // lightning
	case strings.Contains(lower, "sonnet"):
		return "\xf0\x9f\x8e\xb5" // musical note
	default:
		return "\xf0\x9f\xa4\x96" // robot
	}
}
