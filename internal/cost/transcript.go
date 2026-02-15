package cost

import (
	"encoding/json"
	"strings"
	"time"
)

// transcriptEntry holds parsed fields from a single JSONL transcript line.
type transcriptEntry struct {
	Model            string
	InputTokens      int
	OutputTokens     int
	CacheWriteTokens int
	CacheReadTokens  int
	Timestamp        time.Time
}

// rawTranscriptLine is the minimal JSON structure we unmarshal.
// encoding/json ignores fields we don't declare.
type rawTranscriptLine struct {
	Type    string `json:"type"`
	Message struct {
		Model string `json:"model"`
		Usage struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
	Timestamp string `json:"timestamp"`
}

// parseTranscriptEntry parses a single JSONL line and returns a transcriptEntry
// if it represents an assistant message with a valid model and usage data.
func parseTranscriptEntry(line []byte) (transcriptEntry, bool) {
	var raw rawTranscriptLine
	if err := json.Unmarshal(line, &raw); err != nil {
		return transcriptEntry{}, false
	}
	if raw.Type != "assistant" {
		return transcriptEntry{}, false
	}
	if raw.Message.Model == "" || strings.HasPrefix(raw.Message.Model, "<") {
		return transcriptEntry{}, false
	}
	ts, err := time.Parse(time.RFC3339Nano, raw.Timestamp)
	if err != nil {
		return transcriptEntry{}, false
	}
	return transcriptEntry{
		Model:            raw.Message.Model,
		InputTokens:      raw.Message.Usage.InputTokens,
		OutputTokens:     raw.Message.Usage.OutputTokens,
		CacheWriteTokens: raw.Message.Usage.CacheCreationInputTokens,
		CacheReadTokens:  raw.Message.Usage.CacheReadInputTokens,
		Timestamp:        ts,
	}, true
}
