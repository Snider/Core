// ai_updates.go implements task update and completion commands.

package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
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

var taskUpdateCmd = &cobra.Command{
	Use:   "task:update [task-id]",
	Short: i18n.T("cmd.ai.task_update.short"),
	Long:  i18n.T("cmd.ai.task_update.long"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		if taskUpdateStatus == "" && taskUpdateProgress == 0 && taskUpdateNotes == "" {
			return fmt.Errorf(i18n.T("cmd.ai.task_update.flag_required"))
		}

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.load_config"), err)
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
			return fmt.Errorf("%s: %w", i18n.T("cmd.ai.error.update_task"), err)
		}

		fmt.Printf("%s %s\n", successStyle.Render(">>"), i18n.T("cmd.ai.task_update.success", map[string]interface{}{"ID": taskID}))
		return nil
	},
}

var taskCompleteCmd = &cobra.Command{
	Use:   "task:complete [task-id]",
	Short: i18n.T("cmd.ai.task_complete.short"),
	Long:  i18n.T("cmd.ai.task_complete.long"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.load_config"), err)
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
			return fmt.Errorf("%s: %w", i18n.T("cmd.ai.error.complete_task"), err)
		}

		if taskCompleteFailed {
			fmt.Printf("%s %s\n", errorStyle.Render(">>"), i18n.T("cmd.ai.task_complete.failed", map[string]interface{}{"ID": taskID}))
		} else {
			fmt.Printf("%s %s\n", successStyle.Render(">>"), i18n.T("cmd.ai.task_complete.success", map[string]interface{}{"ID": taskID}))
		}
		return nil
	},
}

func init() {
	// task:update command flags
	taskUpdateCmd.Flags().StringVar(&taskUpdateStatus, "status", "", i18n.T("cmd.ai.task_update.flag.status"))
	taskUpdateCmd.Flags().IntVar(&taskUpdateProgress, "progress", 0, i18n.T("cmd.ai.task_update.flag.progress"))
	taskUpdateCmd.Flags().StringVar(&taskUpdateNotes, "notes", "", i18n.T("cmd.ai.task_update.flag.notes"))

	// task:complete command flags
	taskCompleteCmd.Flags().StringVar(&taskCompleteOutput, "output", "", i18n.T("cmd.ai.task_complete.flag.output"))
	taskCompleteCmd.Flags().BoolVar(&taskCompleteFailed, "failed", false, i18n.T("cmd.ai.task_complete.flag.failed"))
	taskCompleteCmd.Flags().StringVar(&taskCompleteErrorMsg, "error", "", i18n.T("cmd.ai.task_complete.flag.error"))
}

func addTaskUpdateCommand(parent *cobra.Command) {
	parent.AddCommand(taskUpdateCmd)
}

func addTaskCompleteCommand(parent *cobra.Command) {
	parent.AddCommand(taskCompleteCmd)
}
