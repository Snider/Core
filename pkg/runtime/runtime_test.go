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
	assert.NotNil(t, runtime.Config, "Config service should be initialized")
	assert.NotNil(t, runtime.Display, "Display service should be initialized")
	assert.NotNil(t, runtime.Help, "Help service should be initialized")
	assert.NotNil(t, runtime.Crypt, "Crypt service should be initialized")
	assert.NotNil(t, runtime.I18n, "I18n service should be initialized")
	assert.NotNil(t, runtime.Workspace, "Workspace service should be initialized")
}

// TestRuntimeServiceInitialization ensures that individual services within the Runtime are accessible.
func TestRuntimeServiceInitialization(t *testing.T) {
	runtime, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, runtime)

	// Test access to a specific service, e.g., Config
	configService := runtime.Config
	assert.NotNil(t, configService, "Config service should be accessible")

	// Test access to another service, e.g., Display
	displayService := runtime.Display
	assert.NotNil(t, displayService, "Display service should be accessible")

	// Add more assertions for other services as needed
}

// Removed TestRuntimeOptions and TestRuntimeCore as these methods no longer exist on the Runtime struct.
