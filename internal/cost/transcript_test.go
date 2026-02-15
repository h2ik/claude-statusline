package cost

import (
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
