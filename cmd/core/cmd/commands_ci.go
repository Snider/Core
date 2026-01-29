//go:build ci

package cmd

import "github.com/leaanthony/clir"

// registerCommands adds only CI/release commands for the minimal binary.
// Build with: go build -tags ci
func registerCommands(app *clir.Cli) {
	// CI/Release commands only - minimal attack surface
	AddBuildCommand(app)
	AddCIReleaseCommand(app)
	AddSDKCommand(app)

	// Doctor for environment verification
	AddDoctorCommand(app)
}
