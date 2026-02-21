package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/claude"
	"github.com/h2ik/claude-statusline/internal/component"
	"github.com/h2ik/claude-statusline/internal/components"
	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
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

	// Create registry and register components
	registry := component.NewRegistry()

	// Line 1 components
	registry.Register(components.NewRepoInfo(r))

	// Line 2 components
	registry.Register(components.NewModelInfo(r))
	registry.Register(components.NewBedrockModel(r, c, cfg, claudeSettings))
	registry.Register(components.NewCommits(r))
	registry.Register(components.NewSubmodules(r))
	registry.Register(components.NewVersionInfo(r, c))
	registry.Register(components.NewTimeDisplay(r))

	// Line 3 components
	registry.Register(components.NewCostMonthly(r, scanner))
	registry.Register(components.NewCostWeekly(r, scanner))
	registry.Register(components.NewCostDaily(r, scanner))
	registry.Register(components.NewCostLive(r, h))
	registry.Register(components.NewContextWindow(r, cfg))
	registry.Register(components.NewSessionMode(r))

	// Line 4 components
	registry.Register(components.NewBurnRate(r))
	registry.Register(components.NewCacheEfficiency(r))
	registry.Register(components.NewBlockProjection(r))
	registry.Register(components.NewCodeProductivity(r, cfg))

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
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			termWidth = n
		}
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
