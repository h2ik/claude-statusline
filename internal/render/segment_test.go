package render

import (
	"testing"
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
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat == (SegmentCategory{}) {
			t.Errorf("component %q has no segment category", name)
		}
	}
}

func TestSegmentCategory_UnknownComponentGetsDim(t *testing.T) {
	cat := SegmentCategoryFor("unknown_component", &ThemeMocha)
	if cat.Background != ThemeMocha.Overlay0 {
		t.Errorf("unknown component should use Overlay0 background, got %v", cat.Background)
	}
}

func TestSegmentCategory_InfoGroupIsBlue(t *testing.T) {
	for _, name := range []string{"repo_info", "model_info", "bedrock_model"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Blue {
			t.Errorf("component %q should have blue background, got %v", name, cat.Background)
		}
		if cat.Foreground != ThemeMocha.Base {
			t.Errorf("component %q should have Base foreground, got %v", name, cat.Foreground)
		}
	}
}

func TestSegmentCategory_CostGroupIsPeach(t *testing.T) {
	for _, name := range []string{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "burn_rate"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Peach {
			t.Errorf("component %q should have peach background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_MetricsGroupIsTeal(t *testing.T) {
	for _, name := range []string{"context_window", "cache_efficiency", "block_projection"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Teal {
			t.Errorf("component %q should have teal background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_ActivityGroupIsGreen(t *testing.T) {
	for _, name := range []string{"code_productivity", "commits"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Green {
			t.Errorf("component %q should have green background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_MetaGroupIsMauve(t *testing.T) {
	for _, name := range []string{"version_info", "session_mode"} {
		cat := SegmentCategoryFor(name, &ThemeMocha)
		if cat.Background != ThemeMocha.Mauve {
			t.Errorf("component %q should have mauve background, got %v", name, cat.Background)
		}
	}
}

func TestSegmentCategory_LatteAutoInvert(t *testing.T) {
	cat := SegmentCategoryFor("repo_info", &ThemeLatte)
	if cat.Background != ThemeLatte.Blue {
		t.Errorf("Latte repo_info should have Latte Blue background, got %v", cat.Background)
	}
	if cat.Foreground != ThemeLatte.Base {
		t.Errorf("Latte repo_info should have Latte Base foreground, got %v", cat.Foreground)
	}
}

func TestSegmentCategory_CrossThemeDiffers(t *testing.T) {
	catMocha := SegmentCategoryFor("repo_info", &ThemeMocha)
	catLatte := SegmentCategoryFor("repo_info", &ThemeLatte)
	if catMocha.Background == catLatte.Background {
		t.Error("Mocha and Latte should have different blue backgrounds for repo_info")
	}
}
