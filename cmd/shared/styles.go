// Package shared provides common utilities and styles for CLI commands.
//
// This package contains:
//   - Terminal styling using lipgloss with Tailwind colours
//   - Common helper functions (truncation, confirmation prompts)
//   - Git and GitHub CLI utilities
package shared

import "github.com/charmbracelet/lipgloss"

// Terminal styles using Tailwind colour palette.
// These are shared across command packages for consistent output.
var (
	// RepoNameStyle highlights repository names (blue, bold).
	RepoNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3b82f6")) // blue-500

	// SuccessStyle indicates successful operations (green, bold).
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22c55e")) // green-500

	// ErrorStyle indicates errors and failures (red, bold).
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ef4444")) // red-500

	// WarningStyle indicates warnings and cautions (amber, bold).
	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f59e0b")) // amber-500

	// DimStyle for secondary/muted text (gray).
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500

	// ValueStyle for data values and output (light gray).
	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")) // gray-200

	// LinkStyle for URLs and clickable references (blue, underlined).
	LinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")). // blue-500
			Underline(true)

	// HeaderStyle for section headers (light gray, bold).
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#e2e8f0")) // gray-200
)
