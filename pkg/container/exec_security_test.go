package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeShellArg(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ls", "'ls'"},
		{"foo bar", "'foo bar'"},
		{"it's", "'it'\\''s'"},
		{"; rm -rf /", "'; rm -rf /'"},
		{"$(whoami)", "'$(whoami)'"},
		{"`whoami`", "'`whoami`'"},
		{"\"quoted\"", "'\"quoted\"'"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, escapeShellArg(tt.input))
		})
	}
}
