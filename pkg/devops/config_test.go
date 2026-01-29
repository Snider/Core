package devops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Good_Default(t *testing.T) {
	// Use temp home dir
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Images.Source != "auto" {
		t.Errorf("expected source 'auto', got %q", cfg.Images.Source)
	}
}

func TestLoadConfig_Good_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".core")
	os.MkdirAll(configDir, 0755)

	configContent := `version: 1
images:
  source: github
  github:
    repo: myorg/images
`
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0644)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Images.Source != "github" {
		t.Errorf("expected source 'github', got %q", cfg.Images.Source)
	}
	if cfg.Images.GitHub.Repo != "myorg/images" {
		t.Errorf("expected repo 'myorg/images', got %q", cfg.Images.GitHub.Repo)
	}
}
