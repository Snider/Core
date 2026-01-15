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

	t.Run("New service fails with invalid JSON in config", func(t *testing.T) {
		tempHomeDir, cleanup := setupTestEnv(t)
		defer cleanup()

		// Create config directory and write invalid JSON
		configDir := filepath.Join(tempHomeDir, appName, "config")
		if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
			t.Fatalf("Failed to create test config dir: %v", err)
		}
		configPath := filepath.Join(configDir, configFileName)

		invalidJSON := `{"language": invalid_value}`
		if err := os.WriteFile(configPath, []byte(invalidJSON), 0644); err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		_, err := New()
		if err == nil {
			t.Error("New() should fail when config file contains invalid JSON")
		}
		if err != nil && !contains(err.Error(), "failed to unmarshal config") {
			t.Errorf("Expected unmarshal error, got: %v", err)
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

	t.Run("Get with key not found", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		var value string
		err = s.Get("nonexistent_key", &value)
		if err == nil {
			t.Error("Get() should fail for nonexistent key")
		}
		if err != nil && err.Error() != "key 'nonexistent_key' not found in config" {
			t.Errorf("Expected 'key not found' error, got: %v", err)
		}
	})

	t.Run("Get with non-pointer output", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		var value string
		err = s.Get("language", value) // Not a pointer
		if err == nil {
			t.Error("Get() should fail for non-pointer output")
		}
		if err != nil && err.Error() != "output argument must be a non-nil pointer" {
			t.Errorf("Expected 'non-nil pointer' error, got: %v", err)
		}
	})

	t.Run("Get with nil pointer output", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		var value *string // nil pointer
		err = s.Get("language", value)
		if err == nil {
			t.Error("Get() should fail for nil pointer output")
		}
		if err != nil && err.Error() != "output argument must be a non-nil pointer" {
			t.Errorf("Expected 'non-nil pointer' error, got: %v", err)
		}
	})

	t.Run("Get with type mismatch", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		var value int // language is a string, not int
		err = s.Get("language", &value)
		if err == nil {
			t.Error("Get() should fail for type mismatch")
		}
		if err != nil && !contains(err.Error(), "cannot assign config value of type") {
			t.Errorf("Expected type mismatch error, got: %v", err)
		}
	})

	t.Run("Set with key not found", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		err = s.Set("nonexistent_key", "value")
		if err == nil {
			t.Error("Set() should fail for nonexistent key")
		}
		if err != nil && err.Error() != "key 'nonexistent_key' not found in config" {
			t.Errorf("Expected 'key not found' error, got: %v", err)
		}
	})

	t.Run("Set with type mismatch", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		err = s.Set("language", 123) // language expects string, not int
		if err == nil {
			t.Error("Set() should fail for type mismatch")
		}
		if err != nil && !contains(err.Error(), "type mismatch") {
			t.Errorf("Expected type mismatch error, got: %v", err)
		}
	})
}

// TestSaveStruct tests the SaveStruct function.
func TestSaveStruct(t *testing.T) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	t.Run("saves struct successfully", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		data := TestData{Name: "test", Value: 42}
		err = s.SaveStruct("test_data", data)
		if err != nil {
			t.Fatalf("SaveStruct() failed: %v", err)
		}

		// Verify file was created
		expectedPath := filepath.Join(s.ConfigDir, "test_data.json")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Expected file to be created at %s", expectedPath)
		}

		// Verify content
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("Failed to read saved file: %v", err)
		}
		if !contains(string(content), "\"name\": \"test\"") {
			t.Errorf("Saved content missing expected data: %s", content)
		}
	})

	t.Run("handles unmarshalable data", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		// Channels cannot be marshaled to JSON
		badData := make(chan int)
		err = s.SaveStruct("bad_data", badData)
		if err == nil {
			t.Error("SaveStruct() should fail for unmarshalable data")
		}
	})
}

// TestLoadStruct tests the LoadStruct function.
func TestLoadStruct(t *testing.T) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	t.Run("loads struct successfully", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		// First save some data
		original := TestData{Name: "loaded", Value: 99}
		err = s.SaveStruct("load_test", original)
		if err != nil {
			t.Fatalf("SaveStruct() failed: %v", err)
		}

		// Now load it
		var loaded TestData
		err = s.LoadStruct("load_test", &loaded)
		if err != nil {
			t.Fatalf("LoadStruct() failed: %v", err)
		}

		if loaded.Name != original.Name || loaded.Value != original.Value {
			t.Errorf("Loaded data mismatch: expected %+v, got %+v", original, loaded)
		}
	})

	t.Run("returns nil for nonexistent file", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		var data TestData
		err = s.LoadStruct("nonexistent_file", &data)
		if err != nil {
			t.Errorf("LoadStruct() should return nil for nonexistent file, got: %v", err)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		// Write invalid JSON
		invalidPath := filepath.Join(s.ConfigDir, "invalid.json")
		if err := os.WriteFile(invalidPath, []byte("not valid json"), 0644); err != nil {
			t.Fatalf("Failed to write invalid JSON: %v", err)
		}

		var data TestData
		err = s.LoadStruct("invalid", &data)
		if err == nil {
			t.Error("LoadStruct() should fail for invalid JSON")
		}
	})

	t.Run("returns error when file is a directory", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := New()
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		// Create a directory where the file should be
		dirPath := filepath.Join(s.ConfigDir, "dir_as_file.json")
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		var data TestData
		err = s.LoadStruct("dir_as_file", &data)
		if err == nil {
			t.Error("LoadStruct() should fail when path is a directory")
		}
	})
}

// contains is a helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
