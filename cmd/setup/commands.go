// Package setup provides workspace bootstrap and package cloning commands.
//
// Two modes of operation:
//
// REGISTRY MODE (repos.yaml exists):
//   - Clones all repositories defined in repos.yaml into packages/
//   - Skips repos that already exist
//   - Supports filtering by type with --only
//
// BOOTSTRAP MODE (no repos.yaml):
//   - Clones core-devops to set up the workspace foundation
//   - Presents an interactive wizard to select packages (unless --all)
//   - Clones selected packages
//
// Flags:
//   - --registry: Path to repos.yaml (auto-detected if not specified)
//   - --only: Filter by repo type (foundation, module, product)
//   - --dry-run: Preview what would be cloned
//   - --all: Skip wizard, clone all packages (non-interactive)
//   - --name: Project directory name for bootstrap mode
//   - --build: Run build after cloning
//
// Uses gh CLI with HTTPS when authenticated, falls back to SSH.
package setup

import "github.com/spf13/cobra"

// AddCommands registers the 'setup' command and all subcommands.
func AddCommands(root *cobra.Command) {
	AddSetupCommand(root)
}
