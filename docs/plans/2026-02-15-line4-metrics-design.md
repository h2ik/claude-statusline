# Line 4: Block Metrics, Code Stats, and Context

## Summary

Add four new components to a fourth statusline row: `burn_rate`,
`cache_efficiency`, `block_projection`, and `code_productivity`. Each
component computes its value from the existing Claude Code stdin JSON
contract — no new external data sources required.

These components are most useful for direct Anthropic API users. Bedrock
users will see graceful degradation (empty strings) where data is
unavailable.

## Components

### burn_rate

Displays current spending velocity in dollars per minute.

- **Formula:** `total_cost_usd / (total_duration_ms / 60000)`
- **Display:** `🔥 $0.12/min` (Peach color)
- **Guard:** Returns empty string when `total_duration_ms == 0`

### cache_efficiency

Displays the cache hit ratio as a percentage of total token usage.

- **Formula:** `cache_read_tokens / (input_tokens + cache_read_tokens + cache_creation_tokens) * 100`
- **Display:** `💾 87% cache`
- **Color coding:**
  - Green: >= 70%
  - Yellow: 40-69%
  - Red: < 40%
- **Guard:** Returns empty string when all token counts are zero

### block_projection

Displays rate limit utilization from the 5-hour and 7-day windows.

- **Data:** `five_hour.utilization`, `seven_day.utilization`
- **Display:** `⏳ 5h: 45% │ 7d: 12%`
- **Color coding per utilization value:**
  - Green: < 50%
  - Yellow: 50-74%
  - Red: >= 75%
- **Guard:** Returns empty string when both utilization values are zero
  (Bedrock API case)

### code_productivity

Displays code output metrics with two configurable sub-metrics.

- **Lines per minute:** `(lines_added + lines_removed) / (total_duration_ms / 60000)`
- **Cost per line:** `total_cost_usd / (lines_added + lines_removed)`
- **Display:** `✏️ 12 lines/min │ $0.03/line`
- **Config toggles:**
  - `show_velocity` (default: true) — lines per minute
  - `show_cost_per_line` (default: true) — dollars per line
- **Guard:** Returns empty string when no lines have changed

## Configuration

### Default TOML layout

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
```

### ComponentConfig additions

```go
ShowVelocity    *bool `toml:"show_velocity,omitempty"`
ShowCostPerLine *bool `toml:"show_cost_per_line,omitempty"`
```

## Architecture

Each component follows the existing pattern:

1. Implements the `Component` interface (`Name()`, `Render()`)
2. Receives `*render.Renderer` (and config where needed) via constructor
3. Returns a styled string or empty string for graceful degradation
4. Registered in `main.go`, added to Line 4 in the default layout

No new packages, no state files, no external dependencies.

## Future Enhancement

For more accurate burn rate trending and block projection, consider adding
session snapshot recording: periodic writes to a JSONL state file that
enable time-series analysis rather than single-point-in-time computation.
This would allow burn rate to show acceleration/deceleration and block
projection to extrapolate from observed trends.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| State model | Single-point math | Simpler, no I/O overhead, sufficient for v1 |
| Architecture | Independent components | Matches existing pattern, composable, testable |
| Bedrock gaps | Graceful empty | Consistent with other component behavior |
| Productivity metrics | Both with toggle | Users choose what matters to them |
| Thoughts component | Dropped | No mapping in stdin contract |
