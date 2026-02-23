package render

import "github.com/charmbracelet/lipgloss"

// SegmentCategory defines the background and foreground colors for a powerline segment.
type SegmentCategory struct {
	Background lipgloss.Color
	Foreground lipgloss.Color
}

// componentGroup maps a component name to its semantic group.
// Returns "dim" for unknown components.
func componentGroup(name string) string {
	switch name {
	case "repo_info", "model_info", "bedrock_model":
		return "info"
	case "cost_monthly", "cost_weekly", "cost_daily", "cost_live", "burn_rate":
		return "cost"
	case "context_window", "cache_efficiency", "block_projection":
		return "metrics"
	case "code_productivity", "commits":
		return "activity"
	case "version_info", "session_mode":
		return "meta"
	default:
		return "dim"
	}
}

// SegmentCategoryFor returns the SegmentCategory for a given component name,
// using the colors from the provided theme. Unknown components fall back to Dim.
func SegmentCategoryFor(componentName string, theme *Theme) SegmentCategory {
	switch componentGroup(componentName) {
	case "info":
		return SegmentCategory{Background: theme.Blue, Foreground: theme.Base}
	case "cost":
		return SegmentCategory{Background: theme.Peach, Foreground: theme.Base}
	case "metrics":
		return SegmentCategory{Background: theme.Teal, Foreground: theme.Base}
	case "activity":
		return SegmentCategory{Background: theme.Green, Foreground: theme.Base}
	case "meta":
		return SegmentCategory{Background: theme.Mauve, Foreground: theme.Base}
	default: // dim
		return SegmentCategory{Background: theme.Overlay0, Foreground: theme.Text}
	}
}
