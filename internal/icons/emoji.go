package icons

// EmojiSet returns standard Unicode emoji characters.
// This is the default icon set for backward compatibility.
type EmojiSet struct{}

var emojiMap = map[string]string{
	Brain:      "\xf0\x9f\xa7\xa0", // 🧠
	Fire:       "\xf0\x9f\x94\xa5", // 🔥
	FloppyDisk: "\xf0\x9f\x92\xbe", // 💾
	Warning:    "\xe2\x9a\xa0\xef\xb8\x8f", // ⚠️
	ChartUp:    "📈",
	ChartBar:   "📊",
	Calendar:   "📅",
	Hourglass:  "⏳",
	Pencil:     "✏️",
	Lightning:  "\xe2\x9a\xa1", // ⚡
	Music:      "\xf0\x9f\x8e\xb5", // 🎵
	Robot:      "\xf0\x9f\xa4\x96", // 🤖
	CheckMark:  "✅",
	Folder:     "📁",
	Link:       "\xf0\x9f\x94\x97", // 🔗
	Clock:      "\xf0\x9f\x95\x90", // 🕐
	Book:       "\xf0\x9f\x93\x9a", // 📚
	Graduation: "\xf0\x9f\x8e\x93", // 🎓
	Sparkles:   "\xe2\x9c\xa8", // ✨
}

// Get returns the emoji character for the given icon name.
// Returns an empty string if the name is unknown.
func (e *EmojiSet) Get(name string) string {
	return emojiMap[name]
}
