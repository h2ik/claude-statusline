package components

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// VersionInfo renders the installed Claude CLI version, cached for 15 minutes
// to avoid shelling out on every render cycle.
type VersionInfo struct {
	renderer *render.Renderer
	cache    *cache.Cache
}

// NewVersionInfo creates a new VersionInfo component with the given renderer and cache.
func NewVersionInfo(r *render.Renderer, c *cache.Cache) *VersionInfo {
	return &VersionInfo{renderer: r, cache: c}
}

// Name returns the component identifier used for registry lookup.
func (c *VersionInfo) Name() string {
	return "version_info"
}

// Render produces the version info string from the given input.
func (c *VersionInfo) Render(in *input.StatusLineInput) string {
	version := c.getClaudeVersion()
	if version == "" {
		return ""
	}

	return fmt.Sprintf("%s%s",
		c.renderer.Dimmed("CC:"),
		c.renderer.Text(version),
	)
}

// getClaudeVersion returns the installed Claude CLI version, using cache when available.
func (c *VersionInfo) getClaudeVersion() string {
	// Check cache
	cached, err := c.cache.Get("claude-version", 15*time.Minute)
	if err == nil {
		return string(cached)
	}

	// Run claude --version
	cmd := exec.Command("claude", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "claude ")
	version = strings.TrimPrefix(version, "v")

	// Cache it
	_ = c.cache.Set("claude-version", []byte(version), 15*time.Minute)

	return version
}
