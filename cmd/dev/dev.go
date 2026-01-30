// Package dev provides multi-repo development workflow commands.
//
// Git Operations:
//   - work: Combined status, commit, and push workflow
//   - health: Quick health check across all repos
//   - commit: Claude-assisted commit message generation
//   - push: Push repos with unpushed commits
//   - pull: Pull repos that are behind remote
//
// GitHub Integration (requires gh CLI):
//   - issues: List open issues across repos
//   - reviews: List PRs needing review
//   - ci: Check GitHub Actions CI status
//   - impact: Analyse dependency impact of changes
//
// API Tools:
//   - api sync: Synchronize public service APIs
//
// Dev Environment (VM management):
//   - install: Download dev environment image
//   - boot: Start dev environment VM
//   - stop: Stop dev environment VM
//   - status: Check dev VM status
//   - shell: Open shell in dev VM
//   - serve: Mount project and start dev server
//   - test: Run tests in dev environment
//   - claude: Start sandboxed Claude session
//   - update: Check for and apply updates
package dev

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style aliases from shared package
var (
	successStyle  = shared.SuccessStyle
	errorStyle    = shared.ErrorStyle
	warningStyle  = shared.WarningStyle
	dimStyle      = shared.DimStyle
	valueStyle    = shared.ValueStyle
	headerStyle   = shared.HeaderStyle
	repoNameStyle = shared.RepoNameStyle
)

// Table styles for status display
var (
	cellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	dirtyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")). // red-500
			Padding(0, 1)

	aheadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22c55e")). // green-500
			Padding(0, 1)

	cleanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")). // gray-500
			Padding(0, 1)
)

// AddCommands registers the 'dev' command and all subcommands.
func AddCommands(app *clir.Cli) {
	devCmd := app.NewSubCommand("dev", "Multi-repo development workflow")
	devCmd.LongDescription("Manage multiple git repositories and GitHub integration.\n\n" +
		"Uses repos.yaml to discover repositories. Falls back to scanning\n" +
		"the current directory if no registry is found.\n\n" +
		"Git Operations:\n" +
		"  work      Combined status -> commit -> push workflow\n" +
		"  health    Quick repo health summary\n" +
		"  commit    Claude-assisted commit messages\n" +
		"  push      Push repos with unpushed commits\n" +
		"  pull      Pull repos behind remote\n\n" +
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
	addWorkCommand(devCmd)
	addHealthCommand(devCmd)
	addCommitCommand(devCmd)
	addPushCommand(devCmd)
	addPullCommand(devCmd)

	// GitHub integration
	addIssuesCommand(devCmd)
	addReviewsCommand(devCmd)
	addCICommand(devCmd)
	addImpactCommand(devCmd)

	// API tools
	addAPICommands(devCmd)

	// Dev environment
	addVMCommands(devCmd)
}
