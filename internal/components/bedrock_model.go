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

	name, region := c.resolveBedrockARN(modelName)
	icon := c.icons.Get(icons.Brain)
	if region != "" && c.config.GetBool("bedrock_model", "show_region", true) {
		return fmt.Sprintf("%s %s %s", icon, c.renderer.Text(name), c.renderer.Dimmed("("+region+")"))
	}
	return fmt.Sprintf("%s %s", icon, c.renderer.Text(name))
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
	if c.settings != nil {
		cmd.Env = append(os.Environ(), c.settings.CommandEnv()...)
	}

	output, err := cmd.Output()
	if err != nil {
		name := "Bedrock Model"
		_ = c.cache.Set("bedrock:"+arn, []byte(name+"\t"+region), 24*time.Hour)
		return name, region
	}

	modelARN := strings.TrimSpace(string(output))
	friendlyName := c.getFriendlyName(modelARN)

	_ = c.cache.Set("bedrock:"+arn, []byte(friendlyName+"\t"+region), 24*time.Hour)

	return friendlyName, region
}

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

	_ = c.cache.Set("bedrock:model-catalog", output, 24*time.Hour)
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
