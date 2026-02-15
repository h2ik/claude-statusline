package input

import (
	"encoding/json"
	"fmt"
	"io"
)

type StatusLineInput struct {
	Workspace      Workspace     `json:"workspace"`
	Model          ModelInfo     `json:"model"`
	SessionID      string        `json:"session_id"`
	TranscriptPath string        `json:"transcript_path"`
	OutputStyle    OutputStyle   `json:"output_style"`
	ContextWindow  ContextWindow `json:"context_window"`
	Cost           CostInfo      `json:"cost"`
	CurrentUsage   UsageInfo     `json:"current_usage"`
	FiveHour       UsageLimit    `json:"five_hour"`
	SevenDay       UsageLimit    `json:"seven_day"`
	MCP            MCPInfo       `json:"mcp"`
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
	TotalCostUSD       float64 `json:"total_cost_usd"`
	TotalDurationMS    int     `json:"total_duration_ms"`
	TotalAPIDurationMS int     `json:"total_api_duration_ms"`
	TotalLinesAdded    int     `json:"total_lines_added"`
	TotalLinesRemoved  int     `json:"total_lines_removed"`
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

	if input.Workspace.CurrentDir == "" {
		return nil, fmt.Errorf("workspace.current_dir is required")
	}

	return &input, nil
}
