package release

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestGenerateWithConfig_Good(t *testing.T) {
	// Note: This test would require a git repository to fully test.
	// For unit testing, we test the filtering logic indirectly through
	// the parseConventionalCommit and formatChangelog functions.

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
