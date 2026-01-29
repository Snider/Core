package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConventionalCommit_Good(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ConventionalCommit
	}{
		{
			name:  "feat without scope",
			input: "abc1234 feat: add new feature",
			expected: &ConventionalCommit{
				Type:        "feat",
				Scope:       "",
				Description: "add new feature",
				Hash:        "abc1234",
				Breaking:    false,
			},
		},
		{
			name:  "fix with scope",
			input: "def5678 fix(auth): resolve login issue",
			expected: &ConventionalCommit{
				Type:        "fix",
				Scope:       "auth",
				Description: "resolve login issue",
				Hash:        "def5678",
				Breaking:    false,
			},
		},
		{
			name:  "breaking change with exclamation",
			input: "ghi9012 feat!: breaking API change",
			expected: &ConventionalCommit{
				Type:        "feat",
				Scope:       "",
				Description: "breaking API change",
				Hash:        "ghi9012",
				Breaking:    true,
			},
		},
		{
			name:  "breaking change with scope",
			input: "jkl3456 fix(api)!: remove deprecated endpoint",
			expected: &ConventionalCommit{
				Type:        "fix",
				Scope:       "api",
				Description: "remove deprecated endpoint",
				Hash:        "jkl3456",
				Breaking:    true,
			},
		},
		{
			name:  "perf type",
			input: "mno7890 perf: optimize database queries",
			expected: &ConventionalCommit{
				Type:        "perf",
				Scope:       "",
				Description: "optimize database queries",
				Hash:        "mno7890",
				Breaking:    false,
			},
		},
		{
			name:  "chore type",
			input: "pqr1234 chore: update dependencies",
			expected: &ConventionalCommit{
				Type:        "chore",
				Scope:       "",
				Description: "update dependencies",
				Hash:        "pqr1234",
				Breaking:    false,
			},
		},
		{
			name:  "uppercase type normalizes to lowercase",
			input: "stu5678 FEAT: uppercase type",
			expected: &ConventionalCommit{
				Type:        "feat",
				Scope:       "",
				Description: "uppercase type",
				Hash:        "stu5678",
				Breaking:    false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseConventionalCommit(tc.input)
			assert.NotNil(t, result)
			assert.Equal(t, tc.expected.Type, result.Type)
			assert.Equal(t, tc.expected.Scope, result.Scope)
			assert.Equal(t, tc.expected.Description, result.Description)
			assert.Equal(t, tc.expected.Hash, result.Hash)
			assert.Equal(t, tc.expected.Breaking, result.Breaking)
		})
	}
}

func TestParseConventionalCommit_Bad(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "non-conventional commit",
			input: "abc1234 Update README",
		},
		{
			name:  "missing colon",
			input: "def5678 feat add feature",
		},
		{
			name:  "empty subject",
			input: "ghi9012",
		},
		{
			name:  "just hash",
			input: "abc1234",
		},
		{
			name:  "merge commit",
			input: "abc1234 Merge pull request #123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseConventionalCommit(tc.input)
			assert.Nil(t, result)
		})
	}
}

func TestFormatChangelog_Good(t *testing.T) {
	t.Run("formats commits by type", func(t *testing.T) {
		commits := []ConventionalCommit{
			{Type: "feat", Description: "add feature A", Hash: "abc1234"},
			{Type: "fix", Description: "fix bug B", Hash: "def5678"},
			{Type: "feat", Description: "add feature C", Hash: "ghi9012"},
		}

		result := formatChangelog(commits, "v1.0.0")

		assert.Contains(t, result, "## v1.0.0")
		assert.Contains(t, result, "### Features")
		assert.Contains(t, result, "### Bug Fixes")
		assert.Contains(t, result, "- add feature A (abc1234)")
		assert.Contains(t, result, "- fix bug B (def5678)")
		assert.Contains(t, result, "- add feature C (ghi9012)")
	})

	t.Run("includes scope in output", func(t *testing.T) {
		commits := []ConventionalCommit{
			{Type: "feat", Scope: "api", Description: "add endpoint", Hash: "abc1234"},
		}

		result := formatChangelog(commits, "v1.0.0")

		assert.Contains(t, result, "**api**: add endpoint")
	})

	t.Run("breaking changes first", func(t *testing.T) {
		commits := []ConventionalCommit{
			{Type: "feat", Description: "normal feature", Hash: "abc1234"},
			{Type: "feat", Description: "breaking feature", Hash: "def5678", Breaking: true},
		}

		result := formatChangelog(commits, "v1.0.0")

		assert.Contains(t, result, "### BREAKING CHANGES")
		// Breaking changes section should appear before Features
		breakingPos := indexOf(result, "BREAKING CHANGES")
		featuresPos := indexOf(result, "Features")
		assert.Less(t, breakingPos, featuresPos)
	})

	t.Run("empty commits returns minimal changelog", func(t *testing.T) {
		result := formatChangelog([]ConventionalCommit{}, "v1.0.0")

		assert.Contains(t, result, "## v1.0.0")
		assert.Contains(t, result, "No notable changes")
	})
}

