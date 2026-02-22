package icons

// Icon name constants used across all components.
const (
	Brain      = "brain"
	Fire       = "fire"
	FloppyDisk = "floppy_disk"
	Warning    = "warning"
	ChartUp    = "chart_up"
	ChartBar   = "chart_bar"
	Calendar   = "calendar"
	Hourglass  = "hourglass"
	Pencil     = "pencil"
	Lightning  = "lightning"
	Music      = "music"
	Robot      = "robot"
	CheckMark  = "check_mark"
	Folder     = "folder"
	Link       = "link"
	Clock      = "clock"
	Book       = "book"
	Graduation = "graduation"
	Sparkles   = "sparkles"
)

// AllIcons lists every known icon name for testing and validation.
var AllIcons = []string{
	Brain, Fire, FloppyDisk, Warning, ChartUp, ChartBar, Calendar,
	Hourglass, Pencil, Lightning, Music, Robot, CheckMark, Folder,
	Link, Clock, Book, Graduation, Sparkles,
}

// IconSet provides icon glyphs by name. Two implementations exist:
// EmojiSet (default, wide Unicode emoji) and NerdFontSet (Nerd Font glyphs).
type IconSet interface {
	Get(name string) string
}

// New returns an IconSet for the given style name.
// Recognized values: "nerd-font". Everything else (including "") returns
// the default EmojiSet for backward compatibility.
func New(style string) IconSet {
	switch style {
	case "nerd-font":
		return &NerdFontSet{}
	default:
		return &EmojiSet{}
	}
}
