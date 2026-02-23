package render

import (
	"strings"
	"testing"
)

func TestPowerlineStyle_RenderLine_LeftOnly(t *testing.T) {
	s := NewPowerlineStyle(New(nil))
	line := LineData{
		Left:      []string{"~/projects", "Claude"},
		LeftNames: []string{"repo_info", "model_info"},
	}
	result := s.RenderLine(line, 80)
	if result == "" {
		t.Error("expected non-empty output")
	}
	// repo_info and model_info are both "Info" category (blue) -- they merge, NO arrow between them
	// But there IS a trailing arrow after the last segment
	if !strings.Contains(result, "\ue0b0") {
		t.Error("expected trailing powerline arrow")
	}
}

func TestPowerlineStyle_RenderLine_DifferentCategories(t *testing.T) {
	s := NewPowerlineStyle(New(nil))
	// repo_info is Info (blue), cost_daily is Cost (peach) -- different categories, should have arrow between
	line := LineData{
		Left:      []string{"~/projects", "$0.89"},
		LeftNames: []string{"repo_info", "cost_daily"},
	}
	result := s.RenderLine(line, 80)
	// Should have arrows between the two different-category segments PLUS trailing arrow = at least 2
	arrowCount := strings.Count(result, "\ue0b0")
	if arrowCount < 2 {
		t.Errorf("expected at least 2 forward arrows for different categories + trailing, got %d", arrowCount)
	}
}

func TestPowerlineStyle_RenderLine_RightOnly(t *testing.T) {
	s := NewPowerlineStyle(New(nil))
	// time_display is Dim, version_info is Meta -- different categories
	line := LineData{
		Right:      []string{"14:32", "CC:0.3.2"},
		RightNames: []string{"time_display", "version_info"},
	}
	result := s.RenderLine(line, 80)
	if !strings.Contains(result, "\ue0b2") {
		t.Error("expected reverse powerline arrow for right-side segments")
	}
}

func TestPowerlineStyle_RenderLine_EmptyLine(t *testing.T) {
	s := NewPowerlineStyle(New(nil))
	line := LineData{}
	result := s.RenderLine(line, 80)
	if result != "" {
		t.Errorf("expected empty string for empty line, got %q", result)
	}
}

func TestPowerlineStyle_RenderLine_FiltersEmptyComponents(t *testing.T) {
	s := NewPowerlineStyle(New(nil))
	line := LineData{
		Left:      []string{"~/projects", "", "5 commits"},
		LeftNames: []string{"repo_info", "model_info", "commits"},
	}
	result := s.RenderLine(line, 80)
	if result == "" {
		t.Error("expected non-empty output")
	}
}

func TestPowerlineStyle_RenderLine_SameCategoryMerges(t *testing.T) {
	s := NewPowerlineStyle(New(nil))
	// cost_daily and cost_live are both Cost (peach) -- should merge into one segment
	line := LineData{
		Left:      []string{"$0.89", "$2.47"},
		LeftNames: []string{"cost_daily", "cost_live"},
	}
	result := s.RenderLine(line, 80)
	// Only 1 trailing arrow (no arrow BETWEEN same-category segments)
	arrowCount := strings.Count(result, "\ue0b0")
	if arrowCount != 1 {
		t.Errorf("expected exactly 1 trailing arrow (no arrow between same-category), got %d", arrowCount)
	}
}

func TestPowerlineStyle_RenderLine_LeftAndRight(t *testing.T) {
	s := NewPowerlineStyle(New(nil))
	line := LineData{
		Left:       []string{"~/projects"},
		LeftNames:  []string{"repo_info"},
		Right:      []string{"14:32"},
		RightNames: []string{"time_display"},
	}
	result := s.RenderLine(line, 120)
	// Should have both forward and reverse arrows
	if !strings.Contains(result, "\ue0b0") {
		t.Error("expected forward arrow on left side")
	}
	if !strings.Contains(result, "\ue0b2") {
		t.Error("expected reverse arrow on right side")
	}
}

func TestPowerlineStyle_ImplementsStyle(t *testing.T) {
	var _ Style = NewPowerlineStyle(New(nil))
}

func TestBuildSegments_MergesSameCategory(t *testing.T) {
	names := []string{"cost_daily", "cost_live", "repo_info"}
	contents := []string{"$0.89", "$2.47", "~/repo"}
	segs := buildSegments(names, contents)
	// cost_daily+cost_live merge into 1, repo_info is separate = 2 segments
	if len(segs) != 2 {
		t.Errorf("expected 2 segments, got %d", len(segs))
	}
	if len(segs[0].parts) != 2 {
		t.Errorf("expected first segment to have 2 parts (merged costs), got %d", len(segs[0].parts))
	}
}

func TestBuildSegments_SkipsEmpty(t *testing.T) {
	names := []string{"repo_info", "model_info", "cost_daily"}
	contents := []string{"~/repo", "", "$0.89"}
	segs := buildSegments(names, contents)
	// repo_info is Info, model_info is skipped (empty), cost_daily is Cost = 2 segments
	if len(segs) != 2 {
		t.Errorf("expected 2 segments (empty skipped), got %d", len(segs))
	}
}
