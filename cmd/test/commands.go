// Package testcmd provides Go test running commands with enhanced output.
//
// Note: Package named testcmd to avoid conflict with Go's test package.
//
// Features:
//   - Colour-coded pass/fail/skip output
//   - Per-package coverage breakdown with --coverage
//   - JSON output for CI/agents with --json
//   - Filters linker warnings on macOS
//
// Flags: --verbose, --coverage, --short, --pkg, --run, --race, --json
package testcmd

import "github.com/leaanthony/clir"

// AddCommands registers the 'test' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddTestCommand(app)
}
