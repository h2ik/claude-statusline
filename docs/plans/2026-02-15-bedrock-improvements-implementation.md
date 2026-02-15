# Bedrock Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix Bedrock auth by reading ~/.claude/settings.json and replace hardcoded model names with dynamic AWS API resolution.

**Architecture:** New `internal/claude/` package parses settings.json and exposes AWS env vars. `BedrockModel` injects those vars into AWS CLI calls and uses a cached `list-foundation-models` response for friendly model names, falling back to a small static map offline.

**Tech Stack:** Go stdlib (`encoding/json`, `os/exec`), existing `internal/cache` package, AWS CLI.

---

### Task 1: Create `internal/claude/` settings reader

**Files:**
- Create: `internal/claude/settings.go`
- Create: `internal/claude/settings_test.go`

**Step 1: Write the failing tests**

Create `internal/claude/settings_test.go`:

```go
package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSettings_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	content := `{
		"env": {
			"CLAUDE_CODE_USE_BEDROCK": "true",
			"AWS_PROFILE": "my-profile",
			"AWS_REGION": "us-west-2"
		}
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.UseBedrock {
		t.Error("expected UseBedrock=true")
	}
	if s.AWSProfile != "my-profile" {
		t.Errorf("expected AWSProfile='my-profile', got %q", s.AWSProfile)
	}
	if s.AWSRegion != "us-west-2" {
		t.Errorf("expected AWSRegion='us-west-2', got %q", s.AWSRegion)
	}
}

func TestLoadSettings_MissingFile(t *testing.T) {
	s, err := LoadSettings("/nonexistent/path/settings.json")
	if err != nil {
		t.Fatalf("missing file should not error, got: %v", err)
	}
	if s.UseBedrock {
		t.Error("expected UseBedrock=false for missing file")
	}
	if s.AWSProfile != "" {
		t.Errorf("expected empty AWSProfile, got %q", s.AWSProfile)
	}
	if s.AWSRegion != "" {
		t.Errorf("expected empty AWSRegion, got %q", s.AWSRegion)
	}
}

func TestLoadSettings_NoEnvBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	content := `{"apiKey": "sk-something"}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.UseBedrock {
		t.Error("expected UseBedrock=false when no env block")
	}
}

func TestLoadSettings_PartialEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	content := `{"env": {"AWS_PROFILE": "only-profile"}}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.AWSProfile != "only-profile" {
		t.Errorf("expected AWSProfile='only-profile', got %q", s.AWSProfile)
	}
	if s.AWSRegion != "" {
		t.Errorf("expected empty AWSRegion, got %q", s.AWSRegion)
	}
	if s.UseBedrock {
		t.Error("expected UseBedrock=false when not set")
	}
}

func TestCommandEnv_AllSet(t *testing.T) {
	s := &Settings{
		UseBedrock: true,
		AWSProfile: "my-profile",
		AWSRegion:  "us-east-1",
	}

	env := s.CommandEnv()

	found := map[string]bool{}
	for _, e := range env {
		switch e {
		case "AWS_PROFILE=my-profile":
			found["profile"] = true
		case "AWS_REGION=us-east-1":
			found["region"] = true
		}
	}

	if !found["profile"] {
		t.Error("expected AWS_PROFILE in CommandEnv")
	}
	if !found["region"] {
		t.Error("expected AWS_REGION in CommandEnv")
	}
}

