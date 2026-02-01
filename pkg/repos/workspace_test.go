package repos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadWorkspaceConfig_Good(t *testing.T) {
	// Setup temp dir
	tmpDir := t.TempDir()
	coreDir := filepath.Join(tmpDir, ".core")
	err := os.MkdirAll(coreDir, 0755)
	assert.NoError(t, err)

	// Write valid config
	configContent := `
version: 1
active: core-php
packages_dir: ./custom-packages
`
	err = os.WriteFile(filepath.Join(coreDir, "workspace.yaml"), []byte(configContent), 0644)
	assert.NoError(t, err)

	// Load
	cfg, err := LoadWorkspaceConfig(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "core-php", cfg.Active)
	assert.Equal(t, "./custom-packages", cfg.PackagesDir)
}

func TestLoadWorkspaceConfig_Default(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Load non-existent
	cfg, err := LoadWorkspaceConfig(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "./packages", cfg.PackagesDir)
}

func TestLoadWorkspaceConfig_BadVersion(t *testing.T) {
	tmpDir := t.TempDir()
	coreDir := filepath.Join(tmpDir, ".core")
	os.MkdirAll(coreDir, 0755)

	configContent := `version: 2`
	os.WriteFile(filepath.Join(coreDir, "workspace.yaml"), []byte(configContent), 0644)

	_, err := LoadWorkspaceConfig(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported workspace config version")
}
