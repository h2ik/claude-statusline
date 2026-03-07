package cost

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// transcriptEntry holds parsed fields from a single JSONL transcript line.
type transcriptEntry struct {
	MessageID        string
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
		ID    string `json:"id"`
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
		MessageID:        raw.Message.ID,
		Model:            raw.Message.Model,
		InputTokens:      raw.Message.Usage.InputTokens,
		OutputTokens:     raw.Message.Usage.OutputTokens,
		CacheWriteTokens: raw.Message.Usage.CacheCreationInputTokens,
		CacheReadTokens:  raw.Message.Usage.CacheReadInputTokens,
		Timestamp:        ts,
	}, true
}

// scanFile reads a single JSONL file and returns the total USD cost of all
// assistant entries whose timestamp is after the cutoff.
//
// Claude Code transcripts emit multiple assistant entries per API call as the
// response streams in (each carrying the same message ID). To avoid counting
// the same turn multiple times, we deduplicate by message ID — keeping only
// the last entry for each ID, which has the final token counts. Entries
// without a message ID are counted individually.
func scanFile(path string, cutoff time.Time) float64 {
	f, err := os.Open(path)
	if err != nil {
		return 0.0
	}
	defer func() { _ = f.Close() }()

	// Track the last entry per message ID so streaming duplicates collapse.
	type entryData struct {
		inputTokens      int
		outputTokens     int
		cacheWriteTokens int
		cacheReadTokens  int
		model            string
	}
	deduped := make(map[string]entryData)
	var noIDTotal float64

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		entry, ok := parseTranscriptEntry(line)
		if !ok {
			continue
		}
		if !entry.Timestamp.After(cutoff) {
			continue
		}
		if entry.MessageID == "" {
			noIDTotal += CalculateEntryCost(
				entry.InputTokens, entry.OutputTokens,
				entry.CacheWriteTokens, entry.CacheReadTokens,
				entry.Model,
			)
			continue
		}
		// Last write wins — later entries for the same ID have final token counts.
		deduped[entry.MessageID] = entryData{
			inputTokens:      entry.InputTokens,
			outputTokens:     entry.OutputTokens,
			cacheWriteTokens: entry.CacheWriteTokens,
			cacheReadTokens:  entry.CacheReadTokens,
			model:            entry.Model,
		}
	}

	var total float64
	for _, e := range deduped {
		total += CalculateEntryCost(
			e.inputTokens, e.outputTokens,
			e.cacheWriteTokens, e.cacheReadTokens,
			e.model,
		)
	}
	return total + noIDTotal
}

// ScanTranscripts walks the root directory (typically ~/.claude/projects/)
// recursively, summing costs from all .jsonl files within the given duration.
// Skips tool-results directories. Uses mtime pre-filtering to skip stale files.
func ScanTranscripts(root string, duration time.Duration) float64 {
	return ScanTranscriptsSince(root, time.Now().Add(-duration))
}

// ScanTranscriptsSince walks the root directory recursively, summing costs
// from all .jsonl files whose entries have timestamps after the given cutoff.
// Skips tool-results directories. Uses mtime pre-filtering to skip stale files.
func ScanTranscriptsSince(root string, cutoff time.Time) float64 {
	var total float64

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == "tool-results" {
			return filepath.SkipDir
		}
		if info.IsDir() || filepath.Ext(path) != ".jsonl" {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			return nil
		}
		total += scanFile(path, cutoff)
		return nil
	})

	return total
}
