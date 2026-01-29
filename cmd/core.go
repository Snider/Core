// Package cmd implements the core CLI application.
//
// The CLI provides commands for:
//   - Multi-repo development workflows (dev)
//   - AI agent task management (ai)
//   - Go and PHP development tools (go, php)
//   - Build and release automation (build, ci)
//   - SDK validation and API compatibility (sdk)
//   - Package and environment management (pkg, vm)
//   - Documentation and testing (docs, test)
//   - Environment health checks (doctor)
//   - Repository setup and cloning (setup)
//
// Two build variants exist:
//   - Default build: Full development toolset
//   - CI build (-tags ci): Minimal release toolset
package cmd

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

// Terminal styles using Tailwind color palette.
var (
	// coreStyle is used for primary headings and the CLI name.
	coreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")). // blue-500
			Bold(true)

	// subPkgStyle is used for subcommand names and secondary headings.
	subPkgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")). // gray-200
			Bold(true)

	// linkStyle is used for URLs and clickable references.
	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")). // blue-500
			Underline(true)
)

// Execute initialises and runs the CLI application.
// Commands are registered based on build tags (see core_ci.go and core_dev.go).
func Execute() error {
	app := clir.NewCli("core", "CLI tool for development and production", "0.1.0")
	registerCommands(app)
	return app.Run()
}
