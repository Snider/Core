//go:build minimal

// minimal.go imports only core packages for a minimal binary.
//
// Build with: go build -tags minimal
//
// This variant includes only the absolute essentials:
//   - doctor: Environment verification
//
// Use this for the smallest possible binary with just health checks.

package variants

import (
	// Commands via self-registration
	_ "github.com/host-uk/core/pkg/doctor"
)
