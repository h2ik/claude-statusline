package component

import "github.com/h2ik/claude-statusline/internal/input"

// Component defines the interface that all status line components must implement.
// Each component has a name (used for registry lookup and configuration) and a
// Render method that produces the component's output string.
type Component interface {
    Name() string
    Render(input *input.StatusLineInput) string
}
