# Go Statusline Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite the shell-based Claude Code statusline into a fast Go binary with lipgloss rendering.

**Architecture:** Component interface pattern. Read JSON from stdin, run 12 registered components across 3 hardcoded lines, render with lipgloss using Catppuccin Mocha theme, write to stdout. File-based caching for AWS and version lookups. Append-only JSONL for cost history.

**Tech Stack:** Go 1.21+, lipgloss, standard library (no AWS SDK, no git library)

---

## Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `main.go`
- Create: `.gitignore`

**Step 1: Initialize Go module**

Run: `go mod init github.com/h2ik/claude-statusline`

**Step 2: Add lipgloss dependency**

Run: `go get github.com/charmbracelet/lipgloss@latest`

**Step 3: Create main.go stub**

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stdout, "claude-statusline v0.1.0")
}
```

**Step 4: Create .gitignore**

```
# Binaries
claude-statusline
*.exe
*.dll
*.so
*.dylib

# Test binaries
*.test

# Output of go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
```

**Step 5: Test build**

Run: `go build -o claude-statusline .`
Expected: Binary created successfully

**Step 6: Commit**

```bash
git add go.mod go.sum main.go .gitignore
git commit -m "chore: Initialize Go module and project scaffold"
```

---

## Task 2: Input Package (JSON Parsing)

**Files:**
- Create: `internal/input/input.go`
- Create: `internal/input/input_test.go`

**Step 1: Write test for parsing valid JSON**

```go
package input

import (
	"strings"
	"testing"
)

func TestParseInput_ValidJSON(t *testing.T) {
	json := `{
		"workspace": {"current_dir": "/tmp/test"},
		"model": {"display_name": "Claude Opus 4.6"},
		"session_id": "test-123",
		"cost": {"total_cost_usd": 0.45}
	}`

	input, err := ParseInput(strings.NewReader(json))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if input.Workspace.CurrentDir != "/tmp/test" {
		t.Errorf("expected /tmp/test, got %s", input.Workspace.CurrentDir)
	}

	if input.Model.DisplayName != "Claude Opus 4.6" {
		t.Errorf("expected Claude Opus 4.6, got %s", input.Model.DisplayName)
	}

	if input.SessionID != "test-123" {
		t.Errorf("expected test-123, got %s", input.SessionID)
	}

	if input.Cost.TotalCostUSD != 0.45 {
		t.Errorf("expected 0.45, got %f", input.Cost.TotalCostUSD)
	}
}

