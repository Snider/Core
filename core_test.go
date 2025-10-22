package core

import (
	"fmt"
	"testing"
)

func TestCoreService_Greet(t *testing.T) {
	service := NewCoreService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple greeting",
			input:    "World",
			expected: "Hello, World!",
		},
		{
			name:     "greeting with name",
			input:    "Alice",
			expected: "Hello, Alice!",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "Hello, !",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.Greet(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCoreService_GetVersion(t *testing.T) {
	service := NewCoreService()
	version := service.GetVersion()
	if version == "" {
		t.Error("expected non-empty version string")
	}
	if version != Version {
		t.Errorf("expected version %q, got %q", Version, version)
	}
}

func TestCoreService_ServiceName(t *testing.T) {
	service := NewCoreService()
	name := service.ServiceName()
	expectedName := "github.com/Snider/Core"
	if name != expectedName {
		t.Errorf("expected service name %q, got %q", expectedName, name)
	}
}

func TestCoreService_Add(t *testing.T) {
	service := NewCoreService()

	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "positive numbers",
			a:        2,
			b:        3,
			expected: 5,
		},
		{
			name:     "negative numbers",
			a:        -5,
			b:        3,
			expected: -2,
		},
		{
			name:     "zero values",
			a:        0,
			b:        0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCoreService_Multiply(t *testing.T) {
	service := NewCoreService()

	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "positive numbers",
			a:        2,
			b:        3,
			expected: 6,
		},
		{
			name:     "negative and positive",
			a:        -5,
			b:        3,
			expected: -15,
		},
		{
			name:     "multiplication by zero",
			a:        5,
			b:        0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.Multiply(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Multiply(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCoreService_Calculate(t *testing.T) {
	service := NewCoreService()

	tests := []struct {
		name      string
		operation string
		a         int
		b         int
		expected  int
		wantErr   bool
	}{
		{
			name:      "add operation",
			operation: "add",
			a:         10,
			b:         5,
			expected:  15,
			wantErr:   false,
		},
		{
			name:      "subtract operation",
			operation: "subtract",
			a:         10,
			b:         5,
			expected:  5,
			wantErr:   false,
		},
		{
			name:      "multiply operation",
			operation: "multiply",
			a:         10,
			b:         5,
			expected:  50,
			wantErr:   false,
		},
		{
			name:      "divide operation",
			operation: "divide",
			a:         10,
			b:         5,
			expected:  2,
			wantErr:   false,
		},
		{
			name:      "divide by zero",
			operation: "divide",
			a:         10,
			b:         0,
			expected:  0,
			wantErr:   true,
		},
		{
			name:      "unknown operation",
			operation: "modulo",
			a:         10,
			b:         5,
			expected:  0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Calculate(tt.operation, tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("Calculate(%q, %d, %d) = %d; expected %d", tt.operation, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkCoreService_Greet(b *testing.B) {
	service := NewCoreService()
	for i := 0; i < b.N; i++ {
		_ = service.Greet("World")
	}
}

func BenchmarkCoreService_Add(b *testing.B) {
	service := NewCoreService()
	for i := 0; i < b.N; i++ {
		_ = service.Add(42, 24)
	}
}

func BenchmarkCoreService_Multiply(b *testing.B) {
	service := NewCoreService()
	for i := 0; i < b.N; i++ {
		_ = service.Multiply(42, 24)
	}
}

func BenchmarkCoreService_Calculate(b *testing.B) {
	service := NewCoreService()
	for i := 0; i < b.N; i++ {
		_, _ = service.Calculate("add", 42, 24)
	}
}

// Example tests
func ExampleCoreService_Greet() {
	service := NewCoreService()
	greeting := service.Greet("World")
	fmt.Println(greeting)
	// Output: Hello, World!
}

func ExampleCoreService_Add() {
	service := NewCoreService()
	result := service.Add(2, 3)
	fmt.Println(result)
	// Output: 5
}

func ExampleCoreService_Calculate() {
	service := NewCoreService()
	result, _ := service.Calculate("multiply", 6, 7)
	fmt.Println(result)
	// Output: 42
}
