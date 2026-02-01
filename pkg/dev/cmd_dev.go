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
// CI/Workflow Management:
//   - workflow list: Show table of repos vs workflows
//   - workflow sync: Copy workflow template to all repos
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
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

func init() {
	cli.RegisterCommands(AddDevCommands)
}

// Style aliases from shared package
var (
	successStyle  = cli.SuccessStyle
	errorStyle    = cli.ErrorStyle
	warningStyle  = cli.WarningStyle
	dimStyle      = cli.DimStyle
	valueStyle    = cli.ValueStyle
	headerStyle   = cli.HeaderStyle
	repoNameStyle = cli.RepoStyle
)

// Table styles for status display (extends shared styles with cell padding)
var (
	dirtyStyle = cli.NewStyle().Foreground(cli.ColourRed500)
	aheadStyle = cli.NewStyle().Foreground(cli.ColourAmber500)
	cleanStyle = cli.NewStyle().Foreground(cli.ColourGreen500)
)

// AddDevCommands registers the 'dev' command and all subcommands.
func AddDevCommands(root *cli.Command) {
	devCmd := &cli.Command{
		Use:   "dev",
		Short: i18n.T("cmd.dev.short"),
		Long:  i18n.T("cmd.dev.long"),
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

	// CI/Workflow management
	addWorkflowCommands(devCmd)

	// API tools
	addAPICommands(devCmd)

	// Dev environment
	addVMCommands(devCmd)
}
