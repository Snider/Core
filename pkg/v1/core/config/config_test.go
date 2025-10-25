package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Snider/Core"
)

// setupTestEnv creates a temporary home directory for testing.
func setupTestEnv(t *testing.T) (string, func()) {
	tempHomeDir, err := os.MkdirTemp("", "test_home")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHomeDir)

	cleanup := func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(tempHomeDir)
	}

	return tempHomeDir, cleanup
}

// newTestCore creates a new, empty core instance for testing.
func newTestCore(t *testing.T) *core.Core {
	c := core.Service()
	if c == nil {
		t.Fatalf("core.Service() returned a nil instance, which is not expected for a test setup")
	}
	return c
}

func TestRegister(t *testing.T) {
	tempHomeDir, cleanup := setupTestEnv(t)
	defer cleanup()

	c := newTestCore(t)

	if err := Register(c); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	mod := c.Mod("config")
	if mod == nil {
		t.Fatalf("Failed to get config module from core instance")
	}

	cfg, ok := mod.(*Config)
	if !ok {
		t.Fatalf("Module is not of type *Config")
	}

	expectedUserHomeDir := filepath.Join(tempHomeDir, appName)
	expectedConfigDir := filepath.Join(expectedUserHomeDir, "config")
	expectedDataDir := filepath.Join(expectedUserHomeDir, "data")
	expectedWorkspacesDir := filepath.Join(expectedUserHomeDir, "workspaces")

	tests := []struct {
		name     string
		actual   string
		expected string
	}{
		{"UserHomeDir", cfg.UserHomeDir, expectedUserHomeDir},
		{"ConfigDir", cfg.ConfigDir, expectedConfigDir},
		{"DataDir", cfg.DataDir, expectedDataDir},
		{"WorkspacesDir", cfg.WorkspacesDir, expectedWorkspacesDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("Mismatch for %s: got %q, want %q", tt.name, tt.actual, tt.expected)
			}
		})
	}
}
