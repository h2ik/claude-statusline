# claude-statusline

Fast Go-based statusline for Claude Code, replacing the shell-based implementation.

## Features

- **Fast:** Single binary, no subprocess spawning for most operations
- **Cached:** AWS Bedrock resolution and version checks are cached
- **Cost tracking:** Multi-period cost tracking (30day/7day/daily/live)
- **Styled:** Catppuccin Mocha theme via lipgloss

## Installation

### Homebrew (recommended)

```bash
brew install --cask h2ik/tap/claude-statusline
```

The cask installs the `claude-statusline` binary to your system.

### From source

```bash
go build -o claude-statusline .
cp claude-statusline ~/.local/bin/
```

## Claude Code Setup

Add the following to your `~/.claude/settings.json` to enable the statusline:

```json
{
  "statusLine": {
    "type": "command",
    "command": "claude-statusline"
  }
}
```

If you installed from source to a non-PATH location, use the full path instead:

```json
{
  "statusLine": {
    "type": "command",
    "command": "/Users/YOUR_USER/.local/bin/claude-statusline"
  }
}
```

Claude Code pipes JSON to stdin on each render cycle. The statusline reads this
input, resolves model/cost/repo data, and writes styled output to stdout.

## Layout

**Line 1:** Repository info (path, branch, clean/dirty status, worktree)

**Line 2:** Model info, commits today, submodules, version, time

**Line 3:** Cost tracking (30day, 7day, daily, live), context window, output style

#### Line 4: Block Metrics + Code Stats

- `burn_rate` - Current spending velocity ($/min)
- `cache_efficiency` - Cache hit ratio with color coding
- `block_projection` - Rate limit utilization (5h/7d windows)
- `code_productivity` - Lines per minute and cost per line (configurable)

**Note:** Line 4 components are most useful for direct Anthropic API users.
Bedrock users will see graceful degradation (empty components) where rate
limit data is unavailable.

## Configuration

The statusline reads configuration from `~/.claude/statusline/config.toml`.
A default config is created automatically on first run.

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
