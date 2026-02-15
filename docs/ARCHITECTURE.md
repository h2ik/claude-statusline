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

## Cost Tracking

### Transcript Scanning (Period Costs)

Period cost components (30DAY, 7DAY, DAY) compute costs by scanning Claude Code's native JSONL transcript files at `~/.claude/projects/`. This approach reads the authoritative source of token usage rather than relying on self-reported cost values.

**Pipeline:** `ScanTranscripts` walks the directory tree → `scanFile` reads each `.jsonl` → `parseTranscriptEntry` extracts assistant messages → `CalculateEntryCost` applies per-model pricing → totals accumulate.

**Optimizations:**
- **mtime pre-filtering:** Files not modified within the target duration are skipped without being opened
- **tool-results exclusion:** `tool-results/` subdirectories are skipped via `filepath.SkipDir`
- **5-minute TTL cache:** `TranscriptScanner` caches computed totals per-duration via the file-based cache, avoiding repeated filesystem walks

**Pricing:** `ModelPrice()` resolves rates via exact match → prefix match → Sonnet-tier default. Rates cover input, output, cache write, and cache read tokens per million.

### Live Session Cost

`CostLive` continues using `History` (append-only JSONL at `~/.claude/statusline/costs/history.jsonl`) to display the current session's cost as reported by Claude Code's stdin JSON.

## Caching

File-based cache at `~/.cache/claude-statusline/`:
- Bedrock model resolution: 24h TTL
- Claude version: 15min TTL
- Transcript cost totals: 5min TTL (per duration)

## Configuration

TOML config at `~/.claude/statusline/config.toml`:
- **Layout control:** which components appear on which lines via `[layout] lines`
- **Per-component toggles:** `show_region` (bedrock_model), `show_tokens` (context_window)
- **Auto-generated** with sensible defaults on first run
- **Parsed by** `github.com/BurntSushi/toml`

`main.go` builds layout lines from `cfg.Layout.Lines`. Components query their settings via `cfg.GetBool()`.

## External Commands

- `git` - for repo info, branch, status, commits, submodules, worktree
- `aws` - for Bedrock model resolution (optional)
- `claude` - for version info (optional)

Failures degrade gracefully.
