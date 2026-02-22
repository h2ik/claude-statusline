package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	registry.Register(components.NewRepoInfo(r, ic))

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

	// Determine terminal width for right-side alignment
	termWidth := 80
	if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 0 {
		termWidth = w
	}

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
