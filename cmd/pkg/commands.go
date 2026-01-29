// Package pkg provides GitHub package management for host-uk repositories.
//
// Commands:
//   - search: Search GitHub org for repos (cached for 1 hour)
//   - install: Clone a repo from GitHub to packages/
//   - list: List installed packages from repos.yaml
//   - update: Pull latest changes for packages
//   - outdated: Check which packages have unpulled commits
//
// Uses gh CLI for authenticated GitHub access. Results are cached in
// .core/cache/ within the workspace directory.
package pkg

import "github.com/leaanthony/clir"

// AddCommands registers the 'pkg' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddPkgCommands(app)
}
