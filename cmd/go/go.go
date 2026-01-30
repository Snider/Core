// Package gocmd provides Go development commands.
//
// Note: Package named gocmd because 'go' is a reserved keyword.
package gocmd

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Style aliases for shared styles
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
)

// AddGoCommands adds Go development commands.
func AddGoCommands(root *cobra.Command) {
	goCmd := &cobra.Command{
		Use:   "go",
		Short: i18n.T("cmd.go.short"),
		Long:  i18n.T("cmd.go.long"),
	}

	root.AddCommand(goCmd)
	addGoTestCommand(goCmd)
	addGoCovCommand(goCmd)
	addGoFmtCommand(goCmd)
	addGoLintCommand(goCmd)
	addGoInstallCommand(goCmd)
	addGoModCommand(goCmd)
	addGoWorkCommand(goCmd)
}
