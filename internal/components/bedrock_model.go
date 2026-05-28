package components

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
	"github.com/h2ik/claude-statusline/internal/claude"
	"github.com/h2ik/claude-statusline/internal/config"
	"github.com/h2ik/claude-statusline/internal/icons"
	"github.com/h2ik/claude-statusline/internal/input"
	"github.com/h2ik/claude-statusline/internal/render"
)

// BedrockModel resolves AWS Bedrock inference profile ARNs to human-readable
// model names. It returns an empty string for non-Bedrock models so that
// model_info can handle those instead.
type BedrockModel struct {
	renderer *render.Renderer
	cache    *cache.Cache
	config   *config.Config
	settings *claude.Settings
	icons    icons.IconSet
}

// NewBedrockModel creates a new BedrockModel component with the given renderer, cache, config, and optional Claude settings.
func NewBedrockModel(r *render.Renderer, c *cache.Cache, cfg *config.Config, s *claude.Settings, ic icons.IconSet) *BedrockModel {
	return &BedrockModel{renderer: r, cache: c, config: cfg, settings: s, icons: ic}
}

// Name returns the component identifier used for registry lookup.
func (c *BedrockModel) Name() string {
	return "bedrock_model"
}

// Render produces the bedrock model string from the given input.
// Returns empty string when the model is NOT a Bedrock ARN -- this component
// only activates for Bedrock users. Non-Bedrock users get model_info instead.
func (c *BedrockModel) Render(in *input.StatusLineInput) string {
	modelName := in.Model.DisplayName

	if !strings.HasPrefix(modelName, "arn:aws:bedrock:") {
		return ""
	}

	// Strip trailing context window suffix (e.g. "[1m]") from the ARN
	arn := modelName
	if idx := strings.Index(arn, "["); idx != -1 {
		arn = arn[:idx]
	}

	name, region := c.resolveBedrockARN(arn)
	icon := c.getIcon(name)
	if region != "" && c.config.GetBool("bedrock_model", "show_region", true) {
		return fmt.Sprintf("%s %s %s", icon, c.renderer.Teal(name), c.renderer.Dimmed("("+region+")"))
	}
	return fmt.Sprintf("%s %s", icon, c.renderer.Teal(name))
}

// resolveBedrockARN resolves a Bedrock inference profile ARN to a friendly name
// and region, using the cache to avoid repeated AWS CLI calls.
// The cache stores "name\tregion" so both values survive round-trips.
func (c *BedrockModel) resolveBedrockARN(arn string) (string, string) {
	parts := strings.Split(arn, ":")
	region := ""
	if len(parts) >= 4 {
		region = parts[3]
	}

	cached, err := c.cache.Get("bedrock:v2:"+arn, 24*time.Hour)
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
	if c.settings != nil {
		cmd.Env = append(os.Environ(), c.settings.CommandEnv()...)
	}

	output, err := cmd.Output()
	if err != nil {
		// AWS CLI failed — try the profile catalog as a fallback.
		// This resolves application-inference-profile ARNs (opaque IDs)
		// by mapping them to their underlying model ARN.
		if profiles := c.loadProfileCatalog(); profiles != nil {
			for _, p := range profiles {
				if p.ARN == arn {
					friendlyName := c.getFriendlyName(p.ModelARN)
					_ = c.cache.Set("bedrock:v2:"+arn, []byte(friendlyName+"\t"+region), 24*time.Hour)
					return friendlyName, region
				}
			}
		}

		// Last resort: try matching the original ARN against the static
		// fallback (works for inference-profile ARNs that contain model
		// slugs like "claude-opus-4-6" in their ID).
		if name := c.getFriendlyName(arn); name != arn {
			_ = c.cache.Set("bedrock:v2:"+arn, []byte(name+"\t"+region), 24*time.Hour)
			return name, region
		}

		return "Bedrock Model", region
	}

	modelARN := strings.TrimSpace(string(output))
	friendlyName := c.getFriendlyName(modelARN)

	_ = c.cache.Set("bedrock:v2:"+arn, []byte(friendlyName+"\t"+region), 24*time.Hour)

	return friendlyName, region
}

