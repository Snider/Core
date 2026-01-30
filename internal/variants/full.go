//go:build !ci && !php && !minimal

// full.go imports all packages for the full development binary.
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

package variants

import (
	// Commands via self-registration
	_ "github.com/host-uk/core/pkg/ai"
	_ "github.com/host-uk/core/pkg/build/buildcmd"
	_ "github.com/host-uk/core/pkg/ci"
	_ "github.com/host-uk/core/pkg/dev"
	_ "github.com/host-uk/core/pkg/docs"
	_ "github.com/host-uk/core/pkg/doctor"
	_ "github.com/host-uk/core/pkg/go"
	_ "github.com/host-uk/core/pkg/php"
	_ "github.com/host-uk/core/pkg/pkgcmd"
	_ "github.com/host-uk/core/pkg/sdk"
	_ "github.com/host-uk/core/pkg/setup"
	_ "github.com/host-uk/core/pkg/test"
	_ "github.com/host-uk/core/pkg/vm"
)
