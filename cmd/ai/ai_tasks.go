// ai_tasks.go implements task listing and viewing commands.

package ai

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/leaanthony/clir"
)

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
		"  core ai tasks\n" +
		"  core ai tasks --status pending --priority high\n" +
		"  core ai tasks --labels bug,urgent")

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
		"  core ai task abc123           # Show task details\n" +
		"  core ai task abc123 --claim   # Show and claim the task\n" +
		"  core ai task abc123 --context # Show task with gathered context\n" +
		"  core ai task --auto           # Auto-select highest priority pending task")

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
	fmt.Printf("%s\n", dimStyle.Render("Use 'core ai task <id>' to view details"))
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
