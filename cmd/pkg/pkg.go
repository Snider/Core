// Package pkg provides package management commands for core-* repos.
package pkg

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
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
func AddPkgCommands(root *cobra.Command) {
	pkgCmd := &cobra.Command{
		Use:   "pkg",
		Short: "Package management for core-* repos",
		Long: "Manage host-uk/core-* packages and repositories.\n\n" +
			"Commands:\n" +
			"  search    Search GitHub for packages\n" +
			"  install   Clone a package from GitHub\n" +
			"  list      List installed packages\n" +
			"  update    Update installed packages\n" +
			"  outdated  Check for outdated packages",
	}

	root.AddCommand(pkgCmd)
	addPkgSearchCommand(pkgCmd)
	addPkgInstallCommand(pkgCmd)
	addPkgListCommand(pkgCmd)
	addPkgUpdateCommand(pkgCmd)
	addPkgOutdatedCommand(pkgCmd)
}
