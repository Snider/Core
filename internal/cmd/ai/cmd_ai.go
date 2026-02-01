// cmd_ai.go defines styles and the AddAgenticCommands function for AI task management.

package ai

import (
	"github.com/host-uk/core/pkg/cli"
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
	taskPriorityHighStyle     = cli.NewStyle().Foreground(cli.ColourRed500)
	taskPriorityMediumStyle   = cli.NewStyle().Foreground(cli.ColourAmber500)
	taskPriorityLowStyle      = cli.NewStyle().Foreground(cli.ColourBlue400)
	taskStatusPendingStyle    = cli.DimStyle
	taskStatusInProgressStyle = cli.NewStyle().Foreground(cli.ColourBlue500)
	taskStatusCompletedStyle  = cli.SuccessStyle
	taskStatusBlockedStyle    = cli.ErrorStyle
)

// Task-specific styles (aliases to shared where possible)
var (
	taskIDStyle    = cli.TitleStyle                                 // Bold + blue
	taskTitleStyle = cli.ValueStyle                                 // Light gray
	taskLabelStyle = cli.NewStyle().Foreground(cli.ColourViolet500) // Violet for labels
)

// AddAgenticCommands adds the agentic task management commands to the ai command.
func AddAgenticCommands(parent *cli.Command) {
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
