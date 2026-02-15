# Claude Statusline: Go Rewrite Design

## Problem

The existing shell-based statusline (`claude-code-statusline`) spawns dozens of
subprocesses per render: `jq`, `git`, `aws`, `curl`, `date`, plus bash
function overhead. Each render is slow enough to notice.

## Goal

Rewrite the three lines of statusline output the user actually displays into a
single Go binary. Use `lipgloss` for styling. Keep it fast, small, and simple.

## Data Flow

```
Claude Code --> stdin (JSON) --> Go binary --> components --> lipgloss --> stdout
```

1. Claude Code pipes a JSON payload to the binary's stdin.
2. The binary unmarshals it into a typed struct.
3. Each registered component receives the struct and returns a styled string.
4. The renderer joins components per line with separators and prints to stdout.

## Input Contract

Claude Code sends JSON to stdin. Required field: `workspace.current_dir`. All
other fields degrade gracefully when absent.

```json
{
  "workspace": { "current_dir": "/path/to/repo" },
  "model": { "display_name": "Claude Opus 4.6" },
  "session_id": "uuid",
  "output_style": { "name": "default" },
  "context_window": {
    "used_percentage": 45,
    "remaining_percentage": 55,
    "context_window_size": 200000
  },
  "cost": {
    "total_cost_usd": 0.45,
    "total_lines_added": 120,
    "total_lines_removed": 30
  },
  "current_usage": {
    "input_tokens": 10000,
    "cache_read_input_tokens": 5000,
    "cache_creation_input_tokens": 2000
  }
}
```

## Project Structure

```
claude-statusline/
  main.go
  go.mod / go.sum
  internal/
    input/
      input.go            # JSON structs + stdin parsing
    component/
      component.go        # Component interface + registry
      registry.go         # Line definitions, component wiring
    render/
      render.go           # Lipgloss theme (Catppuccin Mocha), line builder
    cache/
      cache.go            # File-based cache with TTL
    git/
      git.go              # Branch, status, commits, submodules, worktree
    cost/
      cost.go             # Session + multi-period calculations
      history.go          # Append-only JSONL history file
    components/
      repo_info.go        # Line 1
      bedrock_model.go    # Line 2
      commits.go          # Line 2
      submodules.go       # Line 2
      version_info.go     # Line 2
      time_display.go     # Line 2
      cost_monthly.go     # Line 3
      cost_weekly.go      # Line 3
      cost_daily.go       # Line 3
      cost_live.go        # Line 3
      context_window.go   # Line 3
      session_mode.go     # Line 3
```

## Component Interface

```go
type Component interface {
    Name() string
    Render(input *input.StatusLineInput) string
}
```

Each component returns a lipgloss-styled string. A nil or empty return means
the component has nothing to display and the renderer skips it.

## Line Layout (Hardcoded for Now)

| Line | Components | Separator |
|------|-----------|-----------|
| 1 | repo_info | ` \| ` |
| 2 | bedrock_model, commits, submodules, version_info, time_display | ` \| ` |
| 3 | cost_monthly, cost_weekly, cost_daily, cost_live, context_window, session_mode | ` \| ` |

This layout is hardcoded initially. The registry makes it trivial to
rearrange later.

## Component Details

### Line 1: repo_info

Runs `git rev-parse` and `git status --porcelain` in `workspace.current_dir`.
Displays directory path (with `~` substitution), branch name, clean/dirty
indicator, and worktree marker when applicable.

Output: `~/myproject (main) [checkmark]`

### Line 2: bedrock_model

Checks whether `.model.display_name` contains a Bedrock ARN. If so, shells
out to `aws bedrock get-inference-profile` to resolve it. Caches the result
for 24 hours. Falls back to the raw model name when the CLI is unavailable.

Output: `[brain] Claude Opus 4.6 (us-east-2)`

### Line 2: commits

Runs `git rev-list --count --since="today 00:00"` in the current directory.

Output: `Commits: 5`

### Line 2: submodules

Runs `git submodule status`. Hidden when count is zero.

Output: `SUB: 2` or empty

### Line 2: version_info

Runs `claude --version` with a 15-minute file cache.

Output: `CC:1.0.27`

### Line 2: time_display

Uses Go's `time.Now().Format("15:04")`. No external commands.

Output: `[clock] 14:45`

### Line 3: cost_monthly, cost_weekly, cost_daily

Read from the cost history file at `~/.claude/statusline/costs/history.jsonl`.
Each invocation appends the current session's cost entry (deduplicated by
session ID). Calculations scan entries within the relevant window.

Output: `30DAY $12.45`, `7DAY $5.23`, `DAY $1.87`

### Line 3: cost_live

Reads `.cost.total_cost_usd` directly from the JSON input.

Output: `[fire]LIVE $0.45`

### Line 3: context_window

Reads `.context_window.used_percentage` from input. Color shifts from green
to yellow at 50%, red at 75%, with a warning indicator at 90%.

Output: `[brain] 45% (90K/200K)`

### Line 3: session_mode

Reads `.output_style.name` from input. Maps known styles to emojis.

Output: `[book] Style: explanatory`

## Theme: Catppuccin Mocha (Hardcoded)

| Role | Color | Hex |
|------|-------|-----|
| Dimmed/labels | Overlay0 | `#6c7086` |
| Text/values | Text | `#cdd6f4` |
| Green (clean) | Green | `#a6e3a1` |
| Red (critical) | Red | `#f38ba8` |
| Yellow (warning) | Yellow | `#f9e2af` |
| Blue (paths) | Blue | `#89b4fa` |
| Mauve (accent) | Mauve | `#cba6f7` |
| Peach (costs) | Peach | `#fab387` |
| Teal (secondary) | Teal | `#94e2d5` |

## Caching

File-based cache in `~/.cache/claude-statusline/`. Each entry is a file whose
mtime determines freshness.

| Item | TTL | Key |
|------|-----|-----|
| Bedrock model | 24 hours | `bedrock_<sha256(arn)>` |
| Claude version | 15 minutes | `claude_version` |

Cost history is append-only persistent data, not a cache. It lives at
`~/.claude/statusline/costs/history.jsonl`.

## Error Handling

- Missing JSON fields return zero values; components render fallbacks or
  return empty strings.
- Git failures (not a repo, command missing) degrade to showing just the
  directory path.
- AWS CLI unavailable: display the raw model name from JSON input.
- Cost history file missing: show `$0.00` for all period costs.
- Component panics: recover, log to stderr, return empty string. No single
  component failure should block the others.

## Dependencies

- `github.com/charmbracelet/lipgloss` for terminal styling
- Go standard library for everything else
- External commands: `git`, `aws` (optional), `claude` (optional)

## Future Considerations

- The hardcoded 3-line layout will likely change. The component registry and
  interface make rearrangement straightforward.
- Config-driven themes or layout could be added later via a small TOML file.
- Additional components can implement the `Component` interface and register
  themselves without touching existing code.
