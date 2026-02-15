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
