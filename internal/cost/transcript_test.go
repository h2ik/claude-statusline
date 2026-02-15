package cost

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseTranscriptEntry_ValidAssistant(t *testing.T) {
	line := `{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":200,"cache_read_input_tokens":1000}},"timestamp":"2026-02-15T14:48:28.920Z"}`

	entry, ok := parseTranscriptEntry([]byte(line))
	if !ok {
		t.Fatal("expected ok=true for valid assistant entry")
	}
	if entry.Model != "claude-opus-4-5-20251101" {
		t.Errorf("model: got %q, want %q", entry.Model, "claude-opus-4-5-20251101")
	}
	if entry.InputTokens != 100 {
		t.Errorf("input_tokens: got %d, want 100", entry.InputTokens)
	}
	if entry.OutputTokens != 50 {
		t.Errorf("output_tokens: got %d, want 50", entry.OutputTokens)
	}
	if entry.CacheWriteTokens != 200 {
		t.Errorf("cache_write_tokens: got %d, want 200", entry.CacheWriteTokens)
	}
	if entry.CacheReadTokens != 1000 {
		t.Errorf("cache_read_tokens: got %d, want 1000", entry.CacheReadTokens)
	}
	expected := time.Date(2026, 2, 15, 14, 48, 28, 920000000, time.UTC)
	if !entry.Timestamp.Equal(expected) {
		t.Errorf("timestamp: got %v, want %v", entry.Timestamp, expected)
	}
}

func TestParseTranscriptEntry_SkipsUserType(t *testing.T) {
	line := `{"type":"user","message":{"role":"user","content":"hello"},"timestamp":"2026-02-15T14:48:28.920Z"}`
	_, ok := parseTranscriptEntry([]byte(line))
	if ok {
		t.Error("expected ok=false for user type")
	}
}

func TestParseTranscriptEntry_SkipsAssistantWithoutModel(t *testing.T) {
	line := `{"type":"assistant","message":{"model":"","usage":{}},"timestamp":"2026-02-15T14:48:28.920Z"}`
	_, ok := parseTranscriptEntry([]byte(line))
	if ok {
		t.Error("expected ok=false for empty model")
	}
}

func TestParseTranscriptEntry_SkipsSyntheticModel(t *testing.T) {
	line := `{"type":"assistant","message":{"model":"<synthetic>","usage":{"input_tokens":0,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T14:48:28.920Z"}`
	_, ok := parseTranscriptEntry([]byte(line))
	if ok {
		t.Error("expected ok=false for <synthetic> model")
	}
}

func TestParseTranscriptEntry_SkipsMalformedJSON(t *testing.T) {
	_, ok := parseTranscriptEntry([]byte(`{not valid json`))
	if ok {
		t.Error("expected ok=false for malformed JSON")
	}
}

func TestParseTranscriptEntry_HandlesExtraFields(t *testing.T) {
	line := `{"type":"assistant","message":{"model":"claude-opus-4-6","id":"msg_123","usage":{"input_tokens":3,"cache_creation_input_tokens":28787,"cache_read_input_tokens":0,"output_tokens":220,"server_tool_use":{"web_search_requests":0},"service_tier":"standard","cache_creation":{"ephemeral_1h_input_tokens":0,"ephemeral_5m_input_tokens":28787}}},"timestamp":"2026-02-15T14:07:12.083Z"}`
	entry, ok := parseTranscriptEntry([]byte(line))
	if !ok {
		t.Fatal("expected ok=true for entry with extra fields")
	}
	if entry.InputTokens != 3 {
		t.Errorf("input_tokens: got %d, want 3", entry.InputTokens)
	}
	if entry.OutputTokens != 220 {
		t.Errorf("output_tokens: got %d, want 220", entry.OutputTokens)
	}
}

func TestScanFile_ComputesCost(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")
	lines := []string{
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T10:00:00.000Z"}`,
		`{"type":"user","message":{"role":"user","content":"hello"},"timestamp":"2026-02-15T10:00:01.000Z"}`,
		`{"type":"assistant","message":{"model":"claude-haiku-4-5-20251001","usage":{"input_tokens":2000,"output_tokens":100,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T11:00:00.000Z"}`,
	}
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)

	cutoff := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	cost := scanFile(path, cutoff)
	// Opus: (1000*5+500*25)/1M=0.0175, Haiku: (2000*1+100*5)/1M=0.0025
	expected := 0.0175 + 0.0025
	if cost < expected-0.0001 || cost > expected+0.0001 {
		t.Errorf("expected %f, got %f", expected, cost)
	}
}