// getIcon returns an icon based on the resolved model family name.
func (c *BedrockModel) getIcon(name string) string {
	lower := strings.ToLower(name)

	switch {
	case strings.Contains(lower, "opus"):
		return c.icons.Get(icons.Brain)
	case strings.Contains(lower, "haiku"):
		return c.icons.Get(icons.Lightning)
	case strings.Contains(lower, "sonnet"):
		return c.icons.Get(icons.Music)
	default:
		return c.icons.Get(icons.Robot)
	}
}

// profileEntry maps an inference profile ARN to its underlying model ARN.
type profileEntry struct {
	ARN      string `json:"arn"`
	ModelARN string `json:"modelArn"`
}

// loadProfileCatalog returns the cached inference profile catalog, or fetches
// it from the AWS API. Maps application-inference-profile ARNs to their
// underlying foundation-model ARNs.
func (c *BedrockModel) loadProfileCatalog() []profileEntry {
	cached, err := c.cache.Get("bedrock:v2:profile-catalog", 24*time.Hour)
	if err == nil {
		var profiles []profileEntry
		if json.Unmarshal(cached, &profiles) == nil {
			return profiles
		}
	}

	args := []string{"bedrock", "list-inference-profiles",
		"--type-equals", "APPLICATION",
		"--query", "inferenceProfileSummaries[].{arn:inferenceProfileArn,modelArn:models[0].modelArn}",
		"--output", "json"}

	if c.settings != nil && c.settings.AWSRegion != "" {
		args = append(args, "--region", c.settings.AWSRegion)
	}

	cmd := exec.Command("aws", args...)
	if c.settings != nil {
		cmd.Env = append(os.Environ(), c.settings.CommandEnv()...)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var profiles []profileEntry
	if json.Unmarshal(output, &profiles) != nil {
		return nil
	}

	_ = c.cache.Set("bedrock:v2:profile-catalog", output, 24*time.Hour)
	return profiles
}

// modelEntry represents a single model from the AWS API response.
type modelEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// loadModelCatalog returns the cached model catalog, or fetches it from
// the AWS API. Returns nil if the catalog is unavailable.
func (c *BedrockModel) loadModelCatalog() []modelEntry {
	cached, err := c.cache.Get("bedrock:v2:model-catalog", 24*time.Hour)
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
	if c.settings != nil {
		cmd.Env = append(os.Environ(), c.settings.CommandEnv()...)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var models []modelEntry
	if json.Unmarshal(output, &models) != nil {
		return nil
	}

	_ = c.cache.Set("bedrock:v2:model-catalog", output, 24*time.Hour)
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

	// Static fallback for offline/no-creds scenarios.
	// Ordered most-specific first so "claude-opus-4-6" matches before "claude-opus-4".
	fallback := []struct {
		key  string
		name string
	}{
		{"claude-opus-4-8", "Claude Opus 4.8"},
		{"claude-opus-4-7", "Claude Opus 4.7"},
		{"claude-opus-4-6", "Claude Opus 4.6"},
		{"claude-opus-4-5", "Claude Opus 4.5"},
		{"claude-opus-4", "Claude Opus 4"},
		{"claude-sonnet-4-7", "Claude Sonnet 4.7"},
		{"claude-sonnet-4-6", "Claude Sonnet 4.6"},
		{"claude-sonnet-4-5", "Claude Sonnet 4.5"},
		{"claude-sonnet-4", "Claude Sonnet 4"},
		{"claude-haiku-4-5", "Claude Haiku 4.5"},
		{"claude-haiku-4", "Claude Haiku 4"},
		{"claude-3-5-sonnet", "Claude 3.5 Sonnet"},
		{"claude-3-5-haiku", "Claude 3.5 Haiku"},
		{"claude-3-haiku", "Claude 3 Haiku"},
		{"claude-3-opus", "Claude 3 Opus"},
	}

	for _, f := range fallback {
		if strings.Contains(modelARN, f.key) {
			return f.name
		}
	}

	return modelARN
}
