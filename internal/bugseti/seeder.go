// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SeederService prepares context for issues using the seed-agent-developer skill.
type SeederService struct {
	config *ConfigService
}

// NewSeederService creates a new SeederService.
func NewSeederService(config *ConfigService) *SeederService {
	return &SeederService{
		config: config,
	}
}

// ServiceName returns the service name for Wails.
func (s *SeederService) ServiceName() string {
	return "SeederService"
}

// SeedIssue prepares context for an issue by calling the seed-agent-developer skill.
func (s *SeederService) SeedIssue(issue *Issue) (*IssueContext, error) {
	if issue == nil {
		return nil, fmt.Errorf("issue is nil")
	}

	// Create a temporary workspace for the issue
	workDir, err := s.prepareWorkspace(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare workspace: %w", err)
	}

	// Try to use the seed-agent-developer skill via plugin system
	ctx, err := s.runSeedSkill(issue, workDir)
	if err != nil {
		log.Printf("Seed skill failed, using fallback: %v", err)
		// Fallback to basic context preparation
		ctx = s.prepareBasicContext(issue)
	}

	ctx.PreparedAt = time.Now()
	return ctx, nil
}

// prepareWorkspace creates a temporary workspace and clones the repo.
func (s *SeederService) prepareWorkspace(issue *Issue) (string, error) {
	// Create workspace directory
	baseDir := s.config.GetWorkspaceDir()
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "bugseti")
	}

	// Create issue-specific directory
	workDir := filepath.Join(baseDir, sanitizeRepoName(issue.Repo), fmt.Sprintf("issue-%d", issue.Number))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace: %w", err)
	}

	// Check if repo already cloned
	if _, err := os.Stat(filepath.Join(workDir, ".git")); os.IsNotExist(err) {
		// Clone the repository
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gh", "repo", "clone", issue.Repo, workDir, "--", "--depth=1")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to clone repo: %s: %w", stderr.String(), err)
		}
	}

	return workDir, nil
}

// runSeedSkill executes the seed-agent-developer skill to prepare context.
func (s *SeederService) runSeedSkill(issue *Issue, workDir string) (*IssueContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Look for the plugin script
	pluginPaths := []string{
		"/home/shared/hostuk/claude-plugins/agentic-flows/skills/seed-agent-developer/scripts/analyze-issue.sh",
		filepath.Join(os.Getenv("HOME"), ".claude/plugins/agentic-flows/skills/seed-agent-developer/scripts/analyze-issue.sh"),
	}

	var scriptPath string
	for _, p := range pluginPaths {
		if _, err := os.Stat(p); err == nil {
			scriptPath = p
			break
		}
	}

	if scriptPath == "" {
		return nil, fmt.Errorf("seed-agent-developer skill not found")
	}

	// Run the analyze-issue script
	cmd := exec.CommandContext(ctx, "bash", scriptPath)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ISSUE_NUMBER=%d", issue.Number),
		fmt.Sprintf("ISSUE_REPO=%s", issue.Repo),
		fmt.Sprintf("ISSUE_TITLE=%s", issue.Title),
		fmt.Sprintf("ISSUE_URL=%s", issue.URL),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("seed skill failed: %s: %w", stderr.String(), err)
	}

	// Parse the output as JSON
	var result struct {
		Summary       string   `json:"summary"`
		RelevantFiles []string `json:"relevant_files"`
		SuggestedFix  string   `json:"suggested_fix"`
		RelatedIssues []string `json:"related_issues"`
		Complexity    string   `json:"complexity"`
		EstimatedTime string   `json:"estimated_time"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		// If not JSON, treat as plain text summary
		return &IssueContext{
			Summary:    stdout.String(),
			Complexity: "unknown",
		}, nil
	}

	return &IssueContext{
		Summary:       result.Summary,
		RelevantFiles: result.RelevantFiles,
		SuggestedFix:  result.SuggestedFix,
		RelatedIssues: result.RelatedIssues,
		Complexity:    result.Complexity,
		EstimatedTime: result.EstimatedTime,
	}, nil
}

// prepareBasicContext creates a basic context without the seed skill.
func (s *SeederService) prepareBasicContext(issue *Issue) *IssueContext {
	// Extract potential file references from issue body
	files := extractFileReferences(issue.Body)

	// Estimate complexity based on labels and body length
	complexity := estimateComplexity(issue)

	return &IssueContext{
		Summary:       fmt.Sprintf("Issue #%d in %s: %s", issue.Number, issue.Repo, issue.Title),
		RelevantFiles: files,
		Complexity:    complexity,
		EstimatedTime: estimateTime(complexity),
	}
}

// sanitizeRepoName converts owner/repo to a safe directory name.
func sanitizeRepoName(repo string) string {
	return strings.ReplaceAll(repo, "/", "-")
}

// extractFileReferences finds file paths mentioned in text.
func extractFileReferences(text string) []string {
	var files []string
	seen := make(map[string]bool)

	// Common file patterns
	patterns := []string{
		`.go`, `.js`, `.ts`, `.py`, `.rs`, `.java`, `.cpp`, `.c`, `.h`,
		`.json`, `.yaml`, `.yml`, `.toml`, `.xml`, `.md`,
	}

	words := strings.Fields(text)
	for _, word := range words {
		// Clean up the word
		word = strings.Trim(word, "`,\"'()[]{}:")

		// Check if it looks like a file path
		for _, ext := range patterns {
			if strings.HasSuffix(word, ext) && !seen[word] {
				files = append(files, word)
				seen[word] = true
				break
			}
		}
	}

	return files
}

// estimateComplexity guesses issue complexity from content.
func estimateComplexity(issue *Issue) string {
	bodyLen := len(issue.Body)
	labelScore := 0

	for _, label := range issue.Labels {
		lower := strings.ToLower(label)
		switch {
		case strings.Contains(lower, "good first issue"), strings.Contains(lower, "beginner"):
			labelScore -= 2
		case strings.Contains(lower, "easy"):
			labelScore -= 1
		case strings.Contains(lower, "complex"), strings.Contains(lower, "hard"):
			labelScore += 2
		case strings.Contains(lower, "refactor"):
			labelScore += 1
		}
	}

	// Combine body length and label score
	score := labelScore
	if bodyLen > 2000 {
		score += 2
	} else if bodyLen > 500 {
		score += 1
	}

	switch {
	case score <= -1:
		return "easy"
	case score <= 1:
		return "medium"
	default:
		return "hard"
	}
}

// estimateTime suggests time based on complexity.
func estimateTime(complexity string) string {
	switch complexity {
	case "easy":
		return "15-30 minutes"
	case "medium":
		return "1-2 hours"
	case "hard":
		return "2-4 hours"
	default:
		return "unknown"
	}
}

// GetWorkspaceDir returns the workspace directory for an issue.
func (s *SeederService) GetWorkspaceDir(issue *Issue) string {
	baseDir := s.config.GetWorkspaceDir()
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "bugseti")
	}
	return filepath.Join(baseDir, sanitizeRepoName(issue.Repo), fmt.Sprintf("issue-%d", issue.Number))
}

// CleanupWorkspace removes the workspace for an issue.
func (s *SeederService) CleanupWorkspace(issue *Issue) error {
	workDir := s.GetWorkspaceDir(issue)
	return os.RemoveAll(workDir)
}
