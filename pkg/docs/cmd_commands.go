// Package docs provides documentation management commands for multi-repo workspaces.
//
// Commands:
//   - list: Scan repos for README.md, CLAUDE.md, CHANGELOG.md, docs/
//   - sync: Copy docs/ files from all repos to core-php/docs/packages/
//
// Works with repos.yaml to discover repositories and sync documentation
// to a central location for unified documentation builds.
package docs

import "github.com/host-uk/core/pkg/cli"

func init() {
	cli.RegisterCommands(AddDocsCommands)
}

// AddDocsCommands registers the 'docs' command and all subcommands.
func AddDocsCommands(root *cli.Command) {
	root.AddCommand(docsCmd)
}
