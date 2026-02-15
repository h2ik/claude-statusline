package components

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestRepoInfo_Render(t *testing.T) {
	r := render.New()
	c := NewRepoInfo(r)

	if c.Name() != "repo_info" {
		t.Errorf("expected 'repo_info', got %s", c.Name())
	}

	homeDir, _ := os.UserHomeDir()
	testDir := filepath.Join(homeDir, "testdir")

	in := &input.StatusLineInput{
		Workspace: input.Workspace{
			CurrentDir: testDir,
		},
	}

	output := c.Render(in)

	if !strings.Contains(output, "~/testdir") {
		t.Errorf("expected ~/testdir in output, got: %s", output)
	}
}
