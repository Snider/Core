package core

import (
	"fmt"
	"testing"
)

func TestGreeter(t *testing.T) {
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
			greeter := NewGreeter(tt.input)
			result := greeter.Greet()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("expected non-empty version string")
	}
	if version != Version {
		t.Errorf("expected version %q, got %q", Version, version)
	}
}

func TestAdd(t *testing.T) {
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
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
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
			result := Multiply(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Multiply(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGreeter(b *testing.B) {
	greeter := NewGreeter("World")
	for i := 0; i < b.N; i++ {
		_ = greeter.Greet()
	}
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Add(42, 24)
	}
}

func BenchmarkMultiply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Multiply(42, 24)
	}
}

// Example tests
func ExampleGreeter() {
	greeter := NewGreeter("World")
	greeting := greeter.Greet()
	fmt.Println(greeting)
	// Output: Hello, World!
}

func ExampleAdd() {
	result := Add(2, 3)
	fmt.Println(result)
	// Output: 5
}
