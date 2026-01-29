// Package dev provides multi-repo development workflow commands.
//
// This package manages git operations across multiple repositories defined in
// repos.yaml. It also provides GitHub integration and dev environment management.
//
// Commands:
//   - work: Combined status, commit, and push workflow
//   - health: Quick health check across all repos
//   - commit: Claude-assisted commit message generation
//   - push: Push repos with unpushed commits
//   - pull: Pull repos that are behind remote
//   - sync: Sync all repos with remote (pull + push)
//   - issues: List GitHub issues across repos
//   - reviews: List PRs needing review
//   - ci: Check GitHub Actions CI status
//   - impact: Analyse dependency impact of changes
//   - install/boot/stop: Dev environment VM management
package dev

import "github.com/leaanthony/clir"

// AddCommands registers the 'dev' command and all subcommands.
func AddCommands(app *clir.Cli) {
	devCmd := app.NewSubCommand("dev", "Multi-repo development workflow")
	devCmd.LongDescription("Manage multiple git repositories and GitHub integration.\n\n" +
		"Uses repos.yaml to discover repositories. Falls back to scanning\n" +
		"the current directory if no registry is found.\n\n" +
		"Git Operations:\n" +
		"  work      Combined status → commit → push workflow\n" +
		"  health    Quick repo health summary\n" +
		"  commit    Claude-assisted commit messages\n" +
		"  push      Push repos with unpushed commits\n" +
		"  pull      Pull repos behind remote\n" +
		"  sync      Sync all repos (pull + push)\n\n" +
		"GitHub Integration (requires gh CLI):\n" +
		"  issues    List open issues across repos\n" +
		"  reviews   List PRs awaiting review\n" +
		"  ci        Check GitHub Actions status\n" +
		"  impact    Analyse dependency impact\n\n" +
		"Dev Environment:\n" +
		"  install   Download dev environment image\n" +
		"  boot      Start dev environment VM\n" +
		"  stop      Stop dev environment VM\n" +
		"  shell     Open shell in dev VM\n" +
		"  status    Check dev VM status")

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
