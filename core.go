// Package core provides a Wails v3 service with core utilities and helper functions.
// This is a starter service library that can be extended with additional functionality.
package core

import (
	"fmt"
)

// Version represents the current version of the service.
const Version = "0.1.0"

// CoreService provides core utilities and helper functions as a Wails v3 service.
// This service can be registered with a Wails application and called from the frontend.
type CoreService struct {
}

// NewCoreService creates a new CoreService instance.
func NewCoreService() *CoreService {
	return &CoreService{}
}

// ServiceName returns the name of the service.
// You should use the go module format e.g. github.com/myuser/myservice
func (s *CoreService) ServiceName() string {
	return "github.com/Snider/Core"
}

// GetVersion returns the current service version.
func (s *CoreService) GetVersion() string {
	return Version
}

// Greet returns a greeting message for the given name.
func (s *CoreService) Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Add adds two integers and returns the result.
func (s *CoreService) Add(a, b int) int {
	return a + b
}

// Multiply multiplies two integers and returns the result.
func (s *CoreService) Multiply(a, b int) int {
	return a * b
}

// Calculate performs a calculation based on the operation string.
// Supported operations: "add", "subtract", "multiply", "divide"
func (s *CoreService) Calculate(operation string, a, b int) (int, error) {
	switch operation {
	case "add":
		return a + b, nil
	case "subtract":
		return a - b, nil
	case "multiply":
		return a * b, nil
	case "divide":
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	default:
		return 0, fmt.Errorf("unknown operation: %s", operation)
	}
}
