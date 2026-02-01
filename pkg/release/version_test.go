package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGitRepo creates a temporary directory with an initialized git repository.
func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	return dir
}

// createCommit creates a commit in the given directory.
func createCommit(t *testing.T, dir, message string) {
	t.Helper()

	// Create or modify a file
	filePath := filepath.Join(dir, "test.txt")
	content, _ := os.ReadFile(filePath)
	content = append(content, []byte(message+"\n")...)
	require.NoError(t, os.WriteFile(filePath, content, 0644))

	// Stage and commit
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
}

// createTag creates a tag in the given directory.
func createTag(t *testing.T, dir, tag string) {
	t.Helper()
	cmd := exec.Command("git", "tag", tag)
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
}

func TestDetermineVersion_Good(t *testing.T) {
	t.Run("returns tag when HEAD has tag", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")
		createTag(t, dir, "v1.0.0")

		version, err := DetermineVersion(dir)
		require.NoError(t, err)
		assert.Equal(t, "v1.0.0", version)
	})

	t.Run("normalizes tag without v prefix", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")
		createTag(t, dir, "1.0.0")

		version, err := DetermineVersion(dir)
		require.NoError(t, err)
		assert.Equal(t, "v1.0.0", version)
	})

	t.Run("increments patch when commits after tag", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")
		createTag(t, dir, "v1.0.0")
		createCommit(t, dir, "feat: new feature")

		version, err := DetermineVersion(dir)
		require.NoError(t, err)
		assert.Equal(t, "v1.0.1", version)
	})

	t.Run("returns v0.0.1 when no tags exist", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")

		version, err := DetermineVersion(dir)
		require.NoError(t, err)
		assert.Equal(t, "v0.0.1", version)
	})

	t.Run("handles multiple tags with increments", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: first")
		createTag(t, dir, "v1.0.0")
		createCommit(t, dir, "feat: second")
		createTag(t, dir, "v1.0.1")
		createCommit(t, dir, "feat: third")

		version, err := DetermineVersion(dir)
		require.NoError(t, err)
		assert.Equal(t, "v1.0.2", version)
	})
}

func TestDetermineVersion_Bad(t *testing.T) {
	t.Run("returns v0.0.1 for empty repo", func(t *testing.T) {
		dir := setupGitRepo(t)

		// No commits, git describe will fail
		version, err := DetermineVersion(dir)
		require.NoError(t, err)
		assert.Equal(t, "v0.0.1", version)
	})
}

func TestGetTagOnHead_Good(t *testing.T) {
	t.Run("returns tag when HEAD has tag", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")
		createTag(t, dir, "v1.2.3")

		tag, err := getTagOnHead(dir)
		require.NoError(t, err)
		assert.Equal(t, "v1.2.3", tag)
	})

	t.Run("returns latest tag when multiple tags on HEAD", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")
		createTag(t, dir, "v1.0.0")
		createTag(t, dir, "v1.0.0-beta")

		tag, err := getTagOnHead(dir)
		require.NoError(t, err)
		// Git returns one of the tags
		assert.Contains(t, []string{"v1.0.0", "v1.0.0-beta"}, tag)
	})
}

func TestGetTagOnHead_Bad(t *testing.T) {
	t.Run("returns error when HEAD has no tag", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")

		_, err := getTagOnHead(dir)
		assert.Error(t, err)
	})

	t.Run("returns error when commits after tag", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")
		createTag(t, dir, "v1.0.0")
		createCommit(t, dir, "feat: new feature")

		_, err := getTagOnHead(dir)
		assert.Error(t, err)
	})
}

func TestGetLatestTag_Good(t *testing.T) {
	t.Run("returns latest tag", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")
		createTag(t, dir, "v1.0.0")

		tag, err := getLatestTag(dir)
		require.NoError(t, err)
		assert.Equal(t, "v1.0.0", tag)
	})

	t.Run("returns most recent tag after multiple commits", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: first")
		createTag(t, dir, "v1.0.0")
		createCommit(t, dir, "feat: second")
		createTag(t, dir, "v1.1.0")
		createCommit(t, dir, "feat: third")

		tag, err := getLatestTag(dir)
		require.NoError(t, err)
		assert.Equal(t, "v1.1.0", tag)
	})
}

func TestGetLatestTag_Bad(t *testing.T) {
	t.Run("returns error when no tags exist", func(t *testing.T) {
		dir := setupGitRepo(t)
		createCommit(t, dir, "feat: initial commit")

		_, err := getLatestTag(dir)
		assert.Error(t, err)
	})

	t.Run("returns error for empty repo", func(t *testing.T) {
		dir := setupGitRepo(t)

		_, err := getLatestTag(dir)
		assert.Error(t, err)
	})
}

func TestIncrementMinor_Bad(t *testing.T) {
	t.Run("returns fallback for invalid version", func(t *testing.T) {
		result := IncrementMinor("not-valid")
		assert.Equal(t, "not-valid.1", result)
	})
}

func TestIncrementMajor_Bad(t *testing.T) {
	t.Run("returns fallback for invalid version", func(t *testing.T) {
		result := IncrementMajor("not-valid")
		assert.Equal(t, "not-valid.1", result)
	})
}