func TestParseInput_MissingWorkspace(t *testing.T) {
	json := `{"model": {"display_name": "Claude"}}`

	_, err := ParseInput(strings.NewReader(json))
	if err == nil {
		t.Fatal("expected error for missing workspace, got nil")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/input/... -v`
Expected: FAIL with "undefined: ParseInput"

**Step 3: Implement input structs and parser**

```go
package input

import (
	"encoding/json"
	"fmt"
	"io"
)

type StatusLineInput struct {
	Workspace     Workspace     `json:"workspace"`
	Model         ModelInfo     `json:"model"`
	SessionID     string        `json:"session_id"`
	TranscriptPath string       `json:"transcript_path"`
	OutputStyle   OutputStyle   `json:"output_style"`
	ContextWindow ContextWindow `json:"context_window"`
	Cost          CostInfo      `json:"cost"`
	CurrentUsage  UsageInfo     `json:"current_usage"`
	FiveHour      UsageLimit    `json:"five_hour"`
	SevenDay      UsageLimit    `json:"seven_day"`
	MCP           MCPInfo       `json:"mcp"`
}

type Workspace struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

type ModelInfo struct {
	DisplayName string `json:"display_name"`
}

type OutputStyle struct {
	Name string `json:"name"`
}

type ContextWindow struct {
	UsedPercentage      int `json:"used_percentage"`
	RemainingPercentage int `json:"remaining_percentage"`
	ContextWindowSize   int `json:"context_window_size"`
}

type CostInfo struct {
	TotalCostUSD        float64 `json:"total_cost_usd"`
	TotalDurationMS     int     `json:"total_duration_ms"`
	TotalAPIDurationMS  int     `json:"total_api_duration_ms"`
	TotalLinesAdded     int     `json:"total_lines_added"`
	TotalLinesRemoved   int     `json:"total_lines_removed"`
}

type UsageInfo struct {
	InputTokens              int `json:"input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

type UsageLimit struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

type MCPInfo struct {
	Servers []interface{} `json:"servers"`
}

func ParseInput(r io.Reader) (*StatusLineInput, error) {
	var input StatusLineInput

	if err := json.NewDecoder(r).Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate required field
	if input.Workspace.CurrentDir == "" {
		return nil, fmt.Errorf("workspace.current_dir is required")
	}

	return &input, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/input/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/input/
git commit -m "feat(input): Add JSON input parsing with validation"
```

---

## Task 3: Cache Package

**Files:**
- Create: `internal/cache/cache.go`
- Create: `internal/cache/cache_test.go`

**Step 1: Write test for cache operations**

```go
package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	key := "test-key"
	value := []byte("test value")
	ttl := 1 * time.Hour

	if err := c.Set(key, value, ttl); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := c.Get(key, ttl)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got) != string(value) {
		t.Errorf("expected %s, got %s", value, got)
	}
}

func TestCache_GetExpired(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	key := "expire-test"
	value := []byte("old value")

	// Write cache file with old mtime
	path := c.path(key)
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, value, 0644)

	// Set mtime to 2 hours ago
	oldTime := time.Now().Add(-2 * time.Hour)
	os.Chtimes(path, oldTime, oldTime)

	// Try to get with 1 hour TTL
	_, err := c.Get(key, 1*time.Hour)
	if err == nil {
		t.Fatal("expected error for expired cache, got nil")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cache/... -v`
Expected: FAIL with "undefined: New"

**Step 3: Implement cache**

```go
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	dir string
}

func New(dir string) *Cache {
	return &Cache{dir: dir}
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) error {
	path := c.path(key)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	if err := os.WriteFile(path, value, 0644); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

func (c *Cache) Get(key string, ttl time.Duration) ([]byte, error) {
	path := c.path(key)

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}

	age := time.Since(info.ModTime())
	if age > ttl {
		return nil, fmt.Errorf("cache expired (age: %v, ttl: %v)", age, ttl)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return data, nil
}

func (c *Cache) path(key string) string {
	hash := sha256.Sum256([]byte(key))
	filename := hex.EncodeToString(hash[:])
	return filepath.Join(c.dir, filename)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/cache/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cache/
git commit -m "feat(cache): Add file-based cache with TTL support"
```

---

## Task 4: Render Package (Lipgloss Theme)

**Files:**
- Create: `internal/render/render.go`
- Create: `internal/render/render_test.go`

**Step 1: Write test for rendering lines**

```go
package render

import (
	"strings"
	"testing"
)

func TestRenderer_RenderLines(t *testing.T) {
	r := New()

	lines := [][]string{
		{"component1", "component2"},
		{"component3"},
	}

	output := r.RenderLines(lines)

	// Should have 2 lines
	lineCount := strings.Count(output, "\n")
	if lineCount != 1 { // 2 lines = 1 newline
		t.Errorf("expected 1 newline, got %d", lineCount)
	}

	// Should contain separator
	if !strings.Contains(output, " │ ") {
		t.Error("expected separator ' │ ' in output")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/render/... -v`
Expected: FAIL with "undefined: New"

**Step 3: Implement renderer with Catppuccin Mocha**

```go
package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha colors (hardcoded)
var (
	ColorOverlay0 = lipgloss.Color("#6c7086") // Dimmed/labels
	ColorText     = lipgloss.Color("#cdd6f4") // Text/values
	ColorGreen    = lipgloss.Color("#a6e3a1") // Clean/good
	ColorRed      = lipgloss.Color("#f38ba8") // Critical
	ColorYellow   = lipgloss.Color("#f9e2af") // Warning
	ColorBlue     = lipgloss.Color("#89b4fa") // Paths/info
	ColorMauve    = lipgloss.Color("#cba6f7") // Accent
	ColorPeach    = lipgloss.Color("#fab387") // Costs
	ColorTeal     = lipgloss.Color("#94e2d5") // Secondary
)

type Renderer struct {
	separator string
}

func New() *Renderer {
	return &Renderer{
		separator: " │ ",
	}
}

func (r *Renderer) RenderLines(lines [][]string) string {
	var output []string

	for _, components := range lines {
		// Filter out empty components
		var nonEmpty []string
		for _, c := range components {
			if strings.TrimSpace(c) != "" {
				nonEmpty = append(nonEmpty, c)
			}
		}

		if len(nonEmpty) > 0 {
			line := strings.Join(nonEmpty, r.separator)
			output = append(output, line)
		}
	}

	return strings.Join(output, "\n")
}

// Style helpers
func (r *Renderer) Dimmed(s string) string {
	return lipgloss.NewStyle().Foreground(ColorOverlay0).Render(s)
}

func (r *Renderer) Text(s string) string {
	return lipgloss.NewStyle().Foreground(ColorText).Render(s)
}

func (r *Renderer) Green(s string) string {
	return lipgloss.NewStyle().Foreground(ColorGreen).Render(s)
}

func (r *Renderer) Red(s string) string {
	return lipgloss.NewStyle().Foreground(ColorRed).Render(s)
}

func (r *Renderer) Yellow(s string) string {
	return lipgloss.NewStyle().Foreground(ColorYellow).Render(s)
}

func (r *Renderer) Blue(s string) string {
	return lipgloss.NewStyle().Foreground(ColorBlue).Render(s)
}

func (r *Renderer) Mauve(s string) string {
	return lipgloss.NewStyle().Foreground(ColorMauve).Render(s)
}

func (r *Renderer) Peach(s string) string {
	return lipgloss.NewStyle().Foreground(ColorPeach).Render(s)
}

func (r *Renderer) Teal(s string) string {
	return lipgloss.NewStyle().Foreground(ColorTeal).Render(s)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/render/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/render/
git commit -m "feat(render): Add lipgloss renderer with Catppuccin Mocha theme"
```

---

## Task 5: Component Interface and Registry

**Files:**
- Create: `internal/component/component.go`
- Create: `internal/component/registry.go`
- Create: `internal/component/registry_test.go`

**Step 1: Write test for component registration**

```go
package component

import (
	"testing"

	"github.com/h2ik/claude-statusline/internal/input"
)

type mockComponent struct {
	name   string
	output string
}

func (m *mockComponent) Name() string {
	return m.name
}

func (m *mockComponent) Render(in *input.StatusLineInput) string {
	return m.output
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	comp := &mockComponent{name: "test", output: "hello"}
	r.Register(comp)

	if c := r.Get("test"); c == nil {
		t.Fatal("expected component to be registered")
	}
}

func TestRegistry_RenderLine(t *testing.T) {
	r := NewRegistry()

	r.Register(&mockComponent{name: "c1", output: "first"})
	r.Register(&mockComponent{name: "c2", output: "second"})

	in := &input.StatusLineInput{}
	components := []string{"c1", "c2"}

	output := r.RenderLine(in, components)

	if len(output) != 2 {
		t.Errorf("expected 2 components, got %d", len(output))
	}

	if output[0] != "first" {
		t.Errorf("expected 'first', got %s", output[0])
	}

	if output[1] != "second" {
		t.Errorf("expected 'second', got %s", output[1])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/component/... -v`
Expected: FAIL with "undefined: NewRegistry"

**Step 3: Implement component interface and registry**

```go
// component.go
package component

import "github.com/h2ik/claude-statusline/internal/input"

type Component interface {
	Name() string
	Render(input *input.StatusLineInput) string
}
```

```go
// registry.go
package component

import "github.com/h2ik/claude-statusline/internal/input"

type Registry struct {
	components map[string]Component
}

func NewRegistry() *Registry {
	return &Registry{
		components: make(map[string]Component),
	}
}

func (r *Registry) Register(c Component) {
	r.components[c.Name()] = c
}

func (r *Registry) Get(name string) Component {
	return r.components[name]
}

func (r *Registry) RenderLine(in *input.StatusLineInput, names []string) []string {
	var output []string

	for _, name := range names {
		if c := r.Get(name); c != nil {
			if rendered := c.Render(in); rendered != "" {
				output = append(output, rendered)
			}
		}
	}

	return output
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/component/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/component/
git commit -m "feat(component): Add Component interface and Registry"
```

---

## Task 6: Git Package

**Files:**
- Create: `internal/git/git.go`
- Create: `internal/git/git_test.go`

**Step 1: Write test for git operations**

```go
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupGitRepo(t *testing.T) string {
	dir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Configure git
	exec.Command("git", "config", "user.email", "test@test.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = dir
	cmd.Run()

	return dir
}

func TestGetBranch(t *testing.T) {
	dir := setupGitRepo(t)

	branch, err := GetBranch(dir)
	if err != nil {
		t.Fatalf("GetBranch failed: %v", err)
	}

	// Default branch is usually 'main' or 'master'
	if branch != "main" && branch != "master" {
		t.Logf("got branch: %s (acceptable)", branch)
	}
}

func TestIsClean(t *testing.T) {
	dir := setupGitRepo(t)

	clean, err := IsClean(dir)
	if err != nil {
		t.Fatalf("IsClean failed: %v", err)
	}

	if !clean {
		t.Error("expected clean status")
	}

	// Modify a file
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("modified"), 0644)

	clean, err = IsClean(dir)
	if err != nil {
		t.Fatalf("IsClean failed: %v", err)
	}

	if clean {
		t.Error("expected dirty status after modification")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/git/... -v`
Expected: FAIL with "undefined: GetBranch"

**Step 3: Implement git operations**

```go
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func IsClean(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}

	return len(strings.TrimSpace(string(output))) == 0, nil
}

func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	return cmd.Run() == nil
}

func GetCommitsToday(dir string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", "--since=today 00:00", "HEAD")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("git rev-list failed: %w", err)
	}

	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count); err != nil {
		return 0, fmt.Errorf("parse count failed: %w", err)
	}

	return count, nil
}

func GetSubmoduleCount(dir string) (int, error) {
	cmd := exec.Command("git", "submodule", "status")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		// Not an error if no submodules
		return 0, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}

	return len(lines), nil
}

func IsWorktree(dir string) (bool, string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return false, "", err
	}

	gitDir := strings.TrimSpace(string(output))

	// If .git is a file, it's a worktree
	if strings.Contains(gitDir, ".git/worktrees/") {
		parts := strings.Split(gitDir, "/")
		for i, part := range parts {
			if part == "worktrees" && i+1 < len(parts) {
				return true, parts[i+1], nil
			}
		}
		return true, "", nil
	}

	return false, "", nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/git/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/git/
git commit -m "feat(git): Add git operations (branch, status, commits, worktree)"
```

---

## Task 7: Cost Package

**Files:**
- Create: `internal/cost/cost.go`
- Create: `internal/cost/history.go`
- Create: `internal/cost/cost_test.go`

**Step 1: Write test for cost tracking**

```go
package cost

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistory_Append(t *testing.T) {
	dir := t.TempDir()
	historyFile := filepath.Join(dir, "history.jsonl")

	h := NewHistory(historyFile)

	entry := Entry{
		SessionID: "test-123",
		Cost:      0.45,
		Timestamp: time.Now(),
	}

	if err := h.Append(entry); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(historyFile); err != nil {
		t.Fatalf("history file not created: %v", err)
	}
}

func TestCalculatePeriodCost(t *testing.T) {
	dir := t.TempDir()
	historyFile := filepath.Join(dir, "history.jsonl")

	h := NewHistory(historyFile)

	// Add entries over different days
	now := time.Now()
	entries := []Entry{
		{SessionID: "s1", Cost: 1.0, Timestamp: now.Add(-25 * time.Hour)}, // yesterday
		{SessionID: "s2", Cost: 2.0, Timestamp: now.Add(-1 * time.Hour)},  // today
		{SessionID: "s3", Cost: 3.0, Timestamp: now},                      // now
	}

	for _, e := range entries {
		h.Append(e)
	}

	// Calculate last 24 hours (should be 2.0 + 3.0 = 5.0)
	cost, err := h.CalculatePeriod(24 * time.Hour)
	if err != nil {
		t.Fatalf("CalculatePeriod failed: %v", err)
	}

	if cost < 4.9 || cost > 5.1 {
		t.Errorf("expected ~5.0, got %f", cost)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cost/... -v`
Expected: FAIL with "undefined: NewHistory"

**Step 3: Implement cost history**

```go
// cost.go
package cost

import "time"

type Entry struct {
	SessionID string    `json:"session_id"`
	Cost      float64   `json:"cost"`
	Timestamp time.Time `json:"timestamp"`
}
```

```go
// history.go
package cost

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type History struct {
	path string
}

func NewHistory(path string) *History {
	return &History{path: path}
}

func (h *History) Append(entry Entry) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	// Open file for appending
	f, err := os.OpenFile(h.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}
	defer f.Close()

	// Encode JSON line
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

func (h *History) CalculatePeriod(duration time.Duration) (float64, error) {
	f, err := os.Open(h.path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0.0, nil
		}
		return 0.0, fmt.Errorf("open failed: %w", err)
	}
	defer f.Close()

	cutoff := time.Now().Add(-duration)
	total := 0.0
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed lines
		}

		// Only count entries within the period
		if entry.Timestamp.After(cutoff) {
			// Deduplicate by session ID
			if !seen[entry.SessionID] {
				total += entry.Cost
				seen[entry.SessionID] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0.0, fmt.Errorf("scan failed: %w", err)
	}

	return total, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/cost/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cost/
git commit -m "feat(cost): Add cost history tracking with period calculations"
```

---

## Task 8: Component Implementations (Part 1: Repo Info)

**Files:**
- Create: `internal/components/repo_info.go`
- Create: `internal/components/repo_info_test.go`

**Step 1: Write test for repo_info component**

```go
package components

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func TestRepoInfo_Render(t *testing.T) {
	r := render.New()
	c := NewRepoInfo(r)

	if c.Name() != "repo_info" {
		t.Errorf("expected 'repo_info', got %s", c.Name())
	}

	homeDir, _ := os.UserHomeDir()
	testDir := filepath.Join(homeDir, "testdir")

	in := &input.StatusLineInput{
		Workspace: input.Workspace{
			CurrentDir: testDir,
		},
	}

	output := c.Render(in)

	// Should contain ~/testdir
	if !strings.Contains(output, "~/testdir") {
		t.Errorf("expected ~/testdir in output, got: %s", output)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/components/... -v`
Expected: FAIL with "undefined: NewRepoInfo"

**Step 3: Implement repo_info component**

```go
package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2ik/claude-statusline/internal/git"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type RepoInfo struct {
	renderer *render.Renderer
}

func NewRepoInfo(r *render.Renderer) *RepoInfo {
	return &RepoInfo{renderer: r}
}

func (c *RepoInfo) Name() string {
	return "repo_info"
}

func (c *RepoInfo) Render(in *input.StatusLineInput) string {
	dir := in.Workspace.CurrentDir

	// Convert to ~ notation
	homeDir, _ := os.UserHomeDir()
	displayDir := strings.Replace(dir, homeDir, "~", 1)

	// Check if it's a git repo
	if !git.IsGitRepo(dir) {
		return c.renderer.Blue(displayDir)
	}

	// Get branch
	branch, err := git.GetBranch(dir)
	if err != nil {
		return c.renderer.Blue(displayDir)
	}

	// Get clean/dirty status
	clean, err := git.IsClean(dir)
	if err != nil {
		clean = false
	}

	statusEmoji := "✅"
	statusColor := c.renderer.Green
	if !clean {
		statusEmoji = "📁"
		statusColor = c.renderer.Yellow
	}

	// Check for worktree
	isWT, wtName, _ := git.IsWorktree(dir)
	wtIndicator := ""
	if isWT && wtName != "" {
		wtIndicator = c.renderer.Teal(fmt.Sprintf(" [WT:%s]", wtName))
	}

	return fmt.Sprintf("%s %s %s%s",
		c.renderer.Blue(displayDir),
		c.renderer.Mauve(fmt.Sprintf("(%s)", branch)),
		statusColor(statusEmoji),
		wtIndicator,
	)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/components/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/components/repo_info*
git commit -m "feat(components): Add repo_info component"
```

---

## Task 9: Component Implementations (Part 2: Bedrock Model)

**Files:**
- Create: `internal/components/bedrock_model.go`

**Step 1: Implement bedrock_model component (no test due to AWS dependency)**

```go
package components

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type BedrockModel struct {
	renderer *render.Renderer
	cache    *cache.Cache
}

func NewBedrockModel(r *render.Renderer, c *cache.Cache) *BedrockModel {
	return &BedrockModel{renderer: r, cache: c}
}

func (c *BedrockModel) Name() string {
	return "bedrock_model"
}

func (c *BedrockModel) Render(in *input.StatusLineInput) string {
	modelName := in.Model.DisplayName

	// Check if it's a Bedrock ARN
	if !strings.HasPrefix(modelName, "arn:aws:bedrock:") {
		return fmt.Sprintf("🧠 %s", c.renderer.Text(modelName))
	}

	// Try to resolve via AWS CLI with caching
	resolved := c.resolveBedrockARN(modelName)
	return fmt.Sprintf("🧠 %s", c.renderer.Text(resolved))
}

func (c *BedrockModel) resolveBedrockARN(arn string) string {
	// Check cache
	cached, err := c.cache.Get("bedrock:"+arn, 24*time.Hour)
	if err == nil {
		return string(cached)
	}

	// Call AWS CLI
	cmd := exec.Command("aws", "bedrock", "get-inference-profile",
		"--inference-profile-identifier", arn,
		"--query", "modelId",
		"--output", "text")

	output, err := cmd.Output()
	if err != nil {
		// Fallback to raw ARN
		return arn
	}

	modelID := strings.TrimSpace(string(output))

	// Try to get friendly name
	friendlyName := c.getFriendlyName(modelID)

	// Extract region from ARN
	parts := strings.Split(arn, ":")
	region := ""
	if len(parts) >= 4 {
		region = parts[3]
	}

	result := friendlyName
	if region != "" {
		result = fmt.Sprintf("%s (%s)", friendlyName, region)
	}

	// Cache the result
	c.cache.Set("bedrock:"+arn, []byte(result), 24*time.Hour)

	return result
}

func (c *BedrockModel) getFriendlyName(modelID string) string {
	// Simple hardcoded mapping for common models
	mapping := map[string]string{
		"anthropic.claude-opus-4-6":   "Claude Opus 4.6",
		"anthropic.claude-3-5-sonnet": "Claude 3.5 Sonnet",
		"anthropic.claude-3-haiku":    "Claude 3 Haiku",
	}

	for key, name := range mapping {
		if strings.Contains(modelID, key) {
			return name
		}
	}

	return modelID
}
```

**Step 2: Commit**

```bash
git add internal/components/bedrock_model.go
git commit -m "feat(components): Add bedrock_model component with AWS CLI resolution"
```

---

## Task 10: Component Implementations (Part 3: Simple Components)

**Files:**
- Create: `internal/components/model_info.go`
- Create: `internal/components/commits.go`
- Create: `internal/components/submodules.go`
- Create: `internal/components/version_info.go`
- Create: `internal/components/time_display.go`

**Step 1: Implement all simple components**

```go
// model_info.go
package components

import (
	"fmt"
	"strings"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type ModelInfo struct {
	renderer *render.Renderer
}

func NewModelInfo(r *render.Renderer) *ModelInfo {
	return &ModelInfo{renderer: r}
}

func (c *ModelInfo) Name() string {
	return "model_info"
}

func (c *ModelInfo) Render(in *input.StatusLineInput) string {
	name := in.Model.DisplayName
	if name == "" {
		name = "Claude"
	}

	emoji := c.getEmoji(name)

	return fmt.Sprintf("%s %s", emoji, c.renderer.Teal(name))
}

func (c *ModelInfo) getEmoji(name string) string {
	lower := strings.ToLower(name)

	switch {
	case strings.Contains(lower, "opus"):
		return "🧠"
	case strings.Contains(lower, "haiku"):
		return "⚡"
	case strings.Contains(lower, "sonnet"):
		return "🎵"
	default:
		return "🤖"
	}
}
```

```go
// commits.go
package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/git"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type Commits struct {
	renderer *render.Renderer
}

func NewCommits(r *render.Renderer) *Commits {
	return &Commits{renderer: r}
}

func (c *Commits) Name() string {
	return "commits"
}

func (c *Commits) Render(in *input.StatusLineInput) string {
	count, err := git.GetCommitsToday(in.Workspace.CurrentDir)
	if err != nil || count == 0 {
		return ""
	}

	return fmt.Sprintf("💾 %s %d",
		c.renderer.Dimmed("Commits:"),
		count,
	)
}
```

```go
// submodules.go
package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/git"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type Submodules struct {
	renderer *render.Renderer
}

func NewSubmodules(r *render.Renderer) *Submodules {
	return &Submodules{renderer: r}
}

func (c *Submodules) Name() string {
	return "submodules"
}

func (c *Submodules) Render(in *input.StatusLineInput) string {
	count, err := git.GetSubmoduleCount(in.Workspace.CurrentDir)
	if err != nil || count == 0 {
		return ""
	}

	return fmt.Sprintf("🔗 %s%d",
		c.renderer.Dimmed("SUB:"),
		count,
	)
}
```

```go
// version_info.go
package components

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type VersionInfo struct {
	renderer *render.Renderer
	cache    *cache.Cache
}

func NewVersionInfo(r *render.Renderer, c *cache.Cache) *VersionInfo {
	return &VersionInfo{renderer: r, cache: c}
}

func (c *VersionInfo) Name() string {
	return "version_info"
}

func (c *VersionInfo) Render(in *input.StatusLineInput) string {
	version := c.getClaudeVersion()
	if version == "" {
		return ""
	}

	return fmt.Sprintf("%s%s",
		c.renderer.Dimmed("CC:"),
		c.renderer.Text(version),
	)
}

func (c *VersionInfo) getClaudeVersion() string {
	// Check cache
	cached, err := c.cache.Get("claude-version", 15*time.Minute)
	if err == nil {
		return string(cached)
	}

	// Run claude --version
	cmd := exec.Command("claude", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "claude ")
	version = strings.TrimPrefix(version, "v")

	// Cache it
	c.cache.Set("claude-version", []byte(version), 15*time.Minute)

	return version
}
```

```go
// time_display.go
package components

import (
	"time"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type TimeDisplay struct {
	renderer *render.Renderer
}

func NewTimeDisplay(r *render.Renderer) *TimeDisplay {
	return &TimeDisplay{renderer: r}
}

func (c *TimeDisplay) Name() string {
	return "time_display"
}

func (c *TimeDisplay) Render(in *input.StatusLineInput) string {
	now := time.Now().Format("15:04")
	return "🕐 " + c.renderer.Text(now)
}
```

**Step 2: Commit**

```bash
git add internal/components/model_info.go internal/components/commits.go internal/components/submodules.go internal/components/version_info.go internal/components/time_display.go
git commit -m "feat(components): Add model_info, commits, submodules, version_info, time_display"
```

---

## Task 11: Component Implementations (Part 4: Cost Components)

**Files:**
- Create: `internal/components/cost_monthly.go`
- Create: `internal/components/cost_weekly.go`
- Create: `internal/components/cost_daily.go`
- Create: `internal/components/cost_live.go`

**Step 1: Implement cost components**

```go
// cost_monthly.go
package components

import (
	"fmt"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type CostMonthly struct {
	renderer *render.Renderer
	history  *cost.History
}

func NewCostMonthly(r *render.Renderer, h *cost.History) *CostMonthly {
	return &CostMonthly{renderer: r, history: h}
}

func (c *CostMonthly) Name() string {
	return "cost_monthly"
}

func (c *CostMonthly) Render(in *input.StatusLineInput) string {
	total, _ := c.history.CalculatePeriod(30 * 24 * time.Hour)

	return fmt.Sprintf("📈 %s $%.2f",
		c.renderer.Dimmed("30DAY"),
		total,
	)
}
```

```go
// cost_weekly.go
package components

import (
	"fmt"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type CostWeekly struct {
	renderer *render.Renderer
	history  *cost.History
}

func NewCostWeekly(r *render.Renderer, h *cost.History) *CostWeekly {
	return &CostWeekly{renderer: r, history: h}
}

func (c *CostWeekly) Name() string {
	return "cost_weekly"
}

func (c *CostWeekly) Render(in *input.StatusLineInput) string {
	total, _ := c.history.CalculatePeriod(7 * 24 * time.Hour)

	return fmt.Sprintf("📊 %s $%.2f",
		c.renderer.Dimmed("7DAY"),
		total,
	)
}
```

```go
// cost_daily.go
package components

import (
	"fmt"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type CostDaily struct {
	renderer *render.Renderer
	history  *cost.History
}

func NewCostDaily(r *render.Renderer, h *cost.History) *CostDaily {
	return &CostDaily{renderer: r, history: h}
}

func (c *CostDaily) Name() string {
	return "cost_daily"
}

func (c *CostDaily) Render(in *input.StatusLineInput) string {
	total, _ := c.history.CalculatePeriod(24 * time.Hour)

	return fmt.Sprintf("📅 %s $%.2f",
		c.renderer.Dimmed("DAY"),
		total,
	)
}
```

```go
// cost_live.go
package components

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type CostLive struct {
	renderer *render.Renderer
	history  *cost.History
}

func NewCostLive(r *render.Renderer, h *cost.History) *CostLive {
	return &CostLive{renderer: r, history: h}
}

func (c *CostLive) Name() string {
	return "cost_live"
}

func (c *CostLive) Render(in *input.StatusLineInput) string {
	// Append current session cost to history
	if in.SessionID != "" && in.Cost.TotalCostUSD > 0 {
		entry := cost.Entry{
			SessionID: in.SessionID,
			Cost:      in.Cost.TotalCostUSD,
			Timestamp: time.Now(),
		}
		c.history.Append(entry)
	}

	// Display live session cost
	return fmt.Sprintf("🔥%s $%.2f",
		c.renderer.Dimmed("LIVE"),
		in.Cost.TotalCostUSD,
	)
}
```

**Step 2: Commit**

```bash
git add internal/components/cost_*.go
git commit -m "feat(components): Add cost tracking components (monthly, weekly, daily, live)"
```

---

## Task 12: Component Implementations (Part 5: Context & Session)

**Files:**
- Create: `internal/components/context_window.go`
- Create: `internal/components/session_mode.go`

**Step 1: Implement context_window and session_mode**

```go
// context_window.go
package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type ContextWindow struct {
	renderer *render.Renderer
}

func NewContextWindow(r *render.Renderer) *ContextWindow {
	return &ContextWindow{renderer: r}
}

func (c *ContextWindow) Name() string {
	return "context_window"
}

func (c *ContextWindow) Render(in *input.StatusLineInput) string {
	pct := in.ContextWindow.UsedPercentage

	if pct == 0 {
		return ""
	}

	// Color based on percentage
	var colorFunc func(string) string
	warning := ""

	if pct >= 90 {
		colorFunc = c.renderer.Red
		warning = " ⚠️"
	} else if pct >= 75 {
		colorFunc = c.renderer.Red
	} else if pct >= 50 {
		colorFunc = c.renderer.Yellow
	} else {
		colorFunc = c.renderer.Green
	}

	// Format with tokens if available
	tokens := ""
	if in.ContextWindow.ContextWindowSize > 0 {
		used := float64(pct) / 100.0 * float64(in.ContextWindow.ContextWindowSize)
		tokens = fmt.Sprintf(" (%.0fK/%dK)",
			used/1000.0,
			in.ContextWindow.ContextWindowSize/1000,
		)
	}

	return fmt.Sprintf("🧠 %s%s%s",
		colorFunc(fmt.Sprintf("%d%%", pct)),
		colorFunc(tokens),
		warning,
	)
}
```

```go
// session_mode.go
package components

import (
	"fmt"

	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

type SessionMode struct {
	renderer *render.Renderer
}

func NewSessionMode(r *render.Renderer) *SessionMode {
	return &SessionMode{renderer: r}
}

func (c *SessionMode) Name() string {
	return "session_mode"
}

func (c *SessionMode) Render(in *input.StatusLineInput) string {
	style := in.OutputStyle.Name

	if style == "" || style == "default" {
		return ""
	}

	emoji := c.getEmoji(style)

	return fmt.Sprintf("%s %s %s",
		emoji,
		c.renderer.Dimmed("Style:"),
		c.renderer.Text(style),
	)
}

func (c *SessionMode) getEmoji(style string) string {
	mapping := map[string]string{
		"explanatory": "📚",
		"learning":    "🎓",
	}

	if emoji, ok := mapping[style]; ok {
		return emoji
	}

	return "✨"
}
```

**Step 2: Commit**

```bash
git add internal/components/context_window.go internal/components/session_mode.go
git commit -m "feat(components): Add context_window and session_mode components"
```

---

## Task 13: Wire Everything Together in main.go

**Files:**
- Modify: `main.go`

**Step 1: Update main.go to orchestrate everything**

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/component"
	"github.com/h2ik/claude-statusline/internal/components"
	"github.com/h2ik/claude-statusline/internal/cost"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

func main() {
	// Read JSON from stdin
	in, err := input.ParseInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse input: %v\n", err)
		os.Exit(1)
	}

	// Initialize infrastructure
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".cache", "claude-statusline")
	costDir := filepath.Join(homeDir, ".claude", "statusline", "costs")

	c := cache.New(cacheDir)
	r := render.New()
	h := cost.NewHistory(filepath.Join(costDir, "history.jsonl"))

	// Create registry and register components
	registry := component.NewRegistry()

	// Line 1 components
	registry.Register(components.NewRepoInfo(r))

	// Line 2 components
	registry.Register(components.NewModelInfo(r))
	registry.Register(components.NewBedrockModel(r, c))
	registry.Register(components.NewCommits(r))
	registry.Register(components.NewSubmodules(r))
	registry.Register(components.NewVersionInfo(r, c))
	registry.Register(components.NewTimeDisplay(r))

	// Line 3 components
	registry.Register(components.NewCostMonthly(r, h))
	registry.Register(components.NewCostWeekly(r, h))
	registry.Register(components.NewCostDaily(r, h))
	registry.Register(components.NewCostLive(r, h))
	registry.Register(components.NewContextWindow(r))
	registry.Register(components.NewSessionMode(r))

	// Define line layout
	lines := [][]string{
		{"repo_info"},
		{"bedrock_model", "commits", "submodules", "version_info", "time_display"},
		{"cost_monthly", "cost_weekly", "cost_daily", "cost_live", "context_window", "session_mode"},
	}

	// Render each line
	var renderedLines [][]string
	for _, lineComponents := range lines {
		rendered := registry.RenderLine(in, lineComponents)
		if len(rendered) > 0 {
			renderedLines = append(renderedLines, rendered)
		}
	}

	// Output final result
	output := r.RenderLines(renderedLines)
	fmt.Fprintln(os.Stdout, output)
}
```

**Step 2: Test the full pipeline**

Run: `go build -o claude-statusline .`

Create test JSON:
```bash
cat > /tmp/test-input.json <<'EOF'
{
  "workspace": {"current_dir": "/Users/jon.whitcraft/Projects/h2ik/claude-statusline"},
  "model": {"display_name": "Claude Opus 4.6"},
  "session_id": "test-123",
  "output_style": {"name": "default"},
  "context_window": {"used_percentage": 45, "context_window_size": 200000},
  "cost": {"total_cost_usd": 0.45},
  "current_usage": {"input_tokens": 10000}
}
EOF
```

Run: `cat /tmp/test-input.json | ./claude-statusline`

Expected: Three lines of colorized statusline output

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat(main): Wire all components together with registry"
```

---

## Task 14: Update Claude Code Settings

**Files:**
- Modify: `~/.claude/settings.json`

**Step 1: Build and install binary**

Run: `go build -o claude-statusline .`
Run: `mkdir -p ~/.local/bin && cp claude-statusline ~/.local/bin/`

**Step 2: Update settings.json**

Edit `~/.claude/settings.json` to change the statusLine command:

```json
{
  "statusLine": {
    "type": "command",
    "command": "/Users/jon.whitcraft/.local/bin/claude-statusline"
  }
}
```

**Step 3: Test in a real Claude Code session**

Open a new Claude Code session and verify the statusline renders correctly.

**Step 4: Commit settings change (if tracked)**

If your settings are version controlled:
```bash
git add ~/.claude/settings.json
git commit -m "chore: Update Claude Code statusline to Go binary"
```

---

## Task 15: Documentation and README

**Files:**
- Create: `README.md`
- Create: `docs/ARCHITECTURE.md`

**Step 1: Write README.md**

```markdown
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
```

**Step 2: Write docs/ARCHITECTURE.md**

```markdown
# Architecture

## Overview

Component-based architecture with a registry pattern. Each component implements a simple interface and is responsible for rendering one piece of the statusline.

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
```

**Step 3: Commit**

```bash
git add README.md docs/ARCHITECTURE.md
git commit -m "docs: Add README and architecture documentation"
```

---

## Execution Complete

Plan saved to `docs/plans/2026-02-15-go-statusline-implementation.md`.

**Next Steps:**

1. **Subagent-Driven (this session)** - Stay here, dispatch fresh subagent per task, review between tasks
2. **Parallel Session (separate)** - Open new session in worktree, use `@superpowers:executing-plans` for batch execution

Which approach do you prefer?
