# GitHub Actions & Goreleaser Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add CI (test/lint) and release (goreleaser + Homebrew) automation via GitHub Actions.

**Architecture:** Two workflow files in `.github/workflows/`, one goreleaser config at project root, and version ldflags wired into `main.go`. CI runs on PRs and main pushes. Release runs on version tags.

**Tech Stack:** GitHub Actions, golangci-lint, goreleaser v2, Homebrew tap.

---

### Task 1: Add version variables to main.go

**Files:**
- Modify: `main.go`

**Step 1: Add version variables**

Add these package-level variables right after the `import` block in `main.go`, before `func main()`:

```go
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)
```

These are set via ldflags at build time by goreleaser. The defaults allow the binary to work during local `go build`.

**Step 2: Verify build still works**

Run: `go build -o /dev/null .`
Expected: Build succeeds (vars are unused for now, but Go allows unused package-level vars).

Run: `go vet ./...`
Expected: No issues.

**Step 3: Commit**

```
build: Add version/commit/date variables for ldflags injection
```

---

### Task 2: Create CI workflow

**Files:**
- Create: `.github/workflows/ci.yml`

**Step 1: Create the workflow file**

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Vet
        run: go vet ./...

      - name: Lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

      - name: Test
        run: go test ./... -v
```

**Step 2: Validate YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"`
Expected: No error (valid YAML).

**Step 3: Commit**

```
ci: Add test and lint workflow for PRs and main
```

---

### Task 3: Create goreleaser config

**Files:**
- Create: `.goreleaser.yml`

**Step 1: Create the goreleaser config**

Create `.goreleaser.yml`:

```yaml
version: 2

project_name: claude-statusline

builds:
  - binary: claude-statusline
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.ShortCommit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}

checksum:
  name_template: checksums.txt

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^chore:'
      - '^style:'
  groups:
    - title: Features
      regexp: '^feat'
      order: 0
    - title: Bug Fixes
      regexp: '^fix'
      order: 1
    - title: Other
      order: 999

brews:
  - repository:
      owner: h2ik
      name: homebrew-tap
      token: "{{ .Env.TAP_GITHUB_TOKEN }}"
    name: claude-statusline
    homepage: https://github.com/h2ik/claude-statusline
    description: Statusline for Claude Code terminal integration
    license: MIT
    install: |
      bin.install "claude-statusline"
    hooks:
      pre:
        install: |
          system_command "/usr/bin/xattr", args: ["-cr", "#{staged_path}/claude-statusline"]
```

**Step 2: Validate goreleaser config**

Run: `goreleaser check` (if goreleaser is installed locally)
Or: `python3 -c "import yaml; yaml.safe_load(open('.goreleaser.yml'))"`
Expected: Valid config / valid YAML.

**Step 3: Commit**

```
build: Add goreleaser config with Homebrew tap and xattr hook
```

---

### Task 4: Create release workflow

**Files:**
- Create: `.github/workflows/release.yml`

**Step 1: Create the workflow file**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          TAP_GITHUB_TOKEN: ${{ secrets.TAP_GITHUB_TOKEN }}
```

**Step 2: Validate YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))"`
Expected: No error.

**Step 3: Commit**

```
ci: Add goreleaser release workflow for tag pushes
```

---

### Task 5: Verify everything together

**Files:** None (verification only)

**Step 1: Verify project builds with ldflags**

Run: `go build -ldflags "-s -w -X main.version=test -X main.commit=abc123 -X main.date=2026-02-15" -o /dev/null .`
Expected: Build succeeds.

**Step 2: Verify all tests still pass**

Run: `go test ./... -v`
Expected: All tests PASS.

**Step 3: Verify go vet is clean**

Run: `go vet ./...`
Expected: No issues.

**Step 4: Verify goreleaser config (if installed)**

Run: `goreleaser check 2>&1 || echo "goreleaser not installed, skipping"`
Expected: Valid config or skip message.

**Step 5: List all new files**

Run: `find .github -type f && ls .goreleaser.yml`
Expected:
```
.github/workflows/ci.yml
.github/workflows/release.yml
.goreleaser.yml
```

**Step 6: Commit fixups if needed**

If any fixes were required:
```
fix(ci): Address issues from CI/release verification
```
