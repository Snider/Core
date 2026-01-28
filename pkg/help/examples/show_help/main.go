package main

import (
	"fmt"
	"log"

	"github.com/host-uk/core/pkg/help" // Assuming this is the import path for the help module
)

// MockLogger is a mock implementation of the help.Logger interface.
type MockLogger struct{}

func (m *MockLogger) Info(message string, args ...any)  { fmt.Println("INFO:", message) }
func (m *MockLogger) Error(message string, args ...any) { fmt.Println("ERROR:", message) }

// MockApp is a mock implementation of the help.App interface.
type MockApp struct {
	logger help.Logger
}

func (m *MockApp) Logger() help.Logger { return m.logger }

// MockCore is a mock implementation of the help.Core interface.
type MockCore struct {
	app help.App
}

func (m *MockCore) ACTION(msg map[string]any) error {
	fmt.Printf("ACTION called with: %v\n", msg)
	return nil
}

func (m *MockCore) App() help.App { return m.app }

// MockDisplay is a mock implementation of the help.Display interface.
type MockDisplay struct{}

// This example demonstrates how to use the Show() function in the refactored help module.
func main() {
	// 1. Initialize the help service.
	helpService, err := help.New(help.Options{})
	if err != nil {
		log.Fatalf("Failed to create help service: %v", err)
	}

	// 2. Create mock implementations of the required interfaces.
	mockLogger := &MockLogger{}
	mockApp := &MockApp{logger: mockLogger}
	mockCore := &MockCore{app: mockApp}
	mockDisplay := &MockDisplay{}

	// 3. Initialize the help service with the mock dependencies.
	helpService.Init(mockCore, mockDisplay)
	fmt.Println("Simulating a call to helpService.Show()")

	// 4. Call the Show() method.
	err = helpService.Show()
	if err != nil {
		log.Fatalf("Failed to show help window: %v", err)
	}

	fmt.Println("Successfully called helpService.Show().")
}
