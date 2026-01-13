package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Snider/Core/pkg/core"
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
	c, err := core.New()
	if err != nil {
		t.Fatalf("core.New() failed: %v", err)
	}
	if c == nil {
		t.Fatalf("core.New() returned a nil instance")
	}
	return c
}

func TestConfigService(t *testing.T) {
	t.Run("New service creates default config", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		serviceInstance, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		// Check that the config file was created
		if _, err := os.Stat(serviceInstance.ConfigPath); os.IsNotExist(err) {
			t.Errorf("config.json was not created at %s", serviceInstance.ConfigPath)
		}

		// Check default values
		if serviceInstance.Language != "en" {
			t.Errorf("Expected default language 'en', got '%s'", serviceInstance.Language)
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

		serviceInstance, err := New()
		if err != nil {
			t.Fatalf("New() failed while loading existing config: %v", err)
		}

		if serviceInstance.Language != "fr" {
			t.Errorf("Expected language 'fr', got '%s'", serviceInstance.Language)
		}
		// A check for IsFeatureEnabled would require a proper core instance and service registration.
		// This is a simplified check for now.
		found := false
		for _, f := range serviceInstance.Features {
			if f == "beta-testing" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected 'beta-testing' feature to be enabled")
		}
	})

	t.Run("Set and Get", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		key := "language"
		expectedValue := "de"
		if err := s.Set(key, expectedValue); err != nil {
			t.Fatalf("Set() failed: %v", err)
		}

		var actualValue string
		if err := s.Get(key, &actualValue); err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if actualValue != expectedValue {
			t.Errorf("Expected value '%s', got '%s'", expectedValue, actualValue)
		}
	})

	t.Run("HandleIPCEvents with ActionServiceStartup", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		c := newTestCore(t)
		serviceAny, err := Register(c)
		if err != nil {
			t.Fatalf("Register() failed: %v", err)
		}

		s := serviceAny.(*Service)
		err = s.HandleIPCEvents(c, core.ActionServiceStartup{})
		if err != nil {
			t.Errorf("HandleIPCEvents(ActionServiceStartup) should not error, got: %v", err)
		}
	})

	t.Run("HandleIPCEvents with unknown message type", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		c := newTestCore(t)
		serviceAny, err := Register(c)
		if err != nil {
			t.Fatalf("Register() failed: %v", err)
		}

		s := serviceAny.(*Service)
		// Pass an arbitrary type as unknown message
		err = s.HandleIPCEvents(c, "unknown message")
		if err != nil {
			t.Errorf("HandleIPCEvents(unknown) should not error, got: %v", err)
		}
	})
}
