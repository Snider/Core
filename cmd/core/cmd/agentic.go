package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/agentic"
	"github.com/leaanthony/clir"
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

	cmd := parent.NewSubCommand("task", "Show task details or auto-select a task")
	cmd.LongDescription("Shows details of a specific task or auto-selects the highest priority task.\n\n" +
		"Examples:\n" +
		"  core dev task abc123           # Show task details\n" +
		"  core dev task abc123 --claim   # Show and claim the task\n" +
		"  core dev task --auto           # Auto-select highest priority pending task")

	cmd.BoolFlag("auto", "Auto-select highest priority pending task", &autoSelect)
	cmd.BoolFlag("claim", "Claim the task after showing details", &claim)

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

		printTaskDetails(task)

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
