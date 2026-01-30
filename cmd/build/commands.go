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
//
// Subcommands:
//   - build: Auto-detect and build the current project
//   - build from-path: Build from a local static web app directory
//   - build pwa: Build from a live PWA URL
//   - build sdk: Generate API SDKs from OpenAPI spec
package build

import "github.com/leaanthony/clir"

// AddCommands registers the 'build' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddBuildCommand(app)
}
