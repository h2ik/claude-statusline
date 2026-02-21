package render

import "github.com/charmbracelet/lipgloss"

// colorDark is the Catppuccin Mocha Base color, used as foreground on bright backgrounds.
var colorDark = lipgloss.Color("#1e1e2e")

// SegmentCategory defines the background and foreground colors for a powerline segment.
type SegmentCategory struct {
	Background lipgloss.Color
	Foreground lipgloss.Color
}

// segmentCategories maps component names to their semantic color category.
var segmentCategories = map[string]SegmentCategory{
	// Info: repo, model identity
	"repo_info":     {Background: lipgloss.Color("#89b4fa"), Foreground: colorDark},
	"model_info":    {Background: lipgloss.Color("#89b4fa"), Foreground: colorDark},
	"bedrock_model": {Background: lipgloss.Color("#89b4fa"), Foreground: colorDark},

	// Cost: all spending-related
	"cost_monthly": {Background: lipgloss.Color("#fab387"), Foreground: colorDark},
	"cost_weekly":  {Background: lipgloss.Color("#fab387"), Foreground: colorDark},
	"cost_daily":   {Background: lipgloss.Color("#fab387"), Foreground: colorDark},
	"cost_live":    {Background: lipgloss.Color("#fab387"), Foreground: colorDark},
	"burn_rate":    {Background: lipgloss.Color("#fab387"), Foreground: colorDark},

	// Metrics: utilization and efficiency
	"context_window":   {Background: lipgloss.Color("#94e2d5"), Foreground: colorDark},
	"cache_efficiency": {Background: lipgloss.Color("#94e2d5"), Foreground: colorDark},
	"block_projection": {Background: lipgloss.Color("#94e2d5"), Foreground: colorDark},

	// Activity: productivity and commit output
	"code_productivity": {Background: lipgloss.Color("#a6e3a1"), Foreground: colorDark},
	"commits":           {Background: lipgloss.Color("#a6e3a1"), Foreground: colorDark},

	// Meta: version and session info
	"version_info": {Background: lipgloss.Color("#cba6f7"), Foreground: colorDark},
	"session_mode": {Background: lipgloss.Color("#cba6f7"), Foreground: colorDark},

	// Dim: time and submodule count
	"time_display": {Background: ColorOverlay0, Foreground: ColorText},
	"submodules":   {Background: ColorOverlay0, Foreground: ColorText},
}

// SegmentCategoryFor returns the SegmentCategory for a given component name.
// Unknown components fall back to the Dim category.
func SegmentCategoryFor(componentName string) SegmentCategory {
	if cat, ok := segmentCategories[componentName]; ok {
		return cat
	}
	return SegmentCategory{Background: ColorOverlay0, Foreground: ColorText}
}
