// cmd_tasks.go implements task listing and viewing commands.

package ai

import (
	"context"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/host-uk/core/pkg/ai"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
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

var tasksCmd = &cli.Command{
	Use:   "tasks",
	Short: i18n.T("cmd.ai.tasks.short"),
	Long:  i18n.T("cmd.ai.tasks.long"),
	RunE: func(cmd *cli.Command, args []string) error {
		limit := tasksLimit
		if limit == 0 {
			limit = 20
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return cli.WrapVerb(err, "load", "config")
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
			return cli.WrapVerb(err, "list", "tasks")
		}

		if len(tasks) == 0 {
			cli.Text(i18n.T("cmd.ai.tasks.none_found"))
			return nil
		}

		printTaskList(tasks)
		return nil
	},
}

var taskCmd = &cli.Command{
	Use:   "task [task-id]",
	Short: i18n.T("cmd.ai.task.short"),
	Long:  i18n.T("cmd.ai.task.long"),
	RunE: func(cmd *cli.Command, args []string) error {
		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return cli.WrapVerb(err, "load", "config")
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
				return cli.WrapVerb(err, "list", "tasks")
			}

			if len(tasks) == 0 {
				cli.Text(i18n.T("cmd.ai.task.no_pending"))
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
				return cli.Err("%s", i18n.T("cmd.ai.task.id_required"))
			}

			task, err = client.GetTask(ctx, taskID)
			if err != nil {
				return cli.WrapVerb(err, "get", "task")
			}
		}

		// Show context if requested
		if taskShowContext {
			cwd, _ := os.Getwd()
			taskCtx, err := agentic.BuildTaskContext(task, cwd)
			if err != nil {
				cli.Print("%s %s: %s\n", errorStyle.Render(">>"), i18n.T("i18n.fail.build", "context"), err)
			} else {
				cli.Text(taskCtx.FormatContext())
			}
		} else {
			printTaskDetails(task)
		}

		if taskClaim && task.Status == agentic.StatusPending {
			cli.Blank()
			cli.Print("%s %s\n", dimStyle.Render(">>"), i18n.T("cmd.ai.task.claiming"))

			claimedTask, err := client.ClaimTask(ctx, task.ID)
			if err != nil {
				return cli.WrapVerb(err, "claim", "task")
			}

			// Record task claim event
			_ = ai.Record(ai.Event{
				Type:    "task.claimed",
				AgentID: cfg.AgentID,
				Data:    map[string]any{"task_id": task.ID, "title": task.Title},
			})

			cli.Print("%s %s\n", successStyle.Render(">>"), i18n.T("i18n.done.claim", "task"))
			cli.Print("   %s %s\n", i18n.Label("status"), formatTaskStatus(claimedTask.Status))
		}

		return nil
	},
}

func initTasksFlags() {
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

func addTasksCommand(parent *cli.Command) {
	initTasksFlags()
	parent.AddCommand(tasksCmd)
}

func addTaskCommand(parent *cli.Command) {
	parent.AddCommand(taskCmd)
}

func printTaskList(tasks []agentic.Task) {
	cli.Print("\n%s\n\n", i18n.T("cmd.ai.tasks.found", map[string]interface{}{"Count": len(tasks)}))

	for _, task := range tasks {
		id := taskIDStyle.Render(task.ID)
		title := taskTitleStyle.Render(truncate(task.Title, 50))
		priority := formatTaskPriority(task.Priority)
		status := formatTaskStatus(task.Status)

		line := cli.Sprintf("  %s  %s  %s  %s", id, priority, status, title)

		if len(task.Labels) > 0 {
			labels := taskLabelStyle.Render("[" + strings.Join(task.Labels, ", ") + "]")
			line += " " + labels
		}

		cli.Text(line)
	}

	cli.Blank()
	cli.Print("%s\n", dimStyle.Render(i18n.T("cmd.ai.tasks.hint")))
}

func printTaskDetails(task *agentic.Task) {
	cli.Blank()
	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.id")), taskIDStyle.Render(task.ID))
	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.title")), taskTitleStyle.Render(task.Title))
	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.priority")), formatTaskPriority(task.Priority))
	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("status")), formatTaskStatus(task.Status))

	if task.Project != "" {
		cli.Print("%s %s\n", dimStyle.Render(i18n.Label("project")), task.Project)
	}

	if len(task.Labels) > 0 {
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.labels")), taskLabelStyle.Render(strings.Join(task.Labels, ", ")))
	}

	if task.ClaimedBy != "" {
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.claimed_by")), task.ClaimedBy)
	}

	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.created")), formatAge(task.CreatedAt))

	cli.Blank()
	cli.Print("%s\n", dimStyle.Render(i18n.T("cmd.ai.label.description")))
	cli.Text(task.Description)

	if len(task.Files) > 0 {
		cli.Blank()
		cli.Print("%s\n", dimStyle.Render(i18n.T("cmd.ai.label.related_files")))
		for _, f := range task.Files {
			cli.Print("  - %s\n", f)
		}
	}

	if len(task.Dependencies) > 0 {
		cli.Blank()
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.ai.label.blocked_by")), strings.Join(task.Dependencies, ", "))
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
