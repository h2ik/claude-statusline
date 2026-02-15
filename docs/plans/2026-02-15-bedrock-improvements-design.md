# Bedrock Improvements Design

## Problem

The `BedrockModel` component has two gaps:

1. **No Claude settings.json integration.** The AWS CLI calls don't read `~/.claude/settings.json` for `AWS_PROFILE`, `AWS_REGION`, or `CLAUDE_CODE_USE_BEDROCK`. Users whose Bedrock auth depends on these settings get failures.
2. **Hardcoded model names.** `getFriendlyName` uses a static map of 8 known model fragments. New Anthropic models require code changes to display correctly.

## Design

### 1. New Package: `internal/claude/`

A standalone settings reader, reusable by future components.

```go
// Settings holds parsed values from ~/.claude/settings.json
type Settings struct {
    UseBedrock bool
    AWSProfile string
    AWSRegion  string
}
```

**`LoadSettings(path string) (*Settings, error)`** reads `~/.claude/settings.json`, parses the `.env` block with `encoding/json`, and extracts the three env vars. Returns zero-value `Settings` (not error) if the file doesn't exist -- graceful degradation.

**`(s *Settings) CommandEnv() []string`** returns `KEY=VALUE` pairs for non-empty settings, suitable for appending to `exec.Cmd.Env`.

### 2. BedrockModel Auth Fix

`BedrockModel` gains a `*claude.Settings` field, passed in via constructor.

All `exec.Command("aws", ...)` calls set:

```go
cmd.Env = append(os.Environ(), c.settings.CommandEnv()...)
```

This overlays Claude-specific vars on the user's environment. If settings are empty, behavior is unchanged from today.

The `--region` flag is explicitly passed to AWS CLI calls when `settings.AWSRegion` is set.

### 3. Dynamic Model Resolution

Replace the hardcoded model map with a cached AWS API lookup.

**`loadModelCatalog() map[string]string`:**
- Checks cache for `bedrock:model-catalog` with 24h TTL
- On miss, runs `aws bedrock list-foundation-models --query "modelSummaries[].{id:modelId,name:modelName}" --output json` with `CommandEnv()`
- Parses JSON into `map[modelId]modelName`
- Caches the serialized map as JSON
- Returns `nil` on failure (no CLI, no creds, offline)

**Updated `getFriendlyName(modelARN string)`:**
1. Load model catalog (cached)
2. Extract model ID fragment from the ARN
3. Look up in catalog -- exact match, then contains match
4. Fall back to a small hardcoded map (safety net for offline)
5. If still nothing, return the raw ARN

### 4. Wiring in `main.go`

```go
claudeSettings, _ := claude.LoadSettings(filepath.Join(homeDir, ".claude", "settings.json"))
registry.Register(components.NewBedrockModel(r, c, cfg, claudeSettings))
```

Error ignored intentionally -- missing settings.json must not crash the statusline.

## Testing

### `internal/claude/`
- `TestLoadSettings_ValidFile` -- parse settings, verify all fields
- `TestLoadSettings_MissingFile` -- returns zero-value, no error
- `TestLoadSettings_NoEnvBlock` -- settings.json with no `.env` key
- `TestLoadSettings_PartialEnv` -- only some vars set
- `TestCommandEnv_AllSet` -- returns KEY=VALUE pairs
- `TestCommandEnv_Empty` -- returns empty slice

### `bedrock_model.go`
- Existing tests pass (nil/empty settings)
- `TestGetFriendlyName_FromCatalog` -- pre-seeded cache, verify lookup
- `TestGetFriendlyName_FallbackToHardcoded` -- no catalog, uses static map
- `TestGetFriendlyName_RawARNFallback` -- unknown model, returns raw ARN

No integration tests that call the real AWS CLI.

## Decisions

- **Standalone package** over embedding in component -- reusable, clean separation
- **AWS CLI** over SDK -- keeps binary small, no new dependencies
- **list-foundation-models** over per-model lookups -- single cached call covers all models
- **Hardcoded map as fallback** -- works offline, safety net for API failures
