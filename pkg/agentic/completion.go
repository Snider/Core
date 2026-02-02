// Package agentic provides AI collaboration features for task management.
package agentic

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/host-uk/core/pkg/log"
)

// PROptions contains options for creating a pull request.
type PROptions struct {
	// Title is the PR title.
	Title string `json:"title"`
	// Body is the PR description.
	Body string `json:"body"`
	// Draft marks the PR as a draft.
	Draft bool `json:"draft"`
	// Labels are labels to add to the PR.
	Labels []string `json:"labels"`
	// Base is the base branch (defaults to main).
	Base string `json:"base"`
}

// AutoCommit creates a git commit with a task reference.
// The commit message follows the format:
//
//	feat(scope): description
//
//	Task: #123
//	Co-Authored-By: Claude <noreply@anthropic.com>
func AutoCommit(ctx context.Context, task *Task, dir string, message string) error {
	const op = "agentic.AutoCommit"

	if task == nil {
		return log.E(op, "task is required", nil)
	}

	if message == "" {
		return log.E(op, "commit message is required", nil)
	}

	// Build full commit message
	fullMessage := buildCommitMessage(task, message)

	// Stage all changes
	if _, err := runGitCommandCtx(ctx, dir, "add", "-A"); err != nil {
		return log.E(op, "failed to stage changes", err)
	}

	// Create commit
	if _, err := runGitCommandCtx(ctx, dir, "commit", "-m", fullMessage); err != nil {
		return log.E(op, "failed to create commit", err)
	}

	return nil
}

// buildCommitMessage formats a commit message with task reference.
func buildCommitMessage(task *Task, message string) string {
	var sb strings.Builder

	// Write the main message
	sb.WriteString(message)
	sb.WriteString("\n\n")

	// Add task reference
	sb.WriteString("Task: #")
	sb.WriteString(task.ID)
	sb.WriteString("\n")

	// Add co-author
	sb.WriteString("Co-Authored-By: Claude <noreply@anthropic.com>\n")

	return sb.String()
}

// CreatePR creates a pull request using the gh CLI.
func CreatePR(ctx context.Context, task *Task, dir string, opts PROptions) (string, error) {
	const op = "agentic.CreatePR"

	if task == nil {
		return "", log.E(op, "task is required", nil)
	}

	// Build title if not provided
	title := opts.Title
	if title == "" {
		title = task.Title
	}

	// Build body if not provided
	body := opts.Body
	if body == "" {
		body = buildPRBody(task)
	}

	// Build gh command arguments
	args := []string{"pr", "create", "--title", title, "--body", body}

	if opts.Draft {
		args = append(args, "--draft")
	}

	if opts.Base != "" {
		args = append(args, "--base", opts.Base)
	}

	for _, label := range opts.Labels {
		args = append(args, "--label", label)
	}

	// Run gh pr create
	output, err := runCommandCtx(ctx, dir, "gh", args...)
	if err != nil {
		return "", log.E(op, "failed to create PR", err)
	}

	// Extract PR URL from output
	prURL := strings.TrimSpace(output)

	return prURL, nil
}

// buildPRBody creates a PR body from task details.
func buildPRBody(task *Task) string {
	var sb strings.Builder

	sb.WriteString("## Summary\n\n")
	sb.WriteString(task.Description)
	sb.WriteString("\n\n")

	sb.WriteString("## Task Reference\n\n")
	sb.WriteString("- Task ID: #")
	sb.WriteString(task.ID)
	sb.WriteString("\n")
	sb.WriteString("- Priority: ")
	sb.WriteString(string(task.Priority))
	sb.WriteString("\n")

	if len(task.Labels) > 0 {
		sb.WriteString("- Labels: ")
		sb.WriteString(strings.Join(task.Labels, ", "))
		sb.WriteString("\n")
	}

	sb.WriteString("\n---\n")
	sb.WriteString("Generated with AI assistance\n")

	return sb.String()
}

// SyncStatus syncs the task status back to the agentic service.
func SyncStatus(ctx context.Context, client *Client, task *Task, update TaskUpdate) error {
	const op = "agentic.SyncStatus"

	if client == nil {
		return log.E(op, "client is required", nil)
	}

	if task == nil {
		return log.E(op, "task is required", nil)
	}

	return client.UpdateTask(ctx, task.ID, update)
}

