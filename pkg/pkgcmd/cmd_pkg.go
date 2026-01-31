// Package pkgcmd provides package management commands for core-* repos.
package pkgcmd

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

func init() {
	cli.RegisterCommands(AddPkgCommands)
}

// Style and utility aliases
var (
		repoNameStyle = cli.RepoStyle
	successStyle    = cli.SuccessStyle
	errorStyle      = cli.ErrorStyle
	dimStyle        = cli.DimStyle
	ghAuthenticated = cli.GhAuthenticated
	gitClone        = cli.GitClone
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
