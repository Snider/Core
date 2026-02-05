// Package gitea provides CLI commands for managing a Gitea instance.
//
// Commands:
//   - config: Configure Gitea connection (URL, token)
//   - repos: List repositories
//   - issues: List and create issues
//   - prs: List pull requests
//   - mirror: Create GitHub-to-Gitea mirrors
//   - sync: Sync GitHub repos to Gitea upstream branches
package gitea

import (
	"github.com/host-uk/core/pkg/cli"
)

func init() {
	cli.RegisterCommands(AddGiteaCommands)
}

// Style aliases from shared package.
var (
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
	warningStyle = cli.WarningStyle
	dimStyle     = cli.DimStyle
	valueStyle   = cli.ValueStyle
	repoStyle    = cli.RepoStyle
	numberStyle  = cli.NumberStyle
	infoStyle    = cli.InfoStyle
)

// AddGiteaCommands registers the 'gitea' command and all subcommands.
func AddGiteaCommands(root *cli.Command) {
	giteaCmd := &cli.Command{
		Use:   "gitea",
		Short: "Gitea instance management",
		Long:  "Manage repositories, issues, and pull requests on your Gitea instance.",
	}
	root.AddCommand(giteaCmd)

	addConfigCommand(giteaCmd)
	addReposCommand(giteaCmd)
	addIssuesCommand(giteaCmd)
	addPRsCommand(giteaCmd)
	addMirrorCommand(giteaCmd)
	addSyncCommand(giteaCmd)
}
