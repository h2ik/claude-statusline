package components

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// ============================================================
// CostMonthly tests
// ============================================================

func TestCostMonthly_Name(t *testing.T) {
	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(t.TempDir(), ca)
	c := NewCostMonthly(r, s, icons.New("emoji"))

	if c.Name() != "cost_monthly" {
		t.Errorf("expected 'cost_monthly', got %q", c.Name())
	}
}

func TestCostMonthly_Render_EmptyHistory(t *testing.T) {
	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(t.TempDir(), ca)
	c := NewCostMonthly(r, s, icons.New("emoji"))

	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$0.00") {
		t.Errorf("expected '$0.00' for empty history, got: %s", output)
	}
	if !strings.Contains(output, "30DAY") {
		t.Errorf("expected '30DAY' label in output, got: %s", output)
	}
}

func TestCostMonthly_Render_WithEntries(t *testing.T) {
	projectsDir := t.TempDir()
	projDir := filepath.Join(projectsDir, "-Users-test")
	_ = os.MkdirAll(projDir, 0755)
	_ = os.WriteFile(filepath.Join(projDir, "session.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"`+time.Now().Add(-24*time.Hour).Format(time.RFC3339Nano)+`"}`+"\n",
	), 0644)

	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(projectsDir, ca)
	c := NewCostMonthly(r, s, icons.New("emoji"))
	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$0.02") {
		t.Errorf("expected cost around $0.0175 for transcript entry, got: %s", output)
	}
}

// ============================================================
// CostWeekly tests
// ============================================================

func TestCostWeekly_Name(t *testing.T) {
	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(t.TempDir(), ca)
	c := NewCostWeekly(r, s, icons.New("emoji"))

	if c.Name() != "cost_weekly" {
		t.Errorf("expected 'cost_weekly', got %q", c.Name())
	}
}

func TestCostWeekly_Render_EmptyHistory(t *testing.T) {
	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(t.TempDir(), ca)
	c := NewCostWeekly(r, s, icons.New("emoji"))

	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$0.00") {
		t.Errorf("expected '$0.00' for empty history, got: %s", output)
	}
	if !strings.Contains(output, "7DAY") {
		t.Errorf("expected '7DAY' label in output, got: %s", output)
	}
}

func TestCostWeekly_Render_FiltersOldEntries(t *testing.T) {
	projectsDir := t.TempDir()
	projDir := filepath.Join(projectsDir, "-Users-test")
	_ = os.MkdirAll(projDir, 0755)
	_ = os.WriteFile(filepath.Join(projDir, "recent.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"`+time.Now().Add(-48*time.Hour).Format(time.RFC3339Nano)+`"}`+"\n",
	), 0644)
	_ = os.WriteFile(filepath.Join(projDir, "old.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":10000,"output_tokens":5000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"`+time.Now().Add(-10*24*time.Hour).Format(time.RFC3339Nano)+`"}`+"\n",
	), 0644)

	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(projectsDir, ca)
	c := NewCostWeekly(r, s, icons.New("emoji"))
	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$0.02") {
		t.Errorf("expected cost around $0.0175 (only recent entry), got: %s", output)
	}
}

// ============================================================
// CostDaily (CostToday) tests
// ============================================================

func TestCostDaily_Name(t *testing.T) {
	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(t.TempDir(), ca)
	c := NewCostDaily(r, s, icons.New("emoji"))

	if c.Name() != "cost_daily" {
		t.Errorf("expected 'cost_daily', got %q", c.Name())
	}
}

func TestCostDaily_Render_EmptyHistory(t *testing.T) {
	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(t.TempDir(), ca)
	c := NewCostDaily(r, s, icons.New("emoji"))

	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$0.00") {
		t.Errorf("expected '$0.00' for empty history, got: %s", output)
	}
	if !strings.Contains(output, "TODAY") {
		t.Errorf("expected 'TODAY' label in output, got: %s", output)
	}
}

func TestCostDaily_Render_IncludesTodayEntries(t *testing.T) {
	projectsDir := t.TempDir()
	projDir := filepath.Join(projectsDir, "-Users-test")
	_ = os.MkdirAll(projDir, 0755)

	// Entry from 1 hour ago (should always be today)
	_ = os.WriteFile(filepath.Join(projDir, "recent.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-haiku-4-5-20251001","usage":{"input_tokens":2000,"output_tokens":100,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"`+time.Now().Add(-1*time.Hour).Format(time.RFC3339Nano)+`"}`+"\n",
	), 0644)
	// Entry from 48 hours ago (never today)
	_ = os.WriteFile(filepath.Join(projDir, "old.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":10000,"output_tokens":5000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"`+time.Now().Add(-48*time.Hour).Format(time.RFC3339Nano)+`"}`+"\n",
	), 0644)

	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(projectsDir, ca)
	c := NewCostDaily(r, s, icons.New("emoji"))
	in := &input.StatusLineInput{}

	output := c.Render(in)
	// Haiku: (2000*1+100*5)/1M = 0.0025
	if !strings.Contains(output, "$0.00") {
		t.Errorf("expected cost around $0.0025 (only today's entry), got: %s", output)
	}
}

func TestCostDaily_Render_ExcludesYesterdayEntries(t *testing.T) {
	projectsDir := t.TempDir()
	projDir := filepath.Join(projectsDir, "-Users-test")
	_ = os.MkdirAll(projDir, 0755)

	// Entry from yesterday at 23:00 — should be excluded even if < 24h ago
	now := time.Now()
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 23, 0, 0, 0, now.Location())
	_ = os.WriteFile(filepath.Join(projDir, "yesterday.jsonl"), []byte(
		`{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","usage":{"input_tokens":10000,"output_tokens":5000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"timestamp":"`+yesterday.Format(time.RFC3339Nano)+`"}`+"\n",
	), 0644)

	r := render.New(nil)
	ca := cache.New(t.TempDir())
	s := cost.NewTranscriptScanner(projectsDir, ca)
	c := NewCostDaily(r, s, icons.New("emoji"))
	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$0.00") {
		t.Errorf("expected $0.00 for yesterday-only entries, got: %s", output)
	}
}

