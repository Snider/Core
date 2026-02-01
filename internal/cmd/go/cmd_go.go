// Package gocmd provides Go development commands.
//
// Note: Package named gocmd because 'go' is a reserved keyword.
package gocmd

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

// Style aliases for shared styles
var (
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
	dimStyle     = cli.DimStyle
)

// AddGoCommands adds Go development commands.
func AddGoCommands(root *cli.Command) {
	goCmd := &cli.Command{
		Use:   "go",
		Short: i18n.T("cmd.go.short"),
		Long:  i18n.T("cmd.go.long"),
	}

	root.AddCommand(goCmd)
	addGoQACommand(goCmd)
	addGoTestCommand(goCmd)
	addGoCovCommand(goCmd)
	addGoFmtCommand(goCmd)
	addGoLintCommand(goCmd)
	addGoInstallCommand(goCmd)
	addGoModCommand(goCmd)
	addGoWorkCommand(goCmd)
}
