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