func TestCommandEnv_Empty(t *testing.T) {
	s := &Settings{}
	env := s.CommandEnv()
	if len(env) != 0 {
		t.Errorf("expected empty CommandEnv for zero-value Settings, got %v", env)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/claude/ -v`
Expected: FAIL — package does not exist yet.

**Step 3: Write minimal implementation**

Create `internal/claude/settings.go`:

```go
package claude

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

// Settings holds parsed values from ~/.claude/settings.json.
type Settings struct {
	UseBedrock bool
	AWSProfile string
	AWSRegion  string
}

// settingsFile represents the top-level structure of settings.json.
type settingsFile struct {
	Env map[string]string `json:"env"`
}

// LoadSettings reads and parses Claude Code's settings.json at the given path.
// Returns zero-value Settings (not an error) if the file does not exist,
// ensuring a missing file never crashes the statusline.
func LoadSettings(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Settings{}, nil
		}
		return nil, err
	}

	var sf settingsFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return &Settings{}, nil
	}

	s := &Settings{
		AWSProfile: sf.Env["AWS_PROFILE"],
		AWSRegion:  sf.Env["AWS_REGION"],
	}

	if sf.Env["CLAUDE_CODE_USE_BEDROCK"] == "true" || sf.Env["CLAUDE_CODE_USE_BEDROCK"] == "1" {
		s.UseBedrock = true
	}

	return s, nil
}

