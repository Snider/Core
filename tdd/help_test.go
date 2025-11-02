package tdd

import (
	"testing"

	"github.com/Snider/Core/pkg/core"
	"github.com/Snider/Core/pkg/help"
	"github.com/stretchr/testify/assert"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MockCore is a mock implementation of the core.Core for testing the help service.
type MockCore struct {
	*core.Core
	actionHistory []core.Message
}

// ACTION captures the message sent by the service.
func (m *MockCore) ACTION(msg core.Message) error {
	m.actionHistory = append(m.actionHistory, msg)
	return nil
}

func setupHelpService(t *testing.T) (*help.Service, *MockCore) {
	s, err := help.New()
	assert.NoError(t, err)

	app := application.New(application.Options{})
	c, err := core.New(core.WithWails(app))
	assert.NoError(t, err)

	mockCore := &MockCore{Core: c}
	// This is a bit of a hack, but it allows us to intercept the ACTION calls.
	c.RegisterAction(func(c *core.Core, msg core.Message) error {
		return mockCore.ACTION(msg)
	})

	s.Runtime = core.NewRuntime(c, help.Options{})
	s.SetDisplay(&MockDisplay{})

	return s, mockCore
}

// MockDisplay is a mock of the display service.
type MockDisplay struct{}

func (m *MockDisplay) OpenWindow(opts ...core.WindowOption) error { return nil }

// TestHelp_Show_Good tests the happy path for the Show method.
func TestHelp_Show_Good(t *testing.T) {
	s, mockCore := setupHelpService(t)

	err := s.Show()
	assert.NoError(t, err)
	assert.Len(t, mockCore.actionHistory, 1)

	msg, ok := mockCore.actionHistory[0].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "display.open_window", msg["action"])
}

// TestHelp_ShowAt_Good tests the happy path for the ShowAt method.
func TestHelp_ShowAt_Good(t *testing.T) {
	s, mockCore := setupHelpService(t)

	err := s.ShowAt("my-anchor")
	assert.NoError(t, err)
	assert.Len(t, mockCore.actionHistory, 1)

	msg, ok := mockCore.actionHistory[0].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "display.open_window", msg["action"])

	opts, ok := msg["options"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "/#my-anchor", opts["URL"])
}

// TestHelp_Show_Bad tests the case where the display service is not initialized.
func TestHelp_Show_Bad(t *testing.T) {
	s, _ := setupHelpService(t)
	// Unset the display service to simulate the bad case.
	// This is not possible in Go, so we'll have to find another way.
	// For now, we'll just assume the display is not set.
	// We'll simulate this by creating a new service without setting the display.
	s, err := help.New()
	assert.NoError(t, err)
	c, _ := core.New()
	s.Runtime = core.NewRuntime(c, help.Options{})

	err = s.Show()
	assert.Error(t, err)
	assert.Equal(t, "display service not initialized", err.Error())
}

// TestHelp_Show_Ugly tests the case where the core runtime is not initialized.
func TestHelp_Show_Ugly(t *testing.T) {
	s, err := help.New()
	assert.NoError(t, err)

	// Here, s.Runtime is nil, which should cause a panic.
	// However, the current implementation checks for a nil core, so it will return an error.
	err = s.Show()
	assert.Error(t, err)
	assert.Equal(t, "core runtime not initialized", err.Error())
}
