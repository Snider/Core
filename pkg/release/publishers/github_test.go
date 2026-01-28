package publishers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitHubRepo_Good(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SSH URL",
			input:    "git@github.com:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS URL with .git",
			input:    "https://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS URL without .git",
			input:    "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "SSH URL without .git",
			input:    "git@github.com:owner/repo",
			expected: "owner/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseGitHubRepo(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseGitHubRepo_Bad(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "GitLab URL",
			input: "https://gitlab.com/owner/repo.git",
		},
		{
			name:  "Bitbucket URL",
			input: "git@bitbucket.org:owner/repo.git",
		},
		{
			name:  "Random URL",
			input: "https://example.com/something",
		},
		{
			name:  "Not a URL",
			input: "owner/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseGitHubRepo(tc.input)
			assert.Error(t, err)
		})
	}
}

func TestGitHubPublisher_Name_Good(t *testing.T) {
	t.Run("returns github", func(t *testing.T) {
		p := NewGitHubPublisher()
		assert.Equal(t, "github", p.Name())
	})
}

func TestNewRelease_Good(t *testing.T) {
	t.Run("creates release struct", func(t *testing.T) {
		r := NewRelease("v1.0.0", nil, "changelog", "/project")
		assert.Equal(t, "v1.0.0", r.Version)
		assert.Equal(t, "changelog", r.Changelog)
		assert.Equal(t, "/project", r.ProjectDir)
		assert.Nil(t, r.Artifacts)
	})
}

func TestNewPublisherConfig_Good(t *testing.T) {
	t.Run("creates config struct", func(t *testing.T) {
		cfg := NewPublisherConfig("github", true, false, nil)
		assert.Equal(t, "github", cfg.Type)
		assert.True(t, cfg.Prerelease)
		assert.False(t, cfg.Draft)
		assert.Nil(t, cfg.Extended)
	})

	t.Run("creates config with extended", func(t *testing.T) {
		ext := map[string]any{"key": "value"}
		cfg := NewPublisherConfig("docker", false, false, ext)
		assert.Equal(t, "docker", cfg.Type)
		assert.Equal(t, ext, cfg.Extended)
	})
}

func TestBuildCreateArgs_Good(t *testing.T) {
	p := NewGitHubPublisher()

	t.Run("basic args", func(t *testing.T) {
		release := &Release{
			Version:   "v1.0.0",
			Changelog: "## v1.0.0\n\nChanges",
		}
		cfg := PublisherConfig{
			Type: "github",
		}

		args := p.buildCreateArgs(release, cfg, "owner/repo")

		assert.Contains(t, args, "release")
		assert.Contains(t, args, "create")
		assert.Contains(t, args, "v1.0.0")
		assert.Contains(t, args, "--repo")
		assert.Contains(t, args, "owner/repo")
		assert.Contains(t, args, "--title")
		assert.Contains(t, args, "--notes")
	})

	t.Run("with draft flag", func(t *testing.T) {
		release := &Release{
			Version: "v1.0.0",
		}
		cfg := PublisherConfig{
			Type:  "github",
			Draft: true,
		}

		args := p.buildCreateArgs(release, cfg, "owner/repo")

		assert.Contains(t, args, "--draft")
	})

	t.Run("with prerelease flag", func(t *testing.T) {
		release := &Release{
			Version: "v1.0.0",
		}
		cfg := PublisherConfig{
			Type:       "github",
			Prerelease: true,
		}

		args := p.buildCreateArgs(release, cfg, "owner/repo")

		assert.Contains(t, args, "--prerelease")
	})

	t.Run("generates notes when no changelog", func(t *testing.T) {
		release := &Release{
			Version:   "v1.0.0",
			Changelog: "",
		}
		cfg := PublisherConfig{
			Type: "github",
		}

		args := p.buildCreateArgs(release, cfg, "owner/repo")

		assert.Contains(t, args, "--generate-notes")
	})
}
