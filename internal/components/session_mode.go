package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// SessionMode displays the current output style when it is not the default.
type SessionMode struct {
	renderer *render.Renderer
	icons    icons.IconSet
}

// NewSessionMode creates a new SessionMode component.
func NewSessionMode(r *render.Renderer, ic icons.IconSet) *SessionMode {
	return &SessionMode{renderer: r, icons: ic}
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

	icon := c.getIcon(style)

	return fmt.Sprintf("%s %s %s",
		icon,
		c.renderer.Dimmed("Style:"),
		c.renderer.Text(style),
	)
}

// getIcon returns an icon for the given style name.
func (c *SessionMode) getIcon(style string) string {
	mapping := map[string]string{
		"explanatory": icons.Book,
		"learning":    icons.Graduation,
	}

	if iconName, ok := mapping[style]; ok {
		return c.icons.Get(iconName)
	}

	return c.icons.Get(icons.Sparkles)
}
