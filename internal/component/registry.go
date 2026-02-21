package component

import (
	"fmt"
	"os"

	"github.com/h2ik/claude-statusline/internal/input"
)

// Registry holds named components and renders them in order.
type Registry struct {
	components map[string]Component
}

// NewRegistry creates a Registry ready to accept component registrations.
func NewRegistry() *Registry {
	return &Registry{components: make(map[string]Component)}
}

// Register adds a component to the registry, keyed by its Name().
func (r *Registry) Register(c Component) {
	r.components[c.Name()] = c
}

// Get retrieves a component by name, returning nil if not found.
func (r *Registry) Get(name string) Component {
	return r.components[name]
}

// RenderLine iterates over the requested component names in order, renders
// each one, and returns only the non-empty results. Panics in individual
// components are recovered so one broken component cannot crash the binary.
func (r *Registry) RenderLine(in *input.StatusLineInput, names []string) []string {
	var output []string
	for _, name := range names {
		if c := r.Get(name); c != nil {
			if rendered := safeRender(c, in); rendered != "" {
				output = append(output, rendered)
			}
		}
	}
	return output
}

// RenderNamedLine iterates over the requested component names in order, renders
// each one, and returns parallel slices of names and rendered outputs for only the
// non-empty results. This lets callers (e.g. powerline style) know which component
// produced which output.
func (r *Registry) RenderNamedLine(in *input.StatusLineInput, names []string) (outNames, outContent []string) {
	for _, name := range names {
		if c := r.Get(name); c != nil {
			if rendered := safeRender(c, in); rendered != "" {
				outNames = append(outNames, name)
				outContent = append(outContent, rendered)
			}
		}
	}
	return outNames, outContent
}

// safeRender calls c.Render and recovers from panics, logging to stderr
// and returning an empty string so other components continue rendering.
func safeRender(c Component, in *input.StatusLineInput) (result string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "component %s panicked: %v\n", c.Name(), r)
			result = ""
		}
	}()
	return c.Render(in)
}
