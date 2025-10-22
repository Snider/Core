// Package core provides core utilities and helper functions.
// This is a starter library that can be extended with additional functionality.
package core

import (
	"fmt"
)

// Version represents the current version of the library.
const Version = "0.1.0"

// Greeter provides greeting functionality.
type Greeter struct {
	name string
}

// NewGreeter creates a new Greeter instance with the given name.
func NewGreeter(name string) *Greeter {
	return &Greeter{
		name: name,
	}
}

// Greet returns a greeting message.
func (g *Greeter) Greet() string {
	return fmt.Sprintf("Hello, %s!", g.name)
}

// GetVersion returns the current library version.
func GetVersion() string {
	return Version
}

// Add is a simple utility function that adds two integers.
func Add(a, b int) int {
	return a + b
}

// Multiply is a simple utility function that multiplies two integers.
func Multiply(a, b int) int {
	return a * b
}
