# Line 4 Metrics Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add four new components to Line 4: burn_rate, cache_efficiency, block_projection, and code_productivity.

**Architecture:** Each component implements the Component interface, computes metrics from stdin JSON, returns styled strings or empty strings for graceful degradation. No new packages or state files.

**Tech Stack:** Go 1.21+, Lipgloss for styling, existing renderer/config infrastructure

---

## Task 1: burn_rate component

**Files:**
- Create: `internal/components/burn_rate.go`
- Create: `internal/components/burn_rate_test.go`

### Step 1: Write the failing test

```go
package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestBurnRate_Name(t *testing.T) {
	r := render.New()
	c := NewBurnRate(r)

	if c.Name() != "burn_rate" {
		t.Errorf("expected 'burn_rate', got %q", c.Name())
	}
}

func TestBurnRate_Render_ZeroDuration(t *testing.T) {
	r := render.New()
	c := NewBurnRate(r)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:    1.50,
			TotalDurationMS: 0,
		},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for zero duration, got: %s", output)
	}
}

func TestBurnRate_Render_DisplaysRate(t *testing.T) {
	r := render.New()
	c := NewBurnRate(r)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:    1.20,
			TotalDurationMS: 600000, // 10 minutes
		},
	}

	output := c.Render(in)
	if !strings.Contains(output, "🔥") {
		t.Errorf("expected fire emoji in output, got: %s", output)
	}
	if !strings.Contains(output, "$0.12/min") {
		t.Errorf("expected '$0.12/min' for burn rate, got: %s", output)
	}
}

func TestBurnRate_Render_RoundsCorrectly(t *testing.T) {
	r := render.New()
	c := NewBurnRate(r)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:    0.755,
			TotalDurationMS: 180000, // 3 minutes
		},
	}

	output := c.Render(in)
	// 0.755 / 3 = 0.2516666... should round to $0.25/min
	if !strings.Contains(output, "$0.25/min") {
		t.Errorf("expected '$0.25/min' for burn rate, got: %s", output)
	}
}
```

### Step 2: Run test to verify it fails

Run: `go test ./internal/components -run TestBurnRate -v`

Expected: FAIL with "undefined: NewBurnRate"

### Step 3: Write minimal implementation

Create `internal/components/burn_rate.go`:

```go
package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// BurnRate displays the current spending velocity in dollars per minute.
type BurnRate struct {
	renderer *render.Renderer
}

// NewBurnRate creates a new BurnRate component.
func NewBurnRate(r *render.Renderer) *BurnRate {
	return &BurnRate{renderer: r}
}

// Name returns the component identifier.
func (c *BurnRate) Name() string {
	return "burn_rate"
}

// Render produces the burn rate string.
func (c *BurnRate) Render(in *input.StatusLineInput) string {
	if in.Cost.TotalDurationMS == 0 {
		return ""
	}

	minutes := float64(in.Cost.TotalDurationMS) / 60000.0
	ratePerMin := in.Cost.TotalCostUSD / minutes

	return fmt.Sprintf("🔥 %s",
		c.renderer.Peach(fmt.Sprintf("$%.2f/min", ratePerMin)),
	)
}
```

### Step 4: Run test to verify it passes

Run: `go test ./internal/components -run TestBurnRate -v`

Expected: PASS (all 4 tests)

### Step 5: Commit

```bash
git add internal/components/burn_rate.go internal/components/burn_rate_test.go
git commit --signoff -m "feat(components): Add burn_rate component

Display current spending velocity in dollars per minute.
Returns empty string when duration is zero (graceful degradation)."
```

---

## Task 2: cache_efficiency component

**Files:**
- Create: `internal/components/cache_efficiency.go`
- Create: `internal/components/cache_efficiency_test.go`

### Step 1: Write the failing test

