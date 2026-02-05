// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SubmitService handles the PR submission flow.
type SubmitService struct {
	config *ConfigService
	notify *NotifyService
	stats  *StatsService
}

// NewSubmitService creates a new SubmitService.
func NewSubmitService(config *ConfigService, notify *NotifyService, stats *StatsService) *SubmitService {
	return &SubmitService{
		config: config,
		notify: notify,
		stats:  stats,
	}
}

// ServiceName returns the service name for Wails.
func (s *SubmitService) ServiceName() string {
	return "SubmitService"
}

// PRSubmission contains the data for a pull request submission.
type PRSubmission struct {
	Issue     *Issue   `json:"issue"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Branch    string   `json:"branch"`
	CommitMsg string   `json:"commitMsg"`
	Files     []string `json:"files"`
	WorkDir   string   `json:"workDir"`
}

// PRResult contains the result of a PR submission.
type PRResult struct {
	Success   bool   `json:"success"`
	PRURL     string `json:"prUrl,omitempty"`
	PRNumber  int    `json:"prNumber,omitempty"`
	Error     string `json:"error,omitempty"`
	ForkOwner string `json:"forkOwner,omitempty"`
}

// Submit creates a pull request for the given issue.
// Flow: Fork -> Branch -> Commit -> PR
func (s *SubmitService) Submit(submission *PRSubmission) (*PRResult, error) {
	if submission == nil || submission.Issue == nil {
		return nil, fmt.Errorf("invalid submission")
	}

	issue := submission.Issue
	workDir := submission.WorkDir
	if workDir == "" {
		return nil, fmt.Errorf("work directory not specified")
	}

	// Step 1: Ensure we have a fork
	forkOwner, err := s.ensureFork(issue.Repo)
	if err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("fork failed: %v", err)}, err
	}

	// Step 2: Create branch
	branch := submission.Branch
	if branch == "" {
		branch = fmt.Sprintf("bugseti/issue-%d", issue.Number)
	}
	if err := s.createBranch(workDir, branch); err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("branch creation failed: %v", err)}, err
	}

	// Step 3: Stage and commit changes
	commitMsg := submission.CommitMsg
	if commitMsg == "" {
		commitMsg = fmt.Sprintf("fix: resolve issue #%d\n\n%s\n\nFixes #%d", issue.Number, issue.Title, issue.Number)
	}
	if err := s.commitChanges(workDir, submission.Files, commitMsg); err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("commit failed: %v", err)}, err
	}

	// Step 4: Push to fork
	if err := s.pushToFork(workDir, forkOwner, branch); err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("push failed: %v", err)}, err
	}

	// Step 5: Create PR
	prTitle := submission.Title
	if prTitle == "" {
		prTitle = fmt.Sprintf("Fix #%d: %s", issue.Number, issue.Title)
	}
	prBody := submission.Body
	if prBody == "" {
		prBody = s.generatePRBody(issue)
	}

	prURL, prNumber, err := s.createPR(issue.Repo, forkOwner, branch, prTitle, prBody)
	if err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("PR creation failed: %v", err)}, err
	}

	// Update stats
	s.stats.RecordPRSubmitted(issue.Repo)

	// Notify user
	s.notify.Notify("BugSETI", fmt.Sprintf("PR #%d submitted for issue #%d", prNumber, issue.Number))

	return &PRResult{
		Success:   true,
		PRURL:     prURL,
		PRNumber:  prNumber,
		ForkOwner: forkOwner,
	}, nil
}

// ensureFork ensures a fork exists for the repo.
func (s *SubmitService) ensureFork(repo string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Check if fork exists
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repo format: %s", repo)
	}

	// Get current user
	cmd := exec.CommandContext(ctx, "gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	username := strings.TrimSpace(string(output))

	// Check if fork exists
	forkRepo := fmt.Sprintf("%s/%s", username, parts[1])
	cmd = exec.CommandContext(ctx, "gh", "repo", "view", forkRepo, "--json", "name")
	if err := cmd.Run(); err != nil {
		// Fork doesn't exist, create it
		log.Printf("Creating fork of %s...", repo)
		cmd = exec.CommandContext(ctx, "gh", "repo", "fork", repo, "--clone=false")
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to create fork: %w", err)
		}
		// Wait a bit for GitHub to process
		time.Sleep(2 * time.Second)
	}

	return username, nil
}

// createBranch creates a new branch in the repository.
func (s *SubmitService) createBranch(workDir, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch latest from upstream
	cmd := exec.CommandContext(ctx, "git", "fetch", "origin")
	cmd.Dir = workDir
	cmd.Run() // Ignore errors

	// Create and checkout new branch
	cmd = exec.CommandContext(ctx, "git", "checkout", "-b", branch)
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// Branch might already exist, try to checkout
		cmd = exec.CommandContext(ctx, "git", "checkout", branch)
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create/checkout branch: %s: %w", stderr.String(), err)
		}
	}

	return nil
}

// commitChanges stages and commits the specified files.
func (s *SubmitService) commitChanges(workDir string, files []string, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stage files
	if len(files) == 0 {
		// Stage all changes
		cmd := exec.CommandContext(ctx, "git", "add", "-A")
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stage changes: %w", err)
		}
	} else {
		// Stage specific files
		args := append([]string{"add"}, files...)
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}
	}

	// Check if there are changes to commit
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	cmd.Dir = workDir
	if err := cmd.Run(); err == nil {
		return fmt.Errorf("no changes to commit")
	}

	// Commit
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %s: %w", stderr.String(), err)
	}

	return nil
}

// pushToFork pushes the branch to the user's fork.
func (s *SubmitService) pushToFork(workDir, forkOwner, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Add fork as remote if not exists
	forkRemote := "fork"
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", forkRemote)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		// Get the origin URL and construct fork URL
		cmd = exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
		cmd.Dir = workDir
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get origin URL: %w", err)
		}

		originURL := strings.TrimSpace(string(output))
		// Replace original owner with fork owner
		var forkURL string
		if strings.HasPrefix(originURL, "https://") {
			// https://github.com/owner/repo.git
			parts := strings.Split(originURL, "/")
			if len(parts) >= 4 {
				parts[len(parts)-2] = forkOwner
				forkURL = strings.Join(parts, "/")
			}
		} else {
			// git@github.com:owner/repo.git
			forkURL = strings.Replace(originURL, ":", fmt.Sprintf(":%s/", forkOwner), 1)
			forkURL = strings.Replace(forkURL, strings.Split(forkURL, "/")[0]+"/", "", 1)
			forkURL = fmt.Sprintf("git@github.com:%s/%s", forkOwner, filepath.Base(originURL))
		}

		cmd = exec.CommandContext(ctx, "git", "remote", "add", forkRemote, forkURL)
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add fork remote: %w", err)
		}
	}

	// Push to fork
	cmd = exec.CommandContext(ctx, "git", "push", "-u", forkRemote, branch)
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push: %s: %w", stderr.String(), err)
	}

	return nil
}

// createPR creates a pull request using GitHub CLI.
func (s *SubmitService) createPR(repo, forkOwner, branch, title, body string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create PR
	cmd := exec.CommandContext(ctx, "gh", "pr", "create",
		"--repo", repo,
		"--head", fmt.Sprintf("%s:%s", forkOwner, branch),
		"--title", title,
		"--body", body,
		"--json", "url,number")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", 0, fmt.Errorf("failed to create PR: %s: %w", stderr.String(), err)
	}

	var result struct {
		URL    string `json:"url"`
		Number int    `json:"number"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", 0, fmt.Errorf("failed to parse PR response: %w", err)
	}

	return result.URL, result.Number, nil
}

