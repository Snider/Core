// cmd_updates.go implements task update and completion commands.

package ai

import (
	"context"
	"time"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/host-uk/core/pkg/ai"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

// task:update command flags
var (
	taskUpdateStatus   string
	taskUpdateProgress int
	taskUpdateNotes    string
)

// task:complete command flags
var (
	taskCompleteOutput   string
	taskCompleteFailed   bool
	taskCompleteErrorMsg string
)

var taskUpdateCmd = &cli.Command{
	Use:   "task:update [task-id]",
	Short: i18n.T("cmd.ai.task_update.short"),
	Long:  i18n.T("cmd.ai.task_update.long"),
	Args:  cli.ExactArgs(1),
	RunE: func(cmd *cli.Command, args []string) error {
		taskID := args[0]

		if taskUpdateStatus == "" && taskUpdateProgress == 0 && taskUpdateNotes == "" {
			return cli.Err("%s", i18n.T("cmd.ai.task_update.flag_required"))
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return cli.WrapVerb(err, "load", "config")
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		update := agentic.TaskUpdate{
			Progress: taskUpdateProgress,
			Notes:    taskUpdateNotes,
		}
		if taskUpdateStatus != "" {
			update.Status = agentic.TaskStatus(taskUpdateStatus)
		}

		if err := client.UpdateTask(ctx, taskID, update); err != nil {
			return cli.WrapVerb(err, "update", "task")
		}

		cli.Print("%s %s\n", successStyle.Render(">>"), i18n.T("i18n.done.update", "task"))
		return nil
	},
}

var taskCompleteCmd = &cli.Command{
	Use:   "task:complete [task-id]",
	Short: i18n.T("cmd.ai.task_complete.short"),
	Long:  i18n.T("cmd.ai.task_complete.long"),
	Args:  cli.ExactArgs(1),
	RunE: func(cmd *cli.Command, args []string) error {
		taskID := args[0]

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return cli.WrapVerb(err, "load", "config")
		}

		client := agentic.NewClientFromConfig(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result := agentic.TaskResult{
			Success:      !taskCompleteFailed,
			Output:       taskCompleteOutput,
			ErrorMessage: taskCompleteErrorMsg,
		}

		if err := client.CompleteTask(ctx, taskID, result); err != nil {
			return cli.WrapVerb(err, "complete", "task")
		}

		// Record task completion event
		_ = ai.Record(ai.Event{
			Type:    "task.completed",
			AgentID: cfg.AgentID,
			Data:    map[string]any{"task_id": taskID, "success": !taskCompleteFailed},
		})

		if taskCompleteFailed {
			cli.Print("%s %s\n", errorStyle.Render(">>"), i18n.T("cmd.ai.task_complete.failed", map[string]interface{}{"ID": taskID}))
		} else {
			cli.Print("%s %s\n", successStyle.Render(">>"), i18n.T("i18n.done.complete", "task"))
		}
		return nil
	},
}

func initUpdatesFlags() {
	// task:update command flags
	taskUpdateCmd.Flags().StringVar(&taskUpdateStatus, "status", "", i18n.T("cmd.ai.task_update.flag.status"))
	taskUpdateCmd.Flags().IntVar(&taskUpdateProgress, "progress", 0, i18n.T("cmd.ai.task_update.flag.progress"))
	taskUpdateCmd.Flags().StringVar(&taskUpdateNotes, "notes", "", i18n.T("cmd.ai.task_update.flag.notes"))

	// task:complete command flags
	taskCompleteCmd.Flags().StringVar(&taskCompleteOutput, "output", "", i18n.T("cmd.ai.task_complete.flag.output"))
	taskCompleteCmd.Flags().BoolVar(&taskCompleteFailed, "failed", false, i18n.T("cmd.ai.task_complete.flag.failed"))
	taskCompleteCmd.Flags().StringVar(&taskCompleteErrorMsg, "error", "", i18n.T("cmd.ai.task_complete.flag.error"))
}

func addTaskUpdateCommand(parent *cli.Command) {
	initUpdatesFlags()
	parent.AddCommand(taskUpdateCmd)
}

func addTaskCompleteCommand(parent *cli.Command) {
	parent.AddCommand(taskCompleteCmd)
}
