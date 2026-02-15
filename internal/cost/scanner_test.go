package cost

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
)

func TestTranscriptScanner_ComputesCost(t *testing.T) {
	projectsDir := t.TempDir()
	projDir := filepath.Join(projectsDir, "-Users-test")
	os.MkdirAll(projDir, 0755)
	os.WriteFile(filepath.Join(projDir, "s1.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T10:00:00.000Z"}`+"\n",
	), 0644)

	c := cache.New(t.TempDir())
	scanner := NewTranscriptScanner(projectsDir, c)
	total := scanner.CalculatePeriod(30 * 24 * time.Hour)

	expected := 0.0175
	if total < expected-0.0001 || total > expected+0.0001 {
		t.Errorf("expected %f, got %f", expected, total)
	}
}

func TestTranscriptScanner_UsesCacheOnSecondCall(t *testing.T) {
	projectsDir := t.TempDir()
	projDir := filepath.Join(projectsDir, "-Users-test")
	os.MkdirAll(projDir, 0755)
	os.WriteFile(filepath.Join(projDir, "s1.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T10:00:00.000Z"}`+"\n",
	), 0644)

	c := cache.New(t.TempDir())
	scanner := NewTranscriptScanner(projectsDir, c)

	first := scanner.CalculatePeriod(30 * 24 * time.Hour)
	os.Remove(filepath.Join(projDir, "s1.jsonl"))
	second := scanner.CalculatePeriod(30 * 24 * time.Hour)

	if first != second {
		t.Errorf("expected cached value %f, got %f", first, second)
	}
}

func TestTranscriptScanner_DifferentDurationsUseDifferentCacheKeys(t *testing.T) {
	projectsDir := t.TempDir()
	projDir := filepath.Join(projectsDir, "-Users-test")
	os.MkdirAll(projDir, 0755)
	os.WriteFile(filepath.Join(projDir, "s1.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"`+time.Now().Add(-48*time.Hour).Format(time.RFC3339Nano)+`"}`+"\n",
	), 0644)

	c := cache.New(t.TempDir())
	scanner := NewTranscriptScanner(projectsDir, c)

	monthly := scanner.CalculatePeriod(30 * 24 * time.Hour)
	daily := scanner.CalculatePeriod(24 * time.Hour)

	if monthly <= 0 {
		t.Errorf("expected monthly > 0, got %f", monthly)
	}
	if daily != 0.0 {
		t.Errorf("expected daily 0.0 for 48h-old entry, got %f", daily)
	}
}

func TestTranscriptScanner_EmptyProjectsDir(t *testing.T) {
	c := cache.New(t.TempDir())
	scanner := NewTranscriptScanner(t.TempDir(), c)
	total := scanner.CalculatePeriod(30 * 24 * time.Hour)
	if total != 0.0 {
		t.Errorf("expected 0.0, got %f", total)
	}
}
