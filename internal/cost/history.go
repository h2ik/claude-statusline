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

// Append writes a single cost entry to the history file.
func (h *History) Append(entry Entry) error {
	if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	f, err := os.OpenFile(h.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

// CalculatePeriod returns the total cost of unique sessions within the given
// duration from now. Entries are deduplicated by SessionID -- only the first
// occurrence of each session is counted.
func (h *History) CalculatePeriod(duration time.Duration) (float64, error) {
	f, err := os.Open(h.path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0.0, nil
		}
		return 0.0, fmt.Errorf("open failed: %w", err)
	}
	defer f.Close()

	cutoff := time.Now().Add(-duration)
	total := 0.0
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		if entry.Timestamp.After(cutoff) {
			if !seen[entry.SessionID] {
				total += entry.Cost
				seen[entry.SessionID] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0.0, fmt.Errorf("scan failed: %w", err)
	}

	return total, nil
}
