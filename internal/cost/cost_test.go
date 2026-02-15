package cost

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistory_Append(t *testing.T) {
	dir := t.TempDir()
	historyFile := filepath.Join(dir, "history.jsonl")
	h := NewHistory(historyFile)

	entry := Entry{
		SessionID: "test-123",
		Cost:      0.45,
		Timestamp: time.Now(),
	}

	if err := h.Append(entry); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	if _, err := os.Stat(historyFile); err != nil {
		t.Fatalf("history file not created: %v", err)
	}
}

func TestCalculatePeriodCost(t *testing.T) {
	dir := t.TempDir()
	historyFile := filepath.Join(dir, "history.jsonl")
	h := NewHistory(historyFile)

	now := time.Now()
	entries := []Entry{
		{SessionID: "s1", Cost: 1.0, Timestamp: now.Add(-25 * time.Hour)},
		{SessionID: "s2", Cost: 2.0, Timestamp: now.Add(-1 * time.Hour)},
		{SessionID: "s3", Cost: 3.0, Timestamp: now},
	}

	for _, e := range entries {
		_ = h.Append(e)
	}

	cost, err := h.CalculatePeriod(24 * time.Hour)
	if err != nil {
		t.Fatalf("CalculatePeriod failed: %v", err)
	}

	if cost < 4.9 || cost > 5.1 {
		t.Errorf("expected ~5.0, got %f", cost)
	}
}
