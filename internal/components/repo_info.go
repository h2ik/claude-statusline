package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/git"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// RepoInfo renders repository information for the status line: directory path
// (with ~ notation), git branch, clean/dirty status, and worktree indicator.
type RepoInfo struct {
	renderer *render.Renderer
	config   *config.Config
	icons    icons.IconSet
}

// NewRepoInfo creates a new RepoInfo component with the given renderer and config.
func NewRepoInfo(r *render.Renderer, cfg *config.Config, ic icons.IconSet) *RepoInfo {
	return &RepoInfo{renderer: r, config: cfg, icons: ic}
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

	pathStyle := c.config.GetString("repo_info", "path_style", "full")
	if pathStyle == "compress" {
		displayDir = compressPath(displayDir, in.Workspace.ProjectDir)
	}

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

	statusIcon := c.icons.Get(icons.CheckMark)
	statusColor := c.renderer.Green
	if !clean {
		statusIcon = c.icons.Get(icons.Folder)
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
		statusColor(statusIcon),
		wtIndicator,
	)
}

// compressPath applies Fish-style path compression. Intermediate directory
// segments above the repo root are shortened to their first character.
// The repo root name and any subdirectories within it are kept in full.
//
// Examples:
//
//	~/Projects/h2ik/claude-statusline/internal  → ~/P/h/claude-statusline/internal
//	~/Projects/h2ik/claude-statusline           → ~/P/h/claude-statusline
//	~/testdir (no projectDir)                   → ~/testdir
func compressPath(displayDir, projectDir string) string {
	// Determine the repo root name from projectDir.
	// projectDir is an absolute path like /home/user/Projects/h2ik/repo-name.
	repoRoot := ""
	if projectDir != "" {
		repoRoot = filepath.Base(projectDir)
	}

	parts := strings.Split(displayDir, "/")

	// Nothing to compress if we have fewer than 3 segments
	// (e.g., "~" + "dirname" = 2 segments — just ~/dirname)
	if len(parts) < 3 {
		return displayDir
	}

	// Find the repo root segment index
	repoIdx := -1
	if repoRoot != "" {
		for i, p := range parts {
			if p == repoRoot {
				repoIdx = i
				break
			}
		}
	}

	// If no repo root found, keep the last segment full and compress everything
	// in between (skip the first segment if it's ~)
	if repoIdx < 0 {
		repoIdx = len(parts) - 1
	}

	// Compress segments between the first segment and the repo root.
	// First segment (~ or root) stays as-is. Repo root and after stay full.
	for i := 1; i < repoIdx; i++ {
		if len(parts[i]) > 0 {
			// Use first rune to handle UTF-8 correctly
			runes := []rune(parts[i])
			parts[i] = string(runes[0])
		}
	}

	return strings.Join(parts, "/")
}
