package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecInContainer_Whitelist(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		expected string // Expected error substring
	}{
		{
			"Allowed command",
			[]string{"ls", "-la"},
			"", // Will fail later with "failed to determine state path" or similar, but NOT whitelist error
		},
		{
			"Disallowed command",
			[]string{"rm", "-rf", "/"},
			"command not allowed: rm",
		},
		{
			"Injection attempt in first arg",
			[]string{"ls; rm", "-rf", "/"},
			"command not allowed: ls; rm",
		},
		{
			"Empty command",
			[]string{},
			"container ID and command required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := execInContainer("test-id", tt.cmd)
			if tt.expected == "" {
				// Should NOT be a whitelist error
				if err != nil {
					assert.NotContains(t, err.Error(), "command not allowed")
				}
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
			}
		})
	}
}
