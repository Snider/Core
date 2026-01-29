// Package doctor provides environment validation commands.
//
// Checks for:
//   - Required tools: git, gh, php, composer, node
//   - Optional tools: pnpm, claude, docker
//   - GitHub access: SSH keys and CLI authentication
//   - Workspace: repos.yaml presence and clone status
//
// Run before 'core setup' to ensure your environment is ready.
// Provides platform-specific installation instructions for missing tools.
package doctor

import "github.com/leaanthony/clir"

// AddCommands registers the 'doctor' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddDoctorCommand(app)
}
