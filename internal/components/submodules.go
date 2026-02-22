package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/git"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// Submodules renders the count of git submodules in the current repo.
// Returns an empty string if the directory is not a git repo or has no submodules.
type Submodules struct {
	renderer *render.Renderer
	icons    icons.IconSet
}

// NewSubmodules creates a new Submodules component with the given renderer.
func NewSubmodules(r *render.Renderer, ic icons.IconSet) *Submodules {
	return &Submodules{renderer: r, icons: ic}
}

// Name returns the component identifier used for registry lookup.
func (c *Submodules) Name() string {
	return "submodules"
}

// Render produces the submodule count string from the given input.
func (c *Submodules) Render(in *input.StatusLineInput) string {
	count, err := git.GetSubmoduleCount(in.Workspace.CurrentDir)
	if err != nil || count == 0 {
		return ""
	}

	return fmt.Sprintf("%s %s%d",
		c.icons.Get(icons.Link),
		c.renderer.Dimmed("SUB:"),
		count,
	)
}
