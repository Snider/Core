// Package gocmd provides Go development commands.
//
// Note: Package named gocmd because 'go' is a reserved keyword.
package gocmd

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style aliases for shared styles
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
)

// AddGoCommands adds Go development commands.
func AddGoCommands(parent *clir.Cli) {
	goCmd := parent.NewSubCommand("go", "Go development tools")
	goCmd.LongDescription("Go development tools with enhanced output and environment setup.\n\n" +
		"Commands:\n" +
		"  test     Run tests\n" +
		"  cov      Run tests with coverage report\n" +
		"  fmt      Format Go code\n" +
		"  lint     Run golangci-lint\n" +
		"  install  Install Go binary\n" +
		"  mod      Module management (tidy, download, verify)\n" +
		"  work     Workspace management")

	addGoTestCommand(goCmd)
	addGoCovCommand(goCmd)
	addGoFmtCommand(goCmd)
	addGoLintCommand(goCmd)
	addGoInstallCommand(goCmd)
	addGoModCommand(goCmd)
	addGoWorkCommand(goCmd)
}