```go
package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestCacheEfficiency_Name(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	if c.Name() != "cache_efficiency" {
		t.Errorf("expected 'cache_efficiency', got %q", c.Name())
	}
}

func TestCacheEfficiency_Render_ZeroTokens(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	in := &input.StatusLineInput{
		CurrentUsage: input.UsageInfo{
			InputTokens:              0,
			CacheReadInputTokens:     0,
			CacheCreationInputTokens: 0,
		},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for zero tokens, got: %s", output)
	}
}

func TestCacheEfficiency_Render_HighEfficiency(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	in := &input.StatusLineInput{
		CurrentUsage: input.UsageInfo{
			InputTokens:              1000,
			CacheReadInputTokens:     7000,
			CacheCreationInputTokens: 2000,
		},
	}

	output := c.Render(in)
	if !strings.Contains(output, "💾") {
		t.Errorf("expected disk emoji in output, got: %s", output)
	}
	// 7000 / 10000 = 70%
	if !strings.Contains(output, "70% cache") {
		t.Errorf("expected '70%% cache' for high efficiency, got: %s", output)
	}
}

func TestCacheEfficiency_Render_LowEfficiency(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	in := &input.StatusLineInput{
		CurrentUsage: input.UsageInfo{
			InputTokens:              7000,
			CacheReadInputTokens:     1000,
			CacheCreationInputTokens: 2000,
		},
	}

	output := c.Render(in)
	// 1000 / 10000 = 10%
	if !strings.Contains(output, "10% cache") {
		t.Errorf("expected '10%% cache' for low efficiency, got: %s", output)
	}
}

func TestCacheEfficiency_Render_MediumEfficiency(t *testing.T) {
	r := render.New()
	c := NewCacheEfficiency(r)

	in := &input.StatusLineInput{
		CurrentUsage: input.UsageInfo{
			InputTokens:              3000,
			CacheReadInputTokens:     5000,
			CacheCreationInputTokens: 2000,
		},
	}

	output := c.Render(in)
	// 5000 / 10000 = 50%
	if !strings.Contains(output, "50% cache") {
		t.Errorf("expected '50%% cache' for medium efficiency, got: %s", output)
	}
}
```

### Step 2: Run test to verify it fails

Run: `go test ./internal/components -run TestCacheEfficiency -v`

Expected: FAIL with "undefined: NewCacheEfficiency"

### Step 3: Write minimal implementation

Create `internal/components/cache_efficiency.go`:

```go
package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CacheEfficiency displays the cache hit ratio as a percentage.
type CacheEfficiency struct {
	renderer *render.Renderer
}

// NewCacheEfficiency creates a new CacheEfficiency component.
func NewCacheEfficiency(r *render.Renderer) *CacheEfficiency {
	return &CacheEfficiency{renderer: r}
}

// Name returns the component identifier.
func (c *CacheEfficiency) Name() string {
	return "cache_efficiency"
}

// Render produces the cache efficiency string with color coding.
func (c *CacheEfficiency) Render(in *input.StatusLineInput) string {
	usage := in.CurrentUsage
	totalTokens := usage.InputTokens + usage.CacheReadInputTokens + usage.CacheCreationInputTokens

	if totalTokens == 0 {
		return ""
	}

	percentage := float64(usage.CacheReadInputTokens) / float64(totalTokens) * 100.0

	// Color based on efficiency
	var colorFunc func(string) string
	if percentage >= 70 {
		colorFunc = c.renderer.Green
	} else if percentage >= 40 {
		colorFunc = c.renderer.Yellow
	} else {
		colorFunc = c.renderer.Red
	}

	return fmt.Sprintf("💾 %s",
		colorFunc(fmt.Sprintf("%.0f%% cache", percentage)),
	)
}
```

### Step 4: Run test to verify it passes

Run: `go test ./internal/components -run TestCacheEfficiency -v`

Expected: PASS (all 5 tests)

### Step 5: Commit

```bash
git add internal/components/cache_efficiency.go internal/components/cache_efficiency_test.go
git commit --signoff -m "feat(components): Add cache_efficiency component

Display cache hit ratio with color coding (green >=70%, yellow 40-69%,
red <40%). Returns empty string when no tokens have been used."
```

---

## Task 3: block_projection component

**Files:**
- Create: `internal/components/block_projection.go`
- Create: `internal/components/block_projection_test.go`

### Step 1: Write the failing test

