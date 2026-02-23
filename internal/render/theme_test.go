package render

import (
	"testing"
)

func TestThemeByName_KnownThemes(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantBlue string
	}{
		{"catppuccin-mocha", "catppuccin-mocha", "#89b4fa"},
		{"catppuccin-latte", "catppuccin-latte", "#1e66f5"},
		{"catppuccin-frappe", "catppuccin-frappe", "#8caaee"},
		{"catppuccin-macchiato", "catppuccin-macchiato", "#8aadf4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme, ok := ThemeByName(tt.name)
			if !ok {
				t.Errorf("ThemeByName(%q) returned ok=false, want true", tt.name)
			}
			if theme.Name != tt.wantName {
				t.Errorf("theme.Name = %q, want %q", theme.Name, tt.wantName)
			}
			if string(theme.Blue) != tt.wantBlue {
				t.Errorf("theme.Blue = %q, want %q", theme.Blue, tt.wantBlue)
			}
		})
	}
}

func TestThemeByName_UnknownReturnsMocha(t *testing.T) {
	theme, ok := ThemeByName("solarized-dark")
	if ok {
		t.Error("ThemeByName(unknown) returned ok=true, want false")
	}
	if theme.Name != "catppuccin-mocha" {
		t.Errorf("ThemeByName(unknown) returned theme %q, want catppuccin-mocha", theme.Name)
	}
}

func TestThemeByName_EmptyReturnsMocha(t *testing.T) {
	theme, ok := ThemeByName("")
	if ok {
		t.Error("ThemeByName('') returned ok=true, want false")
	}
	if theme.Name != "catppuccin-mocha" {
		t.Errorf("ThemeByName('') returned theme %q, want catppuccin-mocha", theme.Name)
	}
}

func TestTheme_MochaBaseIsDark(t *testing.T) {
	if string(ThemeMocha.Base) != "#1e1e2e" {
		t.Errorf("ThemeMocha.Base = %q, want #1e1e2e", ThemeMocha.Base)
	}
}

func TestTheme_LatteBaseIsLight(t *testing.T) {
	if string(ThemeLatte.Base) != "#eff1f5" {
		t.Errorf("ThemeLatte.Base = %q, want #eff1f5", ThemeLatte.Base)
	}
}

func TestTheme_AllFlavorsHaveAllFields(t *testing.T) {
	themes := []*Theme{&ThemeMocha, &ThemeLatte, &ThemeFrappe, &ThemeMacchiato}
	fields := []struct {
		name  string
		value func(*Theme) string
	}{
		{"Base", func(t *Theme) string { return string(t.Base) }},
		{"Overlay0", func(t *Theme) string { return string(t.Overlay0) }},
		{"Text", func(t *Theme) string { return string(t.Text) }},
		{"Green", func(t *Theme) string { return string(t.Green) }},
		{"Red", func(t *Theme) string { return string(t.Red) }},
		{"Yellow", func(t *Theme) string { return string(t.Yellow) }},
		{"Blue", func(t *Theme) string { return string(t.Blue) }},
		{"Mauve", func(t *Theme) string { return string(t.Mauve) }},
		{"Peach", func(t *Theme) string { return string(t.Peach) }},
		{"Teal", func(t *Theme) string { return string(t.Teal) }},
	}
	for _, theme := range themes {
		for _, field := range fields {
			if field.value(theme) == "" {
				t.Errorf("theme %q has empty %s field", theme.Name, field.name)
			}
		}
	}
}
