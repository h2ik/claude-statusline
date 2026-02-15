# GitHub Actions & Goreleaser Design

## Problem

The project has no CI or release automation. Tests and lint only run locally. Releases are manual.

## Design

### Workflow 1: CI (test + lint)

`.github/workflows/ci.yml` triggers on PRs and pushes to `main`.

Single job on `ubuntu-latest` with Go 1.23:
1. Checkout
2. Setup Go 1.23
3. Cache Go modules
4. `go vet ./...`
5. `golangci-lint` via `golangci/golangci-lint-action`
6. `go test ./... -v`

### Workflow 2: Release (goreleaser)

`.github/workflows/release.yml` triggers on tag pushes matching `v*`.

Steps:
1. Checkout with `fetch-depth: 0` (goreleaser needs full history)
2. Setup Go 1.23
3. Run goreleaser via `goreleaser/goreleaser-action@v6`
4. Secrets: `GITHUB_TOKEN` for releases, `TAP_GITHUB_TOKEN` (PAT with repo scope) for Homebrew formula push to `h2ik/homebrew-tap`

### Goreleaser Config

`.goreleaser.yml` at project root.

**Builds:**
- Single binary `claude-statusline`
- Targets: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
- ldflags: `-s -w -X main.version={{.Version}} -X main.commit={{.ShortCommit}} -X main.date={{.Date}}`

**Archives:** tar.gz for linux/darwin.

**Brews:**
- Repo: `h2ik/homebrew-tap`
- Name: `claude-statusline`
- Pre-install hook: `system_command "/usr/bin/xattr", args: ["-cr", "#{staged_path}/claude-statusline"]` to strip macOS quarantine flag
- Homepage and description filled in

**Changelog:** Auto-generated, grouped by conventional commit type (feat, fix, etc.).

### Version Variables

Three package-level vars in `main.go`, set via ldflags at build time:

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

Defaults allow the binary to work during local development without ldflags.

## Decisions

- **Single Go version** (1.23) — matches go.mod, no matrix needed
- **golangci-lint-action** over manual install — caches results, maintained by golangci team
- **xattr pre-install hook** — unsigned binaries trigger Gatekeeper quarantine on macOS
- **ldflags version trio** — standard pattern for Go CLI tools, goreleaser templates automatically
- **TAP_GITHUB_TOKEN secret** — separate PAT needed since GITHUB_TOKEN can't push to other repos
