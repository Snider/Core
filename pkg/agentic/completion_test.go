package agentic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCommitMessage(t *testing.T) {
	task := &Task{
		ID:    "ABC123",
		Title: "Test Task",
	}

	message := buildCommitMessage(task, "add new feature")

	assert.Contains(t, message, "add new feature")
	assert.Contains(t, message, "Task: #ABC123")
	assert.Contains(t, message, "Co-Authored-By: Claude <noreply@anthropic.com>")
}

func TestBuildPRBody(t *testing.T) {
	task := &Task{
		ID:          "PR-456",
		Title:       "Add authentication",
		Description: "Implement user authentication with OAuth2",
		Priority:    PriorityHigh,
		Labels:      []string{"enhancement", "security"},
	}

	body := buildPRBody(task)

	assert.Contains(t, body, "## Summary")
	assert.Contains(t, body, "Implement user authentication with OAuth2")
	assert.Contains(t, body, "## Task Reference")
	assert.Contains(t, body, "Task ID: #PR-456")
	assert.Contains(t, body, "Priority: high")
	assert.Contains(t, body, "Labels: enhancement, security")
	assert.Contains(t, body, "Generated with AI assistance")
}

func TestBuildPRBody_NoLabels(t *testing.T) {
	task := &Task{
		ID:          "PR-789",
		Title:       "Fix bug",
		Description: "Fix the login bug",
		Priority:    PriorityMedium,
		Labels:      nil,
	}

	body := buildPRBody(task)

	assert.Contains(t, body, "## Summary")
	assert.Contains(t, body, "Fix the login bug")
	assert.NotContains(t, body, "Labels:")
}

func TestGenerateBranchName(t *testing.T) {
	tests := []struct {
		name     string
		task     *Task
		expected string
	}{
		{
			name: "feature task",
			task: &Task{
				ID:     "123",
				Title:  "Add user authentication",
				Labels: []string{"enhancement"},
			},
			expected: "feat/123-add-user-authentication",
		},
		{
			name: "bug fix task",
			task: &Task{
				ID:     "456",
				Title:  "Fix login error",
				Labels: []string{"bug"},
			},
			expected: "fix/456-fix-login-error",
		},
		{
			name: "docs task",
			task: &Task{
				ID:     "789",
				Title:  "Update README",
				Labels: []string{"documentation"},
			},
			expected: "docs/789-update-readme",
		},
		{
			name: "refactor task",
			task: &Task{
				ID:     "101",
				Title:  "Refactor auth module",
				Labels: []string{"refactor"},
			},
			expected: "refactor/101-refactor-auth-module",
		},
		{
			name: "test task",
			task: &Task{
				ID:     "202",
				Title:  "Add unit tests",
				Labels: []string{"test"},
			},
			expected: "test/202-add-unit-tests",
		},
		{
			name: "chore task",
			task: &Task{
				ID:     "303",
				Title:  "Update dependencies",
				Labels: []string{"chore"},
			},
			expected: "chore/303-update-dependencies",
		},
		{
			name: "long title truncated",
			task: &Task{
				ID:     "404",
				Title:  "This is a very long title that should be truncated to fit the branch name limit",
				Labels: nil,
			},
			expected: "feat/404-this-is-a-very-long-title-that-should-be",
		},
		{
			name: "special characters removed",
			task: &Task{
				ID:     "505",
				Title:  "Fix: user's auth (OAuth2) [important]",
				Labels: nil,
			},
			expected: "feat/505-fix-users-auth-oauth2-important",
		},
		{
			name: "no labels defaults to feat",
			task: &Task{
				ID:     "606",
				Title:  "New feature",
				Labels: nil,
			},
			expected: "feat/606-new-feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateBranchName(tt.task)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoCommit_Bad_NilTask(t *testing.T) {
	err := AutoCommit(context.TODO(), nil, ".", "test message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task is required")
}

func TestAutoCommit_Bad_EmptyMessage(t *testing.T) {
	task := &Task{ID: "123", Title: "Test"}
	err := AutoCommit(context.TODO(), task, ".", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit message is required")
}

func TestSyncStatus_Bad_NilClient(t *testing.T) {
	task := &Task{ID: "123", Title: "Test"}
	update := TaskUpdate{Status: StatusInProgress}

	err := SyncStatus(context.TODO(), nil, task, update)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client is required")
}

func TestSyncStatus_Bad_NilTask(t *testing.T) {
	client := &Client{BaseURL: "http://test"}
	update := TaskUpdate{Status: StatusInProgress}

	err := SyncStatus(context.TODO(), client, nil, update)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task is required")
}

func TestCreateBranch_Bad_NilTask(t *testing.T) {
	branch, err := CreateBranch(context.TODO(), nil, ".")
	assert.Error(t, err)
	assert.Empty(t, branch)
	assert.Contains(t, err.Error(), "task is required")
}

func TestCreatePR_Bad_NilTask(t *testing.T) {
	url, err := CreatePR(context.TODO(), nil, ".", PROptions{})
	assert.Error(t, err)
	assert.Empty(t, url)
	assert.Contains(t, err.Error(), "task is required")
}
