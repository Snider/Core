package release

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			name:       "with prerelease",
			input:      "v1.2.3-alpha",
			major:      1, minor: 2, patch: 3,
			prerelease: "alpha",
		},
		{
			name:       "with prerelease and build",
			input:      "v1.2.3-beta.1+build.456",
			major:      1, minor: 2, patch: 3,
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