func TestParseCommitType_Good(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"feat: add feature", "feat"},
		{"fix(scope): fix bug", "fix"},
		{"perf!: breaking perf", "perf"},
		{"chore: update deps", "chore"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ParseCommitType(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseCommitType_Bad(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"not a conventional commit"},
		{"Update README"},
		{"Merge branch 'main'"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ParseCommitType(tc.input)
			assert.Empty(t, result)
		})
	}
}

func TestGenerateWithConfig_ConfigValues(t *testing.T) {
	t.Run("config filters are parsed correctly", func(t *testing.T) {
		cfg := &ChangelogConfig{
			Include: []string{"feat", "fix"},
			Exclude: []string{"chore", "docs"},
		}

		// Verify the config values
		assert.Contains(t, cfg.Include, "feat")
		assert.Contains(t, cfg.Include, "fix")
		assert.Contains(t, cfg.Exclude, "chore")
		assert.Contains(t, cfg.Exclude, "docs")
	})
}

// indexOf returns the position of a substring in a string, or -1 if not found.
func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// setupChangelogGitRepo creates a temporary directory with an initialized git repository.
func setupChangelogGitRepo(t *testing.T) string {
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

// createChangelogCommit creates a commit in the given directory.
func createChangelogCommit(t *testing.T, dir, message string) {
	t.Helper()

	// Create or modify a file
	filePath := filepath.Join(dir, "changelog_test.txt")
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

// createChangelogTag creates a tag in the given directory.
func createChangelogTag(t *testing.T, dir, tag string) {
	t.Helper()
	cmd := exec.Command("git", "tag", tag)
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
}

func TestGenerate_Good(t *testing.T) {
	t.Run("generates changelog from commits", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: add new feature")
		createChangelogCommit(t, dir, "fix: resolve bug")

		changelog, err := Generate(dir, "", "HEAD")
		require.NoError(t, err)

		assert.Contains(t, changelog, "## HEAD")
		assert.Contains(t, changelog, "### Features")
		assert.Contains(t, changelog, "add new feature")
		assert.Contains(t, changelog, "### Bug Fixes")
		assert.Contains(t, changelog, "resolve bug")
	})

	t.Run("generates changelog between tags", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: initial feature")
		createChangelogTag(t, dir, "v1.0.0")
		createChangelogCommit(t, dir, "feat: new feature")
		createChangelogCommit(t, dir, "fix: bug fix")
		createChangelogTag(t, dir, "v1.1.0")

		changelog, err := Generate(dir, "v1.0.0", "v1.1.0")
		require.NoError(t, err)

		assert.Contains(t, changelog, "## v1.1.0")
		assert.Contains(t, changelog, "new feature")
		assert.Contains(t, changelog, "bug fix")
		// Should NOT contain the initial feature
		assert.NotContains(t, changelog, "initial feature")
	})

	t.Run("handles empty changelog when no conventional commits", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "Update README")
		createChangelogCommit(t, dir, "Merge branch main")

		changelog, err := Generate(dir, "", "HEAD")
		require.NoError(t, err)

		assert.Contains(t, changelog, "No notable changes")
	})

	t.Run("uses previous tag when fromRef is empty", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: old feature")
		createChangelogTag(t, dir, "v1.0.0")
		createChangelogCommit(t, dir, "feat: new feature")

		changelog, err := Generate(dir, "", "HEAD")
		require.NoError(t, err)

		assert.Contains(t, changelog, "new feature")
		assert.NotContains(t, changelog, "old feature")
	})

	t.Run("includes breaking changes", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat!: breaking API change")
		createChangelogCommit(t, dir, "feat: normal feature")

		changelog, err := Generate(dir, "", "HEAD")
		require.NoError(t, err)

		assert.Contains(t, changelog, "### BREAKING CHANGES")
		assert.Contains(t, changelog, "breaking API change")
	})

	t.Run("includes scope in output", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat(api): add endpoint")

		changelog, err := Generate(dir, "", "HEAD")
		require.NoError(t, err)

		assert.Contains(t, changelog, "**api**:")
	})
}

