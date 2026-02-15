package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/git"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// Submodules renders the count of git submodules in the current repo.
// Returns an empty string if the directory is not a git repo or has no submodules.
type Submodules struct {
	renderer *render.Renderer
}

// NewSubmodules creates a new Submodules component with the given renderer.
func NewSubmodules(r *render.Renderer) *Submodules {
	return &Submodules{renderer: r}
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

	return fmt.Sprintf("\xf0\x9f\x94\x97 %s%d",
		c.renderer.Dimmed("SUB:"),
		count,
	)
}
