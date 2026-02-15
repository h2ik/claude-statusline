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

// BedrockModel resolves AWS Bedrock inference profile ARNs to human-readable
// model names. It returns an empty string for non-Bedrock models so that
// model_info can handle those instead.
type BedrockModel struct {
	renderer *render.Renderer
	cache    *cache.Cache
}

// NewBedrockModel creates a new BedrockModel component with the given renderer and cache.
func NewBedrockModel(r *render.Renderer, c *cache.Cache) *BedrockModel {
	return &BedrockModel{renderer: r, cache: c}
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
	if region != "" {
		return fmt.Sprintf("🧠 %s %s", c.renderer.Text(name), c.renderer.Dimmed("("+region+")"))
	}
	return fmt.Sprintf("🧠 %s", c.renderer.Text(name))
}

// resolveBedrockARN resolves a Bedrock inference profile ARN to a friendly name
// and region, using the cache to avoid repeated AWS CLI calls.
// The cache stores "name\tregion" so both values survive round-trips.
func (c *BedrockModel) resolveBedrockARN(arn string) (string, string) {
	// Extract region from ARN (always available)
	parts := strings.Split(arn, ":")
	region := ""
	if len(parts) >= 4 {
		region = parts[3]
	}

	// Check cache — stored as "name\tregion"
	cached, err := c.cache.Get("bedrock:"+arn, 24*time.Hour)
	if err == nil {
		fields := strings.SplitN(string(cached), "\t", 2)
		if len(fields) == 2 {
			return fields[0], fields[1]
		}
		return string(cached), region
	}

	// Resolve via AWS CLI
	cmd := exec.Command("aws", "bedrock", "get-inference-profile",
		"--inference-profile-identifier", arn,
		"--query", "models[0].modelArn",
		"--output", "text")

	output, err := cmd.Output()
	if err != nil {
		name := "Bedrock Model"
		c.cache.Set("bedrock:"+arn, []byte(name+"\t"+region), 24*time.Hour)
		return name, region
	}

	modelARN := strings.TrimSpace(string(output))
	friendlyName := c.getFriendlyName(modelARN)

	// Cache as "name\tregion"
	c.cache.Set("bedrock:"+arn, []byte(friendlyName+"\t"+region), 24*time.Hour)

	return friendlyName, region
}

// getFriendlyName maps known model ARN fragments to human-readable names.
func (c *BedrockModel) getFriendlyName(modelARN string) string {
	mapping := map[string]string{
		"claude-opus-4-6":   "Claude Opus 4.6",
		"claude-opus-4":     "Claude Opus 4",
		"claude-sonnet-4-5": "Claude Sonnet 4.5",
		"claude-sonnet-4":   "Claude Sonnet 4",
		"claude-3-5-sonnet": "Claude 3.5 Sonnet",
		"claude-3-5-haiku":  "Claude 3.5 Haiku",
		"claude-3-haiku":    "Claude 3 Haiku",
		"claude-3-opus":     "Claude 3 Opus",
	}

	for key, name := range mapping {
		if strings.Contains(modelARN, key) {
			return name
		}
	}

	return modelARN
}
