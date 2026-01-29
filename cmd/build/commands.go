// Package build provides project build commands with auto-detection.
//
// Supports building:
//   - Go projects (standard and cross-compilation)
//   - Wails desktop applications
//   - Docker images
//   - LinuxKit VM images
//   - Taskfile-based projects
//
// Configuration via .core/build.yaml or command-line flags.
package build

import "github.com/leaanthony/clir"

// AddCommands registers the 'build' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddBuildCommand(app)
}
