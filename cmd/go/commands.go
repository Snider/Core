// Package gocmd provides Go development commands with enhanced output.
//
// Note: Package named gocmd because 'go' is a reserved keyword.
//
// Commands:
//   - test: Run tests with colour-coded coverage summary
//   - cov: Run tests with detailed coverage reports (HTML, thresholds)
//   - fmt: Format code using goimports or gofmt
//   - lint: Run golangci-lint
//   - install: Install binary to $GOPATH/bin
//   - mod: Module management (tidy, download, verify, graph)
//   - work: Workspace management (sync, init, use)
//
// Sets MACOSX_DEPLOYMENT_TARGET to suppress linker warnings on macOS.
package gocmd

import "github.com/spf13/cobra"

// AddCommands registers the 'go' command and all subcommands.
func AddCommands(root *cobra.Command) {
	AddGoCommands(root)
}
