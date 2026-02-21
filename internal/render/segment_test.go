package render

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestSegmentCategory_AllComponentsMapped(t *testing.T) {
	known := []string{
		"repo_info", "model_info", "bedrock_model",
		"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "burn_rate",
		"context_window", "cache_efficiency", "block_projection",
		"code_productivity", "commits",
		"version_info", "session_mode",
		"time_display", "submodules",
	}

	for _, name := range known {
		cat := SegmentCategoryFor(name)
		if cat == (SegmentCategory{}) {
			t.Errorf("component %q has no segment category", name)
		}
	}
}

func TestSegmentCategory_UnknownComponentGetsDim(t *testing.T) {
	cat := SegmentCategoryFor("unknown_component")
	if cat.Background != ColorOverlay0 {
		t.Errorf("unknown component should use Overlay0 background, got %v", cat.Background)
	}
}

func TestSegmentCategory_InfoGroupIsBlue(t *testing.T) {
	for _, name := range []string{"repo_info", "model_info", "bedrock_model"} {
		cat := SegmentCategoryFor(name)
		if cat.Background != lipgloss.Color("#89b4fa") {
			t.Errorf("component %q should have blue background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_CostGroupIsPeach(t *testing.T) {
	for _, name := range []string{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "burn_rate"} {
		cat := SegmentCategoryFor(name)
		if cat.Background != lipgloss.Color("#fab387") {
			t.Errorf("component %q should have peach background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_MetricsGroupIsTeal(t *testing.T) {
	for _, name := range []string{"context_window", "cache_efficiency", "block_projection"} {
		cat := SegmentCategoryFor(name)
		if cat.Background != lipgloss.Color("#94e2d5") {
			t.Errorf("component %q should have teal background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_ActivityGroupIsGreen(t *testing.T) {
	for _, name := range []string{"code_productivity", "commits"} {
		cat := SegmentCategoryFor(name)
		if cat.Background != lipgloss.Color("#a6e3a1") {
			t.Errorf("component %q should have green background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_MetaGroupIsMauve(t *testing.T) {
	for _, name := range []string{"version_info", "session_mode"} {
		cat := SegmentCategoryFor(name)
		if cat.Background != lipgloss.Color("#cba6f7") {
			t.Errorf("component %q should have mauve background, got %v", name, cat.Background)
		}
	}
}
