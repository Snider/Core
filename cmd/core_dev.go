//go:build !ci

// core_dev.go registers commands for the full development binary.
//
// Build with: go build (default)
//
// This is the default build variant with all development tools:
//   - dev: Multi-repo git workflows (commit, push, pull, sync)
//   - ai: AI agent task management
//   - go: Go module and build tools
//   - php: Laravel/Composer development tools
//   - build: Cross-platform compilation
//   - ci: Release publishing
//   - sdk: API compatibility checks
//   - pkg: Package management
//   - vm: LinuxKit VM management
//   - docs: Documentation generation
//   - setup: Repository cloning and setup
//   - doctor: Environment health checks
//   - test: Test runner with coverage

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

// registerCommands adds all development commands.
func registerCommands(app *clir.Cli) {
	// Multi-repo workflow
	dev.AddCommands(app)

	// AI agent tools
	ai.AddCommands(app)

	// Language tooling
	gocmd.AddCommands(app)
	php.AddCommands(app)

	// Build and release
	build.AddCommands(app)
	ci.AddCommands(app)
	sdk.AddCommands(app)

	// Environment management
	pkg.AddCommands(app)
	vm.AddCommands(app)
	docs.AddCommands(app)
	setup.AddCommands(app)
	doctor.AddCommands(app)
	testcmd.AddCommands(app)
}
