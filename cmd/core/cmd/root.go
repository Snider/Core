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
	devCmd := app.NewSubCommand("dev", "Multi-repo development workflow")
	devCmd.LongDescription("Multi-repo git operations and GitHub integration.\n\n" +
		"Commands:\n" +
		"  work      Multi-repo status, commit, push workflow\n" +
		"  health    Quick health check across repos\n" +
		"  commit    Claude-assisted commits\n" +
		"  push      Push repos with unpushed commits\n" +
		"  pull      Pull repos that are behind\n" +
		"  issues    List open issues across repos\n" +
		"  reviews   List PRs needing review\n" +
		"  ci        Check CI status\n" +
		"  impact    Show dependency impact")

	// Git/multi-repo commands under dev
	AddWorkCommand(devCmd)
	AddHealthCommand(devCmd)
	AddCommitCommand(devCmd)
	AddPushCommand(devCmd)
	AddPullCommand(devCmd)
	AddIssuesCommand(devCmd)
	AddReviewsCommand(devCmd)
	AddCICommand(devCmd)
	AddImpactCommand(devCmd)

	// Internal dev tools (API, sync, agentic)
	AddAPICommands(devCmd)
	AddSyncCommand(devCmd)
	AddAgenticCommands(devCmd)
	AddDevCommand(devCmd)

	// Top-level commands
	AddBuildCommand(app)
	AddDocsCommand(app)
	AddSetupCommand(app)
	AddDoctorCommand(app)
	AddPkgCommands(app)
	AddReleaseCommand(app)
	AddContainerCommands(app)
	AddGoCommands(app)
	AddPHPCommands(app)
	AddSDKCommand(app)
	AddTestCommand(app)

	return app.Run()
}
