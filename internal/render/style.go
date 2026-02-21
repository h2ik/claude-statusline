package render

// LineData holds the pre-rendered component strings for a single statusline line.
// Left/Right are the rendered component outputs. LeftNames/RightNames are the
// corresponding component names (used by PowerlineStyle for segment category lookup).
// Slices are parallel: LeftNames[i] is the name for Left[i].
type LineData struct {
	Left       []string
	LeftNames  []string
	Right      []string
	RightNames []string
}

// LineStyle defines how a statusline line is rendered from its component outputs.
// Implementations control separator choice, background colors, and alignment.
// termWidth is the terminal column count (used for padding in powerline mode).
type LineStyle interface {
	RenderLine(line LineData, termWidth int) string
}
