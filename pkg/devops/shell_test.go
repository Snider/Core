package devops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellOptions_Default(t *testing.T) {
	opts := ShellOptions{}
	assert.False(t, opts.Console)
	assert.Nil(t, opts.Command)
}

func TestShellOptions_Console(t *testing.T) {
	opts := ShellOptions{
		Console: true,
	}
	assert.True(t, opts.Console)
	assert.Nil(t, opts.Command)
}

func TestShellOptions_Command(t *testing.T) {
	opts := ShellOptions{
		Command: []string{"ls", "-la"},
	}
	assert.False(t, opts.Console)
	assert.Equal(t, []string{"ls", "-la"}, opts.Command)
}

func TestShellOptions_ConsoleWithCommand(t *testing.T) {
	opts := ShellOptions{
		Console: true,
		Command: []string{"echo", "hello"},
	}
	assert.True(t, opts.Console)
	assert.Equal(t, []string{"echo", "hello"}, opts.Command)
}

func TestShellOptions_EmptyCommand(t *testing.T) {
	opts := ShellOptions{
		Command: []string{},
	}
	assert.False(t, opts.Console)
	assert.Empty(t, opts.Command)
	assert.Len(t, opts.Command, 0)
}
