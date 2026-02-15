package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/component"
	"github.com/h2ik/claude-statusline/internal/components"
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

	c := cache.New(cacheDir)
	r := render.New()
	h := cost.NewHistory(filepath.Join(costDir, "history.jsonl"))

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
	registry.Register(components.NewCostMonthly(r, h))
	registry.Register(components.NewCostWeekly(r, h))
	registry.Register(components.NewCostDaily(r, h))
	registry.Register(components.NewCostLive(r, h))
	registry.Register(components.NewContextWindow(r))
	registry.Register(components.NewSessionMode(r))

	// Define line layout
	lines := [][]string{
		{"repo_info"},
		{"bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"},
		{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"},
	}

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
