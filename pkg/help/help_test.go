package help

import (
	"github.com/Snider/Core/pkg/core"
)

// MockDisplay is a mock implementation of the core.Display interface.
type MockDisplay struct {
	ShowCalled bool
}

func (m *MockDisplay) Show() error {
	m.ShowCalled = true
	return nil
}

func (m *MockDisplay) ShowAt(anchor string) error {
	m.ShowCalled = true
	return nil
}

func (m *MockDisplay) Hide() error                                { return nil }
func (m *MockDisplay) HideAt(anchor string) error                 { return nil }
func (m *MockDisplay) OpenWindow(opts ...core.WindowOption) error { return nil }

// MockCore is a mock implementation of the *core.Core type.
type MockCore struct {
	Core         *core.Core
	ActionCalled bool
}

func (m *MockCore) ACTION(msg core.Message) error {
	m.ActionCalled = true
	return nil
}
