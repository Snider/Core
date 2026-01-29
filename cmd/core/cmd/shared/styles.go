package shared

import "github.com/charmbracelet/lipgloss"

// Common styles used across multiple command packages.
var (
	RepoNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3b82f6"))

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22c55e"))

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ef4444"))

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f59e0b"))

	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280"))

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0"))

	LinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")).
			Underline(true)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#e2e8f0"))
)
