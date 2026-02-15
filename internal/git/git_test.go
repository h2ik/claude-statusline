package git

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

// gitCmd creates an exec.Command for git that is isolated from the user's
// global git configuration and hooks. This prevents things like commit-msg
// hooks (e.g. DCO sign-off) from interfering with test setup.
func gitCmd(dir string, args ...string) *exec.Cmd {
    cmd := exec.Command("git", args...)
    cmd.Dir = dir
    cmd.Env = append(os.Environ(),
        "GIT_CONFIG_GLOBAL=/dev/null",
        "GIT_CONFIG_SYSTEM=/dev/null",
        "GIT_TEMPLATE_DIR=",
    )
    return cmd
}

func setupGitRepo(t *testing.T) string {
    dir := t.TempDir()

    cmd := gitCmd(dir, "init")
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Fatalf("git init failed: %v\n%s", err, out)
    }

    // Configure git IN the temp repo (not globally)
    for _, args := range [][]string{
        {"config", "user.email", "test@test.com"},
        {"config", "user.name", "Test User"},
    } {
        c := gitCmd(dir, args...)
        if err := c.Run(); err != nil {
            t.Fatalf("git config %v failed: %v", args, err)
        }
    }

    testFile := filepath.Join(dir, "test.txt")
    if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
        t.Fatalf("write test file failed: %v", err)
    }

    c := gitCmd(dir, "add", ".")
    if err := c.Run(); err != nil {
        t.Fatalf("git add failed: %v", err)
    }

    c = gitCmd(dir, "commit", "--no-verify", "-m", "initial")
    if out, err := c.CombinedOutput(); err != nil {
        t.Fatalf("git commit failed: %v\n%s", err, out)
    }

    return dir
}

func TestGetBranch(t *testing.T) {
    dir := setupGitRepo(t)
    branch, err := GetBranch(dir)
    if err != nil {
        t.Fatalf("GetBranch failed: %v", err)
    }
    if branch != "main" && branch != "master" {
        t.Logf("got branch: %s (acceptable)", branch)
    }
}

func TestIsClean(t *testing.T) {
    dir := setupGitRepo(t)

    clean, err := IsClean(dir)
    if err != nil {
        t.Fatalf("IsClean failed: %v", err)
    }
    if !clean {
        t.Error("expected clean status")
    }

    testFile := filepath.Join(dir, "test.txt")
    os.WriteFile(testFile, []byte("modified"), 0644)

    clean, err = IsClean(dir)
    if err != nil {
        t.Fatalf("IsClean failed: %v", err)
    }
    if clean {
        t.Error("expected dirty status after modification")
    }
}

func TestIsGitRepo(t *testing.T) {
    dir := setupGitRepo(t)

    if !IsGitRepo(dir) {
        t.Error("expected true for a git repo")
    }

    nonRepo := t.TempDir()
    if IsGitRepo(nonRepo) {
        t.Error("expected false for a non-git directory")
    }
}

func TestGetCommitsToday(t *testing.T) {
    dir := setupGitRepo(t)

    count, err := GetCommitsToday(dir)
    if err != nil {
        t.Fatalf("GetCommitsToday failed: %v", err)
    }
    // We just made a commit, so count should be at least 1
    if count < 1 {
        t.Errorf("expected at least 1 commit today, got %d", count)
    }
}

func TestGetSubmoduleCount(t *testing.T) {
    dir := setupGitRepo(t)

    count, err := GetSubmoduleCount(dir)
    if err != nil {
        t.Fatalf("GetSubmoduleCount failed: %v", err)
    }
    if count != 0 {
        t.Errorf("expected 0 submodules, got %d", count)
    }
}

func TestIsWorktree(t *testing.T) {
    dir := setupGitRepo(t)

    isWt, _, err := IsWorktree(dir)
    if err != nil {
        t.Fatalf("IsWorktree failed: %v", err)
    }
    if isWt {
        t.Error("expected false for a non-worktree repo")
    }
}
