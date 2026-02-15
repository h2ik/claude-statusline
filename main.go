package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/component"
	"github.com/h2ik/claude-statusline/internal/components"
	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
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
	cacheDir := filepath.Join(homeDir, ".cache", "claude-statusline")
	costDir := filepath.Join(homeDir, ".claude", "statusline", "costs")
	projectsDir := filepath.Join(homeDir, ".claude", "projects")

	c := cache.New(cacheDir)
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
	registry.Register(components.NewBedrockModel(r, c))
	registry.Register(components.NewCommits(r))
	registry.Register(components.NewSubmodules(r))
	registry.Register(components.NewVersionInfo(r, c))
	registry.Register(components.NewTimeDisplay(r))

	// Line 3 components
	registry.Register(components.NewCostMonthly(r, scanner))
	registry.Register(components.NewCostWeekly(r, scanner))
	registry.Register(components.NewCostDaily(r, scanner))
	registry.Register(components.NewCostLive(r, h))
	registry.Register(components.NewContextWindow(r))
	registry.Register(components.NewSessionMode(r))

	// Use config-defined layout
	lines := cfg.Layout.Lines

	// Render each line
	var renderedLines [][]string
	for _, lineComponents := range lines {
		rendered := registry.RenderLine(in, lineComponents)
		if len(rendered) > 0 {
			renderedLines = append(renderedLines, rendered)
		}
	}

	// Output final result
	output := r.RenderLines(renderedLines)
	fmt.Fprint(os.Stdout, output)
}