func TestCompareVersions_Ugly(t *testing.T) {
	t.Run("handles both invalid versions", func(t *testing.T) {
		result := CompareVersions("invalid-a", "invalid-b")
		// Should do string comparison for invalid versions
		assert.Equal(t, -1, result) // "invalid-a" < "invalid-b"
	})

	t.Run("invalid a returns -1", func(t *testing.T) {
		result := CompareVersions("invalid", "v1.0.0")
		assert.Equal(t, -1, result)
	})

	t.Run("invalid b returns 1", func(t *testing.T) {
		result := CompareVersions("v1.0.0", "invalid")
		assert.Equal(t, 1, result)
	})
}

func TestIncrementVersion_Good(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "increment patch with v prefix",
			input:    "v1.2.3",
			expected: "v1.2.4",
		},
		{
			name:     "increment patch without v prefix",
			input:    "1.2.3",
			expected: "v1.2.4",
		},
		{
			name:     "increment from zero",
			input:    "v0.0.0",
			expected: "v0.0.1",
		},
		{
			name:     "strips prerelease",
			input:    "v1.2.3-alpha",
			expected: "v1.2.4",
		},
		{
			name:     "strips build metadata",
			input:    "v1.2.3+build123",
			expected: "v1.2.4",
		},
		{
			name:     "strips prerelease and build",
			input:    "v1.2.3-beta.1+build456",
			expected: "v1.2.4",
		},
		{
			name:     "handles large numbers",
			input:    "v10.20.99",
			expected: "v10.20.100",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IncrementVersion(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIncrementVersion_Bad(t *testing.T) {
	t.Run("invalid semver returns original with suffix", func(t *testing.T) {
		result := IncrementVersion("not-a-version")
		assert.Equal(t, "not-a-version.1", result)
	})
}

func TestIncrementMinor_Good(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "increment minor resets patch",
			input:    "v1.2.3",
			expected: "v1.3.0",
		},
		{
			name:     "increment minor from zero",
			input:    "v1.0.5",
			expected: "v1.1.0",
		},
		{
			name:     "handles large numbers",
			input:    "v5.99.50",
			expected: "v5.100.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IncrementMinor(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIncrementMajor_Good(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "increment major resets minor and patch",
			input:    "v1.2.3",
			expected: "v2.0.0",
		},
		{
			name:     "increment major from zero",
			input:    "v0.5.10",
			expected: "v1.0.0",
		},
		{
			name:     "handles large numbers",
			input:    "v99.50.25",
			expected: "v100.0.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IncrementMajor(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseVersion_Good(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		major      int
		minor      int
		patch      int
		prerelease string
		build      string
	}{
		{
			name:  "simple version with v",
			input: "v1.2.3",
			major: 1, minor: 2, patch: 3,
		},
		{
			name:  "simple version without v",
			input: "1.2.3",
			major: 1, minor: 2, patch: 3,
		},
		{
			name:  "with prerelease",
			input: "v1.2.3-alpha",
			major: 1, minor: 2, patch: 3,
			prerelease: "alpha",
		},
		{
			name:  "with prerelease and build",
			input: "v1.2.3-beta.1+build.456",
			major: 1, minor: 2, patch: 3,
			prerelease: "beta.1",
			build:      "build.456",
		},
		{
			name:  "with build only",
			input: "v1.2.3+sha.abc123",
			major: 1, minor: 2, patch: 3,
			build: "sha.abc123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			major, minor, patch, prerelease, build, err := ParseVersion(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.major, major)
			assert.Equal(t, tc.minor, minor)
			assert.Equal(t, tc.patch, patch)
			assert.Equal(t, tc.prerelease, prerelease)
			assert.Equal(t, tc.build, build)
		})
	}
}

func TestParseVersion_Bad(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"not a version", "not-a-version"},
		{"missing minor", "v1"},
		{"missing patch", "v1.2"},
		{"letters in version", "v1.2.x"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, _, _, err := ParseVersion(tc.input)
			assert.Error(t, err)
		})
	}
}

func TestValidateVersion_Good(t *testing.T) {
	validVersions := []string{
		"v1.0.0",
		"1.0.0",
		"v0.0.1",
		"v10.20.30",
		"v1.2.3-alpha",
		"v1.2.3+build",
		"v1.2.3-alpha.1+build.123",
	}

	for _, v := range validVersions {
		t.Run(v, func(t *testing.T) {
			assert.True(t, ValidateVersion(v))
		})
	}
}

func TestValidateVersion_Bad(t *testing.T) {
	invalidVersions := []string{
		"",
		"v1",
		"v1.2",
		"1.2",
		"not-a-version",
		"v1.2.x",
		"version1.0.0",
	}

	for _, v := range invalidVersions {
		t.Run(v, func(t *testing.T) {
			assert.False(t, ValidateVersion(v))
		})
	}
}

func TestCompareVersions_Good(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"equal versions", "v1.0.0", "v1.0.0", 0},
		{"a less than b major", "v1.0.0", "v2.0.0", -1},
		{"a greater than b major", "v2.0.0", "v1.0.0", 1},
		{"a less than b minor", "v1.1.0", "v1.2.0", -1},
		{"a greater than b minor", "v1.2.0", "v1.1.0", 1},
		{"a less than b patch", "v1.0.1", "v1.0.2", -1},
		{"a greater than b patch", "v1.0.2", "v1.0.1", 1},
		{"with and without v prefix", "v1.0.0", "1.0.0", 0},
		{"different scales", "v1.10.0", "v1.9.0", 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CompareVersions(tc.a, tc.b)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNormalizeVersion_Good(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.0.0", "v1.0.0"},
		{"v1.0.0", "v1.0.0"},
		{"0.0.1", "v0.0.1"},
		{"v10.20.30", "v10.20.30"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeVersion(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
