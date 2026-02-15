package cost

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// History manages an append-only JSONL file of cost entries.
type History struct {
	path string
}

// NewHistory creates a History backed by the given file path.
func NewHistory(path string) *History {
	return &History{path: path}
}

// maxHistoryAge is the oldest entry we keep. Anything beyond this is
// never counted by any period component, so we prune it on compaction.
const maxHistoryAge = 31 * 24 * time.Hour

// Append writes a single cost entry to the history file and periodically
// compacts the file by removing entries older than 31 days.
func (h *History) Append(entry Entry) error {
	if err := os.MkdirAll(filepath.Dir(h.path), 0700); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	f, err := os.OpenFile(h.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	// Compact roughly once per hour — check if the compaction marker is stale.
	h.maybeCompact()

	return nil
}

// maybeCompact removes entries older than maxHistoryAge. It uses a sidecar
// file's mtime to avoid running on every single append — at most once per hour.
func (h *History) maybeCompact() {
	marker := h.path + ".compacted"
	if info, err := os.Stat(marker); err == nil {
		if time.Since(info.ModTime()) < time.Hour {
			return
		}
	}

	// Read all entries, keep only recent ones
	f, err := os.Open(h.path)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-maxHistoryAge)
	var kept [][]byte

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if entry.Timestamp.After(cutoff) {
			line := make([]byte, len(scanner.Bytes()))
			copy(line, scanner.Bytes())
			kept = append(kept, line)
		}
	}
	_ = f.Close()

	// Rewrite the file atomically via temp file
	tmp := h.path + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	for _, line := range kept {
		_, _ = out.Write(append(line, '\n'))
	}
	_ = out.Close()

	_ = os.Rename(tmp, h.path)

	// Touch the compaction marker
	_ = os.WriteFile(marker, nil, 0600)
}

// Deprecated: Period costs are now computed by TranscriptScanner which reads
// Claude Code's native JSONL transcript files directly for accurate totals.
//
// CalculatePeriod returns the total cost of unique sessions within the given
// duration from now. When a session appears multiple times (cost_live appends
// on every render), the last entry wins because it has the most up-to-date cost.
func (h *History) CalculatePeriod(duration time.Duration) (float64, error) {
	f, err := os.Open(h.path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0.0, nil
		}
		return 0.0, fmt.Errorf("open failed: %w", err)
	}
	defer func() { _ = f.Close() }()

	cutoff := time.Now().Add(-duration)
	// Track the latest cost per session (last write wins)
	sessionCost := make(map[string]float64)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		if entry.Timestamp.After(cutoff) {
			sessionCost[entry.SessionID] = entry.Cost
		}
	}

	if err := scanner.Err(); err != nil {
		return 0.0, fmt.Errorf("scan failed: %w", err)
	}

	total := 0.0
	for _, cost := range sessionCost {
		total += cost
	}

	return total, nil
}
