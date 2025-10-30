package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNew ensures that New correctly initializes a Runtime instance.
func TestNew(t *testing.T) {
	runtime, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, runtime)

	// Assert that key services are initialized
	assert.NotNil(t, runtime.Core, "Core service should be initialized")
	assert.NotNil(t, runtime.Config, "Config service should be initialized")
	assert.NotNil(t, runtime.Display, "Display service should be initialized")
	assert.NotNil(t, runtime.Help, "Help service should be initialized")
	assert.NotNil(t, runtime.Crypt, "Crypt service should be initialized")
	assert.NotNil(t, runtime.I18n, "I18n service should be initialized")
	assert.NotNil(t, runtime.Workspace, "Workspace service should be initialized")

	// Verify services are properly wired through Core
	configFromCore := runtime.Core.Service("config")
	assert.Equal(t, runtime.Config, configFromCore, "Config from Core should match direct reference")

	displayFromCore := runtime.Core.Service("display")
	assert.Equal(t, runtime.Display, displayFromCore, "Display from Core should match direct reference")

	helpFromCore := runtime.Core.Service("help")
	assert.Equal(t, runtime.Help, helpFromCore, "Help from Core should match direct reference")

	cryptFromCore := runtime.Core.Service("crypt")
	assert.Equal(t, runtime.Crypt, cryptFromCore, "Crypt from Core should match direct reference")

	i18nFromCore := runtime.Core.Service("i18n")
	assert.Equal(t, runtime.I18n, i18nFromCore, "I18n from Core should match direct reference")

	workspaceFromCore := runtime.Core.Service("workspace")
	assert.Equal(t, runtime.Workspace, workspaceFromCore, "Workspace from Core should match direct reference")
}

// TestNewServiceInitializationError is a placeholder for testing error paths in New.
// TODO: Implement this test once the New function is refactored to allow for
// mockable service initializations (e.g., using dependency injection).
// The test should inject a failing service factory and assert that New returns
// the expected error and a nil runtime.
func TestNewServiceInitializationError(t *testing.T) {
	t.Skip("TODO: Implement once service initialization is mockable")
}

// Removed TestRuntimeOptions and TestRuntimeCore as these methods no longer exist on the Runtime struct.
