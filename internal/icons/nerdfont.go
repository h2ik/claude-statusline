package icons

// NerdFontSet returns Nerd Font glyphs that render as single-width characters.
// These require a Nerd Font-patched terminal font to display correctly.
type NerdFontSet struct{}

var nerdFontMap = map[string]string{
	Brain:      "\U000F09D2", // nf-md-brain
	Fire:       "\U000F0238", // nf-md-fire
	FloppyDisk: "\uf0c7",    // nf-fa-save
	Warning:    "\uf071",    // nf-fa-warning
	ChartUp:    "\U000F08D4", // nf-md-chart_line
	ChartBar:   "\U000F0128", // nf-md-chart_bar
	Calendar:   "\uf073",    // nf-fa-calendar
	Hourglass:  "\U000F051F", // nf-md-timer_sand
	Pencil:     "\uf040",    // nf-fa-pencil
	Lightning:  "\uf0e7",    // nf-fa-bolt
	Music:      "\uf001",    // nf-fa-music
	Robot:      "\U000F06A9", // nf-md-robot
	CheckMark:  "\uf00c",    // nf-fa-check
	Folder:     "\uf07b",    // nf-fa-folder
	Link:       "\uf0c1",    // nf-fa-link
	Clock:      "\U000F0150", // nf-md-clock_outline
	Book:       "\uf02d",    // nf-fa-book
	Graduation: "\U000F0474", // nf-md-school
	Sparkles:   "\U000F0674", // nf-md-creation
}

// Get returns the Nerd Font glyph for the given icon name.
// Returns an empty string if the name is unknown.
func (n *NerdFontSet) Get(name string) string {
	return nerdFontMap[name]
}