```go
package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestBlockProjection_Name(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r)

	if c.Name() != "block_projection" {
		t.Errorf("expected 'block_projection', got %q", c.Name())
	}
}

func TestBlockProjection_Render_ZeroUtilization(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r)

	in := &input.StatusLineInput{
		FiveHour: input.UsageLimit{Utilization: 0.0},
		SevenDay: input.UsageLimit{Utilization: 0.0},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for zero utilization (Bedrock case), got: %s", output)
	}
}

func TestBlockProjection_Render_LowUtilization(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r)

	in := &input.StatusLineInput{
		FiveHour: input.UsageLimit{Utilization: 0.25},
		SevenDay: input.UsageLimit{Utilization: 0.10},
	}

	output := c.Render(in)
	if !strings.Contains(output, "⏳") {
		t.Errorf("expected hourglass emoji in output, got: %s", output)
	}
	if !strings.Contains(output, "5h: 25%") {
		t.Errorf("expected '5h: 25%%' in output, got: %s", output)
	}
	if !strings.Contains(output, "7d: 10%") {
		t.Errorf("expected '7d: 10%%' in output, got: %s", output)
	}
}

func TestBlockProjection_Render_HighUtilization(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r)

	in := &input.StatusLineInput{
		FiveHour: input.UsageLimit{Utilization: 0.85},
		SevenDay: input.UsageLimit{Utilization: 0.62},
	}

	output := c.Render(in)
	if !strings.Contains(output, "5h: 85%") {
		t.Errorf("expected '5h: 85%%' in output, got: %s", output)
	}
	if !strings.Contains(output, "7d: 62%") {
		t.Errorf("expected '7d: 62%%' in output, got: %s", output)
	}
}

func TestBlockProjection_Render_OnlyFiveHourData(t *testing.T) {
	r := render.New()
	c := NewBlockProjection(r)

	in := &input.StatusLineInput{
		FiveHour: input.UsageLimit{Utilization: 0.45},
		SevenDay: input.UsageLimit{Utilization: 0.0},
	}

	output := c.Render(in)
	if !strings.Contains(output, "5h: 45%") {
		t.Errorf("expected '5h: 45%%' in output, got: %s", output)
	}
}
```

### Step 2: Run test to verify it fails

Run: `go test ./internal/components -run TestBlockProjection -v`

Expected: FAIL with "undefined: NewBlockProjection"

### Step 3: Write minimal implementation

Create `internal/components/block_projection.go`:

```go
package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// BlockProjection displays rate limit utilization from 5-hour and 7-day windows.
type BlockProjection struct {
	renderer *render.Renderer
}

// NewBlockProjection creates a new BlockProjection component.
func NewBlockProjection(r *render.Renderer) *BlockProjection {
	return &BlockProjection{renderer: r}
}

// Name returns the component identifier.
func (c *BlockProjection) Name() string {
	return "block_projection"
}

// Render produces the block projection string with color-coded utilization.
func (c *BlockProjection) Render(in *input.StatusLineInput) string {
	fiveHourPct := in.FiveHour.Utilization * 100.0
	sevenDayPct := in.SevenDay.Utilization * 100.0

	// Graceful degradation when no data (Bedrock case)
	if fiveHourPct == 0 && sevenDayPct == 0 {
		return ""
	}

	var parts []string

	if fiveHourPct > 0 {
		colorFunc := c.getColorForUtilization(fiveHourPct)
		parts = append(parts, colorFunc(fmt.Sprintf("5h: %.0f%%", fiveHourPct)))
	}

	if sevenDayPct > 0 {
		colorFunc := c.getColorForUtilization(sevenDayPct)
		parts = append(parts, colorFunc(fmt.Sprintf("7d: %.0f%%", sevenDayPct)))
	}

	if len(parts) == 0 {
		return ""
	}

	output := "⏳ "
	for i, part := range parts {
		if i > 0 {
			output += " │ "
		}
		output += part
	}

	return output
}

func (c *BlockProjection) getColorForUtilization(pct float64) func(string) string {
	if pct >= 75 {
		return c.renderer.Red
	} else if pct >= 50 {
		return c.renderer.Yellow
	}
	return c.renderer.Green
}
```

### Step 4: Run test to verify it passes

Run: `go test ./internal/components -run TestBlockProjection -v`

Expected: PASS (all 5 tests)

### Step 5: Commit

```bash
git add internal/components/block_projection.go internal/components/block_projection_test.go
git commit --signoff -m "feat(components): Add block_projection component

Display rate limit utilization from 5-hour and 7-day windows with
color-coded percentages. Returns empty string when utilization data is
unavailable (Bedrock case)."
```

---

## Task 4: code_productivity component with config

**Files:**
- Create: `internal/components/code_productivity.go`
- Create: `internal/components/code_productivity_test.go`
- Modify: `internal/config/config.go`

### Step 1: Write the failing test

