package render

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for a statusline theme.
type Theme struct {
	Name     string
	Base     lipgloss.Color // Background base (dark for dark themes, light for Latte)
	Overlay0 lipgloss.Color // Dimmed/labels
	Text     lipgloss.Color // Default text
	Green    lipgloss.Color // Clean/good
	Red      lipgloss.Color // Critical
	Yellow   lipgloss.Color // Warning
	Blue     lipgloss.Color // Paths/info
	Mauve    lipgloss.Color // Accent
	Peach    lipgloss.Color // Costs
	Teal     lipgloss.Color // Secondary
}

// Catppuccin flavor instances. Use ThemeByName to look up by config string.
var (
	ThemeMocha = Theme{
		Name:     "catppuccin-mocha",
		Base:     lipgloss.Color("#1e1e2e"),
		Overlay0: lipgloss.Color("#6c7086"),
		Text:     lipgloss.Color("#cdd6f4"),
		Green:    lipgloss.Color("#a6e3a1"),
		Red:      lipgloss.Color("#f38ba8"),
		Yellow:   lipgloss.Color("#f9e2af"),
		Blue:     lipgloss.Color("#89b4fa"),
		Mauve:    lipgloss.Color("#cba6f7"),
		Peach:    lipgloss.Color("#fab387"),
		Teal:     lipgloss.Color("#94e2d5"),
	}

	ThemeLatte = Theme{
		Name:     "catppuccin-latte",
		Base:     lipgloss.Color("#eff1f5"),
		Overlay0: lipgloss.Color("#9ca0b0"),
		Text:     lipgloss.Color("#4c4f69"),
		Green:    lipgloss.Color("#40a02b"),
		Red:      lipgloss.Color("#d20f39"),
		Yellow:   lipgloss.Color("#df8e1d"),
		Blue:     lipgloss.Color("#1e66f5"),
		Mauve:    lipgloss.Color("#8839ef"),
		Peach:    lipgloss.Color("#fe640b"),
		Teal:     lipgloss.Color("#179299"),
	}

	ThemeFrappe = Theme{
		Name:     "catppuccin-frappe",
		Base:     lipgloss.Color("#292c3c"),
		Overlay0: lipgloss.Color("#626880"),
		Text:     lipgloss.Color("#c6d0f5"),
		Green:    lipgloss.Color("#a6d189"),
		Red:      lipgloss.Color("#e78284"),
		Yellow:   lipgloss.Color("#e5c890"),
		Blue:     lipgloss.Color("#8caaee"),
		Mauve:    lipgloss.Color("#ca9ee6"),
		Peach:    lipgloss.Color("#ef9f76"),
		Teal:     lipgloss.Color("#81c8be"),
	}

	ThemeMacchiato = Theme{
		Name:     "catppuccin-macchiato",
		Base:     lipgloss.Color("#24273a"),
		Overlay0: lipgloss.Color("#6e738d"),
		Text:     lipgloss.Color("#cad3f5"),
		Green:    lipgloss.Color("#a6da95"),
		Red:      lipgloss.Color("#ed8796"),
		Yellow:   lipgloss.Color("#eed49f"),
		Blue:     lipgloss.Color("#8aadf4"),
		Mauve:    lipgloss.Color("#c6a0f6"),
		Peach:    lipgloss.Color("#f5a97f"),
		Teal:     lipgloss.Color("#8bd5ca"),
	}
)

// ThemeByName returns the theme for the given name.
// If the name is unknown or empty, it returns ThemeMocha and false.
func ThemeByName(name string) (Theme, bool) {
	switch name {
	case "catppuccin-mocha":
		return ThemeMocha, true
	case "catppuccin-latte":
		return ThemeLatte, true
	case "catppuccin-frappe":
		return ThemeFrappe, true
	case "catppuccin-macchiato":
		return ThemeMacchiato, true
	default:
		return ThemeMocha, false
	}
}
