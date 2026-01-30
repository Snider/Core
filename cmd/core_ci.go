//go:build ci

// core_ci.go registers commands for the minimal CI/release binary.
//
// Build with: go build -tags ci
//
// This variant includes only commands needed for CI pipelines:
//   - build: Cross-platform compilation
//   - ci: Release publishing
//   - sdk: API compatibility checks
//   - doctor: Environment verification
//
// Use this build to reduce binary size and attack surface in production.

package cmd

import (
	"github.com/host-uk/core/cmd/build"
	"github.com/host-uk/core/cmd/ci"
	"github.com/host-uk/core/cmd/doctor"
	"github.com/host-uk/core/cmd/sdk"
)

func init() {
	build.AddCommands(rootCmd)
	ci.AddCommands(rootCmd)
	sdk.AddCommands(rootCmd)
	doctor.AddCommands(rootCmd)
}
