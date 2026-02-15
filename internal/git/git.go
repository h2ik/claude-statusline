package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetBranch returns the current branch name for the git repo at dir.
func GetBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// IsClean returns true if the working tree has no uncommitted changes.
func IsClean(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return len(strings.TrimSpace(string(output))) == 0, nil
}

// IsGitRepo returns true if dir is inside a git repository.
func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// GetCommitsToday returns the number of commits made today on the current branch.
func GetCommitsToday(dir string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", "--since=today 00:00", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("git rev-list failed: %w", err)
	}
	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count); err != nil {
		return 0, fmt.Errorf("parse count failed: %w", err)
	}
	return count, nil
}

// GetSubmoduleCount returns the number of git submodules in the repo.
func GetSubmoduleCount(dir string) (int, error) {
	cmd := exec.Command("git", "submodule", "status")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return 0, nil
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}
	return len(lines), nil
}

// IsWorktree returns whether the repo at dir is a git worktree, and if so, the worktree name.
func IsWorktree(dir string) (bool, string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return false, "", err
	}
	gitDir := strings.TrimSpace(string(output))
	if strings.Contains(gitDir, ".git/worktrees/") {
		parts := strings.Split(gitDir, "/")
		for i, part := range parts {
			if part == "worktrees" && i+1 < len(parts) {
				return true, parts[i+1], nil
			}
		}
		return true, "", nil
	}
	return false, "", nil
}
