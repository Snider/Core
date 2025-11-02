package tdd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Snider/Core/pkg/config"
	"github.com/Snider/Core/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const appName = "lethean"
const configFileName = "config.json"

// setupTestEnv creates a temporary home directory for testing and ensures a clean environment.
func setupTestEnv(t *testing.T) (string, func()) {
	tempHomeDir, err := os.MkdirTemp("", "test_home_*")
	require.NoError(t, err, "Failed to create temp home directory")

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHomeDir)

	// Unset XDG vars to ensure HOME is used for path resolution, creating a hermetic test.
	oldXdgData, hadXdgData := os.LookupEnv("XDG_DATA_HOME")
	oldXdgCache, hadXdgCache := os.LookupEnv("XDG_CACHE_HOME")
	require.NoError(t, os.Unsetenv("XDG_DATA_HOME"))
	require.NoError(t, os.Unsetenv("XDG_CACHE_HOME"))

	cleanup := func() {
		os.Setenv("HOME", oldHome)
		if hadXdgData {
			os.Setenv("XDG_DATA_HOME", oldXdgData)
		} else {
			os.Unsetenv("XDG_DATA_HOME")
		}
		if hadXdgCache {
			os.Setenv("XDG_CACHE_HOME", oldXdgCache)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
		os.RemoveAll(tempHomeDir)
	}

	return tempHomeDir, cleanup
}

func TestConfigService(t *testing.T) {
	t.Run("New service creates default config", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		serviceInstance, err := config.New()
		require.NoError(t, err, "New() failed")

		// Check that the config file was created
		assert.FileExists(t, serviceInstance.ConfigPath, "config.json was not created")

		// Check default values
		assert.Equal(t, "en", serviceInstance.Language, "Expected default language 'en'")
	})

	t.Run("New service loads existing config", func(t *testing.T) {
		tempHomeDir, cleanup := setupTestEnv(t)
		defer cleanup()

		// Manually create a config file with non-default values
		configDir := filepath.Join(tempHomeDir, appName, "config")
		require.NoError(t, os.MkdirAll(configDir, os.ModePerm), "Failed to create test config dir")
		configPath := filepath.Join(configDir, configFileName)

		customConfig := `{"language": "fr", "features": ["beta-testing"]}`
		require.NoError(t, os.WriteFile(configPath, []byte(customConfig), 0644), "Failed to write custom config file")

		serviceInstance, err := config.New()
		require.NoError(t, err, "New() failed while loading existing config")

		assert.Equal(t, "fr", serviceInstance.Language, "Expected language 'fr'")
		assert.True(t, serviceInstance.IsFeatureEnabled("beta-testing"), "Expected 'beta-testing' feature to be enabled")
		assert.False(t, serviceInstance.IsFeatureEnabled("alpha-testing"), "Did not expect 'alpha-testing' to be enabled")
	})

	t.Run("Set and Get", func(t *testing.T) {
		_, cleanup := setupTestEnv(t)
		defer cleanup()

		s, err := config.New()
		require.NoError(t, err, "New() failed")

		key := "language"
		expectedValue := "de"
		require.NoError(t, s.Set(key, expectedValue), "Set() failed")

		var actualValue string
		require.NoError(t, s.Get(key, &actualValue), "Get() failed")
		assert.Equal(t, expectedValue, actualValue, "Get() returned unexpected value")
	})
}

func TestIsFeatureEnabled(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	s, err := config.New()
	require.NoError(t, err)

	// Test with no features enabled
	assert.False(t, s.IsFeatureEnabled("beta-feature"))

	// Enable a feature
	s.Features = []string{"beta-feature", "alpha-testing"}

	// Test for an enabled feature
	assert.True(t, s.IsFeatureEnabled("beta-feature"))

	// Test for another enabled feature
	assert.True(t, s.IsFeatureEnabled("alpha-testing"))

	// Test for a disabled feature
	assert.False(t, s.IsFeatureEnabled("gamma-feature"))

	// Test with an empty string
	assert.False(t, s.IsFeatureEnabled(""))

	// Test with a nil slice
	s.Features = nil
	assert.False(t, s.IsFeatureEnabled("beta-feature"))
}

func TestSet_Good(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	s, err := config.New()
	require.NoError(t, err, "New() failed")

	// Test setting a string value
	err = s.Set("language", "de")
	assert.NoError(t, err)
	var lang string
	err = s.Get("language", &lang)
	assert.NoError(t, err)
	assert.Equal(t, "de", lang)

	// Test setting a slice value
	err = s.Set("features", []string{"new-feature"})
	assert.NoError(t, err)
	var features []string
	err = s.Get("features", &features)
	assert.NoError(t, err)
	assert.Equal(t, []string{"new-feature"}, features)
}

func TestSet_Bad(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	s, err := config.New()
	require.NoError(t, err, "New() failed")

	// Test setting a value with the wrong type
	err = s.Set("language", 123)
	assert.Error(t, err)

	// Test setting a non-existent key
	err = s.Set("nonExistentKey", "value")
	assert.Error(t, err)
}

func TestSet_Ugly(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	s, err := config.New()
	require.NoError(t, err, "New() failed")

	// This should not panic
	assert.NotPanics(t, func() {
		err = s.Set("features", nil)
	})
	assert.NoError(t, err)

	// Verify the slice is now nil
	var features []string
	err = s.Get("features", &features)
	assert.NoError(t, err)
	assert.Nil(t, features)
}

func TestRegister_Good(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	c, err := core.New()
	require.NoError(t, err)

	svc, err := config.Register(c)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	configSvc, ok := svc.(*config.Service)
	assert.True(t, ok)
	assert.NotNil(t, configSvc.Runtime)
}
