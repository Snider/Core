package php

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertDeployment_Good(t *testing.T) {
	t.Run("converts all fields", func(t *testing.T) {
		now := time.Now()
		coolify := &CoolifyDeployment{
			ID:          "dep-123",
			Status:      "finished",
			CommitSHA:   "abc123",
			CommitMsg:   "Test commit",
			Branch:      "main",
			CreatedAt:   now,
			FinishedAt:  now.Add(5 * time.Minute),
			Log:         "Build successful",
			DeployedURL: "https://app.example.com",
		}

		status := convertDeployment(coolify)

		assert.Equal(t, "dep-123", status.ID)
		assert.Equal(t, "finished", status.Status)
		assert.Equal(t, "https://app.example.com", status.URL)
		assert.Equal(t, "abc123", status.Commit)
		assert.Equal(t, "Test commit", status.CommitMessage)
		assert.Equal(t, "main", status.Branch)
		assert.Equal(t, now, status.StartedAt)
		assert.Equal(t, now.Add(5*time.Minute), status.CompletedAt)
		assert.Equal(t, "Build successful", status.Log)
	})

	t.Run("handles empty deployment", func(t *testing.T) {
		coolify := &CoolifyDeployment{}
		status := convertDeployment(coolify)

		assert.Empty(t, status.ID)
		assert.Empty(t, status.Status)
	})
}

func TestDeploymentStatus_Struct_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		now := time.Now()
		status := DeploymentStatus{
			ID:            "dep-123",
			Status:        "finished",
			URL:           "https://app.example.com",
			Commit:        "abc123",
			CommitMessage: "Test commit",
			Branch:        "main",
			StartedAt:     now,
			CompletedAt:   now.Add(5 * time.Minute),
			Log:           "Build log",
		}

		assert.Equal(t, "dep-123", status.ID)
		assert.Equal(t, "finished", status.Status)
		assert.Equal(t, "https://app.example.com", status.URL)
		assert.Equal(t, "abc123", status.Commit)
		assert.Equal(t, "Test commit", status.CommitMessage)
		assert.Equal(t, "main", status.Branch)
		assert.Equal(t, "Build log", status.Log)
	})
}

func TestDeployOptions_Struct_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := DeployOptions{
			Dir:          "/project",
			Environment:  EnvProduction,
			Force:        true,
			Wait:         true,
			WaitTimeout:  10 * time.Minute,
			PollInterval: 5 * time.Second,
		}

		assert.Equal(t, "/project", opts.Dir)
		assert.Equal(t, EnvProduction, opts.Environment)
		assert.True(t, opts.Force)
		assert.True(t, opts.Wait)
		assert.Equal(t, 10*time.Minute, opts.WaitTimeout)
		assert.Equal(t, 5*time.Second, opts.PollInterval)
	})
}

func TestStatusOptions_Struct_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := StatusOptions{
			Dir:          "/project",
			Environment:  EnvStaging,
			DeploymentID: "dep-123",
		}

		assert.Equal(t, "/project", opts.Dir)
		assert.Equal(t, EnvStaging, opts.Environment)
		assert.Equal(t, "dep-123", opts.DeploymentID)
	})
}

func TestRollbackOptions_Struct_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := RollbackOptions{
			Dir:          "/project",
			Environment:  EnvProduction,
			DeploymentID: "dep-old",
			Wait:         true,
			WaitTimeout:  5 * time.Minute,
		}

		assert.Equal(t, "/project", opts.Dir)
		assert.Equal(t, EnvProduction, opts.Environment)
		assert.Equal(t, "dep-old", opts.DeploymentID)
		assert.True(t, opts.Wait)
		assert.Equal(t, 5*time.Minute, opts.WaitTimeout)
	})
}

func TestEnvironment_Constants(t *testing.T) {
	t.Run("constants are defined", func(t *testing.T) {
		assert.Equal(t, Environment("production"), EnvProduction)
		assert.Equal(t, Environment("staging"), EnvStaging)
	})
}

func TestGetAppIDForEnvironment_Edge(t *testing.T) {
	t.Run("staging without staging ID falls back to production", func(t *testing.T) {
		config := &CoolifyConfig{
			AppID: "prod-123",
			// No StagingAppID set
		}

		id := getAppIDForEnvironment(config, EnvStaging)
		assert.Equal(t, "prod-123", id)
	})

	t.Run("staging with staging ID uses staging", func(t *testing.T) {
		config := &CoolifyConfig{
			AppID:        "prod-123",
			StagingAppID: "staging-456",
		}

		id := getAppIDForEnvironment(config, EnvStaging)
		assert.Equal(t, "staging-456", id)
	})

	t.Run("production uses production ID", func(t *testing.T) {
		config := &CoolifyConfig{
			AppID:        "prod-123",
			StagingAppID: "staging-456",
		}

		id := getAppIDForEnvironment(config, EnvProduction)
		assert.Equal(t, "prod-123", id)
	})

	t.Run("unknown environment uses production", func(t *testing.T) {
		config := &CoolifyConfig{
			AppID: "prod-123",
		}

		id := getAppIDForEnvironment(config, "unknown")
		assert.Equal(t, "prod-123", id)
	})
}

func TestIsDeploymentComplete_Edge(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"finished", true},
		{"success", true},
		{"failed", true},
		{"error", true},
		{"cancelled", true},
		{"queued", false},
		{"building", false},
		{"deploying", false},
		{"pending", false},
		{"rolling_back", false},
		{"", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsDeploymentComplete(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsDeploymentSuccessful_Edge(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"finished", true},
		{"success", true},
		{"failed", false},
		{"error", false},
		{"cancelled", false},
		{"queued", false},
		{"building", false},
		{"deploying", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsDeploymentSuccessful(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
