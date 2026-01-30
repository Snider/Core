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
	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
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

// Table styles for status display (extends shared styles with cell padding)
var (
	dirtyStyle = shared.GitDirtyStyle.Padding(0, 1)
	aheadStyle = shared.GitAheadStyle.Padding(0, 1)
	cleanStyle = shared.GitCleanStyle.Padding(0, 1)
)

// AddCommands registers the 'dev' command and all subcommands.
func AddCommands(root *cobra.Command) {
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Multi-repo development workflow",
		Long: `Manage multiple git repositories and GitHub integration.

Uses repos.yaml to discover repositories. Falls back to scanning
the current directory if no registry is found.

Git Operations:
  work      Combined status -> commit -> push workflow
  health    Quick repo health summary
  commit    Claude-assisted commit messages
  push      Push repos with unpushed commits
  pull      Pull repos behind remote

GitHub Integration (requires gh CLI):
  issues    List open issues across repos
  reviews   List PRs awaiting review
  ci        Check GitHub Actions status
  impact    Analyse dependency impact

Dev Environment:
  install   Download dev environment image
  boot      Start dev environment VM
  stop      Stop dev environment VM
  shell     Open shell in dev VM
  status    Check dev VM status`,
	}
	root.AddCommand(devCmd)

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
