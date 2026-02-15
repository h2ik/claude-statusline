package components

import (
	"fmt"
	"os"
	"strings"

	"github.com/h2ik/claude-statusline/internal/git"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// RepoInfo renders repository information for the status line: directory path
// (with ~ notation), git branch, clean/dirty status, and worktree indicator.
type RepoInfo struct {
	renderer *render.Renderer
}

// NewRepoInfo creates a new RepoInfo component with the given renderer.
func NewRepoInfo(r *render.Renderer) *RepoInfo {
	return &RepoInfo{renderer: r}
}

// Name returns the component identifier used for registry lookup.
func (c *RepoInfo) Name() string {
	return "repo_info"
}

// Render produces the repo info string from the given input.
func (c *RepoInfo) Render(in *input.StatusLineInput) string {
	dir := in.Workspace.CurrentDir

	homeDir, _ := os.UserHomeDir()
	displayDir := strings.Replace(dir, homeDir, "~", 1)

	if !git.IsGitRepo(dir) {
		return c.renderer.Blue(displayDir)
	}

	branch, err := git.GetBranch(dir)
	if err != nil {
		return c.renderer.Blue(displayDir)
	}

	clean, err := git.IsClean(dir)
	if err != nil {
		clean = false
	}

	statusEmoji := "✅"
	statusColor := c.renderer.Green
	if !clean {
		statusEmoji = "📁"
		statusColor = c.renderer.Yellow
	}

	isWT, wtName, _ := git.IsWorktree(dir)
	wtIndicator := ""
	if isWT && wtName != "" {
		wtIndicator = c.renderer.Teal(fmt.Sprintf(" [WT:%s]", wtName))
	}

	return fmt.Sprintf("%s %s %s%s",
		c.renderer.Blue(displayDir),
		c.renderer.Mauve(fmt.Sprintf("(%s)", branch)),
		statusColor(statusEmoji),
		wtIndicator,
	)
}
