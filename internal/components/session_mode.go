package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// SessionMode displays the current output style when it is not the default.
type SessionMode struct {
	renderer *render.Renderer
}

// NewSessionMode creates a new SessionMode component.
func NewSessionMode(r *render.Renderer) *SessionMode {
	return &SessionMode{renderer: r}
}

// Name returns the component identifier.
func (c *SessionMode) Name() string {
	return "session_mode"
}

// Render produces the session mode string with an appropriate emoji.
func (c *SessionMode) Render(in *input.StatusLineInput) string {
	style := in.OutputStyle.Name

	if style == "" || style == "default" {
		return ""
	}

	emoji := c.getEmoji(style)

	return fmt.Sprintf("%s %s %s",
		emoji,
		c.renderer.Dimmed("Style:"),
		c.renderer.Text(style),
	)
}

// getEmoji returns an emoji for the given style name.
func (c *SessionMode) getEmoji(style string) string {
	mapping := map[string]string{
		"explanatory": "\xf0\x9f\x93\x9a",
		"learning":    "\xf0\x9f\x8e\x93",
	}

	if emoji, ok := mapping[style]; ok {
		return emoji
	}

	return "\xe2\x9c\xa8"
}
