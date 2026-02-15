# claude-statusline

Fast Go-based statusline for Claude Code, replacing the shell-based implementation.

## Features

- **Fast:** Single binary, no subprocess spawning for most operations
- **Cached:** AWS Bedrock resolution and version checks are cached
- **Cost tracking:** Multi-period cost tracking (30day/7day/daily/live)
- **Styled:** Catppuccin Mocha theme via lipgloss

## Installation

```bash
go build -o claude-statusline .
cp claude-statusline ~/.local/bin/
```

Update `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "/Users/YOUR_USER/.local/bin/claude-statusline"
  }
}
```

## Layout

**Line 1:** Repository info (path, branch, clean/dirty status, worktree)

**Line 2:** Model info, commits today, submodules, version, time

**Line 3:** Cost tracking (30day, 7day, daily, live), context window, output style

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
