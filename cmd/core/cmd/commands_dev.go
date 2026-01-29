//go:build !ci

package cmd

import "github.com/leaanthony/clir"

// registerCommands adds all commands for the full development binary.
// Build with: go build (default) or go build -tags dev
func registerCommands(app *clir.Cli) {
	// Dev workflow commands (multi-repo git operations)
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

	AddWorkCommand(devCmd)
	AddHealthCommand(devCmd)
	AddCommitCommand(devCmd)
	AddPushCommand(devCmd)
	AddPullCommand(devCmd)
	AddIssuesCommand(devCmd)
	AddReviewsCommand(devCmd)
	AddCICommand(devCmd)
	AddImpactCommand(devCmd)
	AddAPICommands(devCmd)
	AddSyncCommand(devCmd)
	AddAgenticCommands(devCmd)
	AddDevCommand(devCmd)

	// Language-specific development tools
	AddGoCommands(app)
	AddPHPCommands(app)

	// CI/Release commands (also available in ci build)
	AddBuildCommand(app)
	AddCIReleaseCommand(app)
	AddSDKCommand(app)

	// Package/environment management (dev only)
	AddPkgCommands(app)
	AddContainerCommands(app)
	AddDocsCommand(app)
	AddSetupCommand(app)
	AddDoctorCommand(app)
	AddTestCommand(app)
}
