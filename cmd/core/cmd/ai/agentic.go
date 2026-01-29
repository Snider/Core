// Package ai provides AI agent tools and task management commands.
package ai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/core/cmd/shared"
	"github.com/host-uk/core/pkg/agentic"
	"github.com/leaanthony/clir"
)

// Style aliases for shared styles
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
	truncate     = shared.Truncate
	formatAge    = shared.FormatAge
)

var (
	taskIDStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3b82f6")) // blue-500

	taskTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")) // gray-200

	taskPriorityHighStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ef4444")) // red-500

	taskPriorityMediumStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f59e0b")) // amber-500

	taskPriorityLowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22c55e")) // green-500

	taskStatusPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6b7280")) // gray-500

	taskStatusInProgressStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#3b82f6")) // blue-500

	taskStatusCompletedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#22c55e")) // green-500

	taskStatusBlockedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ef4444")) // red-500

	taskLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a78bfa")) // violet-400
)

// AddAgenticCommands adds the agentic task management commands to the dev command.
func AddAgenticCommands(parent *clir.Command) {
	// core dev tasks - list available tasks
	addTasksCommand(parent)

	// core dev task <id> - show task details and claim
	addTaskCommand(parent)

	// core dev task:update <id> - update task
	addTaskUpdateCommand(parent)

	// core dev task:complete <id> - mark task complete
	addTaskCompleteCommand(parent)

	// core dev task:commit <id> - auto-commit with task reference
	addTaskCommitCommand(parent)

	// core dev task:pr <id> - create PR for task
	addTaskPRCommand(parent)
}

