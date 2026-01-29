package devops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClaudeOptions_Default(t *testing.T) {
	opts := ClaudeOptions{}
	assert.False(t, opts.NoAuth)
	assert.Nil(t, opts.Auth)
	assert.Empty(t, opts.Model)
}

func TestClaudeOptions_Custom(t *testing.T) {
	opts := ClaudeOptions{
		NoAuth: true,
		Auth:   []string{"gh", "anthropic"},
		Model:  "opus",
	}
	assert.True(t, opts.NoAuth)
	assert.Equal(t, []string{"gh", "anthropic"}, opts.Auth)
	assert.Equal(t, "opus", opts.Model)
}

func TestFormatAuthList_Good_NoAuth(t *testing.T) {
	opts := ClaudeOptions{NoAuth: true}
	result := formatAuthList(opts)
	assert.Equal(t, " (none)", result)
}

func TestFormatAuthList_Good_Default(t *testing.T) {
	opts := ClaudeOptions{}
	result := formatAuthList(opts)
	assert.Equal(t, ", gh, anthropic, git", result)
}

func TestFormatAuthList_Good_CustomAuth(t *testing.T) {
	opts := ClaudeOptions{
		Auth: []string{"gh"},
	}
	result := formatAuthList(opts)
	assert.Equal(t, ", gh", result)
}

func TestFormatAuthList_Good_MultipleAuth(t *testing.T) {
	opts := ClaudeOptions{
		Auth: []string{"gh", "ssh", "git"},
	}
	result := formatAuthList(opts)
	assert.Equal(t, ", gh, ssh, git", result)
}

func TestFormatAuthList_Good_EmptyAuth(t *testing.T) {
	opts := ClaudeOptions{
		Auth: []string{},
	}
	result := formatAuthList(opts)
	assert.Equal(t, ", gh, anthropic, git", result)
}