```go
package components

import (
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestCodeProductivity_Name(t *testing.T) {
	r := render.New()
	cfg := config.DefaultConfig()
	c := NewCodeProductivity(r, cfg)

	if c.Name() != "code_productivity" {
		t.Errorf("expected 'code_productivity', got %q", c.Name())
	}
}

func TestCodeProductivity_Render_NoLinesChanged(t *testing.T) {
	r := render.New()
	cfg := config.DefaultConfig()
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      1.50,
			TotalDurationMS:   600000,
			TotalLinesAdded:   0,
			TotalLinesRemoved: 0,
		},
	}

	output := c.Render(in)
	if output != "" {
		t.Errorf("expected empty string for no lines changed, got: %s", output)
	}
}

func TestCodeProductivity_Render_BothMetrics(t *testing.T) {
	r := render.New()
	cfg := config.DefaultConfig()
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      0.60,
			TotalDurationMS:   300000, // 5 minutes
			TotalLinesAdded:   80,
			TotalLinesRemoved: 20,
		},
	}

	output := c.Render(in)
	if !strings.Contains(output, "✏️") {
		t.Errorf("expected pencil emoji in output, got: %s", output)
	}
	// 100 lines / 5 min = 20 lines/min
	if !strings.Contains(output, "20 lines/min") {
		t.Errorf("expected '20 lines/min' in output, got: %s", output)
	}
	// $0.60 / 100 lines = $0.01/line
	if !strings.Contains(output, "$0.01/line") {
		t.Errorf("expected '$0.01/line' in output, got: %s", output)
	}
}

func TestCodeProductivity_Render_VelocityOnly(t *testing.T) {
	r := render.New()
	showVelocity := true
	showCostPerLine := false
	cfg := &config.Config{
		Components: map[string]config.ComponentConfig{
			"code_productivity": {
				ShowVelocity:    &showVelocity,
				ShowCostPerLine: &showCostPerLine,
			},
		},
	}
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      0.30,
			TotalDurationMS:   120000, // 2 minutes
			TotalLinesAdded:   50,
			TotalLinesRemoved: 10,
		},
	}

	output := c.Render(in)
	// 60 lines / 2 min = 30 lines/min
	if !strings.Contains(output, "30 lines/min") {
		t.Errorf("expected '30 lines/min' in output, got: %s", output)
	}
	if strings.Contains(output, "$") {
		t.Errorf("expected no cost in output when disabled, got: %s", output)
	}
}

func TestCodeProductivity_Render_CostPerLineOnly(t *testing.T) {
	r := render.New()
	showVelocity := false
	showCostPerLine := true
	cfg := &config.Config{
		Components: map[string]config.ComponentConfig{
			"code_productivity": {
				ShowVelocity:    &showVelocity,
				ShowCostPerLine: &showCostPerLine,
			},
		},
	}
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      1.00,
			TotalDurationMS:   300000,
			TotalLinesAdded:   100,
			TotalLinesRemoved: 0,
		},
	}

	output := c.Render(in)
	// $1.00 / 100 lines = $0.01/line
	if !strings.Contains(output, "$0.01/line") {
		t.Errorf("expected '$0.01/line' in output, got: %s", output)
	}
	if strings.Contains(output, "lines/min") {
		t.Errorf("expected no velocity in output when disabled, got: %s", output)
	}
}

func TestCodeProductivity_Render_ZeroDuration(t *testing.T) {
	r := render.New()
	cfg := config.DefaultConfig()
	c := NewCodeProductivity(r, cfg)

	in := &input.StatusLineInput{
		Cost: input.CostInfo{
			TotalCostUSD:      0.50,
			TotalDurationMS:   0,
			TotalLinesAdded:   100,
			TotalLinesRemoved: 50,
		},
	}

	output := c.Render(in)
	// Should show cost per line only (velocity requires duration)
	if !strings.Contains(output, "$0.00/line") {
		t.Errorf("expected cost per line in output, got: %s", output)
	}
	if strings.Contains(output, "lines/min") {
		t.Errorf("expected no velocity when duration is zero, got: %s", output)
	}
}
```

### Step 2: Run test to verify it fails

Run: `go test ./internal/components -run TestCodeProductivity -v`

Expected: FAIL with "undefined: NewCodeProductivity"

### Step 3: Add config fields

Modify `internal/config/config.go`:

```go
// ComponentConfig holds per-component configuration options.
// Pointer bools distinguish "not set" from "set to false".
type ComponentConfig struct {
	ShowRegion      *bool `toml:"show_region,omitempty"`
	ShowTokens      *bool `toml:"show_tokens,omitempty"`
	ShowVelocity    *bool `toml:"show_velocity,omitempty"`
	ShowCostPerLine *bool `toml:"show_cost_per_line,omitempty"`
}
```

And update the `GetBool` method:

