// ai_updates.go implements task update and completion commands.

package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/leaanthony/clir"
)

func addTaskUpdateCommand(parent *clir.Command) {
	var status string
	var progress int
	var notes string

	cmd := parent.NewSubCommand("task:update", "Update task status or progress")
	cmd.LongDescription("Updates a task's status, progress, or adds notes.\n\n" +
		"Examples:\n" +
		"  core ai task:update abc123 --status in_progress\n" +
		"  core ai task:update abc123 --progress 50 --notes 'Halfway done'")

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
		"  core ai task:complete abc123 --output 'Feature implemented'\n" +
		"  core ai task:complete abc123 --failed --error 'Build failed'")

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
