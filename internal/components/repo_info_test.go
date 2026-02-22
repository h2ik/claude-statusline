package components

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestRepoInfo_Render(t *testing.T) {
	r := render.New()
	cfg := config.DefaultConfig()
	c := NewRepoInfo(r, cfg, icons.New("emoji"))

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

func TestCompressPath(t *testing.T) {
	tests := []struct {
		name       string
		displayDir string
		projectDir string
		want       string
	}{
		{
			name:       "full path with repo root",
			displayDir: "~/Projects/h2ik/claude-statusline",
			projectDir: "/home/user/Projects/h2ik/claude-statusline",
			want:       "~/P/h/claude-statusline",
		},
		{
			name:       "path with subdirs inside repo",
			displayDir: "~/Projects/h2ik/claude-statusline/internal/components",
			projectDir: "/home/user/Projects/h2ik/claude-statusline",
			want:       "~/P/h/claude-statusline/internal/components",
		},
		{
			name:       "single dir under home",
			displayDir: "~/testdir",
			projectDir: "",
			want:       "~/testdir",
		},
		{
			name:       "just home",
			displayDir: "~",
			projectDir: "",
			want:       "~",
		},
		{
			name:       "no project dir with deep path",
			displayDir: "~/a/b/c/d",
			projectDir: "",
			want:       "~/a/b/c/d",
		},
		{
			name:       "no project dir with three segments",
			displayDir: "~/foo/bar",
			projectDir: "",
			want:       "~/f/bar",
		},
		{
			name:       "absolute path no tilde",
			displayDir: "/var/log/nginx",
			projectDir: "",
			want:       "/v/l/nginx",
		},
		{
			name:       "repo at first level under home",
			displayDir: "~/claude-statusline",
			projectDir: "/home/user/claude-statusline",
			want:       "~/claude-statusline",
		},
		{
			name:       "repo at first level with subdir",
			displayDir: "~/claude-statusline/internal",
			projectDir: "/home/user/claude-statusline",
			want:       "~/claude-statusline/internal",
		},
		{
			name:       "project dir matches repo root with deep nesting above",
			displayDir: "~/a/b/c/myrepo/src/main",
			projectDir: "/home/user/a/b/c/myrepo",
			want:       "~/a/b/c/myrepo/src/main",
		},
		{
			name:       "UTF-8 directory names",
			displayDir: "~/Projets/données/mon-repo",
			projectDir: "/home/user/Projets/données/mon-repo",
			want:       "~/P/d/mon-repo",
		},
		{
			name:       "empty display dir",
			displayDir: "",
			projectDir: "",
			want:       "",
		},
		{
			name:       "root only",
			displayDir: "/",
			projectDir: "",
			want:       "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compressPath(tt.displayDir, tt.projectDir)
			if got != tt.want {
				t.Errorf("compressPath(%q, %q) = %q, want %q",
					tt.displayDir, tt.projectDir, got, tt.want)
			}
		})
	}
}
