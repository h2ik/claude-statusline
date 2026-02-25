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

	// Try resolving via AWS CLI. Use the region from the ARN itself (not settings)
	// since the inference profile must be queried in its own region.
	friendlyName := c.resolveViaAWSCLI(arn, region)

	_ = c.cache.Set("bedrock:"+arn, []byte(friendlyName+"\t"+region), 24*time.Hour)

	return friendlyName, region
}

// inferenceProfile holds the fields we care about from both get-inference-profile
// and list-inference-profiles responses.
type inferenceProfile struct {
	Name string `json:"inferenceProfileName"`
	ARN  string `json:"inferenceProfileArn"`
	Models []struct {
		ModelARN string `json:"modelArn"`
	} `json:"models"`
}

// resolveViaAWSCLI attempts to resolve an inference profile ARN to a friendly
// model name using the AWS CLI. It tries get-inference-profile first, then
// falls back to list-inference-profiles if permissions are insufficient.
func (c *BedrockModel) resolveViaAWSCLI(arn, arnRegion string) string {
	region := c.awsRegionArg(arnRegion)

	// Try get-inference-profile first (direct lookup)
	if profile := c.getInferenceProfile(arn, region); profile != nil {
		if name := c.nameFromProfile(profile); name != "" {
			return name
		}
	}

	// Fall back to list-inference-profiles (works with broader permissions)
	if profile := c.findInferenceProfile(arn, region); profile != nil {
		if name := c.nameFromProfile(profile); name != "" {
			return name
		}
	}

	return c.getFriendlyName(arn)
}

// awsRegionArg returns the region to use for AWS CLI calls.
func (c *BedrockModel) awsRegionArg(arnRegion string) string {
	if arnRegion != "" {
		return arnRegion
	}
	if c.settings != nil && c.settings.AWSRegion != "" {
		return c.settings.AWSRegion
	}
	return ""
}

// runAWS executes an AWS CLI command and returns the raw JSON output.
func (c *BedrockModel) runAWS(args []string, region string) ([]byte, error) {
	if region != "" {
		args = append(args, "--region", region)
	}
	cmd := exec.Command("aws", args...)
	if c.settings != nil {
		cmd.Env = append(os.Environ(), c.settings.CommandEnv()...)
	}
	return cmd.Output()
}

// getInferenceProfile tries the direct get-inference-profile API call.
func (c *BedrockModel) getInferenceProfile(arn, region string) *inferenceProfile {
	output, err := c.runAWS([]string{
		"bedrock", "get-inference-profile",
		"--inference-profile-identifier", arn,
		"--output", "json",
	}, region)
	if err != nil {
		return nil
	}
	var profile inferenceProfile
	if json.Unmarshal(output, &profile) != nil {
		return nil
	}
	return &profile
}

// findInferenceProfile searches list-inference-profiles for a matching ARN.
// It checks APPLICATION profiles first (custom), then SYSTEM_DEFINED.
func (c *BedrockModel) findInferenceProfile(arn, region string) *inferenceProfile {
	for _, profileType := range []string{"APPLICATION", "SYSTEM_DEFINED"} {
		output, err := c.runAWS([]string{
			"bedrock", "list-inference-profiles",
			"--type-equals", profileType,
			"--output", "json",
		}, region)
		if err != nil {
			continue
		}
		var resp struct {
			Profiles []inferenceProfile `json:"inferenceProfileSummaries"`
		}
		if json.Unmarshal(output, &resp) != nil {
			continue
		}
		for i := range resp.Profiles {
			if resp.Profiles[i].ARN == arn {
				return &resp.Profiles[i]
			}
		}
	}
	return nil
}

// nameFromProfile extracts a friendly model name from a resolved profile.
func (c *BedrockModel) nameFromProfile(p *inferenceProfile) string {
	// Try the underlying model ARN first (most specific)
	if len(p.Models) > 0 && p.Models[0].ModelARN != "" {
		if name := c.getFriendlyName(p.Models[0].ModelARN); name != p.Models[0].ModelARN && name != "Bedrock Model" {
			return name
		}
	}
	// Fall back to the inference profile's own name
	if p.Name != "" {
		return p.Name
	}
	return ""
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

	// Static fallback for offline/no-creds scenarios.
	// Order matters: more specific patterns first to avoid partial matches.
	fallback := map[string]string{
		"claude-opus-4-6":   "Claude Opus 4.6",
		"claude-opus-4":     "Claude Opus 4",
		"claude-sonnet-4-5": "Claude Sonnet 4.5",
		"claude-sonnet-4":   "Claude Sonnet 4",
		"claude-3-7-sonnet": "Claude 3.7 Sonnet",
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

	// If nothing matched and the input looks like a Bedrock ARN with an opaque
	// inference-profile ID, return a generic label instead of the raw ARN.
	if strings.HasPrefix(modelARN, "arn:aws:bedrock:") {
		return "Bedrock Model"
	}

	return modelARN
}
