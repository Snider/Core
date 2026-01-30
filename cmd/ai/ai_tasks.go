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
	"github.com/spf13/cobra"
)

// tasks command flags
var (
	tasksStatus   string
	tasksPriority string
	tasksLabels   string
	tasksLimit    int
	tasksProject  string
)

// task command flags
var (
	taskAutoSelect  bool
	taskClaim       bool
	taskShowContext bool
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List available tasks from core-agentic",
	Long: `Lists tasks from the core-agentic service.

Configuration is loaded from:
  1. Environment variables (AGENTIC_TOKEN, AGENTIC_BASE_URL)
  2. .env file in current directory
  3. ~/.core/agentic.yaml

Examples:
  core ai tasks
  core ai tasks --status pending --priority high
  core ai tasks --labels bug,urgent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit := tasksLimit
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
			Project: tasksProject,
		}

		if tasksStatus != "" {
			opts.Status = agentic.TaskStatus(tasksStatus)
		}
		if tasksPriority != "" {
			opts.Priority = agentic.TaskPriority(tasksPriority)
		}
		if tasksLabels != "" {
			opts.Labels = strings.Split(tasksLabels, ",")
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
	},
}

var taskCmd = &cobra.Command{
	Use:   "task [task-id]",
	Short: "Show task details or auto-select a task",
	Long: `Shows details of a specific task or auto-selects the highest priority task.

Examples:
  core ai task abc123           # Show task details
  core ai task abc123 --claim   # Show and claim the task
  core ai task abc123 --context # Show task with gathered context
  core ai task --auto           # Auto-select highest priority pending task`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var task *agentic.Task

		// Get the task ID from args
		var taskID string
		if len(args) > 0 {
			taskID = args[0]
		}

		if taskAutoSelect {
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
			taskClaim = true // Auto-select implies claiming
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
		if taskShowContext {
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

		if taskClaim && task.Status == agentic.StatusPending {
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
	},
}

func init() {
	// tasks command flags
	tasksCmd.Flags().StringVar(&tasksStatus, "status", "", "Filter by status (pending, in_progress, completed, blocked)")
	tasksCmd.Flags().StringVar(&tasksPriority, "priority", "", "Filter by priority (critical, high, medium, low)")
	tasksCmd.Flags().StringVar(&tasksLabels, "labels", "", "Filter by labels (comma-separated)")
	tasksCmd.Flags().IntVar(&tasksLimit, "limit", 20, "Max number of tasks to return")
	tasksCmd.Flags().StringVar(&tasksProject, "project", "", "Filter by project")

	// task command flags
	taskCmd.Flags().BoolVar(&taskAutoSelect, "auto", false, "Auto-select highest priority pending task")
	taskCmd.Flags().BoolVar(&taskClaim, "claim", false, "Claim the task after showing details")
	taskCmd.Flags().BoolVar(&taskShowContext, "context", false, "Show gathered context for AI collaboration")
}

func addTasksCommand(parent *cobra.Command) {
	parent.AddCommand(tasksCmd)
}

func addTaskCommand(parent *cobra.Command) {
	parent.AddCommand(taskCmd)
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
