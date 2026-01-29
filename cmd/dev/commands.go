// Package dev provides multi-repo development workflow commands.
//
// Commands include git operations (work, commit, push, pull), GitHub
// integration (issues, reviews, ci, impact), and dev environment management.
package dev

import "github.com/leaanthony/clir"

// AddCommands registers the 'dev' command and all subcommands.
func AddCommands(app *clir.Cli) {
	devCmd := app.NewSubCommand("dev", "Multi-repo development workflow")
	devCmd.LongDescription("Multi-repo git operations and GitHub integration.\n\n" +
		"Git Operations:\n" +
		"  work      Multi-repo status, commit, push workflow\n" +
		"  health    Quick health check across repos\n" +
		"  commit    Claude-assisted commits\n" +
		"  push      Push repos with unpushed commits\n" +
		"  pull      Pull repos that are behind\n\n" +
		"GitHub Integration:\n" +
		"  issues    List open issues across repos\n" +
		"  reviews   List PRs needing review\n" +
		"  ci        Check CI status\n" +
		"  impact    Show dependency impact")

	// Git operations
	AddWorkCommand(devCmd)
	AddHealthCommand(devCmd)
	AddCommitCommand(devCmd)
	AddPushCommand(devCmd)
	AddPullCommand(devCmd)

	// GitHub integration
	AddIssuesCommand(devCmd)
	AddReviewsCommand(devCmd)
	AddCICommand(devCmd)
	AddImpactCommand(devCmd)

	// API tools
	AddAPICommands(devCmd)

	// Dev environment
	AddDevCommand(devCmd)
}
