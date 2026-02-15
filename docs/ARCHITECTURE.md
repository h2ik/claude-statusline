# Architecture

## Overview

Component-based architecture with a registry pattern. Each component implements a simple interface and renders one piece of the statusline.

## Components

All components implement:

```go
type Component interface {
    Name() string
    Render(input *StatusLineInput) string
}
```

Components return styled strings via lipgloss. Empty strings are filtered out.

## Data Flow

1. Read JSON from stdin
2. Parse into `StatusLineInput` struct
3. For each line, call `registry.RenderLine()` with component names
4. Renderer joins components with separators
5. Print to stdout

## Caching

File-based cache at `~/.cache/claude-statusline/`:
- Bedrock model resolution: 24h TTL
- Claude version: 15min TTL

Cost history is persistent (not a cache) at `~/.claude/statusline/costs/history.jsonl`.

## External Commands

- `git` - for repo info, branch, status, commits, submodules, worktree
- `aws` - for Bedrock model resolution (optional)
- `claude` - for version info (optional)

Failures degrade gracefully.
