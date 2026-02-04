//go:build !ci && !php && !minimal

// full.go imports all packages for the full development binary.
//
// Build with: go build (default)
//
// This is the default build variant with all development tools:
//   - dev: Multi-repo git workflows (commit, push, pull, sync)
//   - ai: AI agent task management + RAG + metrics
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
//   - qa: Quality assurance workflows
//   - monitor: Security monitoring aggregation

package variants

import (
	// Commands via self-registration
	_ "github.com/host-uk/core/internal/cmd/ai"
	_ "github.com/host-uk/core/internal/cmd/ci"
	_ "github.com/host-uk/core/internal/cmd/deploy"
	_ "github.com/host-uk/core/internal/cmd/dev"
	_ "github.com/host-uk/core/internal/cmd/docs"
	_ "github.com/host-uk/core/internal/cmd/doctor"
	_ "github.com/host-uk/core/internal/cmd/gitcmd"
	_ "github.com/host-uk/core/internal/cmd/go"
	_ "github.com/host-uk/core/internal/cmd/help"
	_ "github.com/host-uk/core/internal/cmd/monitor"
	_ "github.com/host-uk/core/internal/cmd/php"
	_ "github.com/host-uk/core/internal/cmd/pkgcmd"
	_ "github.com/host-uk/core/internal/cmd/qa"
	_ "github.com/host-uk/core/internal/cmd/sdk"
	_ "github.com/host-uk/core/internal/cmd/security"
	_ "github.com/host-uk/core/internal/cmd/setup"
	_ "github.com/host-uk/core/internal/cmd/test"
	_ "github.com/host-uk/core/internal/cmd/updater"
	_ "github.com/host-uk/core/internal/cmd/vm"
	_ "github.com/host-uk/core/internal/cmd/workspace"
	_ "github.com/host-uk/core/pkg/build/buildcmd"
)