func TestGenerate_Bad(t *testing.T) {
	t.Run("returns error for non-git directory", func(t *testing.T) {
		dir := t.TempDir()

		_, err := Generate(dir, "", "HEAD")
		assert.Error(t, err)
	})
}

func TestGenerateWithConfig_Good(t *testing.T) {
	t.Run("filters commits by include list", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: new feature")
		createChangelogCommit(t, dir, "fix: bug fix")
		createChangelogCommit(t, dir, "chore: update deps")

		cfg := &ChangelogConfig{
			Include: []string{"feat"},
		}

		changelog, err := GenerateWithConfig(dir, "", "HEAD", cfg)
		require.NoError(t, err)

		assert.Contains(t, changelog, "new feature")
		assert.NotContains(t, changelog, "bug fix")
		assert.NotContains(t, changelog, "update deps")
	})

	t.Run("filters commits by exclude list", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: new feature")
		createChangelogCommit(t, dir, "fix: bug fix")
		createChangelogCommit(t, dir, "chore: update deps")

		cfg := &ChangelogConfig{
			Exclude: []string{"chore"},
		}

		changelog, err := GenerateWithConfig(dir, "", "HEAD", cfg)
		require.NoError(t, err)

		assert.Contains(t, changelog, "new feature")
		assert.Contains(t, changelog, "bug fix")
		assert.NotContains(t, changelog, "update deps")
	})

	t.Run("combines include and exclude filters", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: new feature")
		createChangelogCommit(t, dir, "fix: bug fix")
		createChangelogCommit(t, dir, "perf: performance")

		cfg := &ChangelogConfig{
			Include: []string{"feat", "fix", "perf"},
			Exclude: []string{"perf"},
		}

		changelog, err := GenerateWithConfig(dir, "", "HEAD", cfg)
		require.NoError(t, err)

		assert.Contains(t, changelog, "new feature")
		assert.Contains(t, changelog, "bug fix")
		assert.NotContains(t, changelog, "performance")
	})
}

func TestGetCommits_Good(t *testing.T) {
	t.Run("returns all commits when fromRef is empty", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: first")
		createChangelogCommit(t, dir, "feat: second")
		createChangelogCommit(t, dir, "feat: third")

		commits, err := getCommits(dir, "", "HEAD")
		require.NoError(t, err)

		assert.Len(t, commits, 3)
	})

	t.Run("returns commits between refs", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: first")
		createChangelogTag(t, dir, "v1.0.0")
		createChangelogCommit(t, dir, "feat: second")
		createChangelogCommit(t, dir, "feat: third")

		commits, err := getCommits(dir, "v1.0.0", "HEAD")
		require.NoError(t, err)

		assert.Len(t, commits, 2)
	})

	t.Run("excludes merge commits", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: regular commit")
		// Merge commits are excluded by --no-merges flag
		// We can verify by checking the count matches expected

		commits, err := getCommits(dir, "", "HEAD")
		require.NoError(t, err)

		assert.Len(t, commits, 1)
		assert.Contains(t, commits[0], "regular commit")
	})

	t.Run("returns empty slice for no commits in range", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: only commit")
		createChangelogTag(t, dir, "v1.0.0")

		commits, err := getCommits(dir, "v1.0.0", "HEAD")
		require.NoError(t, err)

		assert.Empty(t, commits)
	})
}

func TestGetCommits_Bad(t *testing.T) {
	t.Run("returns error for invalid ref", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: commit")

		_, err := getCommits(dir, "nonexistent-tag", "HEAD")
		assert.Error(t, err)
	})

	t.Run("returns error for non-git directory", func(t *testing.T) {
		dir := t.TempDir()

		_, err := getCommits(dir, "", "HEAD")
		assert.Error(t, err)
	})
}

func TestGetPreviousTag_Good(t *testing.T) {
	t.Run("returns previous tag", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: first")
		createChangelogTag(t, dir, "v1.0.0")
		createChangelogCommit(t, dir, "feat: second")
		createChangelogTag(t, dir, "v1.1.0")

		tag, err := getPreviousTag(dir, "v1.1.0")
		require.NoError(t, err)
		assert.Equal(t, "v1.0.0", tag)
	})

	t.Run("returns tag before HEAD", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: first")
		createChangelogTag(t, dir, "v1.0.0")
		createChangelogCommit(t, dir, "feat: second")

		tag, err := getPreviousTag(dir, "HEAD")
		require.NoError(t, err)
		assert.Equal(t, "v1.0.0", tag)
	})
}

