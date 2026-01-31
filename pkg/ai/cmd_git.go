// cmd_git.go implements git integration commands for task commits and PRs.

package ai

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
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

var taskCommitCmd = &cli.Command{
	Use:   "task:commit [task-id]",
	Short: i18n.T("cmd.ai.task_commit.short"),
	Long:  i18n.T("cmd.ai.task_commit.long"),
	Args:  cli.ExactArgs(1),
	RunE: func(cmd *cli.Command, args []string) error {
		taskID := args[0]

		if taskCommitMessage == "" {
			return cli.Err("commit message required")
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return cli.WrapVerb(err, "load", "config")
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get task details
		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			return cli.WrapVerb(err, "get", "task")
		}

		// Build commit message with optional scope
		commitType := inferCommitType(task.Labels)
		var fullMessage string
		if taskCommitScope != "" {
			fullMessage = cli.Sprintf("%s(%s): %s", commitType, taskCommitScope, taskCommitMessage)
		} else {
			fullMessage = cli.Sprintf("%s: %s", commitType, taskCommitMessage)
		}

		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return cli.WrapVerb(err, "get", "working directory")
		}

		// Check for uncommitted changes
		hasChanges, err := agentic.HasUncommittedChanges(ctx, cwd)
		if err != nil {
			return cli.WrapVerb(err, "check", "git status")
		}

		if !hasChanges {
			cli.Println("No changes to commit")
			return nil
		}

		// Create commit
		cli.Print("%s %s\n", dimStyle.Render(">>"), i18n.ProgressSubject("create", "commit for "+taskID))
		if err := agentic.AutoCommit(ctx, task, cwd, fullMessage); err != nil {
			return cli.WrapAction(err, "commit")
		}

		cli.Print("%s %s %s\n", successStyle.Render(">>"), i18n.T("i18n.done.commit")+":", fullMessage)

		// Push if requested
		if taskCommitPush {
			cli.Print("%s %s\n", dimStyle.Render(">>"), i18n.Progress("push"))
			if err := agentic.PushChanges(ctx, cwd); err != nil {
				return cli.WrapAction(err, "push")
			}
			cli.Print("%s %s\n", successStyle.Render(">>"), i18n.T("i18n.done.push", "changes"))
		}

		return nil
	},
}

var taskPRCmd = &cli.Command{
	Use:   "task:pr [task-id]",
	Short: i18n.T("cmd.ai.task_pr.short"),
	Long:  i18n.T("cmd.ai.task_pr.long"),
	Args:  cli.ExactArgs(1),
	RunE: func(cmd *cli.Command, args []string) error {
		taskID := args[0]

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return cli.WrapVerb(err, "load", "config")
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Get task details
		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			return cli.WrapVerb(err, "get", "task")
		}

		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return cli.WrapVerb(err, "get", "working directory")
		}

		// Check current branch
		branch, err := agentic.GetCurrentBranch(ctx, cwd)
		if err != nil {
			return cli.WrapVerb(err, "get", "branch")
		}

		if branch == "main" || branch == "master" {
			return cli.Err("cannot create PR from %s branch", branch)
		}

		// Push current branch
		cli.Print("%s %s\n", dimStyle.Render(">>"), i18n.ProgressSubject("push", branch))
		if err := agentic.PushChanges(ctx, cwd); err != nil {
			// Try setting upstream
			if _, err := runGitCommand(cwd, "push", "-u", "origin", branch); err != nil {
				return cli.WrapVerb(err, "push", "branch")
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
		cli.Print("%s %s\n", dimStyle.Render(">>"), i18n.ProgressSubject("create", "PR"))
		prURL, err := agentic.CreatePR(ctx, task, cwd, opts)
		if err != nil {
			return cli.WrapVerb(err, "create", "PR")
		}

		cli.Print("%s %s\n", successStyle.Render(">>"), i18n.T("i18n.done.create", "PR"))
		cli.Print("   %s %s\n", i18n.Label("url"), prURL)

		return nil
	},
}

func initGitFlags() {
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

func addTaskCommitCommand(parent *cli.Command) {
	initGitFlags()
	parent.AddCommand(taskCommitCmd)
}

func addTaskPRCommand(parent *cli.Command) {
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
			return "", cli.Wrap(err, stderr.String())
		}
		return "", err
	}

	return stdout.String(), nil
}
