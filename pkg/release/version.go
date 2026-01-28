// Package release provides release automation with changelog generation and publishing.
package release

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// semverRegex matches semantic version strings with or without 'v' prefix.
var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9.-]+))?(?:\+([a-zA-Z0-9.-]+))?$`)

// DetermineVersion determines the version for a release.
// It checks in order:
//  1. Git tag on HEAD
//  2. Most recent tag + increment patch
//  3. Default to v0.0.1 if no tags exist
func DetermineVersion(dir string) (string, error) {
	// Check if HEAD has a tag
	headTag, err := getTagOnHead(dir)
	if err == nil && headTag != "" {
		return normalizeVersion(headTag), nil
	}

	// Get most recent tag
	latestTag, err := getLatestTag(dir)
	if err != nil || latestTag == "" {
		// No tags exist, return default
		return "v0.0.1", nil
	}

	// Increment patch version
	return IncrementVersion(latestTag), nil
}

// IncrementVersion increments the patch version of a semver string.
// Examples:
//   - "v1.2.3" -> "v1.2.4"
//   - "1.2.3" -> "v1.2.4"
//   - "v1.2.3-alpha" -> "v1.2.4" (strips prerelease)
func IncrementVersion(current string) string {
	matches := semverRegex.FindStringSubmatch(current)
	if matches == nil {
		// Not a valid semver, return as-is with increment suffix
		return current + ".1"
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	// Increment patch
	patch++

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

// IncrementMinor increments the minor version of a semver string.
// Examples:
//   - "v1.2.3" -> "v1.3.0"
//   - "1.2.3" -> "v1.3.0"
func IncrementMinor(current string) string {
	matches := semverRegex.FindStringSubmatch(current)
	if matches == nil {
		return current + ".1"
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])

	// Increment minor, reset patch
	minor++

	return fmt.Sprintf("v%d.%d.0", major, minor)
}

// IncrementMajor increments the major version of a semver string.
// Examples:
//   - "v1.2.3" -> "v2.0.0"
//   - "1.2.3" -> "v2.0.0"
func IncrementMajor(current string) string {
	matches := semverRegex.FindStringSubmatch(current)
	if matches == nil {
		return current + ".1"
	}

	major, _ := strconv.Atoi(matches[1])

	// Increment major, reset minor and patch
	major++

	return fmt.Sprintf("v%d.0.0", major)
}

// ParseVersion parses a semver string into its components.
// Returns (major, minor, patch, prerelease, build, error).
func ParseVersion(version string) (int, int, int, string, string, error) {
	matches := semverRegex.FindStringSubmatch(version)
	if matches == nil {
		return 0, 0, 0, "", "", fmt.Errorf("invalid semver: %s", version)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	prerelease := matches[4]
	build := matches[5]

	return major, minor, patch, prerelease, build, nil
}

// ValidateVersion checks if a string is a valid semver.
func ValidateVersion(version string) bool {
	return semverRegex.MatchString(version)
}

// normalizeVersion ensures the version starts with 'v'.
func normalizeVersion(version string) string {
	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

// getTagOnHead returns the tag on HEAD, if any.
func getTagOnHead(dir string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--exact-match", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getLatestTag returns the most recent tag in the repository.
func getLatestTag(dir string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// CompareVersions compares two semver strings.
// Returns:
//
//	-1 if a < b
//	 0 if a == b
//	 1 if a > b
func CompareVersions(a, b string) int {
	aMajor, aMinor, aPatch, _, _, errA := ParseVersion(a)
	bMajor, bMinor, bPatch, _, _, errB := ParseVersion(b)

	// Invalid versions are considered less than valid ones
	if errA != nil && errB != nil {
		return strings.Compare(a, b)
	}
	if errA != nil {
		return -1
	}
	if errB != nil {
		return 1
	}

	// Compare major
	if aMajor != bMajor {
		if aMajor < bMajor {
			return -1
		}
		return 1
	}

	// Compare minor
	if aMinor != bMinor {
		if aMinor < bMinor {
			return -1
		}
		return 1
	}

	// Compare patch
	if aPatch != bPatch {
		if aPatch < bPatch {
			return -1
		}
		return 1
	}

	return 0
}