// CommitAndSync commits changes and syncs task status.
func CommitAndSync(ctx context.Context, client *Client, task *Task, dir string, message string, progress int) error {
	const op = "agentic.CommitAndSync"

	// Create commit
	if err := AutoCommit(ctx, task, dir, message); err != nil {
		return log.E(op, "failed to commit", err)
	}

	// Sync status if client provided
	if client != nil {
		update := TaskUpdate{
			Status:   StatusInProgress,
			Progress: progress,
			Notes:    "Committed: " + message,
		}

		if err := SyncStatus(ctx, client, task, update); err != nil {
			// Log but don't fail on sync errors
			return log.E(op, "commit succeeded but sync failed", err)
		}
	}

	return nil
}

// PushChanges pushes committed changes to the remote.
func PushChanges(ctx context.Context, dir string) error {
	const op = "agentic.PushChanges"

	_, err := runGitCommandCtx(ctx, dir, "push")
	if err != nil {
		return log.E(op, "failed to push changes", err)
	}

	return nil
}

// CreateBranch creates a new branch for the task.
func CreateBranch(ctx context.Context, task *Task, dir string) (string, error) {
	const op = "agentic.CreateBranch"

	if task == nil {
		return "", log.E(op, "task is required", nil)
	}

	// Generate branch name from task
	branchName := generateBranchName(task)

	// Create and checkout branch
	_, err := runGitCommandCtx(ctx, dir, "checkout", "-b", branchName)
	if err != nil {
		return "", log.E(op, "failed to create branch", err)
	}

	return branchName, nil
}

// generateBranchName creates a branch name from task details.
func generateBranchName(task *Task) string {
	// Determine prefix based on labels
	prefix := "feat"
	for _, label := range task.Labels {
		switch strings.ToLower(label) {
		case "bug", "bugfix", "fix":
			prefix = "fix"
		case "docs", "documentation":
			prefix = "docs"
		case "refactor":
			prefix = "refactor"
		case "test", "tests":
			prefix = "test"
		case "chore":
			prefix = "chore"
		}
	}

	// Sanitize title for branch name
	title := strings.ToLower(task.Title)
	title = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, title)

	// Remove consecutive dashes
	for strings.Contains(title, "--") {
		title = strings.ReplaceAll(title, "--", "-")
	}
	title = strings.Trim(title, "-")

	// Truncate if too long
	if len(title) > 40 {
		title = title[:40]
		title = strings.TrimRight(title, "-")
	}

	return fmt.Sprintf("%s/%s-%s", prefix, task.ID, title)
}

// runGitCommandCtx runs a git command with context.
func runGitCommandCtx(ctx context.Context, dir string, args ...string) (string, error) {
	return runCommandCtx(ctx, dir, "git", args...)
}

// runCommandCtx runs an arbitrary command with context.
func runCommandCtx(ctx context.Context, dir string, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%w: %s", err, stderr.String())
		}
		return "", err
	}

	return stdout.String(), nil
}

// GetCurrentBranch returns the current git branch name.
func GetCurrentBranch(ctx context.Context, dir string) (string, error) {
	const op = "agentic.GetCurrentBranch"

	output, err := runGitCommandCtx(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", log.E(op, "failed to get current branch", err)
	}

	return strings.TrimSpace(output), nil
}

// HasUncommittedChanges checks if there are uncommitted changes.
func HasUncommittedChanges(ctx context.Context, dir string) (bool, error) {
	const op = "agentic.HasUncommittedChanges"

	output, err := runGitCommandCtx(ctx, dir, "status", "--porcelain")
	if err != nil {
		return false, log.E(op, "failed to get git status", err)
	}

	return strings.TrimSpace(output) != "", nil
}

// GetDiff returns the current diff for staged and unstaged changes.
func GetDiff(ctx context.Context, dir string, staged bool) (string, error) {
	const op = "agentic.GetDiff"

	args := []string{"diff"}
	if staged {
		args = append(args, "--staged")
	}

	output, err := runGitCommandCtx(ctx, dir, args...)
	if err != nil {
		return "", log.E(op, "failed to get diff", err)
	}

	return output, nil
}
