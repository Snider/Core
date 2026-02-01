// Package gitcmd provides git workflow commands as a root-level command.
//
// Git Operations:
//   - health: Show status across repos
//   - commit: Claude-assisted commit message generation
//   - push: Push repos with unpushed commits
//   - pull: Pull repos that are behind remote
//   - work: Combined status, commit, and push workflow
//
// Safe Operations (for AI agents):
//   - file-sync: Sync files across repos with auto commit/push
//   - apply: Run command across repos with auto commit/push
package gitcmd

import (
	"github.com/host-uk/core/internal/cmd/dev"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

func init() {
	cli.RegisterCommands(AddGitCommands)
}

// AddGitCommands registers the 'git' command and all subcommands.
func AddGitCommands(root *cli.Command) {
	gitCmd := &cli.Command{
		Use:   "git",
		Short: i18n.T("cmd.git.short"),
		Long:  i18n.T("cmd.git.long"),
	}
	root.AddCommand(gitCmd)

	// Import git commands from dev package
	dev.AddHealthCommand(gitCmd) // Shows repo status
	dev.AddCommitCommand(gitCmd)
	dev.AddPushCommand(gitCmd)
	dev.AddPullCommand(gitCmd)
	dev.AddWorkCommand(gitCmd)

	// Safe operations for AI agents
	dev.AddFileSyncCommand(gitCmd)
	dev.AddApplyCommand(gitCmd)
}
