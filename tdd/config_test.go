package tdd

import (
	"os"
	"testing"

	"github.com/Snider/Core/pkg/config"
	"github.com/stretchr/testify/assert"
)

// setupTestEnv creates a temporary home directory for testing and ensures a clean environment.
func setupTestEnv(t *testing.T) (string, func()) {
	tempHomeDir, err := os.MkdirTemp("", "test_home_*")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHomeDir)

	// Unset XDG vars to ensure HOME is used for path resolution, creating a hermetic test.
	originalXdgData := os.Getenv("XDG_DATA_HOME")
	originalXdgCache := os.Getenv("XDG_CACHE_HOME")
	os.Unsetenv("XDG_DATA_HOME")
	os.Unsetenv("XDG_CACHE_HOME")

	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("XDG_DATA_HOME", originalXdgData)
		os.Setenv("XDG_CACHE_HOME", originalXdgCache)
		os.RemoveAll(tempHomeDir)
	}

	return tempHomeDir, cleanup
}

// TestConfig_GetSet_Good tests the Get and Set methods with valid keys and values.
func TestConfig_GetSet_Good(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	s, err := config.New()
	assert.NoError(t, err, "config.New() should not return an error")

	key := "language"
	expectedValue := "de"
	err = s.Set(key, expectedValue)
	assert.NoError(t, err, "Set() should not return an error for a valid key and value")

	var actualValue string
	err = s.Get(key, &actualValue)
	assert.NoError(t, err, "Get() should not return an error for a valid key")
	assert.Equal(t, expectedValue, actualValue, "The retrieved value should match the set value")
}

// TestConfig_GetSet_Bad tests Get with a non-existent key and Set with a value of the wrong type.
func TestConfig_GetSet_Bad(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	s, err := config.New()
	assert.NoError(t, err, "config.New() should not return an error")

	// Test Get with a non-existent key
	var value string
	err = s.Get("nonexistent", &value)
	assert.Error(t, err, "Get() should return an error for a non-existent key")

	// Test Set with a value of the wrong type
	err = s.Set("language", 123)
	assert.Error(t, err, "Set() should return an error for a value of the wrong type")
}

// TestConfig_Get_Ugly tests Get with a nil pointer to ensure it returns an error instead of panicking.
func TestConfig_Get_Ugly(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	s, err := config.New()
	assert.NoError(t, err, "config.New() should not return an error")

	var value *string // A nil pointer
	err = s.Get("language", value)
	assert.Error(t, err, "Get() should return an error when a nil pointer is passed")
}