// generatePRBody creates a default PR body for an issue.
func (s *SubmitService) generatePRBody(issue *Issue) string {
	var body strings.Builder

	body.WriteString("## Summary\n\n")
	body.WriteString(fmt.Sprintf("This PR addresses issue #%d.\n\n", issue.Number))

	if issue.Context != nil && issue.Context.Summary != "" {
		body.WriteString("## Context\n\n")
		body.WriteString(issue.Context.Summary)
		body.WriteString("\n\n")
	}

	body.WriteString("## Changes\n\n")
	body.WriteString("<!-- Describe your changes here -->\n\n")

	body.WriteString("## Testing\n\n")
	body.WriteString("<!-- Describe how you tested your changes -->\n\n")

	body.WriteString("---\n\n")
	body.WriteString("*Submitted via [BugSETI](https://github.com/host-uk/core) - Distributed Bug Fixing*\n")

	return body.String()
}

// GetPRStatus checks the status of a submitted PR.
func (s *SubmitService) GetPRStatus(repo string, prNumber int) (*PRStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", "pr", "view",
		"--repo", repo,
		fmt.Sprintf("%d", prNumber),
		"--json", "state,mergeable,reviews,statusCheckRollup")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get PR status: %w", err)
	}

	var result struct {
		State             string `json:"state"`
		Mergeable         string `json:"mergeable"`
		StatusCheckRollup []struct {
			State string `json:"state"`
		} `json:"statusCheckRollup"`
		Reviews []struct {
			State string `json:"state"`
		} `json:"reviews"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PR status: %w", err)
	}

	status := &PRStatus{
		State:     result.State,
		Mergeable: result.Mergeable == "MERGEABLE",
	}

	// Check CI status
	status.CIPassing = true
	for _, check := range result.StatusCheckRollup {
		if check.State != "SUCCESS" && check.State != "NEUTRAL" {
			status.CIPassing = false
			break
		}
	}

	// Check review status
	for _, review := range result.Reviews {
		if review.State == "APPROVED" {
			status.Approved = true
			break
		}
	}

	return status, nil
}

// PRStatus represents the current status of a PR.
type PRStatus struct {
	State     string `json:"state"`
	Mergeable bool   `json:"mergeable"`
	CIPassing bool   `json:"ciPassing"`
	Approved  bool   `json:"approved"`
}
