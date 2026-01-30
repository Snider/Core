// Package gocmd provides Go development commands.
//
// Note: Package named gocmd because 'go' is a reserved keyword.
package gocmd

import (
	"github.com/host-uk/core/cmd/shared"
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
		Short: "Go development tools",
		Long: "Go development tools with enhanced output and environment setup.\n\n" +
			"Commands:\n" +
			"  test     Run tests\n" +
			"  cov      Run tests with coverage report\n" +
			"  fmt      Format Go code\n" +
			"  lint     Run golangci-lint\n" +
			"  install  Install Go binary\n" +
			"  mod      Module management (tidy, download, verify)\n" +
			"  work     Workspace management",
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
