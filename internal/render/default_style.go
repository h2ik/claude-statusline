package render

import "strings"

// DefaultStyle renders lines by joining all components (Left then Right)
// with a configurable separator. Right-side components are appended after
// Left components -- there is no actual right-alignment in this style.
// This preserves the original behavior for users who have not opted into powerline.
type DefaultStyle struct {
	separator string
}

// NewDefaultStyle returns a DefaultStyle with the given separator string.
func NewDefaultStyle(separator string) *DefaultStyle {
	return &DefaultStyle{separator: separator}
}

// RenderLine joins all non-empty Left and Right components with the separator.
// termWidth is unused in the default style.
func (s *DefaultStyle) RenderLine(line LineData, _ int) string {
	all := make([]string, 0, len(line.Left)+len(line.Right))
	all = append(all, line.Left...)
	all = append(all, line.Right...)

	var nonEmpty []string
	for _, c := range all {
		if strings.TrimSpace(c) != "" {
			nonEmpty = append(nonEmpty, c)
		}
	}
	return strings.Join(nonEmpty, s.separator)
}
