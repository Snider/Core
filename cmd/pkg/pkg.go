// Package pkg provides package management commands for core-* repos.
package pkg

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
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
		Short: i18n.T("cmd.pkg.short"),
		Long:  i18n.T("cmd.pkg.long"),
	}

	root.AddCommand(pkgCmd)
	addPkgSearchCommand(pkgCmd)
	addPkgInstallCommand(pkgCmd)
	addPkgListCommand(pkgCmd)
	addPkgUpdateCommand(pkgCmd)
	addPkgOutdatedCommand(pkgCmd)
}
