package help

import (
	"context"
	"embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockLogger is a mock implementation of the Logger interface.
type MockLogger struct {
	InfoCalled  bool
	ErrorCalled bool
}

func (m *MockLogger) Info(message string, args ...any)  { m.InfoCalled = true }
func (m *MockLogger) Error(message string, args ...any) { m.ErrorCalled = true }

// MockApp is a mock implementation of the App interface.
type MockApp struct {
	logger Logger
}

func (m *MockApp) Logger() Logger { return m.logger }

// MockCore is a mock implementation of the Core interface.
type MockCore struct {
	ActionCalled bool
	ActionMsg    map[string]any
	app          App
	ActionErr    error
}

func (m *MockCore) ACTION(msg map[string]any) error {
	m.ActionCalled = true
	m.ActionMsg = msg
	return m.ActionErr
}

func (m *MockCore) App() App { return m.app }

// MockDisplay is a mock implementation of the Display interface.
type MockDisplay struct{}

//go:embed all:public/*
var testAssets embed.FS

func setupService(t *testing.T, opts Options) (*Service, *MockCore, *MockDisplay) {
	s, err := New(opts)
	assert.NoError(t, err)

	mockLogger := &MockLogger{}
	mockApp := &MockApp{logger: mockLogger}
	mockCore := &MockCore{app: mockApp}
	mockDisplay := &MockDisplay{}

	s.Init(mockCore, mockDisplay)

	return s, mockCore, mockDisplay
}

func TestNew(t *testing.T) {
	s, err := New(Options{})
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestNew_WithAssets(t *testing.T) {
	s, err := New(Options{Assets: testAssets})
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, testAssets, s.assets)
}

func TestServiceStartup(t *testing.T) {
	s, _, _ := setupService(t, Options{})
	err := s.ServiceStartup(context.Background())
	assert.NoError(t, err)
}

func TestShow(t *testing.T) {
	s, mockCore, _ := setupService(t, Options{})

	err := s.Show()
	assert.NoError(t, err)
	assert.True(t, mockCore.ActionCalled)

	msg := mockCore.ActionMsg
	assert.Equal(t, "display.open_window", msg["action"])
	assert.Equal(t, "help", msg["name"])
}

func TestShowAt(t *testing.T) {
	s, mockCore, _ := setupService(t, Options{})

	err := s.ShowAt("test-anchor")
	assert.NoError(t, err)
	assert.True(t, mockCore.ActionCalled)

	msg := mockCore.ActionMsg
	assert.Equal(t, "display.open_window", msg["action"])
	assert.Equal(t, "help", msg["name"])

	opts, ok := msg["options"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "/#test-anchor", opts["URL"])
}

func TestShowAt_CustomSource(t *testing.T) {
	s, mockCore, _ := setupService(t, Options{Source: "custom"})

	err := s.ShowAt("test-anchor")
	assert.NoError(t, err)
	assert.True(t, mockCore.ActionCalled)

	msg := mockCore.ActionMsg
	assert.Equal(t, "display.open_window", msg["action"])
	assert.Equal(t, "help", msg["name"])

	opts, ok := msg["options"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "/#test-anchor", opts["URL"])
}

func TestServiceStartup_CoreNotInitialized(t *testing.T) {
	s, _, _ := setupService(t, Options{})
	s.core = nil
	err := s.ServiceStartup(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "core runtime not initialized", err.Error())
}

func ExampleNew() {
	// Create a new help service with default options.
	// This demonstrates the simplest way to create a new service.
	// The service will use the default embedded "mkdocs" content.
	s, err := New(Options{})
	if err != nil {
		fmt.Printf("Error creating new service: %v", err)
		return
	}
	if s != nil {
		fmt.Println("Help service created successfully.")
	}

	// Output: Help service created successfully.
}

func ExampleService_Show() {
	// Create a new service and initialize it with mock dependencies.
	s, _ := New(Options{})
	s.Init(&MockCore{}, &MockDisplay{})

	// Call the Show method. In a real application, this would open a help window.
	// Since we are using a mock core, it will just record the action.
	if err := s.Show(); err != nil {
		fmt.Printf("Error showing help: %v", err)
	} else {
		fmt.Println("Show method called.")
	}

	// Output: Show method called.
}

func ExampleService_ShowAt() {
	// Create a new service and initialize it with mock dependencies.
	s, _ := New(Options{})
	s.Init(&MockCore{}, &MockDisplay{})

	// Call the ShowAt method. In a real application, this would open a help
	// window at a specific anchor.
	if err := s.ShowAt("getting-started"); err != nil {
		fmt.Printf("Error showing help at anchor: %v", err)
	} else {
		fmt.Println("ShowAt method called for 'getting-started'.")
	}

	// Output: ShowAt method called for 'getting-started'.
}

func TestGood_ShowAndShowAt_DispatchesCorrectPayload(t *testing.T) {
	s, mockCore, _ := setupService(t, Options{})

	// Test Show()
	err := s.Show()
	assert.NoError(t, err)
	assert.True(t, mockCore.ActionCalled)

	expectedMsgShow := map[string]any{
		"action": "display.open_window",
		"name":   "help",
		"options": map[string]any{
			"Title":  "Help",
			"Width":  800,
			"Height": 600,
		},
	}
	assert.Equal(t, expectedMsgShow, mockCore.ActionMsg)

	// Reset mock and test ShowAt()
	mockCore.ActionCalled = false
	mockCore.ActionMsg = nil

	err = s.ShowAt("good-anchor")
	assert.NoError(t, err)
	assert.True(t, mockCore.ActionCalled)

	expectedMsgShowAt := map[string]any{
		"action": "display.open_window",
		"name":   "help",
		"options": map[string]any{
			"Title":  "Help",
			"Width":  800,
			"Height": 600,
			"URL":    "/#good-anchor",
		},
	}
	assert.Equal(t, expectedMsgShowAt, mockCore.ActionMsg)
}

func TestBad_ShowAt_EmptyAnchor(t *testing.T) {
	s, mockCore, _ := setupService(t, Options{})

	err := s.ShowAt("")
	assert.NoError(t, err)
	assert.True(t, mockCore.ActionCalled)

	msg := mockCore.ActionMsg
	opts, ok := msg["options"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "/#", opts["URL"])
}

func TestUgly_ActionError_Propagates(t *testing.T) {
	s, mockCore, _ := setupService(t, Options{})

	// Simulate an error from the core.ACTION method
	expectedErr := assert.AnError
	mockCore.ActionErr = expectedErr

	// Test Show()
	err := s.Show()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	// Test ShowAt()
	err = s.ShowAt("any-anchor")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestShow_DisplayNotInitialized(t *testing.T) {
	s, _, _ := setupService(t, Options{})
	s.display = nil
	err := s.Show()
	assert.Error(t, err)
	assert.Equal(t, "wails application not running", err.Error())
}

func TestShow_CoreNotInitialized(t *testing.T) {
	s, _, _ := setupService(t, Options{})
	s.core = nil
	err := s.Show()
	assert.Error(t, err)
	assert.Equal(t, "core runtime not initialized", err.Error())
}

func TestShowAt_DisplayNotInitialized(t *testing.T) {
	s, _, _ := setupService(t, Options{})
	s.display = nil
	err := s.ShowAt("some-anchor")
	assert.Error(t, err)
	assert.Equal(t, "wails application not running", err.Error())
}

func TestShowAt_CoreNotInitialized(t *testing.T) {
	s, _, _ := setupService(t, Options{})
	s.core = nil
	err := s.ShowAt("some-anchor")
	assert.Error(t, err)
	assert.Equal(t, "core runtime not initialized", err.Error())
}
