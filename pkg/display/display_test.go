package display

import (
	"testing"

	"github.com/Snider/Core/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestCore creates a new core instance with essential services for testing.
func newTestCore(t *testing.T) *core.Core {
	// We need a real wails app for the display service to function.
	// This setup will be more complex than for other services.
	// For now, we can use a simplified core instance.
	coreInstance, err := core.New()
	require.NoError(t, err)
	return coreInstance
}

func TestNew(t *testing.T) {
	service, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, service, "New() should return a non-nil service instance")
}

func TestRegister(t *testing.T) {
	coreInstance := newTestCore(t)
	service, err := Register(coreInstance)
	require.NoError(t, err)
	assert.NotNil(t, service, "Register() should return a non-nil service instance")
}

func TestOpenWindow(t *testing.T) {
	// This test is complex to set up properly without a running Wails application.
	// A true functional test would require a more elaborate test harness that
	// can initialize the Wails runtime.

	// For now, we can perform a basic smoke test.
	t.Run("basic window open smoke test", func(t *testing.T) {
		// Skipping this test for now as it requires a running app instance.
		t.Skip("Skipping OpenWindow test as it requires a running Wails application instance.")
	})
}
