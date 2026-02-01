package repos

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WorkspaceConfig holds workspace-level configuration.
type WorkspaceConfig struct {
	Version     int    `yaml:"version"`
	Active      string `yaml:"active"`
	PackagesDir string `yaml:"packages_dir"`
}

// DefaultWorkspaceConfig returns a config with default values.
func DefaultWorkspaceConfig() *WorkspaceConfig {
	return &WorkspaceConfig{
		Version:     1,
		PackagesDir: "./packages",
	}
}

// LoadWorkspaceConfig tries to load workspace.yaml from the given directory's .core subfolder.
func LoadWorkspaceConfig(dir string) (*WorkspaceConfig, error) {
	path := filepath.Join(dir, ".core", "workspace.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultWorkspaceConfig(), nil
		}
		return nil, fmt.Errorf("failed to read workspace config: %w", err)
	}

	config := DefaultWorkspaceConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse workspace config: %w", err)
	}

	if config.Version != 1 {
		return nil, fmt.Errorf("unsupported workspace config version: %d", config.Version)
	}

	return config, nil
}
