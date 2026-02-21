package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	arrowRight = "\ue0b0" // filled right-pointing arrow
	arrowLeft  = "\ue0b2" // filled left-pointing arrow
)

// PowerlineStyle renders statusline components as colored background segments
// separated by Nerd Font powerline arrow characters. Adjacent components in the
// same semantic category are merged into a single segment.
type PowerlineStyle struct {
	lg *lipgloss.Renderer
}

// NewPowerlineStyle creates a PowerlineStyle renderer from an existing Renderer.
func NewPowerlineStyle(r *Renderer) *PowerlineStyle {
	return &PowerlineStyle{lg: r.lg}
}

// segment groups one or more component outputs that share the same SegmentCategory.
type segment struct {
	category SegmentCategory
	parts    []string
}

// buildSegments groups consecutive components by SegmentCategory, merging adjacent
// same-category components into a single segment. Empty components are skipped.
func buildSegments(names, contents []string) []segment {
	var segments []segment
	for i, content := range contents {
		stripped := StripANSI(content)
		if strings.TrimSpace(stripped) == "" {
			continue
		}
		name := ""
		if i < len(names) {
			name = names[i]
		}
		cat := SegmentCategoryFor(name)
		if len(segments) > 0 && segments[len(segments)-1].category == cat {
			segments[len(segments)-1].parts = append(segments[len(segments)-1].parts, stripped)
		} else {
			segments = append(segments, segment{category: cat, parts: []string{stripped}})
		}
	}
	return segments
}

// renderSegmentText renders the text content of a segment with background and
// foreground colors from its category, with horizontal padding.
func (s *PowerlineStyle) renderSegmentText(seg segment) string {
	text := strings.Join(seg.parts, " \u2502 ")
	return s.lg.NewStyle().
		Background(seg.category.Background).
		Foreground(seg.category.Foreground).
		Padding(0, 1).
		Render(text)
}

// renderLeftSegments renders left-aligned segments with forward-pointing arrows
// between different-category segments and a trailing arrow after the last segment.
func (s *PowerlineStyle) renderLeftSegments(segments []segment) string {
	if len(segments) == 0 {
		return ""
	}
	var parts []string
	for i, seg := range segments {
		parts = append(parts, s.renderSegmentText(seg))
		if i < len(segments)-1 {
			arrow := s.lg.NewStyle().
				Foreground(seg.category.Background).
				Background(segments[i+1].category.Background).
				Render(arrowRight)
			parts = append(parts, arrow)
		}
	}
	last := segments[len(segments)-1]
	trailingArrow := s.lg.NewStyle().
		Foreground(last.category.Background).
		Render(arrowRight)
	parts = append(parts, trailingArrow)
	return strings.Join(parts, "")
}

// renderRightSegments renders right-aligned segments with a leading reverse arrow
// before the first segment and reverse arrows between different-category segments.
func (s *PowerlineStyle) renderRightSegments(segments []segment) string {
	if len(segments) == 0 {
		return ""
	}
	var parts []string
	first := segments[0]
	leadingArrow := s.lg.NewStyle().
		Foreground(first.category.Background).
		Render(arrowLeft)
	parts = append(parts, leadingArrow)
	for i, seg := range segments {
		parts = append(parts, s.renderSegmentText(seg))
		if i < len(segments)-1 {
			arrow := s.lg.NewStyle().
				Foreground(segments[i+1].category.Background).
				Background(seg.category.Background).
				Render(arrowLeft)
			parts = append(parts, arrow)
		}
	}
	return strings.Join(parts, "")
}

// RenderLine renders a complete statusline from LineData, with left-aligned segments
// on the left, right-aligned segments on the right, and space-padding in between
// to fill the terminal width.
func (s *PowerlineStyle) RenderLine(line LineData, termWidth int) string {
	leftSegs := buildSegments(line.LeftNames, line.Left)
	rightSegs := buildSegments(line.RightNames, line.Right)
	if len(leftSegs) == 0 && len(rightSegs) == 0 {
		return ""
	}
	leftStr := s.renderLeftSegments(leftSegs)
	rightStr := s.renderRightSegments(rightSegs)
	leftWidth := VisualWidth(leftStr)
	rightWidth := VisualWidth(rightStr)
	padding := termWidth - leftWidth - rightWidth
	if padding < 0 {
		padding = 0
	}
	return leftStr + strings.Repeat(" ", padding) + rightStr
}
