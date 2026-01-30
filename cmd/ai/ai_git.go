// ai_git.go implements git integration commands for task commits and PRs.

package ai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/leaanthony/clir"
)

func addTaskCommitCommand(parent *clir.Command) {
	var message string
	var scope string
	var push bool

	cmd := parent.NewSubCommand("task:commit", "Auto-commit changes with task reference")
	cmd.LongDescription("Creates a git commit with a task reference and co-author attribution.\n\n" +
		"Commit message format:\n" +
		"  feat(scope): description\n" +
		"\n" +
		"  Task: #123\n" +
		"  Co-Authored-By: Claude <noreply@anthropic.com>\n\n" +
		"Examples:\n" +
		"  core ai task:commit abc123 --message 'add user authentication'\n" +
		"  core ai task:commit abc123 -m 'fix login bug' --scope auth\n" +
		"  core ai task:commit abc123 -m 'update docs' --push")

	cmd.StringFlag("message", "Commit message (without task reference)", &message)
	cmd.StringFlag("m", "Commit message (short form)", &message)
	cmd.StringFlag("scope", "Scope for the commit type (e.g., auth, api, ui)", &scope)
	cmd.BoolFlag("push", "Push changes after committing", &push)

	cmd.Action(func() error {
		// Find task ID from args
		args := os.Args
		var taskID string
		for i, arg := range args {
			if arg == "task:commit" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				taskID = args[i+1]
				break
			}
		}

		if taskID == "" {
			return fmt.Errorf("task ID required")
		}

		if message == "" {
			return fmt.Errorf("commit message required (--message or -m)")
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get task details
		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}

		// Build commit message with optional scope
		commitType := inferCommitType(task.Labels)
		var fullMessage string
		if scope != "" {
			fullMessage = fmt.Sprintf("%s(%s): %s", commitType, scope, message)
		} else {
			fullMessage = fmt.Sprintf("%s: %s", commitType, message)
		}

		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		// Check for uncommitted changes
		hasChanges, err := agentic.HasUncommittedChanges(ctx, cwd)
		if err != nil {
			return fmt.Errorf("failed to check git status: %w", err)
		}

		if !hasChanges {
			fmt.Println("No uncommitted changes to commit.")
			return nil
		}

		// Create commit
		fmt.Printf("%s Creating commit for task %s...\n", dimStyle.Render(">>"), taskID)
		if err := agentic.AutoCommit(ctx, task, cwd, fullMessage); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}

		fmt.Printf("%s Committed: %s\n", successStyle.Render(">>"), fullMessage)

		// Push if requested
		if push {
			fmt.Printf("%s Pushing changes...\n", dimStyle.Render(">>"))
			if err := agentic.PushChanges(ctx, cwd); err != nil {
				return fmt.Errorf("failed to push: %w", err)
			}
			fmt.Printf("%s Changes pushed successfully\n", successStyle.Render(">>"))
		}

		return nil
	})
}

func addTaskPRCommand(parent *clir.Command) {
	var title string
	var draft bool
	var labels string
	var base string

	cmd := parent.NewSubCommand("task:pr", "Create a pull request for a task")
	cmd.LongDescription("Creates a GitHub pull request linked to a task.\n\n" +
		"Requires the GitHub CLI (gh) to be installed and authenticated.\n\n" +
		"Examples:\n" +
		"  core ai task:pr abc123\n" +
		"  core ai task:pr abc123 --title 'Add authentication feature'\n" +
		"  core ai task:pr abc123 --draft --labels 'enhancement,needs-review'\n" +
		"  core ai task:pr abc123 --base develop")

	cmd.StringFlag("title", "PR title (defaults to task title)", &title)
	cmd.BoolFlag("draft", "Create as draft PR", &draft)
	cmd.StringFlag("labels", "Labels to add (comma-separated)", &labels)
	cmd.StringFlag("base", "Base branch (defaults to main)", &base)

	cmd.Action(func() error {
		// Find task ID from args
		args := os.Args
		var taskID string
		for i, arg := range args {
			if arg == "task:pr" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				taskID = args[i+1]
				break
			}
		}

		if taskID == "" {
			return fmt.Errorf("task ID required")
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Get task details
		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}

		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		// Check current branch
		branch, err := agentic.GetCurrentBranch(ctx, cwd)
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}

		if branch == "main" || branch == "master" {
			return fmt.Errorf("cannot create PR from %s branch; create a feature branch first", branch)
		}

		// Push current branch
		fmt.Printf("%s Pushing branch %s...\n", dimStyle.Render(">>"), branch)
		if err := agentic.PushChanges(ctx, cwd); err != nil {
			// Try setting upstream
			if _, err := runGitCommand(cwd, "push", "-u", "origin", branch); err != nil {
				return fmt.Errorf("failed to push branch: %w", err)
			}
		}

		// Build PR options
		opts := agentic.PROptions{
			Title: title,
			Draft: draft,
			Base:  base,
		}

		if labels != "" {
			opts.Labels = strings.Split(labels, ",")
		}

		// Create PR
		fmt.Printf("%s Creating pull request...\n", dimStyle.Render(">>"))
		prURL, err := agentic.CreatePR(ctx, task, cwd, opts)
		if err != nil {
			return fmt.Errorf("failed to create PR: %w", err)
		}

		fmt.Printf("%s Pull request created!\n", successStyle.Render(">>"))
		fmt.Printf("   URL: %s\n", prURL)

		return nil
	})
}

// inferCommitType infers the commit type from task labels.
func inferCommitType(labels []string) string {
	for _, label := range labels {
		switch strings.ToLower(label) {
		case "bug", "bugfix", "fix":
			return "fix"
		case "docs", "documentation":
			return "docs"
		case "refactor", "refactoring":
			return "refactor"
		case "test", "tests", "testing":
			return "test"
		case "chore":
			return "chore"
		case "style":
			return "style"
		case "perf", "performance":
			return "perf"
		case "ci":
			return "ci"
		case "build":
			return "build"
		}
	}
	return "feat"
}

// runGitCommand runs a git command in the specified directory.
func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
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
