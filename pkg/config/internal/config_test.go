package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	s, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for the config file
	tempDir, err := os.MkdirTemp("", "config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new config service
	s, err := New()
	assert.NoError(t, err)

	// Override the config path to use the temporary directory
	s.ConfigPath = filepath.Join(tempDir, "config.json")

	// Set some config values
	err = s.Set("language", "fr")
	assert.NoError(t, err)
	err = s.Set("default_route", "/home")
	assert.NoError(t, err)

	// Save the config
	err = s.Save()
	assert.NoError(t, err)

	// Create a new config service to load the saved config
	s2, err := New()
	assert.NoError(t, err)

	// Override the config path to use the temporary directory
	s2.ConfigPath = s.ConfigPath

	// Load the config
	data, err := os.ReadFile(s2.ConfigPath)
	assert.NoError(t, err)
	err = json.Unmarshal(data, s2)
	assert.NoError(t, err)

	// Check that the values were loaded correctly
	var lang string
	err = s2.Get("language", &lang)
	assert.NoError(t, err)
	assert.Equal(t, "fr", lang)

	var route string
	err = s2.Get("default_route", &route)
	assert.NoError(t, err)
	assert.Equal(t, "/home", route)
}

func TestIsFeatureEnabled(t *testing.T) {
	s, err := New()
	assert.NoError(t, err)

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
