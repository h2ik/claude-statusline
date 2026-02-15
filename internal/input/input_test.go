package input

import (
	"strings"
	"testing"
)

func TestParseInput_ValidJSON(t *testing.T) {
	json := `{
		"workspace": {"current_dir": "/tmp/test"},
		"model": {"display_name": "Claude Opus 4.6"},
		"session_id": "test-123",
		"cost": {"total_cost_usd": 0.45}
	}`

	input, err := ParseInput(strings.NewReader(json))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if input.Workspace.CurrentDir != "/tmp/test" {
		t.Errorf("expected /tmp/test, got %s", input.Workspace.CurrentDir)
	}

	if input.Model.DisplayName != "Claude Opus 4.6" {
		t.Errorf("expected Claude Opus 4.6, got %s", input.Model.DisplayName)
	}

	if input.SessionID != "test-123" {
		t.Errorf("expected test-123, got %s", input.SessionID)
	}

	if input.Cost.TotalCostUSD != 0.45 {
		t.Errorf("expected 0.45, got %f", input.Cost.TotalCostUSD)
	}
}

func TestParseInput_MissingWorkspace(t *testing.T) {
	json := `{"model": {"display_name": "Claude"}}`

	_, err := ParseInput(strings.NewReader(json))
	if err == nil {
		t.Fatal("expected error for missing workspace, got nil")
	}
}
