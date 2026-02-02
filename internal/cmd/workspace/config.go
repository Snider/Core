package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	coreio "github.com/host-uk/core/pkg/io"
	"gopkg.in/yaml.v3"
)

// WorkspaceConfig holds workspace-level configuration from .core/workspace.yaml.
type WorkspaceConfig struct {
	Version     int      `yaml:"version"`
	Active      string   `yaml:"active"`       // Active package name
	DefaultOnly []string `yaml:"default_only"` // Default types for setup
	PackagesDir string   `yaml:"packages_dir"` // Where packages are cloned
}

// DefaultConfig returns a config with default values.
func DefaultConfig() *WorkspaceConfig {
	return &WorkspaceConfig{
		Version:     1,
		PackagesDir: "./packages",
	}
}

// LoadConfig tries to load workspace.yaml from the given directory's .core subfolder.
// Returns nil if no config file exists (caller should check for nil).
func LoadConfig(dir string) (*WorkspaceConfig, error) {
	path := filepath.Join(dir, ".core", "workspace.yaml")
	data, err := coreio.Local.Read(path)
	if err != nil {
		// If using Local.Read, it returns error on not found.
		// We can check if file exists first or handle specific error if exposed.
		// Simplest is to check existence first or assume IsNotExist.
		// Since we don't have easy IsNotExist check on coreio error returned yet (uses wrapped error),
		// let's check IsFile first.
		if !coreio.Local.IsFile(path) {
			// Try parent directory
			parent := filepath.Dir(dir)
			if parent != dir {
				return LoadConfig(parent)
			}
			// No workspace.yaml found anywhere - return nil to indicate no config
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read workspace config: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal([]byte(data), config); err != nil {
		return nil, fmt.Errorf("failed to parse workspace config: %w", err)
	}

	if config.Version != 1 {
		return nil, fmt.Errorf("unsupported workspace config version: %d", config.Version)
	}

	return config, nil
}

// SaveConfig saves the configuration to the given directory's .core/workspace.yaml.
func SaveConfig(dir string, config *WorkspaceConfig) error {
	coreDir := filepath.Join(dir, ".core")
	if err := coreio.Local.EnsureDir(coreDir); err != nil {
		return fmt.Errorf("failed to create .core directory: %w", err)
	}

	path := filepath.Join(coreDir, "workspace.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace config: %w", err)
	}

	if err := coreio.Local.Write(path, string(data)); err != nil {
		return fmt.Errorf("failed to write workspace config: %w", err)
	}

	return nil
}

// FindWorkspaceRoot searches for the root directory containing .core/workspace.yaml.
func FindWorkspaceRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if coreio.Local.IsFile(filepath.Join(dir, ".core", "workspace.yaml")) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("not in a workspace")
}
