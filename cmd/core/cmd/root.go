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
	app := clir.NewCli("core", "CLI for Go/PHP development, multi-repo management, and deployment", "0.1.0")

	// Add the top-level commands
	devCmd := app.NewSubCommand("dev", "Development tools for Core Framework")
	AddAPICommands(devCmd)
	AddSyncCommand(devCmd)
	AddAgenticCommands(devCmd)
	AddDevCommand(devCmd)
	AddBuildCommand(app)
	AddTviewCommand(app)
	AddWorkCommand(app)
	AddHealthCommand(app)
	AddIssuesCommand(app)
	AddReviewsCommand(app)
	AddCommitCommand(app)
	AddPushCommand(app)
	AddPullCommand(app)
	AddImpactCommand(app)
	AddDocsCommand(app)
	AddCICommand(app)
	AddSetupCommand(app)
	AddDoctorCommand(app)
	AddSearchCommand(app)
	AddInstallCommand(app)
	AddReleaseCommand(app)
	AddContainerCommands(app)
	AddTemplatesCommand(app)
	AddPHPCommands(app)
	AddSDKCommand(app)
	AddTestCommand(app)

	return app.Run()
}
