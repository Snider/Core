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

func TestConfigServiceGood(t *testing.T) {
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

	t.Run("Save and Load Struct", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		type CustomConfig struct {
			APIKey  string `json:"apiKey"`
			Timeout int    `json:"timeout"`
		}

		key := "custom"
		expectedConfig := CustomConfig{
			APIKey:  "12345",
			Timeout: 30,
		}

		if err := s.SaveStruct(key, expectedConfig); err != nil {
			t.Fatalf("SaveStruct() failed: %v", err)
		}

		var actualConfig CustomConfig
		if err := s.LoadStruct(key, &actualConfig); err != nil {
			t.Fatalf("LoadStruct() failed: %v", err)
		}

		if actualConfig.APIKey != expectedConfig.APIKey {
			t.Errorf("Expected APIKey '%s', got '%s'", expectedConfig.APIKey, actualConfig.APIKey)
		}
		if actualConfig.Timeout != expectedConfig.Timeout {
			t.Errorf("Expected Timeout '%d', got '%d'", expectedConfig.Timeout, actualConfig.Timeout)
		}
	})
}

func TestConfigServiceUgly(t *testing.T) {
	t.Run("LoadStruct with nil value", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		key := "nil-value"
		filePath := filepath.Join(s.ConfigDir, key+".json")
		if err := os.WriteFile(filePath, []byte("null"), 0644); err != nil {
			t.Fatalf("Failed to write nil value file: %v", err)
		}

		type CustomConfig struct {
			APIKey  string `json:"apiKey"`
			Timeout int    `json:"timeout"`
		}

		var actualConfig CustomConfig
		err = s.LoadStruct(key, &actualConfig)
		if err != nil {
			t.Fatalf("LoadStruct() should not have failed with a nil value, but it did: %v", err)
		}
	})

	t.Run("Concurrent access", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		// Run concurrent Set and Get operations
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				s.Set("language", "en")
				done <- true
			}()
			go func() {
				var lang string
				s.Get("language", &lang)
				done <- true
			}()
		}

		for i := 0; i < 20; i++ {
			<-done
		}
	})
}

func TestConfigServiceBad(t *testing.T) {
	t.Run("Load non-existent struct", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		type CustomConfig struct {
			APIKey  string `json:"apiKey"`
			Timeout int    `json:"timeout"`
		}

		var actualConfig CustomConfig
		if err := s.LoadStruct("non-existent", &actualConfig); err != nil {
			t.Fatalf("LoadStruct() failed: %v", err)
		}

		// Expect the struct to be zero-valued
		if actualConfig.APIKey != "" {
			t.Errorf("Expected empty APIKey, got '%s'", actualConfig.APIKey)
		}
		if actualConfig.Timeout != 0 {
			t.Errorf("Expected zero Timeout, got '%d'", actualConfig.Timeout)
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		var value string
		err = s.Get("non-existent", &value)
		if err == nil {
			t.Errorf("Expected an error for non-existent key, but got nil")
		}
	})

	t.Run("Set non-existent key", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		err = s.Set("non-existent", "value")
		if err == nil {
			t.Errorf("Expected an error for non-existent key, but got nil")
		}
	})

	t.Run("SaveStruct with unmarshallable type", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		err = s.SaveStruct("test", make(chan int))
		if err == nil {
			t.Errorf("Expected an error for unmarshallable type, but got nil")
		}
	})

	t.Run("LoadStruct with invalid JSON", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		key := "invalid"
		filePath := filepath.Join(s.ConfigDir, key+".json")
		if err := os.WriteFile(filePath, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("Failed to write invalid json file: %v", err)
		}

		type CustomConfig struct {
			APIKey  string `json:"apiKey"`
			Timeout int    `json:"timeout"`
		}

		var actualConfig CustomConfig
		err = s.LoadStruct(key, &actualConfig)
		if err == nil {
			t.Errorf("Expected an error for invalid JSON, but got nil")
		}
	})

	t.Run("New service with empty config file", func(t *testing.T) {
		tempHomeDir, cleanup := setupTestEnv(t)
		defer cleanup()

		// Manually create an empty config file
		configDir := filepath.Join(tempHomeDir, appName, "config")
		if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
			t.Fatalf("Failed to create test config dir: %v", err)
		}
		configPath := filepath.Join(configDir, configFileName)
		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to write empty config file: %v", err)
		}

		_, err := New()
		if err == nil {
			t.Fatalf("New() should have failed with an empty config file, but it did not")
		}
	})
}
