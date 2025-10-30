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

	// Assert that key services are initialized (e.g., Config, Display, Help, Crypt, I18n, Workspace)
	assert.NotNil(t, runtime.Core, "Core service should be initialized")
	assert.NotNil(t, runtime.Config, "Config service should be initialized")
	assert.NotNil(t, runtime.Display, "Display service should be initialized")
	assert.NotNil(t, runtime.Help, "Help service should be initialized")
	assert.NotNil(t, runtime.Crypt, "Crypt service should be initialized")
	assert.NotNil(t, runtime.I18n, "I18n service should be initialized")
	assert.NotNil(t, runtime.Workspace, "Workspace service should be initialized")
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