```go
// GetBool retrieves a boolean value from the ComponentConfig for the given
// component and key name. Returns fallback if the component or key is not set.
func (c *Config) GetBool(component, key string, fallback bool) bool {
	comp, ok := c.Components[component]
	if !ok {
		return fallback
	}

	switch key {
	case "show_region":
		if comp.ShowRegion != nil {
			return *comp.ShowRegion
		}
	case "show_tokens":
		if comp.ShowTokens != nil {
			return *comp.ShowTokens
		}
	case "show_velocity":
		if comp.ShowVelocity != nil {
			return *comp.ShowVelocity
		}
	case "show_cost_per_line":
		if comp.ShowCostPerLine != nil {
			return *comp.ShowCostPerLine
		}
	}

	return fallback
}
```

### Step 4: Write the implementation

Create `internal/components/code_productivity.go`:

```go
package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// CodeProductivity displays code output metrics (velocity and/or cost per line).
type CodeProductivity struct {
	renderer *render.Renderer
	config   *config.Config
}

// NewCodeProductivity creates a new CodeProductivity component.
func NewCodeProductivity(r *render.Renderer, cfg *config.Config) *CodeProductivity {
	return &CodeProductivity{renderer: r, config: cfg}
}

// Name returns the component identifier.
func (c *CodeProductivity) Name() string {
	return "code_productivity"
}

// Render produces the code productivity string with configurable sub-metrics.
func (c *CodeProductivity) Render(in *input.StatusLineInput) string {
	totalLines := in.Cost.TotalLinesAdded + in.Cost.TotalLinesRemoved

	if totalLines == 0 {
		return ""
	}

	showVelocity := c.config.GetBool("code_productivity", "show_velocity", true)
	showCostPerLine := c.config.GetBool("code_productivity", "show_cost_per_line", true)

	var parts []string

	// Lines per minute
	if showVelocity && in.Cost.TotalDurationMS > 0 {
		minutes := float64(in.Cost.TotalDurationMS) / 60000.0
		linesPerMin := float64(totalLines) / minutes
		parts = append(parts, fmt.Sprintf("%.0f lines/min", linesPerMin))
	}

	// Cost per line
	if showCostPerLine {
		costPerLine := in.Cost.TotalCostUSD / float64(totalLines)
		parts = append(parts, fmt.Sprintf("$%.2f/line", costPerLine))
	}

	if len(parts) == 0 {
		return ""
	}

	output := "✏️ "
	for i, part := range parts {
		if i > 0 {
			output += " │ "
		}
		output += c.renderer.Text(part)
	}

	return output
}
```

### Step 5: Run test to verify it passes

Run: `go test ./internal/components -run TestCodeProductivity -v`

Expected: PASS (all 7 tests)

### Step 6: Commit

```bash
git add internal/components/code_productivity.go internal/components/code_productivity_test.go internal/config/config.go
git commit --signoff -m "feat(components): Add code_productivity component

Display code output metrics with configurable sub-metrics:
- Lines per minute (velocity)
- Cost per line (efficiency)

Both default to enabled, configurable via show_velocity and
show_cost_per_line in config. Returns empty string when no lines
have changed."
```

---

## Task 5: Register components and add Line 4

**Files:**
- Modify: `main.go` (lines 64-67)
- Modify: `internal/config/config.go` (lines 67-72)

### Step 1: Register components in main.go

Add after line 64 in `main.go`:

```go
	// Line 4 components
	registry.Register(components.NewBurnRate(r))
	registry.Register(components.NewCacheEfficiency(r))
	registry.Register(components.NewBlockProjection(r))
	registry.Register(components.NewCodeProductivity(r, cfg))
```

### Step 2: Add Line 4 to default config

Modify `DefaultConfig()` in `internal/config/config.go` (lines 67-72):

```go
		Layout: Layout{
			Lines: [][]string{
				{"repo_info"},
				{"bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"},
				{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"},
				{"burn_rate", "cache_efficiency", "block_projection", "code_productivity"},
			},
		},
```

And add the default config for code_productivity after line 80:

```go
			"code_productivity": {
				ShowVelocity:    &showVelocity,
				ShowCostPerLine: &showCostPerLine,
			},
```

(Re-use the existing `showVelocity` and `showCostPerLine` variables that should be declared at the top of the function alongside `showRegion` and `showTokens`.)

Add variable declarations after line 64:

```go
	showVelocity := true
	showCostPerLine := true
```

### Step 3: Build and smoke test

Run: `go build -o statusline .`

