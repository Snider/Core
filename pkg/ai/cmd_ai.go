// cmd_ai.go defines styles and the AddAgenticCommands function for AI task management.

package ai

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/spf13/cobra"
)

// Style aliases from shared package
var (
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
	dimStyle     = cli.DimStyle
	truncate     = cli.Truncate
	formatAge    = cli.FormatAge
)

// Task priority/status styles from shared
var (
	taskPriorityHighStyle     = cli.PriorityHighStyle
	taskPriorityMediumStyle   = cli.PriorityMediumStyle
	taskPriorityLowStyle      = cli.PriorityLowStyle
	taskStatusPendingStyle    = cli.StatusPendingStyle
	taskStatusInProgressStyle = cli.StatusRunningStyle
	taskStatusCompletedStyle  = cli.StatusSuccessStyle
	taskStatusBlockedStyle    = cli.StatusErrorStyle
)

// Task-specific styles (aliases to shared where possible)
var (
	taskIDStyle    = cli.TitleStyle       // Bold + blue
	taskTitleStyle = cli.ValueStyle       // Light gray
	taskLabelStyle = cli.AccentLabelStyle // Violet for labels
)

// AddAgenticCommands adds the agentic task management commands to the ai command.
func AddAgenticCommands(parent *cobra.Command) {
	// Task listing and viewing
	addTasksCommand(parent)
	addTaskCommand(parent)

	// Task updates
	addTaskUpdateCommand(parent)
	addTaskCompleteCommand(parent)

	// Git integration
	addTaskCommitCommand(parent)
	addTaskPRCommand(parent)
}