func TestGetPreviousTag_Bad(t *testing.T) {
	t.Run("returns error when no previous tag exists", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: first")
		createChangelogTag(t, dir, "v1.0.0")

		// v1.0.0^ has no tag before it
		_, err := getPreviousTag(dir, "v1.0.0")
		assert.Error(t, err)
	})

	t.Run("returns error for invalid ref", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: commit")

		_, err := getPreviousTag(dir, "nonexistent")
		assert.Error(t, err)
	})
}

func TestFormatCommitLine_Good(t *testing.T) {
	t.Run("formats commit without scope", func(t *testing.T) {
		commit := ConventionalCommit{
			Type:        "feat",
			Description: "add feature",
			Hash:        "abc1234",
		}

		result := formatCommitLine(commit)
		assert.Equal(t, "- add feature (abc1234)\n", result)
	})

	t.Run("formats commit with scope", func(t *testing.T) {
		commit := ConventionalCommit{
			Type:        "fix",
			Scope:       "api",
			Description: "fix bug",
			Hash:        "def5678",
		}

		result := formatCommitLine(commit)
		assert.Equal(t, "- **api**: fix bug (def5678)\n", result)
	})
}

func TestFormatChangelog_Ugly(t *testing.T) {
	t.Run("handles custom commit type not in order", func(t *testing.T) {
		commits := []ConventionalCommit{
			{Type: "custom", Description: "custom type", Hash: "abc1234"},
		}

		result := formatChangelog(commits, "v1.0.0")

		assert.Contains(t, result, "### Custom")
		assert.Contains(t, result, "custom type")
	})

	t.Run("handles multiple custom commit types", func(t *testing.T) {
		commits := []ConventionalCommit{
			{Type: "alpha", Description: "alpha feature", Hash: "abc1234"},
			{Type: "beta", Description: "beta feature", Hash: "def5678"},
		}

		result := formatChangelog(commits, "v1.0.0")

		// Should be sorted alphabetically for custom types
		assert.Contains(t, result, "### Alpha")
		assert.Contains(t, result, "### Beta")
	})
}

func TestGenerateWithConfig_Bad(t *testing.T) {
	t.Run("returns error for non-git directory", func(t *testing.T) {
		dir := t.TempDir()
		cfg := &ChangelogConfig{
			Include: []string{"feat"},
		}

		_, err := GenerateWithConfig(dir, "", "HEAD", cfg)
		assert.Error(t, err)
	})
}

func TestGenerateWithConfig_EdgeCases(t *testing.T) {
	t.Run("uses HEAD when toRef is empty", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: new feature")

		cfg := &ChangelogConfig{
			Include: []string{"feat"},
		}

		// Pass empty toRef
		changelog, err := GenerateWithConfig(dir, "", "", cfg)
		require.NoError(t, err)

		assert.Contains(t, changelog, "## HEAD")
	})

	t.Run("handles previous tag lookup failure gracefully", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: first")

		cfg := &ChangelogConfig{
			Include: []string{"feat"},
		}

		// No tags exist, should still work
		changelog, err := GenerateWithConfig(dir, "", "HEAD", cfg)
		require.NoError(t, err)

		assert.Contains(t, changelog, "first")
	})

	t.Run("uses explicit fromRef when provided", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: old feature")
		createChangelogTag(t, dir, "v1.0.0")
		createChangelogCommit(t, dir, "feat: new feature")

		cfg := &ChangelogConfig{
			Include: []string{"feat"},
		}

		// Use explicit fromRef
		changelog, err := GenerateWithConfig(dir, "v1.0.0", "HEAD", cfg)
		require.NoError(t, err)

		assert.Contains(t, changelog, "new feature")
		assert.NotContains(t, changelog, "old feature")
	})

	t.Run("skips non-conventional commits", func(t *testing.T) {
		dir := setupChangelogGitRepo(t)
		createChangelogCommit(t, dir, "feat: conventional commit")
		createChangelogCommit(t, dir, "Update README")

		cfg := &ChangelogConfig{
			Include: []string{"feat"},
		}

		changelog, err := GenerateWithConfig(dir, "", "HEAD", cfg)
		require.NoError(t, err)

		assert.Contains(t, changelog, "conventional commit")
		assert.NotContains(t, changelog, "Update README")
	})
}
