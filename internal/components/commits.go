package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/git"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// Commits renders the number of git commits made today on the current branch.
// Returns an empty string if the directory is not a git repo or there are no
// commits today.
type Commits struct {
	renderer *render.Renderer
}

// NewCommits creates a new Commits component with the given renderer.
func NewCommits(r *render.Renderer) *Commits {
	return &Commits{renderer: r}
}

// Name returns the component identifier used for registry lookup.
func (c *Commits) Name() string {
	return "commits"
}

// Render produces the commits-today string from the given input.
func (c *Commits) Render(in *input.StatusLineInput) string {
	count, err := git.GetCommitsToday(in.Workspace.CurrentDir)
	if err != nil || count == 0 {
		return ""
	}

	return fmt.Sprintf("\xf0\x9f\x92\xbe %s %d",
		c.renderer.Dimmed("Commits:"),
		count,
	)
}
