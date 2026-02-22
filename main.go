package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/claude"
	"github.com/h2ik/claude-statusline/internal/component"
	"github.com/h2ik/claude-statusline/internal/components"
	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"

	"golang.org/x/term"
)

//nolint:unused // set via ldflags at build time by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Read JSON from stdin
	in, err := input.ParseInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse input: %v\n", err)
		os.Exit(1)
	}

	// Initialize infrastructure
	homeDir, _ := os.UserHomeDir()
	claudeSettings, _ := claude.LoadSettings(filepath.Join(homeDir, ".claude", "settings.json"))
	cacheDir := filepath.Join(homeDir, ".cache", "claude-statusline")
	costDir := filepath.Join(homeDir, ".claude", "statusline", "costs")
	projectsDir := filepath.Join(homeDir, ".claude", "projects")

	c := cache.New(cacheDir)
	_ = c.Prune(30 * 24 * time.Hour)
	r := render.New()
	h := cost.NewHistory(filepath.Join(costDir, "history.jsonl"))
	scanner := cost.NewTranscriptScanner(projectsDir, c)

	// Load configuration
	configPath := filepath.Join(homeDir, ".claude", "statusline", "config.toml")
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create icon set from config
	ic := icons.New(cfg.Layout.IconStyle)

	// Create registry and register components
	registry := component.NewRegistry()

	// Line 1 components
	registry.Register(components.NewRepoInfo(r, cfg, ic))

	// Line 2 components
	registry.Register(components.NewModelInfo(r, ic))
	registry.Register(components.NewBedrockModel(r, c, cfg, claudeSettings, ic))
	registry.Register(components.NewCommits(r, ic))
	registry.Register(components.NewSubmodules(r, ic))
	registry.Register(components.NewVersionInfo(r, c))
	registry.Register(components.NewTimeDisplay(r, ic))

	// Line 3 components
	registry.Register(components.NewCostMonthly(r, scanner, ic))
	registry.Register(components.NewCostWeekly(r, scanner, ic))
	registry.Register(components.NewCostDaily(r, scanner, ic))
	registry.Register(components.NewCostLive(r, h, ic))
	registry.Register(components.NewContextWindow(r, cfg, ic))
	registry.Register(components.NewSessionMode(r, ic))

	// Line 4 components
	registry.Register(components.NewBurnRate(r, ic))
	registry.Register(components.NewCacheEfficiency(r, ic))
	registry.Register(components.NewBlockProjection(r, ic))
	registry.Register(components.NewCodeProductivity(r, cfg, ic))

	// Select rendering style
	switch cfg.Layout.Style {
	case "powerline":
		r.SetStyle(render.NewPowerlineStyle(r))
	default:
		if cfg.Layout.Style != "" && cfg.Layout.Style != "default" {
			fmt.Fprintf(os.Stderr, "unknown style %q, falling back to default\n", cfg.Layout.Style)
		}
	}

	// Determine terminal width for right-side alignment.
	// When invoked as a subprocess (e.g. by Claude Code), all fds are pipes
	// so tty detection fails. Try multiple strategies:
	//   1. term.GetSize on stderr/stdout (works in a real terminal)
	//   2. /dev/tty (works for some subprocesses with a controlling terminal)
	//   3. $COLUMNS env var (user-configurable override)
	//   4. Default to 80
	termWidth := detectTerminalWidth() - cfg.Layout.Padding

	// Render each line
	var lineData []render.LineData
	for _, line := range cfg.Layout.Lines {
		leftNames, leftContent := registry.RenderNamedLine(in, line.Left)
		rightNames, rightContent := registry.RenderNamedLine(in, line.Right)
		if len(leftContent) > 0 || len(rightContent) > 0 {
			lineData = append(lineData, render.LineData{
				Left:       leftContent,
				LeftNames:  leftNames,
				Right:      rightContent,
				RightNames: rightNames,
			})
		}
	}

	output := r.RenderOutput(lineData, termWidth)
	_, _ = fmt.Fprint(os.Stdout, output)
}

// detectTerminalWidth tries multiple strategies to determine the terminal width.
func detectTerminalWidth() int {
	// Try stderr then stdout (stdin is consumed by JSON input)
	for _, f := range []*os.File{os.Stderr, os.Stdout} {
		if w, _, err := term.GetSize(int(f.Fd())); err == nil && w > 0 {
			return w
		}
	}

	// Try /dev/tty directly — works for some subprocesses with a controlling terminal
	if tty, err := os.Open("/dev/tty"); err == nil {
		w, _, err := term.GetSize(int(tty.Fd()))
		tty.Close()
		if err == nil && w > 0 {
			return w
		}
	}

	// Ask tmux for pane width — Claude Code runs inside a tmux pane
	// where all fds are pipes, but tmux knows the real dimensions
	if os.Getenv("TMUX") != "" {
		if out, err := exec.Command("tmux", "display-message", "-p", "#{pane_width}").Output(); err == nil {
			if n, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil && n > 0 {
				return n
			}
		}
	}

	// Fall back to COLUMNS env var
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			return n
		}
	}

	return 80
}
