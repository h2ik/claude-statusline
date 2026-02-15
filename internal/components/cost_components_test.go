package components

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// ============================================================
// CostMonthly tests
// ============================================================

func TestCostMonthly_Name(t *testing.T) {
	r := render.New()
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostMonthly(r, h)

	if c.Name() != "cost_monthly" {
		t.Errorf("expected 'cost_monthly', got %q", c.Name())
	}
}

func TestCostMonthly_Render_EmptyHistory(t *testing.T) {
	r := render.New()
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostMonthly(r, h)

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
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "cost.jsonl")

	r := render.New()
	h := cost.NewHistory(histPath)

	// Add entries within the 30-day window
	h.Append(cost.Entry{SessionID: "s1", Cost: 1.50, Timestamp: time.Now().Add(-1 * 24 * time.Hour)})
	h.Append(cost.Entry{SessionID: "s2", Cost: 2.75, Timestamp: time.Now().Add(-5 * 24 * time.Hour)})

	c := NewCostMonthly(r, h)
	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$4.25") {
		t.Errorf("expected '$4.25' for summed entries, got: %s", output)
	}
}

// ============================================================
// CostWeekly tests
// ============================================================

func TestCostWeekly_Name(t *testing.T) {
	r := render.New()
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostWeekly(r, h)

	if c.Name() != "cost_weekly" {
		t.Errorf("expected 'cost_weekly', got %q", c.Name())
	}
}

func TestCostWeekly_Render_EmptyHistory(t *testing.T) {
	r := render.New()
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostWeekly(r, h)

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
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "cost.jsonl")

	r := render.New()
	h := cost.NewHistory(histPath)

	// Entry within 7-day window
	h.Append(cost.Entry{SessionID: "s1", Cost: 3.00, Timestamp: time.Now().Add(-2 * 24 * time.Hour)})
	// Entry outside 7-day window (should be excluded)
	h.Append(cost.Entry{SessionID: "s2", Cost: 10.00, Timestamp: time.Now().Add(-10 * 24 * time.Hour)})

	c := NewCostWeekly(r, h)
	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$3.00") {
		t.Errorf("expected '$3.00' (only recent entry), got: %s", output)
	}
}

// ============================================================
// CostDaily tests
// ============================================================

func TestCostDaily_Name(t *testing.T) {
	r := render.New()
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostDaily(r, h)

	if c.Name() != "cost_daily" {
		t.Errorf("expected 'cost_daily', got %q", c.Name())
	}
}

func TestCostDaily_Render_EmptyHistory(t *testing.T) {
	r := render.New()
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostDaily(r, h)

	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$0.00") {
		t.Errorf("expected '$0.00' for empty history, got: %s", output)
	}
	if !strings.Contains(output, "DAY") {
		t.Errorf("expected 'DAY' label in output, got: %s", output)
	}
}

func TestCostDaily_Render_FiltersOldEntries(t *testing.T) {
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "cost.jsonl")

	r := render.New()
	h := cost.NewHistory(histPath)

	// Entry within 24-hour window
	h.Append(cost.Entry{SessionID: "s1", Cost: 0.50, Timestamp: time.Now().Add(-6 * time.Hour)})
	// Entry outside 24-hour window
	h.Append(cost.Entry{SessionID: "s2", Cost: 5.00, Timestamp: time.Now().Add(-48 * time.Hour)})

	c := NewCostDaily(r, h)
	in := &input.StatusLineInput{}

	output := c.Render(in)
	if !strings.Contains(output, "$0.50") {
		t.Errorf("expected '$0.50' (only recent entry), got: %s", output)
	}
}

// ============================================================
// CostLive tests
// ============================================================

func TestCostLive_Name(t *testing.T) {
	r := render.New()
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostLive(r, h)

	if c.Name() != "cost_live" {
		t.Errorf("expected 'cost_live', got %q", c.Name())
	}
}

func TestCostLive_Render_ZeroCost(t *testing.T) {
	r := render.New()
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostLive(r, h)

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
	r := render.New()
	h := cost.NewHistory(filepath.Join(t.TempDir(), "cost.jsonl"))
	c := NewCostLive(r, h)

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

	r := render.New()
	h := cost.NewHistory(histPath)
	c := NewCostLive(r, h)

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

	r := render.New()
	h := cost.NewHistory(histPath)
	c := NewCostLive(r, h)

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

	r := render.New()
	h := cost.NewHistory(histPath)
	c := NewCostLive(r, h)

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
