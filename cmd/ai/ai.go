// ai.go defines styles and the AddAgenticCommands function for AI task management.

package ai

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style aliases from shared package
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
	truncate     = shared.Truncate
	formatAge    = shared.FormatAge
)

// Task-specific styles
var (
	taskIDStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3b82f6")) // blue-500

	taskTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")) // gray-200

	taskPriorityHighStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ef4444")) // red-500

	taskPriorityMediumStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f59e0b")) // amber-500

	taskPriorityLowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22c55e")) // green-500

	taskStatusPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6b7280")) // gray-500

	taskStatusInProgressStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#3b82f6")) // blue-500

	taskStatusCompletedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#22c55e")) // green-500

	taskStatusBlockedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ef4444")) // red-500

	taskLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a78bfa")) // violet-400
)

// AddAgenticCommands adds the agentic task management commands to the ai command.
func AddAgenticCommands(parent *clir.Command) {
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
