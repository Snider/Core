//go:build ci

package cmd

import (
	"github.com/host-uk/core/cmd/core/cmd/build"
	"github.com/host-uk/core/cmd/core/cmd/ci"
	"github.com/host-uk/core/cmd/core/cmd/doctor"
	"github.com/host-uk/core/cmd/core/cmd/sdk"
	"github.com/leaanthony/clir"
)

// registerCommands adds only CI/release commands for the minimal binary.
// Build with: go build -tags ci
func registerCommands(app *clir.Cli) {
	// CI/Release commands only - minimal attack surface
	build.AddCommands(app)
	ci.AddCommands(app)
	sdk.AddCommands(app)

	// Doctor for environment verification
	doctor.AddCommands(app)
}
