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
	"github.com/host-uk/core/pkg/i18n"
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
	Short: i18n.T("cmd.ai.task_commit.short"),
	Long:  i18n.T("cmd.ai.task_commit.long"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		if taskCommitMessage == "" {
			return fmt.Errorf(i18n.T("cmd.ai.task_commit.message_required"))
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "load config"}), err)
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get task details
		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get task"}), err)
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
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
		}

		// Check for uncommitted changes
		hasChanges, err := agentic.HasUncommittedChanges(ctx, cwd)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "check git status"}), err)
		}

		if !hasChanges {
			fmt.Println(i18n.T("cmd.ai.task_commit.no_changes"))
			return nil
		}

		// Create commit
		fmt.Printf("%s %s\n", dimStyle.Render(">>"), i18n.T("cmd.ai.task_commit.creating", map[string]interface{}{"ID": taskID}))
		if err := agentic.AutoCommit(ctx, task, cwd, fullMessage); err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "commit"}), err)
		}

		fmt.Printf("%s %s %s\n", successStyle.Render(">>"), i18n.T("cmd.ai.task_commit.committed"), fullMessage)

		// Push if requested
		if taskCommitPush {
			fmt.Printf("%s %s\n", dimStyle.Render(">>"), i18n.T("cmd.ai.task_commit.pushing"))
			if err := agentic.PushChanges(ctx, cwd); err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "push"}), err)
			}
			fmt.Printf("%s %s\n", successStyle.Render(">>"), i18n.T("common.success.completed", map[string]any{"Action": "Changes pushed"}))
		}

		return nil
	},
}

var taskPRCmd = &cobra.Command{
	Use:   "task:pr [task-id]",
	Short: i18n.T("cmd.ai.task_pr.short"),
	Long:  i18n.T("cmd.ai.task_pr.long"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "load config"}), err)
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Get task details
		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get task"}), err)
		}

		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
		}

		// Check current branch
		branch, err := agentic.GetCurrentBranch(ctx, cwd)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get current branch"}), err)
		}

		if branch == "main" || branch == "master" {
			return fmt.Errorf(i18n.T("cmd.ai.task_pr.branch_error", map[string]interface{}{"Branch": branch}))
		}

		// Push current branch
		fmt.Printf("%s %s\n", dimStyle.Render(">>"), i18n.T("cmd.ai.task_pr.pushing_branch", map[string]interface{}{"Branch": branch}))
		if err := agentic.PushChanges(ctx, cwd); err != nil {
			// Try setting upstream
			if _, err := runGitCommand(cwd, "push", "-u", "origin", branch); err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "push branch"}), err)
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
		fmt.Printf("%s %s\n", dimStyle.Render(">>"), i18n.T("cmd.ai.task_pr.creating"))
		prURL, err := agentic.CreatePR(ctx, task, cwd, opts)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "create PR"}), err)
		}

		fmt.Printf("%s %s\n", successStyle.Render(">>"), i18n.T("cmd.ai.task_pr.created"))
		fmt.Printf("   %s %s\n", i18n.T("common.label.url"), prURL)

		return nil
	},
}

func init() {
	// task:commit command flags
	taskCommitCmd.Flags().StringVarP(&taskCommitMessage, "message", "m", "", i18n.T("cmd.ai.task_commit.flag.message"))
	taskCommitCmd.Flags().StringVar(&taskCommitScope, "scope", "", i18n.T("cmd.ai.task_commit.flag.scope"))
	taskCommitCmd.Flags().BoolVar(&taskCommitPush, "push", false, i18n.T("cmd.ai.task_commit.flag.push"))

	// task:pr command flags
	taskPRCmd.Flags().StringVar(&taskPRTitle, "title", "", i18n.T("cmd.ai.task_pr.flag.title"))
	taskPRCmd.Flags().BoolVar(&taskPRDraft, "draft", false, i18n.T("cmd.ai.task_pr.flag.draft"))
	taskPRCmd.Flags().StringVar(&taskPRLabels, "labels", "", i18n.T("cmd.ai.task_pr.flag.labels"))
	taskPRCmd.Flags().StringVar(&taskPRBase, "base", "", i18n.T("cmd.ai.task_pr.flag.base"))
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
