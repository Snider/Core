// Package setup provides workspace initialisation commands.
//
// Clones all repositories defined in repos.yaml into the workspace.
// Skips repos that already exist. Supports filtering by type.
//
// Flags:
//   - --registry: Path to repos.yaml (auto-detected if not specified)
//   - --only: Filter by repo type (foundation, module, product)
//   - --dry-run: Preview what would be cloned
//
// Uses gh CLI with HTTPS when authenticated, falls back to SSH.
package setup

import "github.com/leaanthony/clir"

// AddCommands registers the 'setup' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddSetupCommand(app)
}
