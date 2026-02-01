//go:build ci

// ci.go imports packages for the minimal CI/release binary.
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

package variants

import (
	// Commands via self-registration
	_ "github.com/host-uk/core/internal/cmd/ci"
	_ "github.com/host-uk/core/internal/cmd/doctor"
	_ "github.com/host-uk/core/internal/cmd/sdk"
	_ "github.com/host-uk/core/pkg/build/buildcmd"
)
