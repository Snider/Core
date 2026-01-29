package cmd

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

// Define some global lipgloss styles for a Tailwind dark theme
var (
	coreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")). // Tailwind blue-500
			Bold(true)

	subPkgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")). // Tailwind gray-200
			Bold(true)

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")). // Tailwind blue-500
			Underline(true)
)

// Execute creates the root CLI application and runs it.
func Execute() error {
	app := clir.NewCli("core", "CLI tool for development and production", "0.1.0")

	// Register commands based on build tags
	registerCommands(app)

	return app.Run()
}
