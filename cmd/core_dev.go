//go:build !ci

package cmd

import (
	"github.com/host-uk/core/cmd/ai"
	"github.com/host-uk/core/cmd/build"
	"github.com/host-uk/core/cmd/ci"
	"github.com/host-uk/core/cmd/dev"
	"github.com/host-uk/core/cmd/docs"
	"github.com/host-uk/core/cmd/doctor"
	gocmd "github.com/host-uk/core/cmd/go"
	"github.com/host-uk/core/cmd/php"
	"github.com/host-uk/core/cmd/pkg"
	"github.com/host-uk/core/cmd/sdk"
	"github.com/host-uk/core/cmd/setup"
	testcmd "github.com/host-uk/core/cmd/test"
	"github.com/host-uk/core/cmd/vm"
	"github.com/leaanthony/clir"
)

// registerCommands adds all commands for the full development binary.
// Build with: go build (default) or go build -tags dev
func registerCommands(app *clir.Cli) {
	// Dev workflow commands
	dev.AddCommands(app)

	// AI/Agent commands
	ai.AddCommands(app)

	// Language-specific development tools
	gocmd.AddCommands(app)
	php.AddCommands(app)

	// CI/Release commands (also available in ci build)
	build.AddCommands(app)
	ci.AddCommands(app)
	sdk.AddCommands(app)

	// Package/environment management (dev only)
	pkg.AddCommands(app)
	vm.AddCommands(app)
	docs.AddCommands(app)
	setup.AddCommands(app)
	doctor.AddCommands(app)
	testcmd.AddCommands(app)
}
