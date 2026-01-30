// Package pkg provides package management commands for core-* repos.
package pkg

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style and utility aliases
var (
	repoNameStyle   = shared.RepoNameStyle
	successStyle    = shared.SuccessStyle
	errorStyle      = shared.ErrorStyle
	dimStyle        = shared.DimStyle
	ghAuthenticated = shared.GhAuthenticated
	gitClone        = shared.GitClone
)

// AddPkgCommands adds the 'pkg' command and subcommands for package management.
func AddPkgCommands(parent *clir.Cli) {
	pkgCmd := parent.NewSubCommand("pkg", "Package management for core-* repos")
	pkgCmd.LongDescription("Manage host-uk/core-* packages and repositories.\n\n" +
		"Commands:\n" +
		"  search    Search GitHub for packages\n" +
		"  install   Clone a package from GitHub\n" +
		"  list      List installed packages\n" +
		"  update    Update installed packages\n" +
		"  outdated  Check for outdated packages")

	addPkgSearchCommand(pkgCmd)
	addPkgInstallCommand(pkgCmd)
	addPkgListCommand(pkgCmd)
	addPkgUpdateCommand(pkgCmd)
	addPkgOutdatedCommand(pkgCmd)
}
