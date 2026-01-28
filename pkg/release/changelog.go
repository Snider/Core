// Package release provides release automation with changelog generation and publishing.
package release

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ConventionalCommit represents a parsed conventional commit.
type ConventionalCommit struct {
	Type        string // feat, fix, etc.
	Scope       string // optional scope in parentheses
	Description string // commit description
	Hash        string // short commit hash
	Breaking    bool   // has breaking change indicator
}

// commitTypeLabels maps commit types to human-readable labels for the changelog.
var commitTypeLabels = map[string]string{
	"feat":     "Features",
	"fix":      "Bug Fixes",
	"perf":     "Performance Improvements",
	"refactor": "Code Refactoring",
	"docs":     "Documentation",
	"style":    "Styles",
	"test":     "Tests",
	"build":    "Build System",
	"ci":       "Continuous Integration",
	"chore":    "Chores",
	"revert":   "Reverts",
}

// commitTypeOrder defines the order of sections in the changelog.
var commitTypeOrder = []string{
	"feat",
	"fix",
	"perf",
	"refactor",
	"docs",
	"style",
	"test",
	"build",
	"ci",
	"chore",
	"revert",
}

// conventionalCommitRegex matches conventional commit format.
// Examples: "feat: add feature", "fix(scope): fix bug", "feat!: breaking change"
var conventionalCommitRegex = regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?(!)?:\s*(.+)$`)

// Generate generates a markdown changelog from git commits between two refs.
// If fromRef is empty, it uses the previous tag or initial commit.
// If toRef is empty, it uses HEAD.
func Generate(dir, fromRef, toRef string) (string, error) {
	if toRef == "" {
		toRef = "HEAD"
	}

	// If fromRef is empty, try to find previous tag
	if fromRef == "" {
		prevTag, err := getPreviousTag(dir, toRef)
		if err != nil {
			// No previous tag, use initial commit
			fromRef = ""
		} else {
			fromRef = prevTag
		}
	}

	// Get commits between refs
	commits, err := getCommits(dir, fromRef, toRef)
	if err != nil {
		return "", fmt.Errorf("changelog.Generate: failed to get commits: %w", err)
	}

	// Parse conventional commits
	var parsedCommits []ConventionalCommit
	for _, commit := range commits {
		parsed := parseConventionalCommit(commit)
		if parsed != nil {
			parsedCommits = append(parsedCommits, *parsed)
		}
	}

	// Generate markdown
	return formatChangelog(parsedCommits, toRef), nil
}

// GenerateWithConfig generates a changelog with filtering based on config.
func GenerateWithConfig(dir, fromRef, toRef string, cfg *ChangelogConfig) (string, error) {
	if toRef == "" {
		toRef = "HEAD"
	}

	// If fromRef is empty, try to find previous tag
	if fromRef == "" {
		prevTag, err := getPreviousTag(dir, toRef)
		if err != nil {
			fromRef = ""
		} else {
			fromRef = prevTag
		}
	}

	// Get commits between refs
	commits, err := getCommits(dir, fromRef, toRef)
	if err != nil {
		return "", fmt.Errorf("changelog.GenerateWithConfig: failed to get commits: %w", err)
	}

	// Build include/exclude sets
	includeSet := make(map[string]bool)
	excludeSet := make(map[string]bool)
	for _, t := range cfg.Include {
		includeSet[t] = true
	}
	for _, t := range cfg.Exclude {
		excludeSet[t] = true
	}

	// Parse and filter conventional commits
	var parsedCommits []ConventionalCommit
	for _, commit := range commits {
		parsed := parseConventionalCommit(commit)
		if parsed == nil {
			continue
		}

		// Apply filters
		if len(includeSet) > 0 && !includeSet[parsed.Type] {
			continue
		}
		if excludeSet[parsed.Type] {
			continue
		}

		parsedCommits = append(parsedCommits, *parsed)
	}

	return formatChangelog(parsedCommits, toRef), nil
}

// getPreviousTag returns the tag before the given ref.
func getPreviousTag(dir, ref string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0", ref+"^")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getCommits returns a slice of commit strings between two refs.
// Format: "hash subject"
func getCommits(dir, fromRef, toRef string) ([]string, error) {
	var args []string
	if fromRef == "" {
		// All commits up to toRef
		args = []string{"log", "--oneline", "--no-merges", toRef}
	} else {
		// Commits between refs
		args = []string{"log", "--oneline", "--no-merges", fromRef + ".." + toRef}
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			commits = append(commits, line)
		}
	}

	return commits, scanner.Err()
}

// parseConventionalCommit parses a git log --oneline output into a ConventionalCommit.
// Returns nil if the commit doesn't follow conventional commit format.
func parseConventionalCommit(commitLine string) *ConventionalCommit {
	// Split hash and subject
	parts := strings.SplitN(commitLine, " ", 2)
	if len(parts) != 2 {
		return nil
	}

	hash := parts[0]
	subject := parts[1]

	// Match conventional commit format
	matches := conventionalCommitRegex.FindStringSubmatch(subject)
	if matches == nil {
		return nil
	}

	return &ConventionalCommit{
		Type:        strings.ToLower(matches[1]),
		Scope:       matches[2],
		Breaking:    matches[3] == "!",
		Description: matches[4],
		Hash:        hash,
	}
}

// formatChangelog formats parsed commits into markdown.
func formatChangelog(commits []ConventionalCommit, version string) string {
	if len(commits) == 0 {
		return fmt.Sprintf("## %s\n\nNo notable changes.", version)
	}

	// Group commits by type
	grouped := make(map[string][]ConventionalCommit)
	var breaking []ConventionalCommit

	for _, commit := range commits {
		if commit.Breaking {
			breaking = append(breaking, commit)
		}
		grouped[commit.Type] = append(grouped[commit.Type], commit)
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("## %s\n\n", version))

	// Breaking changes first
	if len(breaking) > 0 {
		buf.WriteString("### BREAKING CHANGES\n\n")
		for _, commit := range breaking {
			buf.WriteString(formatCommitLine(commit))
		}
		buf.WriteString("\n")
	}

	// Other sections in order
	for _, commitType := range commitTypeOrder {
		commits, ok := grouped[commitType]
		if !ok || len(commits) == 0 {
			continue
		}

		label, ok := commitTypeLabels[commitType]
		if !ok {
			label = cases.Title(language.English).String(commitType)
		}

		buf.WriteString(fmt.Sprintf("### %s\n\n", label))
		for _, commit := range commits {
			buf.WriteString(formatCommitLine(commit))
		}
		buf.WriteString("\n")
	}

	// Any remaining types not in the order list
	var remainingTypes []string
	for commitType := range grouped {
		found := false
		for _, t := range commitTypeOrder {
			if t == commitType {
				found = true
				break
			}
		}
		if !found {
			remainingTypes = append(remainingTypes, commitType)
		}
	}
	sort.Strings(remainingTypes)

	for _, commitType := range remainingTypes {
		commits := grouped[commitType]
		label := cases.Title(language.English).String(commitType)
		buf.WriteString(fmt.Sprintf("### %s\n\n", label))
		for _, commit := range commits {
			buf.WriteString(formatCommitLine(commit))
		}
		buf.WriteString("\n")
	}

	return strings.TrimSuffix(buf.String(), "\n")
}

// formatCommitLine formats a single commit as a changelog line.
func formatCommitLine(commit ConventionalCommit) string {
	var buf strings.Builder
	buf.WriteString("- ")

	if commit.Scope != "" {
		buf.WriteString(fmt.Sprintf("**%s**: ", commit.Scope))
	}

	buf.WriteString(commit.Description)
	buf.WriteString(fmt.Sprintf(" (%s)\n", commit.Hash))

	return buf.String()
}

// ParseCommitType extracts the type from a conventional commit subject.
// Returns empty string if not a conventional commit.
func ParseCommitType(subject string) string {
	matches := conventionalCommitRegex.FindStringSubmatch(subject)
	if matches == nil {
		return ""
	}
	return strings.ToLower(matches[1])
}
