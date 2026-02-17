# claude-statusline

A fast, informative statusline for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) — built in Go.

![claude-statusline example](assets/example.png)

## What It Does

claude-statusline displays live session data directly in your Claude Code terminal: repository status, model info, cost tracking, context usage, and more. It replaces the default statusline with a richer, faster alternative styled with the Catppuccin Mocha theme.

## Features

- **Fast** — Single compiled binary; no subprocess spawning
- **Cached** — AWS Bedrock resolution and version checks avoid repeated lookups
- **Cost tracking** — See spending across four windows: 30-day, 7-day, daily, and live
- **Styled** — Catppuccin Mocha color theme via [lipgloss](https://github.com/charmbracelet/lipgloss)

## Install

### Homebrew (recommended)

```bash
brew install --cask h2ik/tap/claude-statusline
```

### From source

```bash
go build -o claude-statusline .
cp claude-statusline ~/.local/bin/
```

## Setup

Add this to your `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "claude-statusline"
  }
}
```

If the binary isn't on your `PATH`, use the full path:

```json
{
  "statusLine": {
    "type": "command",
    "command": "/Users/YOUR_USER/.local/bin/claude-statusline"
  }
}
```

Claude Code sends JSON to the statusline on each render cycle. The statusline reads that input, resolves model, cost, and repo data, then writes styled output back.

## Layout

The statusline renders four lines of information:

| Line | Content |
|------|---------|
| **1** | Repository path, branch, clean/dirty status, worktree |
| **2** | Model, commits today, submodules, version, time |
| **3** | Cost (30-day, 7-day, daily, live), context window, output style |
| **4** | Burn rate, cache efficiency, block projection, code productivity |

**Line 4** is most useful for direct Anthropic API users. Bedrock users will see empty components where rate-limit data is unavailable.

## Configuration

The statusline reads its config from `~/.claude/statusline/config.toml`. A default file is created on first run.

```toml
# Claude Code Statusline Configuration

[layout]
lines = [
  ["repo_info"],
  ["bedrock_model", "model_info", "commits", "submodules", "version_info", "time_display"],
  ["cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"],
  ["burn_rate", "cache_efficiency", "block_projection", "code_productivity"],
]

[components.bedrock_model]
show_region = true

[components.context_window]
show_tokens = true

[components.code_productivity]
show_velocity = true
show_cost_per_line = true
```

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build -o claude-statusline .
```

Test with sample input:

```bash
echo '{"workspace":{"current_dir":"'$(pwd)'"},"model":{"display_name":"Claude"}}' | ./claude-statusline
```

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.
