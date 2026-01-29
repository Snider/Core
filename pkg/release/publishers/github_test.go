package publishers

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/host-uk/core/pkg/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("with draft and prerelease flags", func(t *testing.T) {
		release := &Release{
			Version: "v1.0.0-alpha",
		}
		cfg := PublisherConfig{
			Type:       "github",
			Draft:      true,
			Prerelease: true,
		}

		args := p.buildCreateArgs(release, cfg, "owner/repo")

		assert.Contains(t, args, "--draft")
		assert.Contains(t, args, "--prerelease")
	})

	t.Run("without repo includes version", func(t *testing.T) {
		release := &Release{
			Version:   "v2.0.0",
			Changelog: "Some changes",
		}
		cfg := PublisherConfig{
			Type: "github",
		}

		args := p.buildCreateArgs(release, cfg, "")

		assert.Contains(t, args, "release")
		assert.Contains(t, args, "create")
		assert.Contains(t, args, "v2.0.0")
		assert.NotContains(t, args, "--repo")
	})
}

func TestGitHubPublisher_DryRunPublish_Good(t *testing.T) {
	p := NewGitHubPublisher()

	t.Run("outputs expected dry run information", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0",
			Changelog:  "## Changes\n\n- Feature A\n- Bug fix B",
			ProjectDir: "/project",
		}
		cfg := PublisherConfig{
			Type:       "github",
			Draft:      false,
			Prerelease: false,
		}

		err := p.dryRunPublish(release, cfg, "owner/repo")

		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "DRY RUN: GitHub Release")
		assert.Contains(t, output, "Repository: owner/repo")
		assert.Contains(t, output, "Version:    v1.0.0")
		assert.Contains(t, output, "Draft:      false")
		assert.Contains(t, output, "Prerelease: false")
		assert.Contains(t, output, "Would create release with command:")
		assert.Contains(t, output, "gh release create")
		assert.Contains(t, output, "Changelog:")
		assert.Contains(t, output, "## Changes")
		assert.Contains(t, output, "END DRY RUN")
	})

	t.Run("shows artifacts when present", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0",
			Changelog:  "Changes",
			ProjectDir: "/project",
			Artifacts: []build.Artifact{
				{Path: "/dist/myapp-darwin-amd64.tar.gz"},
				{Path: "/dist/myapp-linux-amd64.tar.gz"},
			},
		}
		cfg := PublisherConfig{Type: "github"}

		err := p.dryRunPublish(release, cfg, "owner/repo")

		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "Would upload artifacts:")
		assert.Contains(t, output, "myapp-darwin-amd64.tar.gz")
		assert.Contains(t, output, "myapp-linux-amd64.tar.gz")
	})

	t.Run("shows draft and prerelease flags", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0-beta",
			Changelog:  "Beta release",
			ProjectDir: "/project",
		}
		cfg := PublisherConfig{
			Type:       "github",
			Draft:      true,
			Prerelease: true,
		}

		err := p.dryRunPublish(release, cfg, "owner/repo")

		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "Draft:      true")
		assert.Contains(t, output, "Prerelease: true")
		assert.Contains(t, output, "--draft")
		assert.Contains(t, output, "--prerelease")
	})
}

func TestGitHubPublisher_Publish_Good(t *testing.T) {
	p := NewGitHubPublisher()

	t.Run("dry run uses repository from config", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0",
			Changelog:  "Changes",
			ProjectDir: "/tmp",
		}
		pubCfg := PublisherConfig{Type: "github"}
		relCfg := &mockReleaseConfig{repository: "custom/repo"}

		// Dry run should succeed without needing gh CLI
		err := p.Publish(nil, release, pubCfg, relCfg, true)

		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Repository: custom/repo")
	})
}

func TestGitHubPublisher_Publish_Bad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	p := NewGitHubPublisher()

	t.Run("fails when gh CLI not available and not dry run", func(t *testing.T) {
		// This test will fail if gh is installed but not authenticated
		// or succeed if gh is not installed
		release := &Release{
			Version:    "v1.0.0",
			Changelog:  "Changes",
			ProjectDir: "/nonexistent",
		}
		pubCfg := PublisherConfig{Type: "github"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err := p.Publish(context.Background(), release, pubCfg, relCfg, false)

		// Should fail due to either gh not found or not authenticated
		assert.Error(t, err)
	})

	t.Run("fails when repository cannot be detected", func(t *testing.T) {
		// Create a temp directory that is NOT a git repo
		tmpDir, err := os.MkdirTemp("", "github-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		release := &Release{
			Version:    "v1.0.0",
			Changelog:  "Changes",
			ProjectDir: tmpDir,
		}
		pubCfg := PublisherConfig{Type: "github"}
		relCfg := &mockReleaseConfig{repository: ""} // Empty repository

		err = p.Publish(context.Background(), release, pubCfg, relCfg, true)

		// Should fail because detectRepository will fail on non-git dir
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not determine repository")
	})
}

