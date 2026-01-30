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
	"github.com/spf13/cobra"
)

// task:commit command flags
var (
	taskCommitMessage string
	taskCommitScope   string
	taskCommitPush    bool
)

// task:pr command flags
var (
	taskPRTitle  string
	taskPRDraft  bool
	taskPRLabels string
	taskPRBase   string
)

var taskCommitCmd = &cobra.Command{
	Use:   "task:commit [task-id]",
	Short: "Auto-commit changes with task reference",
	Long: `Creates a git commit with a task reference and co-author attribution.

Commit message format:
  feat(scope): description

  Task: #123
  Co-Authored-By: Claude <noreply@anthropic.com>

Examples:
  core ai task:commit abc123 --message 'add user authentication'
  core ai task:commit abc123 -m 'fix login bug' --scope auth
  core ai task:commit abc123 -m 'update docs' --push`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		if taskCommitMessage == "" {
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
		if taskCommitScope != "" {
			fullMessage = fmt.Sprintf("%s(%s): %s", commitType, taskCommitScope, taskCommitMessage)
		} else {
			fullMessage = fmt.Sprintf("%s: %s", commitType, taskCommitMessage)
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
		if taskCommitPush {
			fmt.Printf("%s Pushing changes...\n", dimStyle.Render(">>"))
			if err := agentic.PushChanges(ctx, cwd); err != nil {
				return fmt.Errorf("failed to push: %w", err)
			}
			fmt.Printf("%s Changes pushed successfully\n", successStyle.Render(">>"))
		}

		return nil
	},
}

var taskPRCmd = &cobra.Command{
	Use:   "task:pr [task-id]",
	Short: "Create a pull request for a task",
	Long: `Creates a GitHub pull request linked to a task.

Requires the GitHub CLI (gh) to be installed and authenticated.

Examples:
  core ai task:pr abc123
  core ai task:pr abc123 --title 'Add authentication feature'
  core ai task:pr abc123 --draft --labels 'enhancement,needs-review'
  core ai task:pr abc123 --base develop`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

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
			Title: taskPRTitle,
			Draft: taskPRDraft,
			Base:  taskPRBase,
		}

		if taskPRLabels != "" {
			opts.Labels = strings.Split(taskPRLabels, ",")
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
	},
}

func init() {
	// task:commit command flags
	taskCommitCmd.Flags().StringVarP(&taskCommitMessage, "message", "m", "", "Commit message (without task reference)")
	taskCommitCmd.Flags().StringVar(&taskCommitScope, "scope", "", "Scope for the commit type (e.g., auth, api, ui)")
	taskCommitCmd.Flags().BoolVar(&taskCommitPush, "push", false, "Push changes after committing")

	// task:pr command flags
	taskPRCmd.Flags().StringVar(&taskPRTitle, "title", "", "PR title (defaults to task title)")
	taskPRCmd.Flags().BoolVar(&taskPRDraft, "draft", false, "Create as draft PR")
	taskPRCmd.Flags().StringVar(&taskPRLabels, "labels", "", "Labels to add (comma-separated)")
	taskPRCmd.Flags().StringVar(&taskPRBase, "base", "", "Base branch (defaults to main)")
}

func addTaskCommitCommand(parent *cobra.Command) {
	parent.AddCommand(taskCommitCmd)
}

func addTaskPRCommand(parent *cobra.Command) {
	parent.AddCommand(taskPRCmd)
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