// CommandEnv returns KEY=VALUE pairs for non-empty settings, suitable for
// appending to exec.Cmd.Env to overlay on os.Environ().
func (s *Settings) CommandEnv() []string {
	if s == nil {
		return nil
	}

	var env []string
	if s.AWSProfile != "" {
		env = append(env, "AWS_PROFILE="+s.AWSProfile)
	}
	if s.AWSRegion != "" {
		env = append(env, "AWS_REGION="+s.AWSRegion)
	}
	return env
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/claude/ -v`
Expected: All 6 tests PASS.

**Step 5: Commit**

```
feat(claude): Add settings.json reader for Bedrock auth
```

---

### Task 2: Update `BedrockModel` to accept and use settings

**Files:**
- Modify: `internal/components/bedrock_model.go`
- Modify: `internal/components/bedrock_model_test.go`
- Modify: `main.go`

**Step 1: Update existing tests to pass nil settings**

In `internal/components/bedrock_model_test.go`, update all `NewBedrockModel` calls to pass a `nil` settings parameter. This ensures backward compatibility is tested. The `nil` case is handled by `CommandEnv()` returning `nil`.

Every call like:
```go
bm := NewBedrockModel(r, c, cfg)
```
becomes:
```go
bm := NewBedrockModel(r, c, cfg, nil)
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/components/ -run TestBedrockModel -v`
Expected: FAIL — `NewBedrockModel` doesn't accept 4 args yet.

**Step 3: Update BedrockModel struct and constructor**

In `internal/components/bedrock_model.go`:

Add import:
```go
"github.com/h2ik/claude-statusline/internal/claude"
```

Update struct:
```go
type BedrockModel struct {
	renderer *render.Renderer
	cache    *cache.Cache
	config   *config.Config
	settings *claude.Settings
}
```

Update constructor:
```go
func NewBedrockModel(r *render.Renderer, c *cache.Cache, cfg *config.Config, s *claude.Settings) *BedrockModel {
	return &BedrockModel{renderer: r, cache: c, config: cfg, settings: s}
}
```

Update `resolveBedrockARN` — add env and region flag to the `exec.Command`:
```go
func (c *BedrockModel) resolveBedrockARN(arn string) (string, string) {
	parts := strings.Split(arn, ":")
	region := ""
	if len(parts) >= 4 {
		region = parts[3]
	}

	cached, err := c.cache.Get("bedrock:"+arn, 24*time.Hour)
	if err == nil {
		fields := strings.SplitN(string(cached), "\t", 2)
		if len(fields) == 2 {
			return fields[0], fields[1]
		}
		return string(cached), region
	}

	args := []string{"bedrock", "get-inference-profile",
		"--inference-profile-identifier", arn,
		"--query", "models[0].modelArn",
		"--output", "text"}

	if c.settings != nil && c.settings.AWSRegion != "" {
		args = append(args, "--region", c.settings.AWSRegion)
	}

	cmd := exec.Command("aws", args...)
	cmd.Env = append(os.Environ(), c.settings.CommandEnv()...)

	output, err := cmd.Output()
	if err != nil {
		name := "Bedrock Model"
		c.cache.Set("bedrock:"+arn, []byte(name+"\t"+region), 24*time.Hour)
		return name, region
	}

	modelARN := strings.TrimSpace(string(output))
	friendlyName := c.getFriendlyName(modelARN)

	c.cache.Set("bedrock:"+arn, []byte(friendlyName+"\t"+region), 24*time.Hour)

	return friendlyName, region
}
```

Add `"os"` to the imports.

**Step 4: Update main.go wiring**

Add import:
```go
"github.com/h2ik/claude-statusline/internal/claude"
```

After `homeDir, _ := os.UserHomeDir()`, add:
```go
claudeSettings, _ := claude.LoadSettings(filepath.Join(homeDir, ".claude", "settings.json"))
```

Update registration:
```go
registry.Register(components.NewBedrockModel(r, c, cfg, claudeSettings))
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/components/ -run TestBedrockModel -v`
Expected: All 4 existing tests PASS.

Run: `go build ./...`
Expected: Build succeeds.

**Step 6: Commit**

```
feat(bedrock): Inject Claude settings.json env vars into AWS CLI calls
```

---

### Task 3: Add dynamic model catalog resolution

**Files:**
- Modify: `internal/components/bedrock_model.go`
- Modify: `internal/components/bedrock_model_test.go`

**Step 1: Write failing tests for catalog-based lookup**

Append to `internal/components/bedrock_model_test.go`:

```go
func TestGetFriendlyName_FromCatalog(t *testing.T) {
	r := render.New()
	cacheDir := t.TempDir()
	c := cache.New(cacheDir)
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	// Pre-seed the cache with a model catalog
	catalog := `[{"id":"anthropic.claude-opus-4-6-v1","name":"Claude Opus 4.6"},{"id":"anthropic.claude-sonnet-4-20250514-v1:0","name":"Claude Sonnet 4"}]`
	c.Set("bedrock:model-catalog", []byte(catalog), 24*time.Hour)

	bm := NewBedrockModel(r, c, cfg, nil)

	// Should match via catalog
	name := bm.getFriendlyName("arn:aws:bedrock:us-east-2:123456:foundation-model/anthropic.claude-opus-4-6-v1")
	if name != "Claude Opus 4.6" {
		t.Errorf("expected 'Claude Opus 4.6', got %q", name)
	}
}

func TestGetFriendlyName_FallbackToHardcoded(t *testing.T) {
	r := render.New()
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	// No catalog in cache — should fall back to hardcoded map
	bm := NewBedrockModel(r, c, cfg, nil)

	name := bm.getFriendlyName("arn:aws:bedrock:us-east-2:123456:foundation-model/anthropic.claude-sonnet-4-20250514-v1:0")
	if name != "Claude Sonnet 4" {
		t.Errorf("expected 'Claude Sonnet 4' from hardcoded fallback, got %q", name)
	}
}

func TestGetFriendlyName_RawARNFallback(t *testing.T) {
	r := render.New()
	c := cache.New(t.TempDir())
	cfg := &config.Config{Components: make(map[string]config.ComponentConfig)}

	bm := NewBedrockModel(r, c, cfg, nil)

	// Totally unknown model — should return the raw ARN
	name := bm.getFriendlyName("arn:aws:bedrock:us-east-2:123456:foundation-model/anthropic.claude-99-turbo-v1:0")
	if name != "arn:aws:bedrock:us-east-2:123456:foundation-model/anthropic.claude-99-turbo-v1:0" {
		t.Errorf("expected raw ARN passthrough, got %q", name)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/components/ -run TestGetFriendlyName -v`
Expected: FAIL — `getFriendlyName` doesn't check the catalog yet.

**Step 3: Implement dynamic model catalog**

Replace `getFriendlyName` and add `loadModelCatalog` in `bedrock_model.go`:

Add `"encoding/json"` to imports.

```go
// modelEntry represents a single model from the AWS API response.
type modelEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// loadModelCatalog returns the cached model catalog, or fetches it from
// the AWS API. Returns nil if the catalog is unavailable.
func (c *BedrockModel) loadModelCatalog() []modelEntry {
	cached, err := c.cache.Get("bedrock:model-catalog", 24*time.Hour)
	if err == nil {
		var models []modelEntry
		if json.Unmarshal(cached, &models) == nil {
			return models
		}
	}

	// Fetch from AWS CLI
	args := []string{"bedrock", "list-foundation-models",
		"--query", "modelSummaries[].{id:modelId,name:modelName}",
		"--output", "json"}

	if c.settings != nil && c.settings.AWSRegion != "" {
		args = append(args, "--region", c.settings.AWSRegion)
	}

	cmd := exec.Command("aws", args...)
	cmd.Env = append(os.Environ(), c.settings.CommandEnv()...)

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var models []modelEntry
	if json.Unmarshal(output, &models) != nil {
		return nil
	}

	c.cache.Set("bedrock:model-catalog", output, 24*time.Hour)
	return models
}

// getFriendlyName resolves a model ARN to a human-readable name.
// Checks the dynamic catalog first, falls back to a static map, then
// returns the raw ARN if nothing matches.
func (c *BedrockModel) getFriendlyName(modelARN string) string {
	// Try dynamic catalog first
	if catalog := c.loadModelCatalog(); catalog != nil {
		for _, m := range catalog {
			if strings.Contains(modelARN, m.ID) {
				return m.Name
			}
		}
	}

	// Static fallback for offline/no-creds scenarios
	fallback := map[string]string{
		"claude-opus-4":     "Claude Opus 4",
		"claude-sonnet-4":   "Claude Sonnet 4",
		"claude-3-5-sonnet": "Claude 3.5 Sonnet",
		"claude-3-5-haiku":  "Claude 3.5 Haiku",
		"claude-3-haiku":    "Claude 3 Haiku",
		"claude-3-opus":     "Claude 3 Opus",
	}

	for key, name := range fallback {
		if strings.Contains(modelARN, key) {
			return name
		}
	}

	return modelARN
}
```

Note: The static fallback map is trimmed — it only has generic entries for older models. The dynamic catalog handles all current and future models.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/components/ -run TestGetFriendlyName -v`
Expected: All 3 new tests PASS.

Run: `go test ./internal/components/ -v`
Expected: All tests PASS (existing + new).

Run: `go build ./...`
Expected: Build succeeds.

**Step 5: Commit**

```
feat(bedrock): Replace hardcoded model map with dynamic AWS catalog
```

---

### Task 4: Run full test suite and verify build

**Files:** None (verification only)

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: All tests PASS across all packages.

**Step 2: Build binary**

Run: `go build -o claude-statusline .`
Expected: Binary builds successfully.

**Step 3: Verify no vet/lint issues**

Run: `go vet ./...`
Expected: No issues.

**Step 4: Commit if any fixups were needed**

If any fixes were required, commit them:
```
fix(bedrock): Address test/lint issues from Bedrock improvements
```

---

### Task 5: Update documentation

**Files:**
- Modify: `docs/ARCHITECTURE.md`

**Step 1: Update ARCHITECTURE.md**

Add under the "External Commands" section, update the `aws` bullet:

```
- `aws` - for Bedrock model resolution and model catalog (optional; reads auth from `~/.claude/settings.json`)
```

Add a new section after "Configuration":

```markdown
## Claude Settings Integration

`internal/claude/` reads `~/.claude/settings.json` to extract AWS env vars
(`AWS_PROFILE`, `AWS_REGION`, `CLAUDE_CODE_USE_BEDROCK`) from the `.env` block.
Settings are loaded once at startup and injected into AWS CLI calls by the
`BedrockModel` component. Missing or unreadable settings degrade gracefully.
```

**Step 2: Commit**

```
docs: Document Claude settings.json integration
```