// ============================================================
// CostLive tests
// ============================================================

func TestCostLive_Name(t *testing.T) {
	r := render.New(nil)
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostLive(r, h, icons.New("emoji"))

	if c.Name() != "cost_live" {
		t.Errorf("expected 'cost_live', got %q", c.Name())
	}
}

func TestCostLive_Render_ZeroCost(t *testing.T) {
	r := render.New(nil)
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostLive(r, h, icons.New("emoji"))

	in := &input.StatusLineInput{
		Cost: input.CostInfo{TotalCostUSD: 0.0},
	}

	output := c.Render(in)
	if !strings.Contains(output, "$0.00") {
		t.Errorf("expected '$0.00' for zero cost, got: %s", output)
	}
	if !strings.Contains(output, "LIVE") {
		t.Errorf("expected 'LIVE' label in output, got: %s", output)
	}
}

func TestCostLive_Render_DisplaysCost(t *testing.T) {
	r := render.New(nil)
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostLive(r, h, icons.New("emoji"))

	in := &input.StatusLineInput{
		SessionID: "session-abc",
		Cost:      input.CostInfo{TotalCostUSD: 1.23},
	}

	output := c.Render(in)
	if !strings.Contains(output, "$1.23") {
		t.Errorf("expected '$1.23' for live cost, got: %s", output)
	}
}

func TestCostLive_Render_AppendsToHistory(t *testing.T) {
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "cost.jsonl")

	r := render.New(nil)
	h := cost.NewHistory(histPath)
	c := NewCostLive(r, h, icons.New("emoji"))

	in := &input.StatusLineInput{
		SessionID: "session-xyz",
		Cost:      input.CostInfo{TotalCostUSD: 2.50},
	}

	c.Render(in)

	// Verify the history file was created and has content
	data, err := os.ReadFile(histPath)
	if err != nil {
		t.Fatalf("expected history file to be created, got error: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "session-xyz") {
		t.Errorf("expected session ID in history file, got: %s", content)
	}
	if !strings.Contains(content, "2.5") {
		t.Errorf("expected cost value in history file, got: %s", content)
	}
}

func TestCostLive_Render_SkipsAppendWhenNoSession(t *testing.T) {
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "cost.jsonl")

	r := render.New(nil)
	h := cost.NewHistory(histPath)
	c := NewCostLive(r, h, icons.New("emoji"))

	in := &input.StatusLineInput{
		SessionID: "",
		Cost:      input.CostInfo{TotalCostUSD: 1.00},
	}

	c.Render(in)

	// History file should NOT be created when session ID is empty
	if _, err := os.Stat(histPath); !os.IsNotExist(err) {
		t.Errorf("expected history file to NOT exist when session ID is empty")
	}
}

func TestCostLive_Render_SkipsAppendWhenZeroCost(t *testing.T) {
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "cost.jsonl")

	r := render.New(nil)
	h := cost.NewHistory(histPath)
	c := NewCostLive(r, h, icons.New("emoji"))

	in := &input.StatusLineInput{
		SessionID: "session-abc",
		Cost:      input.CostInfo{TotalCostUSD: 0.0},
	}

	c.Render(in)

	// History file should NOT be created when cost is zero
	if _, err := os.Stat(histPath); !os.IsNotExist(err) {
		t.Errorf("expected history file to NOT exist when cost is zero")
	}
}
