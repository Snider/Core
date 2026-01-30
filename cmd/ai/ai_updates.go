// ai_updates.go implements task update and completion commands.

package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/agentic"
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
	Short: "Update task status or progress",
	Long: `Updates a task's status, progress, or adds notes.

Examples:
  core ai task:update abc123 --status in_progress
  core ai task:update abc123 --progress 50 --notes 'Halfway done'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		if taskUpdateStatus == "" && taskUpdateProgress == 0 && taskUpdateNotes == "" {
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
			Progress: taskUpdateProgress,
			Notes:    taskUpdateNotes,
		}
		if taskUpdateStatus != "" {
			update.Status = agentic.TaskStatus(taskUpdateStatus)
		}

		if err := client.UpdateTask(ctx, taskID, update); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		fmt.Printf("%s Task %s updated successfully\n", successStyle.Render(">>"), taskID)
		return nil
	},
}

var taskCompleteCmd = &cobra.Command{
	Use:   "task:complete [task-id]",
	Short: "Mark a task as completed",
	Long: `Marks a task as completed with optional output and artifacts.

Examples:
  core ai task:complete abc123 --output 'Feature implemented'
  core ai task:complete abc123 --failed --error 'Build failed'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		cfg, err := agentic.LoadConfig("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
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
			return fmt.Errorf("failed to complete task: %w", err)
		}

		if taskCompleteFailed {
			fmt.Printf("%s Task %s marked as failed\n", errorStyle.Render(">>"), taskID)
		} else {
			fmt.Printf("%s Task %s completed successfully\n", successStyle.Render(">>"), taskID)
		}
		return nil
	},
}

func init() {
	// task:update command flags
	taskUpdateCmd.Flags().StringVar(&taskUpdateStatus, "status", "", "New status (pending, in_progress, completed, blocked)")
	taskUpdateCmd.Flags().IntVar(&taskUpdateProgress, "progress", 0, "Progress percentage (0-100)")
	taskUpdateCmd.Flags().StringVar(&taskUpdateNotes, "notes", "", "Notes about the update")

	// task:complete command flags
	taskCompleteCmd.Flags().StringVar(&taskCompleteOutput, "output", "", "Summary of the completed work")
	taskCompleteCmd.Flags().BoolVar(&taskCompleteFailed, "failed", false, "Mark the task as failed")
	taskCompleteCmd.Flags().StringVar(&taskCompleteErrorMsg, "error", "", "Error message if failed")
}

func addTaskUpdateCommand(parent *cobra.Command) {
	parent.AddCommand(taskUpdateCmd)
}

func addTaskCompleteCommand(parent *cobra.Command) {
	parent.AddCommand(taskCompleteCmd)
}
