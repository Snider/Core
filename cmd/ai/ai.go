// ai.go defines styles and the AddAgenticCommands function for AI task management.

package ai

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
)

// Style aliases from shared package
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
	truncate     = shared.Truncate
	formatAge    = shared.FormatAge
)

// Task priority/status styles from shared
var (
	taskPriorityHighStyle     = shared.PriorityHighStyle
	taskPriorityMediumStyle   = shared.PriorityMediumStyle
	taskPriorityLowStyle      = shared.PriorityLowStyle
	taskStatusPendingStyle    = shared.StatusPendingStyle
	taskStatusInProgressStyle = shared.StatusRunningStyle
	taskStatusCompletedStyle  = shared.StatusSuccessStyle
	taskStatusBlockedStyle    = shared.StatusErrorStyle
)

// Task-specific styles (unique to task display)
var (
	taskIDStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(shared.ColourBlue500)

	taskTitleStyle = lipgloss.NewStyle().
			Foreground(shared.ColourGray200)

	taskLabelStyle = lipgloss.NewStyle().
			Foreground(shared.ColourViolet400)
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
