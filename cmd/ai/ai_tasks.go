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
	"github.com/host-uk/core/pkg/i18n"
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
	Short: i18n.T("cmd.ai.tasks.short"),
	Long:  i18n.T("cmd.ai.tasks.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit := tasksLimit
		if limit == 0 {
			limit = 20
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("cmd.ai.error.load_config"), err)
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
			return fmt.Errorf("%s: %w", i18n.T("cmd.ai.error.list_tasks"), err)
		}

		if len(tasks) == 0 {
			fmt.Println(i18n.T("cmd.ai.tasks.none_found"))
			return nil
		}

		printTaskList(tasks)
		return nil
	},
}

var taskCmd = &cobra.Command{
	Use:   "task [task-id]",
	Short: i18n.T("cmd.ai.task.short"),
	Long:  i18n.T("cmd.ai.task.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("cmd.ai.error.load_config"), err)
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
				return fmt.Errorf("%s: %w", i18n.T("cmd.ai.error.list_tasks"), err)
			}

			if len(tasks) == 0 {
				fmt.Println(i18n.T("cmd.ai.task.no_pending"))
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
				return fmt.Errorf(i18n.T("cmd.ai.task.id_required"))
			}

			task, err = client.GetTask(ctx, taskID)
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.ai.error.get_task"), err)
			}
		}

		// Show context if requested
		if taskShowContext {
			cwd, _ := os.Getwd()
			taskCtx, err := agentic.BuildTaskContext(task, cwd)
			if err != nil {
				fmt.Printf("%s %s: %s\n", errorStyle.Render(">>"), i18n.T("cmd.ai.task.context_failed"), err)
			} else {
				fmt.Println(taskCtx.FormatContext())
			}
		} else {
			printTaskDetails(task)
		}

		if taskClaim && task.Status == agentic.StatusPending {
			fmt.Println()
			fmt.Printf("%s %s\n", dimStyle.Render(">>"), i18n.T("cmd.ai.task.claiming"))

			claimedTask, err := client.ClaimTask(ctx, task.ID)
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.ai.error.claim_task"), err)
			}

			fmt.Printf("%s %s\n", successStyle.Render(">>"), i18n.T("cmd.ai.task.claimed"))
			fmt.Printf("   %s %s\n", i18n.T("cmd.ai.label.status"), formatTaskStatus(claimedTask.Status))
		}

		return nil
	},
}

func init() {
	// tasks command flags
	tasksCmd.Flags().StringVar(&tasksStatus, "status", "", i18n.T("cmd.ai.tasks.flag.status"))
	tasksCmd.Flags().StringVar(&tasksPriority, "priority", "", i18n.T("cmd.ai.tasks.flag.priority"))
	tasksCmd.Flags().StringVar(&tasksLabels, "labels", "", i18n.T("cmd.ai.tasks.flag.labels"))
	tasksCmd.Flags().IntVar(&tasksLimit, "limit", 20, i18n.T("cmd.ai.tasks.flag.limit"))
	tasksCmd.Flags().StringVar(&tasksProject, "project", "", i18n.T("cmd.ai.tasks.flag.project"))

	// task command flags
	taskCmd.Flags().BoolVar(&taskAutoSelect, "auto", false, i18n.T("cmd.ai.task.flag.auto"))
	taskCmd.Flags().BoolVar(&taskClaim, "claim", false, i18n.T("cmd.ai.task.flag.claim"))
	taskCmd.Flags().BoolVar(&taskShowContext, "context", false, i18n.T("cmd.ai.task.flag.context"))
}

func addTasksCommand(parent *cobra.Command) {
	parent.AddCommand(tasksCmd)
}

func addTaskCommand(parent *cobra.Command) {
	parent.AddCommand(taskCmd)
}

func printTaskList(tasks []agentic.Task) {
	fmt.Printf("\n%s\n\n", i18n.T("cmd.ai.tasks.found", map[string]interface{}{"Count": len(tasks)}))

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
	fmt.Printf("%s\n", dimStyle.Render(i18n.T("cmd.ai.tasks.hint")))
}

func printTaskDetails(task *agentic.Task) {
	fmt.Println()
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.id")), taskIDStyle.Render(task.ID))
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.title")), taskTitleStyle.Render(task.Title))
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.priority")), formatTaskPriority(task.Priority))
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.status")), formatTaskStatus(task.Status))

	if task.Project != "" {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.project")), task.Project)
	}

	if len(task.Labels) > 0 {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.labels")), taskLabelStyle.Render(strings.Join(task.Labels, ", ")))
	}

	if task.ClaimedBy != "" {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.claimed_by")), task.ClaimedBy)
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.created")), formatAge(task.CreatedAt))

	fmt.Println()
	fmt.Printf("%s\n", dimStyle.Render(i18n.T("cmd.ai.label.description")))
	fmt.Println(task.Description)

	if len(task.Files) > 0 {
		fmt.Println()
		fmt.Printf("%s\n", dimStyle.Render(i18n.T("cmd.ai.label.related_files")))
		for _, f := range task.Files {
			fmt.Printf("  - %s\n", f)
		}
	}

	if len(task.Dependencies) > 0 {
		fmt.Println()
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.blocked_by")), strings.Join(task.Dependencies, ", "))
	}
}

func formatTaskPriority(p agentic.TaskPriority) string {
	switch p {
	case agentic.PriorityCritical:
		return taskPriorityHighStyle.Render("[" + i18n.T("cmd.ai.priority.critical") + "]")
	case agentic.PriorityHigh:
		return taskPriorityHighStyle.Render("[" + i18n.T("cmd.ai.priority.high") + "]")
	case agentic.PriorityMedium:
		return taskPriorityMediumStyle.Render("[" + i18n.T("cmd.ai.priority.medium") + "]")
	case agentic.PriorityLow:
		return taskPriorityLowStyle.Render("[" + i18n.T("cmd.ai.priority.low") + "]")
	default:
		return dimStyle.Render("[" + string(p) + "]")
	}
}

func formatTaskStatus(s agentic.TaskStatus) string {
	switch s {
	case agentic.StatusPending:
		return taskStatusPendingStyle.Render(i18n.T("cmd.ai.status.pending"))
	case agentic.StatusInProgress:
		return taskStatusInProgressStyle.Render(i18n.T("cmd.ai.status.in_progress"))
	case agentic.StatusCompleted:
		return taskStatusCompletedStyle.Render(i18n.T("cmd.ai.status.completed"))
	case agentic.StatusBlocked:
		return taskStatusBlockedStyle.Render(i18n.T("cmd.ai.status.blocked"))
	default:
		return dimStyle.Render(string(s))
	}
}
