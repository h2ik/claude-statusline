package icons

import "testing"

func TestEmojiSetAllIconsPopulated(t *testing.T) {
	set := &EmojiSet{}
	for _, name := range AllIcons {
		got := set.Get(name)
		if got == "" {
			t.Errorf("EmojiSet.Get(%q) returned empty string", name)
		}
	}
}

func TestNerdFontSetAllIconsPopulated(t *testing.T) {
	set := &NerdFontSet{}
	for _, name := range AllIcons {
		got := set.Get(name)
		if got == "" {
			t.Errorf("NerdFontSet.Get(%q) returned empty string", name)
		}
	}
}

func TestNewFactoryEmojiDefault(t *testing.T) {
	tests := []struct {
		style string
		want  string
	}{
		{"emoji", "*icons.EmojiSet"},
		{"", "*icons.EmojiSet"},
		{"unknown", "*icons.EmojiSet"},
		{"nerd-font", "*icons.NerdFontSet"},
	}

	for _, tt := range tests {
		set := New(tt.style)
		got := typeString(set)
		if got != tt.want {
			t.Errorf("New(%q) returned %s, want %s", tt.style, got, tt.want)
		}
	}
}

func TestUnknownIconNameReturnsEmpty(t *testing.T) {
	emoji := &EmojiSet{}
	nerd := &NerdFontSet{}

	if got := emoji.Get("nonexistent"); got != "" {
		t.Errorf("EmojiSet.Get(nonexistent) = %q, want empty", got)
	}
	if got := nerd.Get("nonexistent"); got != "" {
		t.Errorf("NerdFontSet.Get(nonexistent) = %q, want empty", got)
	}
}

func TestEmojiAndNerdFontDiffer(t *testing.T) {
	emoji := &EmojiSet{}
	nerd := &NerdFontSet{}

	for _, name := range AllIcons {
		e := emoji.Get(name)
		n := nerd.Get(name)
		if e == n {
			t.Errorf("icon %q: emoji and nerd-font return identical value %q", name, e)
		}
	}
}

func typeString(v IconSet) string {
	switch v.(type) {
	case *EmojiSet:
		return "*icons.EmojiSet"
	case *NerdFontSet:
		return "*icons.NerdFontSet"
	default:
		return "unknown"
	}
}