func TestScanFile_FiltersOldEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")
	lines := []string{
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-01-01T10:00:00.000Z"}`,
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T11:00:00.000Z"}`,
	}
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)

	cutoff := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	cost := scanFile(path, cutoff)
	expected := 0.00175 // Only recent entry
	if cost < expected-0.0001 || cost > expected+0.0001 {
		t.Errorf("expected %f, got %f", expected, cost)
	}
}

func TestScanFile_HandlesEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")
	os.WriteFile(path, []byte(""), 0644)
	cost := scanFile(path, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if cost != 0.0 {
		t.Errorf("expected 0.0, got %f", cost)
	}
}

func TestScanFile_HandlesMissingFile(t *testing.T) {
	cost := scanFile("/nonexistent/path.jsonl", time.Now())
	if cost != 0.0 {
		t.Errorf("expected 0.0, got %f", cost)
	}
}

func TestScanFile_HandlesMalformedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "malformed.jsonl")
	lines := []string{
		`{not valid json`,
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T10:00:00.000Z"}`,
		``,
		`another bad line`,
	}
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)

	cutoff := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	cost := scanFile(path, cutoff)
	expected := 0.0175
	if cost < expected-0.0001 || cost > expected+0.0001 {
		t.Errorf("expected %f, got %f", expected, cost)
	}
}

func TestScanTranscripts_WalksDirectoryTree(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-test-project")
	os.MkdirAll(projectDir, 0755)

	os.WriteFile(filepath.Join(projectDir, "abc-123.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T10:00:00.000Z"}`+"\n",
	), 0644)

	subagentDir := filepath.Join(projectDir, "abc-123", "subagents")
	os.MkdirAll(subagentDir, 0755)
	os.WriteFile(filepath.Join(subagentDir, "agent-xyz.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-haiku-4-5-20251001","usage":{"input_tokens":2000,"output_tokens":100,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T11:00:00.000Z"}`+"\n",
	), 0644)

	total := ScanTranscripts(root, 30*24*time.Hour)
	expected := 0.0175 + 0.0025
	if total < expected-0.0001 || total > expected+0.0001 {
		t.Errorf("expected %f, got %f", expected, total)
	}
}

func TestScanTranscripts_SkipsNonJSONL(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-test")
	os.MkdirAll(projectDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "notes.txt"), []byte("not jsonl"), 0644)
	os.WriteFile(filepath.Join(projectDir, "session.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T10:00:00.000Z"}`+"\n",
	), 0644)

	total := ScanTranscripts(root, 30*24*time.Hour)
	expected := 0.0175
	if total < expected-0.0001 || total > expected+0.0001 {
		t.Errorf("expected %f, got %f", expected, total)
	}
}

func TestScanTranscripts_EmptyDirectory(t *testing.T) {
	total := ScanTranscripts(t.TempDir(), 30*24*time.Hour)
	if total != 0.0 {
		t.Errorf("expected 0.0, got %f", total)
	}
}

func TestScanTranscripts_MissingDirectory(t *testing.T) {
	total := ScanTranscripts("/nonexistent/path", 30*24*time.Hour)
	if total != 0.0 {
		t.Errorf("expected 0.0, got %f", total)
	}
}

func TestScanTranscripts_SkipsToolResultsDir(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-test-project")

	toolDir := filepath.Join(projectDir, "abc-123", "tool-results")
	os.MkdirAll(toolDir, 0755)
	os.WriteFile(filepath.Join(toolDir, "result.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":99999,"output_tokens":99999,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T10:00:00.000Z"}`+"\n",
	), 0644)

	os.MkdirAll(projectDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "session.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"2026-02-15T10:00:00.000Z"}`+"\n",
	), 0644)

	total := ScanTranscripts(root, 30*24*time.Hour)
	expected := 0.00175
	if total < expected-0.0001 || total > expected+0.0001 {
		t.Errorf("expected %f, got %f", expected, total)
	}
}