func addTasksCommand(parent *clir.Command) {
	var status string
	var priority string
	var labels string
	var limit int
	var project string

	cmd := parent.NewSubCommand("tasks", "List available tasks from core-agentic")
	cmd.LongDescription("Lists tasks from the core-agentic service.\n\n" +
		"Configuration is loaded from:\n" +
		"  1. Environment variables (AGENTIC_TOKEN, AGENTIC_BASE_URL)\n" +
		"  2. .env file in current directory\n" +
		"  3. ~/.core/agentic.yaml\n\n" +
		"Examples:\n" +
		"  core dev tasks\n" +
		"  core dev tasks --status pending --priority high\n" +
		"  core dev tasks --labels bug,urgent")

	cmd.StringFlag("status", "Filter by status (pending, in_progress, completed, blocked)", &status)
	cmd.StringFlag("priority", "Filter by priority (critical, high, medium, low)", &priority)
	cmd.StringFlag("labels", "Filter by labels (comma-separated)", &labels)
	cmd.IntFlag("limit", "Max number of tasks to return (default 20)", &limit)
	cmd.StringFlag("project", "Filter by project", &project)

	cmd.Action(func() error {
		if limit == 0 {
			limit = 20
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := agentic.NewClientFromConfig(cfg)

		opts := agentic.ListOptions{
			Limit:   limit,
			Project: project,
		}

		if status != "" {
			opts.Status = agentic.TaskStatus(status)
		}
		if priority != "" {
			opts.Priority = agentic.TaskPriority(priority)
		}
		if labels != "" {
			opts.Labels = strings.Split(labels, ",")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tasks, err := client.ListTasks(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		printTaskList(tasks)
		return nil
	})
}

func addTaskCommand(parent *clir.Command) {
	var autoSelect bool
	var claim bool
	var showContext bool

	cmd := parent.NewSubCommand("task", "Show task details or auto-select a task")
	cmd.LongDescription("Shows details of a specific task or auto-selects the highest priority task.\n\n" +
		"Examples:\n" +
		"  core dev task abc123           # Show task details\n" +
		"  core dev task abc123 --claim   # Show and claim the task\n" +
		"  core dev task abc123 --context # Show task with gathered context\n" +
		"  core dev task --auto           # Auto-select highest priority pending task")

	cmd.BoolFlag("auto", "Auto-select highest priority pending task", &autoSelect)
	cmd.BoolFlag("claim", "Claim the task after showing details", &claim)
	cmd.BoolFlag("context", "Show gathered context for AI collaboration", &showContext)

	cmd.Action(func() error {
		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var task *agentic.Task

		// Get the task ID from remaining args
		args := os.Args
		var taskID string

		// Find the task ID in args (after "task" subcommand)
		for i, arg := range args {
			if arg == "task" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				taskID = args[i+1]
				break
			}
		}

		if autoSelect {
			// Auto-select: find highest priority pending task
			tasks, err := client.ListTasks(ctx, agentic.ListOptions{
				Status: agentic.StatusPending,
				Limit:  50,
			})
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			if len(tasks) == 0 {
				fmt.Println("No pending tasks available.")
				return nil
			}

			// Sort by priority (critical > high > medium > low)
			priorityOrder := map[agentic.TaskPriority]int{
				agentic.PriorityCritical: 0,
				agentic.PriorityHigh:     1,
				agentic.PriorityMedium:   2,
				agentic.PriorityLow:      3,
			}

			sort.Slice(tasks, func(i, j int) bool {
				return priorityOrder[tasks[i].Priority] < priorityOrder[tasks[j].Priority]
			})

			task = &tasks[0]
			claim = true // Auto-select implies claiming
		} else {
			if taskID == "" {
				return fmt.Errorf("task ID required (or use --auto)")
			}

			task, err = client.GetTask(ctx, taskID)
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}
		}

		// Show context if requested
		if showContext {
			cwd, _ := os.Getwd()
			taskCtx, err := agentic.BuildTaskContext(task, cwd)
			if err != nil {
				fmt.Printf("%s Failed to build context: %s\n", errorStyle.Render(">>"), err)
			} else {
				fmt.Println(taskCtx.FormatContext())
			}
		} else {
			printTaskDetails(task)
		}

		if claim && task.Status == agentic.StatusPending {
			fmt.Println()
			fmt.Printf("%s Claiming task...\n", dimStyle.Render(">>"))

			claimedTask, err := client.ClaimTask(ctx, task.ID)
			if err != nil {
				return fmt.Errorf("failed to claim task: %w", err)
			}

			fmt.Printf("%s Task claimed successfully!\n", successStyle.Render(">>"))
			fmt.Printf("   Status: %s\n", formatTaskStatus(claimedTask.Status))
		}

		return nil
	})
}

func addTaskUpdateCommand(parent *clir.Command) {
	var status string
	var progress int
	var notes string

	cmd := parent.NewSubCommand("task:update", "Update task status or progress")
	cmd.LongDescription("Updates a task's status, progress, or adds notes.\n\n" +
		"Examples:\n" +
		"  core dev task:update abc123 --status in_progress\n" +
		"  core dev task:update abc123 --progress 50 --notes 'Halfway done'")

	cmd.StringFlag("status", "New status (pending, in_progress, completed, blocked)", &status)
	cmd.IntFlag("progress", "Progress percentage (0-100)", &progress)
	cmd.StringFlag("notes", "Notes about the update", &notes)

	cmd.Action(func() error {
		// Find task ID from args
		args := os.Args
		var taskID string
		for i, arg := range args {
			if arg == "task:update" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				taskID = args[i+1]
				break
			}
		}

		if taskID == "" {
			return fmt.Errorf("task ID required")
		}

		if status == "" && progress == 0 && notes == "" {
			return fmt.Errorf("at least one of --status, --progress, or --notes required")
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		update := agentic.TaskUpdate{
			Progress: progress,
			Notes:    notes,
		}
		if status != "" {
			update.Status = agentic.TaskStatus(status)
		}

		if err := client.UpdateTask(ctx, taskID, update); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		fmt.Printf("%s Task %s updated successfully\n", successStyle.Render(">>"), taskID)
		return nil
	})
}

func addTaskCompleteCommand(parent *clir.Command) {
	var output string
	var failed bool
	var errorMsg string

	cmd := parent.NewSubCommand("task:complete", "Mark a task as completed")
	cmd.LongDescription("Marks a task as completed with optional output and artifacts.\n\n" +
		"Examples:\n" +
		"  core dev task:complete abc123 --output 'Feature implemented'\n" +
		"  core dev task:complete abc123 --failed --error 'Build failed'")

	cmd.StringFlag("output", "Summary of the completed work", &output)
	cmd.BoolFlag("failed", "Mark the task as failed", &failed)
	cmd.StringFlag("error", "Error message if failed", &errorMsg)

	cmd.Action(func() error {
		// Find task ID from args
		args := os.Args
		var taskID string
		for i, arg := range args {
			if arg == "task:complete" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
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

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result := agentic.TaskResult{
			Success:      !failed,
			Output:       output,
			ErrorMessage: errorMsg,
		}

		if err := client.CompleteTask(ctx, taskID, result); err != nil {
			return fmt.Errorf("failed to complete task: %w", err)
		}

		if failed {
			fmt.Printf("%s Task %s marked as failed\n", errorStyle.Render(">>"), taskID)
		} else {
			fmt.Printf("%s Task %s completed successfully\n", successStyle.Render(">>"), taskID)
		}
		return nil
	})
}

func printTaskList(tasks []agentic.Task) {
	fmt.Printf("\n%d task(s) found:\n\n", len(tasks))

	for _, task := range tasks {
		id := taskIDStyle.Render(task.ID)
		title := taskTitleStyle.Render(truncate(task.Title, 50))
		priority := formatTaskPriority(task.Priority)
		status := formatTaskStatus(task.Status)

		line := fmt.Sprintf("  %s  %s  %s  %s", id, priority, status, title)

		if len(task.Labels) > 0 {
			labels := taskLabelStyle.Render("[" + strings.Join(task.Labels, ", ") + "]")
			line += " " + labels
		}

		fmt.Println(line)
	}

	fmt.Println()
	fmt.Printf("%s\n", dimStyle.Render("Use 'core dev task <id>' to view details"))
}

func printTaskDetails(task *agentic.Task) {
	fmt.Println()
	fmt.Printf("%s %s\n", dimStyle.Render("ID:"), taskIDStyle.Render(task.ID))
	fmt.Printf("%s %s\n", dimStyle.Render("Title:"), taskTitleStyle.Render(task.Title))
	fmt.Printf("%s %s\n", dimStyle.Render("Priority:"), formatTaskPriority(task.Priority))
	fmt.Printf("%s %s\n", dimStyle.Render("Status:"), formatTaskStatus(task.Status))

	if task.Project != "" {
		fmt.Printf("%s %s\n", dimStyle.Render("Project:"), task.Project)
	}

	if len(task.Labels) > 0 {
		fmt.Printf("%s %s\n", dimStyle.Render("Labels:"), taskLabelStyle.Render(strings.Join(task.Labels, ", ")))
	}

	if task.ClaimedBy != "" {
		fmt.Printf("%s %s\n", dimStyle.Render("Claimed by:"), task.ClaimedBy)
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Created:"), formatAge(task.CreatedAt))

	fmt.Println()
	fmt.Printf("%s\n", dimStyle.Render("Description:"))
	fmt.Println(task.Description)

	if len(task.Files) > 0 {
		fmt.Println()
		fmt.Printf("%s\n", dimStyle.Render("Related files:"))
		for _, f := range task.Files {
			fmt.Printf("  - %s\n", f)
		}
	}

	if len(task.Dependencies) > 0 {
		fmt.Println()
		fmt.Printf("%s %s\n", dimStyle.Render("Blocked by:"), strings.Join(task.Dependencies, ", "))
	}
}

func formatTaskPriority(p agentic.TaskPriority) string {
	switch p {
	case agentic.PriorityCritical:
		return taskPriorityHighStyle.Render("[CRITICAL]")
	case agentic.PriorityHigh:
		return taskPriorityHighStyle.Render("[HIGH]")
	case agentic.PriorityMedium:
		return taskPriorityMediumStyle.Render("[MEDIUM]")
	case agentic.PriorityLow:
		return taskPriorityLowStyle.Render("[LOW]")
	default:
		return dimStyle.Render("[" + string(p) + "]")
	}
}

func formatTaskStatus(s agentic.TaskStatus) string {
	switch s {
	case agentic.StatusPending:
		return taskStatusPendingStyle.Render("pending")
	case agentic.StatusInProgress:
		return taskStatusInProgressStyle.Render("in_progress")
	case agentic.StatusCompleted:
		return taskStatusCompletedStyle.Render("completed")
	case agentic.StatusBlocked:
		return taskStatusBlockedStyle.Render("blocked")
	default:
		return dimStyle.Render(string(s))
	}
}

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
		"  core dev task:commit abc123 --message 'add user authentication'\n" +
		"  core dev task:commit abc123 -m 'fix login bug' --scope auth\n" +
		"  core dev task:commit abc123 -m 'update docs' --push")

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
		"  core dev task:pr abc123\n" +
		"  core dev task:pr abc123 --title 'Add authentication feature'\n" +
		"  core dev task:pr abc123 --draft --labels 'enhancement,needs-review'\n" +
		"  core dev task:pr abc123 --base develop")

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
