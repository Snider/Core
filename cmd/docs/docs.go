// Package docs provides documentation management commands.
package docs

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style and utility aliases from shared
var (
	repoNameStyle = shared.RepoNameStyle
	successStyle  = shared.SuccessStyle
	errorStyle    = shared.ErrorStyle
	dimStyle      = shared.DimStyle
	headerStyle   = shared.HeaderStyle
	confirm       = shared.Confirm
)

// Package-specific styles
var (
	docsFoundStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22c55e")) // green-500

	docsMissingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6b7280")) // gray-500

	docsFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")) // blue-500
)

// AddDocsCommand adds the 'docs' command to the given parent command.
func AddDocsCommand(parent *clir.Cli) {
	docsCmd := parent.NewSubCommand("docs", "Documentation management")
	docsCmd.LongDescription("Manage documentation across all repos.\n" +
		"Scan for docs, check coverage, and sync to core-php/docs/packages/.")

	// Add subcommands
	addDocsSyncCommand(docsCmd)
	addDocsListCommand(docsCmd)
}
