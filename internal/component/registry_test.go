package component

import (
	"testing"

	"github.com/h2ik/claude-statusline/internal/input"
)

type mockComponent struct {
	name   string
	output string
}

func (m *mockComponent) Name() string                            { return m.name }
func (m *mockComponent) Render(in *input.StatusLineInput) string { return m.output }

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	comp := &mockComponent{name: "test", output: "hello"}
	r.Register(comp)
	if c := r.Get("test"); c == nil {
		t.Fatal("expected component to be registered")
	}
}

func TestRegistry_RenderLine(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockComponent{name: "c1", output: "first"})
	r.Register(&mockComponent{name: "c2", output: "second"})

	in := &input.StatusLineInput{}
	output := r.RenderLine(in, []string{"c1", "c2"})

	if len(output) != 2 {
		t.Errorf("expected 2 components, got %d", len(output))
	}
	if output[0] != "first" {
		t.Errorf("expected 'first', got %s", output[0])
	}
	if output[1] != "second" {
		t.Errorf("expected 'second', got %s", output[1])
	}
}

func TestRegistry_Get_ReturnsNilForUnknown(t *testing.T) {
	r := NewRegistry()
	if c := r.Get("nonexistent"); c != nil {
		t.Fatal("expected nil for unregistered component")
	}
}

func TestRegistry_RenderLine_SkipsMissing(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockComponent{name: "c1", output: "first"})

	in := &input.StatusLineInput{}
	output := r.RenderLine(in, []string{"c1", "missing", "also_missing"})

	if len(output) != 1 {
		t.Errorf("expected 1 component, got %d", len(output))
	}
	if output[0] != "first" {
		t.Errorf("expected 'first', got %s", output[0])
	}
}

func TestRegistry_RenderLine_SkipsEmptyOutput(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockComponent{name: "c1", output: "first"})
	r.Register(&mockComponent{name: "c2", output: ""})

	in := &input.StatusLineInput{}
	output := r.RenderLine(in, []string{"c1", "c2"})

	if len(output) != 1 {
		t.Errorf("expected 1 component (empty output skipped), got %d", len(output))
	}
}
