package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Snider/Core"
)

// setupTestEnv creates a temporary home directory for testing and ensures a clean environment.
func setupTestEnv(t *testing.T) (string, func()) {
	tempHomeDir, err := os.MkdirTemp("", "test_home_*")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHomeDir)

	// Unset XDG vars to ensure HOME is used for path resolution, creating a hermetic test.
	oldXdgData := os.Getenv("XDG_DATA_HOME")
	oldXdgCache := os.Getenv("XDG_CACHE_HOME")
	os.Unsetenv("XDG_DATA_HOME")
	os.Unsetenv("XDG_CACHE_HOME")

	cleanup := func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("XDG_DATA_HOME", oldXdgData)
		os.Setenv("XDG_CACHE_HOME", oldXdgCache)
		os.RemoveAll(tempHomeDir)
	}

	return tempHomeDir, cleanup
}

// newTestCore creates a new, empty core instance for testing.
func newTestCore(t *testing.T) *core.Core {
	c := core.New()
	if c == nil {
		t.Fatalf("core.New() returned a nil instance")
	}
	return c
}

func TestConfigService(t *testing.T) {
	t.Run("New service creates default config", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		c := newTestCore(t)
		serviceInstance, err := New(c)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}
		s, ok := serviceInstance.(*Service)
		if !ok {
			t.Fatalf("Service instance is not of type *Service")
		}

		// Check that the config file was created
		if _, err := os.Stat(s.ConfigPath); os.IsNotExist(err) {
			t.Errorf("config.json was not created at %s", s.ConfigPath)
		}

		// Check default values
		if s.Language != "en" {
			t.Errorf("Expected default language 'en', got '%s'", s.Language)
		}
	})

	t.Run("New service loads existing config", func(t *testing.T) {
		tempHomeDir, cleanup := setupTestEnv(t)
		defer cleanup()

		// Manually create a config file with non-default values
		configDir := filepath.Join(tempHomeDir, appName, "config")
		if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
			t.Fatalf("Failed to create test config dir: %v", err)
		}
		configPath := filepath.Join(configDir, configFileName)

		customConfig := `{"language": "fr", "features": ["beta-testing"]}`
		if err := os.WriteFile(configPath, []byte(customConfig), 0644); err != nil {
			t.Fatalf("Failed to write custom config file: %v", err)
		}

		c := newTestCore(t)
		serviceInstance, err := New(c)
		if err != nil {
			t.Fatalf("New() failed while loading existing config: %v", err)
		}
		s, ok := serviceInstance.(*Service)
		if !ok {
			t.Fatalf("Service instance is not of type *Service")
		}

		if s.Language != "fr" {
			t.Errorf("Expected language 'fr', got '%s'", s.Language)
		}
		if !s.IsFeatureEnabled("beta-testing") {
			t.Errorf("Expected 'beta-testing' feature to be enabled")
		}
	})

	t.Run("EnableFeature and Save", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		c := newTestCore(t)
		serviceInstance, err := New(c)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}
		s, ok := serviceInstance.(*Service)
		if !ok {
			t.Fatalf("Service instance is not of type *Service")
		}

		if err := s.EnableFeature("new-feature"); err != nil {
			t.Fatalf("EnableFeature() failed: %v", err)
		}

		data, err := os.ReadFile(s.ConfigPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var onDiskService Service
		if err := json.Unmarshal(data, &onDiskService); err != nil {
			t.Fatalf("Failed to unmarshal saved config: %v", err)
		}

		found := false
		for _, f := range onDiskService.Features {
			if f == "new-feature" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Enabled feature 'new-feature' was not saved to disk")
		}
	})
}