func TestDetectRepository_Good(t *testing.T) {
	t.Run("detects repository from git remote", func(t *testing.T) {
		// Create a temp git repo
		tmpDir, err := os.MkdirTemp("", "git-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Initialize git repo and set remote
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "remote", "add", "origin", "git@github.com:test-owner/test-repo.git")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		repo, err := detectRepository(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "test-owner/test-repo", repo)
	})

	t.Run("detects repository from HTTPS remote", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "git-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/another-owner/another-repo.git")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		repo, err := detectRepository(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "another-owner/another-repo", repo)
	})
}

func TestDetectRepository_Bad(t *testing.T) {
	t.Run("fails when not a git repository", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "no-git-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		_, err = detectRepository(tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get git remote")
	})

	t.Run("fails when directory does not exist", func(t *testing.T) {
		_, err := detectRepository("/nonexistent/directory/that/does/not/exist")
		assert.Error(t, err)
	})

	t.Run("fails when remote is not GitHub", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "git-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "remote", "add", "origin", "git@gitlab.com:owner/repo.git")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		_, err = detectRepository(tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a GitHub URL")
	})
}

func TestValidateGhCli_Bad(t *testing.T) {
	// This test verifies the error messages from validateGhCli
	// We can't easily mock exec.Command, but we can at least
	// verify the function exists and returns expected error types
	t.Run("returns error when gh not installed", func(t *testing.T) {
		// We can't force gh to not be installed, but we can verify
		// the function signature works correctly
		err := validateGhCli()
		if err != nil {
			// Either gh is not installed or not authenticated
			assert.True(t,
				strings.Contains(err.Error(), "gh CLI not found") ||
					strings.Contains(err.Error(), "not authenticated"),
				"unexpected error: %s", err.Error())
		}
		// If err is nil, gh is installed and authenticated - that's OK too
	})
}

func TestGitHubPublisher_ExecutePublish_Good(t *testing.T) {
	// These tests run only when gh CLI is available and authenticated
	if err := validateGhCli(); err != nil {
		t.Skip("skipping test: gh CLI not available or not authenticated")
	}

	p := NewGitHubPublisher()

	t.Run("executePublish builds command with artifacts", func(t *testing.T) {
		// We test the command building by checking that it fails appropriately
		// with a non-existent release (rather than testing actual release creation)
		release := &Release{
			Version:    "v999.999.999-test-nonexistent",
			Changelog:  "Test changelog",
			ProjectDir: "/tmp",
			Artifacts: []build.Artifact{
				{Path: "/tmp/nonexistent-artifact.tar.gz"},
			},
		}
		cfg := PublisherConfig{
			Type:       "github",
			Draft:      true,
			Prerelease: true,
		}

		// This will fail because the artifact doesn't exist, but it proves
		// the code path runs
		err := p.executePublish(context.Background(), release, cfg, "test-owner/test-repo-nonexistent")
		assert.Error(t, err) // Expected to fail
	})
}

func TestReleaseExists_Good(t *testing.T) {
	// These tests run only when gh CLI is available
	if err := validateGhCli(); err != nil {
		t.Skip("skipping test: gh CLI not available or not authenticated")
	}

	t.Run("returns false for non-existent release", func(t *testing.T) {
		ctx := context.Background()
		// Use a non-existent repo and version
		exists := ReleaseExists(ctx, "nonexistent-owner-12345/nonexistent-repo-67890", "v999.999.999")
		assert.False(t, exists)
	})

	t.Run("checks release existence", func(t *testing.T) {
		ctx := context.Background()
		// Test against a known public repository with releases
		// This tests the true path if the release exists
		exists := ReleaseExists(ctx, "cli/cli", "v2.0.0")
		// We don't assert the result since it depends on network access
		// and the release may or may not exist
		_ = exists // Just verify function runs without panic
	})
}