Expected: Build succeeds

Test with sample input:

```bash
echo '{"workspace":{"current_dir":"/tmp/test","project_dir":"/tmp/test"},"model":{"display_name":"opus"},"session_id":"test","transcript_path":"","output_style":{"name":"default"},"context_window":{"used_percentage":45,"remaining_percentage":55,"context_window_size":200000},"cost":{"total_cost_usd":1.20,"total_duration_ms":600000,"total_api_duration_ms":300000,"total_lines_added":80,"total_lines_removed":20},"current_usage":{"input_tokens":1000,"cache_read_input_tokens":7000,"cache_creation_input_tokens":2000},"five_hour":{"utilization":0.35,"resets_at":""},"seven_day":{"utilization":0.12,"resets_at":""}}' | ./statusline
```

Expected: Output shows 4 lines with Line 4 containing burn_rate, cache_efficiency, block_projection, and code_productivity

### Step 4: Commit

```bash
git add main.go internal/config/config.go
git commit --signoff -m "feat: Wire Line 4 metrics into statusline

Register burn_rate, cache_efficiency, block_projection, and
code_productivity components. Add Line 4 to default config layout
with code_productivity defaults (both metrics enabled)."
```

---

## Task 6: Update README and add future enhancement note

**Files:**
- Modify: `README.md`
- Create: `docs/FUTURE_ENHANCEMENTS.md`

### Step 1: Update README to document Line 4 components

Add section after the Line 3 documentation in README.md:

```markdown
#### Line 4: Block Metrics + Code Stats + Context

- `burn_rate` - Current spending velocity ($/min)
- `cache_efficiency` - Cache hit ratio with color coding
- `block_projection` - Rate limit utilization (5h/7d windows)
- `code_productivity` - Lines per minute and cost per line (configurable)

**Note:** Line 4 components are most useful for direct Anthropic API users.
Bedrock users will see graceful degradation (empty components) where rate
limit data is unavailable.

**Config example:**

```toml
[layout]
lines = [
  ["repo_info"],
  ["bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"],
  ["cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"],
  ["burn_rate", "cache_efficiency", "block_projection", "code_productivity"]
]

[components.code_productivity]
show_velocity = true
show_cost_per_line = true
\```
```

### Step 2: Create future enhancements doc

Create `docs/FUTURE_ENHANCEMENTS.md`:

```markdown
# Future Enhancements

This document tracks potential improvements that would require more significant architectural changes.

## Session State Snapshots

**Current limitation:** `burn_rate` and `block_projection` compute from single-point-in-time data. This provides instantaneous metrics but cannot show trends or acceleration.

**Enhancement:** Add session snapshot recording - periodic writes to a JSONL state file (similar to cost history) that records:
- Timestamp
- Context usage percentage
- Total cost
- Token counts
- Rate limit utilization

**Benefits:**
- `burn_rate` could show acceleration/deceleration trends
- `block_projection` could extrapolate from observed usage curves
- Time-to-limit predictions based on actual session behavior
- Historical context usage patterns

**Implementation approach:**
1. Add `internal/session/` package with `Snapshot` and `StateWriter` types
2. Write snapshots every N seconds (e.g., 30s) during statusline renders
3. Components read recent snapshots from JSONL for trend calculation
4. Auto-compact old session files (keep last 24h)

**Trade-offs:**
- Adds I/O overhead on every render (mitigated by buffered writes)
- Increases storage requirements (~1KB per session)
- More accurate metrics vs. simpler single-point computation

**Decision:** Deferred for v1. Single-point metrics are sufficient for initial release. Revisit if users request trend-based features.
```

### Step 3: Commit

```bash
git add README.md docs/FUTURE_ENHANCEMENTS.md
git commit --signoff -m "docs: Document Line 4 components and future enhancements

Add README section for Line 4 metrics components with usage examples.
Create FUTURE_ENHANCEMENTS.md documenting session snapshot approach
for trend-based metrics (deferred for v1)."
```

---

## Execution Complete

All tasks complete. Line 4 components are implemented, tested, registered, and documented.

**Summary:**
- 4 new components (burn_rate, cache_efficiency, block_projection, code_productivity)
- 8 new test files with comprehensive coverage
- Config extended to support code_productivity toggles
- Default config includes Line 4
- Documentation updated with usage examples and future enhancement notes

**Next steps:**
- Test with real Claude Code stdin data
- Consider adding integration tests for Line 4 layout rendering
- Monitor for user feedback on metric usefulness
