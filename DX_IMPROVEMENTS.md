# Developer Experience (DX) Improvements

This document outlines identified gaps and opportunities for improving the developer experience in the Core repository. Each section represents a potential GitHub issue.

## Table of Contents

1. [Critical Issues](#critical-issues)
2. [High Priority](#high-priority)
3. [Medium Priority](#medium-priority)
4. [Low Priority / Nice to Have](#low-priority--nice-to-have)
5. [Summary Matrix](#summary-matrix)

---

## Critical Issues

### Issue #1: Missing CONTRIBUTING.md

**Priority:** 🔴 Critical  
**Effort:** ~30 minutes  
**Labels:** `documentation`, `good-first-issue`, `dx-improvement`

**Problem:**
No CONTRIBUTING.md file exists despite having a "For New Contributors" section in README. New developers don't know:
- PR conventions and review process
- Commit message format
- Code review expectations
- Testing requirements before submitting

**Impact:**
- Contributors waste time on incorrectly formatted PRs
- Maintainers spend time explaining basic contribution workflow
- Barriers to entry for new contributors

**Proposed Solution:**
Create `.github/CONTRIBUTING.md` with:

```markdown
# Contributing to Core

## Quick Start for Contributors

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/core.git
   cd core
   ```

2. **Set Up Development Environment**
   - See [Development Setup Guide](../docs/development-setup.md)
   - Install: Go 1.25+, Task, Node.js 18+

3. **TDD Workflow**
   ```bash
   task test-gen    # Generate test stubs
   task test        # Run tests (watch them fail)
   # Implement your feature
   task test        # Run tests (watch them pass)
   task review      # CodeRabbit review
   ```

4. **Before Submitting PR**
   - [ ] All tests pass: `task test`
   - [ ] Code formatted: `task fmt`
   - [ ] QA checks pass: `task qa`
   - [ ] Code reviewed: `task review`
   - [ ] Documentation updated if needed

## Code Standards

- Follow the [Service Registration Pattern](docs/architecture.md#service-registration-pattern)
- Use the dual-constructor pattern: `New()` for tests, `Register()` for runtime
- Error handling: Use `core.E()` for contextual errors
- Test naming: Use `_Good`, `_Bad`, `_Ugly` suffix pattern
- See [PACKAGE_STANDARDS.md](docs/PACKAGE_STANDARDS.md) for detailed guidelines

## Commit Messages

Use conventional commits format:
```
feat(pkg): add new feature
fix(cli): correct bug in command
docs: update README
test: add tests for service
```

## Pull Request Process

1. Create a feature branch: `git checkout -b feature/my-feature`
2. Make your changes following TDD workflow
3. Push to your fork
4. Open PR against `dev` branch
5. Address review feedback
6. Maintainers will merge when approved

## Code Review

- PRs require at least one approval
- CodeRabbit automatically reviews all PRs
- Address all review comments or explain why not
- Keep PRs focused and small when possible

## Getting Help

- 💬 Discord: http://discord.dappco.re
- 📖 Documentation: [docs/](docs/)
- ❓ Questions: Open a discussion or ask in Discord
```

**Acceptance Criteria:**
- [ ] `.github/CONTRIBUTING.md` file created
- [ ] Linked from main README.md
- [ ] Referenced in PR template (if exists)

---

### Issue #2: Missing Development Setup Guide

**Priority:** 🔴 Critical  
**Effort:** ~45 minutes  
**Labels:** `documentation`, `dx-improvement`, `onboarding`

**Problem:**
Developers cloning the repository don't have a clear guide on:
- How to install all required dependencies
- How to run the CLI vs GUI in development mode
- How to enable hot-reload during development
- What versions of tools are required
- How to verify their setup is correct

**Impact:**
- New contributors spend hours figuring out setup
- Inconsistent development environments lead to bugs
- High barrier to first contribution

**Proposed Solution:**
Create `docs/development-setup.md`:

```markdown
# Development Setup

Complete guide to setting up your development environment for Core.

## Prerequisites

### Required Tools

| Tool | Version | Check | Install |
|------|---------|-------|---------|
| Go | 1.25+ | `go version` | [go.dev](https://go.dev/dl/) |
| Node.js | 18+ | `node --version` | [nodejs.org](https://nodejs.org/) |
| Task | latest | `task --version` | `go install github.com/go-task/task/v3/cmd/task@latest` |
| Git | 2.30+ | `git --version` | [git-scm.com](https://git-scm.com/) |

### Optional Tools

| Tool | Purpose | Install |
|------|---------|---------|
| `gh` | GitHub CLI integration | [cli.github.com](https://cli.github.com/) |
| Wails | GUI development | `go install github.com/wailsapp/wails/v3/cmd/wails3@latest` |
| Docker | Container builds | [docker.com](https://www.docker.com/) |

## Initial Setup

### 1. Clone Repository

```bash
git clone https://github.com/host-uk/core.git
cd core
```

### 2. Install Dependencies

```bash
# Go dependencies
go mod download

# For GUI development
cd cmd/core-gui
npm install
cd ../..
```

### 3. Verify Setup

```bash
# Build CLI
task cli:build

# Verify it works
./bin/core --version
```

## Development Workflows

### CLI Development

```bash
# Build CLI
task cli:build

# Run directly
./bin/core doctor

# Install to system PATH (optional)
task cli:install
core doctor
```

### GUI Development (Wails)

```bash
# Development mode with hot-reload
task gui:dev

# Build production
task gui:build
```

**Hot Reload:**
- Frontend changes: Automatic reload in `task gui:dev`
- Backend (Go) changes: Restart `task gui:dev`

### Testing

```bash
# Run all tests
task test

# Run specific test
task test:run -- TestServiceRegistration

# With verbose output
task test:verbose

# With coverage
task cov
task cov-view  # Opens in browser
```

### Code Quality

```bash
# Format code
task fmt

# Run linter
task lint

# Full QA (fmt, vet, lint, test)
task qa

# Quick QA (skip tests)
task qa:quick
```

## Common Tasks

| Task | Command | Notes |
|------|---------|-------|
| Run tests | `task test` | All tests |
| Quick QA | `task qa:quick` | fmt, vet, lint only |
| Full QA | `task qa:full` | + race detector, vuln scan |
| Code review | `task review` | CodeRabbit review |
| Build CLI | `task cli:build` | Output: `./bin/core` |
| Run GUI dev | `task gui:dev` | Hot-reload enabled |

## Troubleshooting

### "command not found: task"

```bash
go install github.com/go-task/task/v3/cmd/task@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

### "wails3: command not found" (for GUI dev)

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

### Tests fail with "no such file"

Make sure you're in the repository root directory.

### GUI won't start

Check Node.js version:
```bash
node --version  # Should be 18+
```

Install GUI dependencies:
```bash
cd cmd/core-gui && npm install
```

## Next Steps

- Read [Contributing Guidelines](../.github/CONTRIBUTING.md)
- Review [Package Standards](PACKAGE_STANDARDS.md)
- Check [Architecture Documentation](architecture.md)
- Run `task test` to verify everything works
```

**Acceptance Criteria:**
- [ ] `docs/development-setup.md` created with all setup steps
- [ ] Includes troubleshooting section
- [ ] Linked from main README.md
- [ ] Linked from CONTRIBUTING.md

---

## High Priority

### Issue #3: No CODE_OF_CONDUCT.md

**Priority:** 🟡 High  
**Effort:** ~15 minutes  
**Labels:** `documentation`, `community`, `good-first-issue`

**Problem:**
No code of conduct file exists, making the project less welcoming and unclear about community standards.

**Impact:**
- Potential contributors may hesitate to engage
- No clear guidelines for handling disputes
- Less professional appearance for open source project

**Proposed Solution:**
Add `.github/CODE_OF_CONDUCT.md` using Contributor Covenant:

```markdown
# Contributor Covenant Code of Conduct

## Our Pledge

We as members, contributors, and leaders pledge to make participation in our
community a harassment-free experience for everyone, regardless of age, body
size, visible or invisible disability, ethnicity, sex characteristics, gender
identity and expression, level of experience, education, socio-economic status,
nationality, personal appearance, race, religion, or sexual identity
and orientation.

## Our Standards

Examples of behavior that contributes to a positive environment:

* Using welcoming and inclusive language
* Being respectful of differing viewpoints and experiences
* Gracefully accepting constructive criticism
* Focusing on what is best for the community
* Showing empathy towards other community members

Examples of unacceptable behavior:

* The use of sexualized language or imagery
* Trolling, insulting/derogatory comments, and personal attacks
* Public or private harassment
* Publishing others' private information without permission
* Other conduct which could reasonably be considered inappropriate

## Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be
reported to the project team at [INSERT EMAIL]. All complaints will be
reviewed and investigated promptly and fairly.

## Attribution

This Code of Conduct is adapted from the Contributor Covenant,
version 2.0, available at https://www.contributor-covenant.org/version/2/0/code_of_conduct.html
```

**Acceptance Criteria:**
- [ ] `.github/CODE_OF_CONDUCT.md` created
- [ ] Contact email added for reporting
- [ ] Linked from README.md

---

### Issue #4: CLI Help Text Lacks Examples

**Priority:** 🟡 High  
**Effort:** ~1 hour  
**Labels:** `enhancement`, `cli`, `dx-improvement`

**Problem:**
When users run `core <command> --help`, they see descriptions but no practical examples. This makes it harder to understand how to use commands effectively.

**Current:**
```
$ core build --help
Build project with auto-detection

Usage:
  core build [flags]

Flags:
  --targets string   Build targets
```

**Desired:**
```
$ core build --help
Build project with auto-detection

Usage:
  core build [flags]

Examples:
  # Build for current platform
  core build

  # Cross-compile for multiple platforms
  core build --targets linux/amd64,darwin/arm64,windows/amd64

  # Build with CI mode (checksums, all targets)
  core build --ci

Flags:
  --targets string   Build targets (comma-separated)
  --ci              CI mode with checksums
```

**Impact:**
- Users struggle to understand command usage
- Increased support requests
- Harder to discover features

**Proposed Solution:**
Add `Example` field to all major cobra commands:

```go
&cobra.Command{
    Use:   "build",
    Short: "Build project",
    Long:  "Build project with auto-detection...",
    Example: `  # Build for current platform
  core build

  # Cross-compile for multiple platforms
  core build --targets linux/amd64,darwin/arm64

  # CI mode with checksums
  core build --ci`,
    Run: runBuild,
}
```

**Commands to update:**
- `build`
- `test`
- `dev work`
- `dev health`
- `dev issues`
- `setup`
- All `go` subcommands
- All `php` subcommands

**Acceptance Criteria:**
- [ ] All major commands have `Example` field
- [ ] Examples show common use cases
- [ ] Examples tested and verified to work
- [ ] Help output reviewed for clarity

---

### Issue #5: Missing Examples Directory

**Priority:** 🟡 High  
**Effort:** ~2 hours  
**Labels:** `documentation`, `examples`, `dx-improvement`

**Problem:**
No `examples/` directory with runnable code samples. Developers learning the framework have to piece together patterns from reading production code.

**Impact:**
- Steep learning curve for new developers
- Common patterns not easily discoverable
- More questions in issues/Discord

**Proposed Solution:**
Create `examples/` directory with practical, runnable examples:

```
examples/
├── README.md                    # Overview of examples
├── hello-service/              # Minimal service example
│   ├── main.go
│   ├── service.go
│   ├── service_test.go
│   └── README.md
├── cli-command/                # Adding a custom CLI command
│   ├── main.go
│   ├── cmd_example.go
│   └── README.md
├── error-handling/             # Proper error patterns
│   ├── main.go
│   ├── errors.go
│   └── README.md
├── testing-service/            # Service testing patterns
│   ├── service.go
│   ├── service_test.go
│   └── README.md
└── ipc-actions/               # IPC/Action system usage
    ├── main.go
    ├── service.go
    └── README.md
```

**Example Structure:**

Each example should have:
1. `README.md` explaining the pattern
2. Runnable code (`go run .`)
3. Tests demonstrating testing approach
4. Comments explaining key concepts

**Sample `examples/hello-service/README.md`:**

```markdown
# Hello Service Example

Demonstrates the basic service registration pattern in Core.

## What This Shows

- Creating a simple service
- Registering with Core
- Service lifecycle (OnStartup)
- Using ServiceRuntime[T]

## Run It

```bash
cd examples/hello-service
go run .
```

## Key Patterns

### Dual Constructor

```go
// New() - for standalone/testing
func New() (*Service, error)

// Register() - for Core integration
func Register(c *core.Core) (any, error)
```

### ServiceRuntime

Embed `*core.ServiceRuntime[T]` to get:
- Access to Core instance via `s.Core()`
- Type-safe options via `s.Config()`

## See Also

- [Package Standards](../../docs/PACKAGE_STANDARDS.md)
- [Architecture Guide](../../docs/architecture.md)
```

**Acceptance Criteria:**
- [ ] `examples/` directory created with at least 3 examples
- [ ] Each example has README, code, and tests
- [ ] All examples run successfully
- [ ] Examples referenced from main README
- [ ] Examples linked from getting-started guide

---

## Medium Priority

### Issue #6: Expand Troubleshooting Guide

**Priority:** 🟠 Medium  
**Effort:** ~1 hour  
**Labels:** `documentation`, `dx-improvement`

**Problem:**
While `docs/troubleshooting.md` exists and has good content, it's missing some common developer issues:
- `core doctor` failure scenarios and fixes
- M1 Mac specific issues (CGO, architecture)
- Wails/GUI specific problems
- Go workspace sync issues

**Impact:**
- Developers get stuck on setup issues
- More support questions in Discord/issues
- Time wasted debugging known problems

**Proposed Solution:**
Expand `docs/troubleshooting.md` with:

```markdown
## Development Environment Issues

### "core doctor" shows failed checks

**Check: Go version**
```bash
# Update Go to 1.25+
# macOS: brew upgrade go
# Linux: Download from go.dev
```

**Check: Task not found**
```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

**Check: Git not configured**
```bash
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
```

### Tests fail on Apple Silicon (M1/M2)

**Cause:** CGO architecture mismatch

**Fix:**
```bash
# Ensure Go is native ARM64
go version  # Should show darwin/arm64

# If shows darwin/amd64, reinstall Go
arch -arm64 /bin/bash
brew reinstall go
```

### "go work sync" errors

**Cause:** Go workspace out of sync

**Fix:**
```bash
# Clean and resync
rm go.work go.work.sum
go work init
go work use . cmd/core-gui cmd/examples/*
go work sync
```

### GUI build fails with webpack errors

**Cause:** Node modules out of date or corrupted

**Fix:**
```bash
cd cmd/core-gui
rm -rf node_modules package-lock.json
npm install
npm run build
```
```

**Acceptance Criteria:**
- [ ] Troubleshooting guide expanded with developer-specific issues
- [ ] M1 Mac section added
- [ ] Go workspace troubleshooting added
- [ ] Wails/GUI issues section added
- [ ] Each issue has clear fix steps

---

### Issue #7: Add Task Help Command

**Priority:** 🟠 Medium  
**Effort:** ~30 minutes  
**Labels:** `enhancement`, `dx-improvement`, `good-first-issue`

**Problem:**
`Taskfile.yml` has many tasks but no overview showing common workflows. Developers need to either read the Taskfile or guess at commands.

**Impact:**
- Developers don't discover useful tasks
- Common workflows not obvious
- More time spent in documentation

**Proposed Solution:**
Add `help` task to `Taskfile.yml`:

```yaml
tasks:
  help:
    desc: "Show common development workflows"
    cmds:
      - |
        cat << 'EOF'
        Core Development Tasks
        =====================

        🧪 Testing & QA
          task test              Run all tests
          task test:verbose      Run tests with verbose output
          task test:run -- NAME  Run specific test
          task cov               Generate coverage report
          task cov-view          Open coverage in browser

        🔧 Quality Assurance
          task qa                Standard QA (fmt, vet, lint, test)
          task qa:quick          Quick QA (skip tests)
          task qa:full           Full QA (+ race, vuln, security)
          task fmt               Format code
          task lint              Run linter

        🏗️  Building
          task build             Build with auto-detection
          task cli:build         Build CLI to ./bin/core
          task gui:dev           Run GUI in development mode
          task gui:build         Build GUI for production

        📝 Code Review
          task review            Run CodeRabbit review
          task check             Tidy + test + review

        🌍 Multi-Repo (when in workspace)
          task dev:health        Quick health check
          task dev:status        Show detailed status table
          task dev:work          Commit and push all repos

        🔍 Diagnostics
          task doctor            Check development environment

        📚 Documentation
          task i18n:generate     Regenerate i18n keys
          task i18n:validate     Validate i18n usage

        TDD Workflow:
          1. task test-gen       # Generate test stubs
          2. task test           # Run tests (watch fail)
          3. [implement feature]
          4. task test           # Run tests (watch pass)
          5. task review         # CodeRabbit review

        For detailed help: task --list-all
        EOF
    silent: true
```

**Acceptance Criteria:**
- [ ] `task help` command added
- [ ] Shows organized list of common tasks
- [ ] Includes TDD workflow
- [ ] Mentions `task --list-all` for full list
- [ ] Documented in README

---

### Issue #8: Document Environment Variables

**Priority:** 🟠 Medium  
**Effort:** ~45 minutes  
**Labels:** `documentation`, `dx-improvement`

**Problem:**
No documentation on what environment variables the CLI respects. Developers have to read source code to discover configuration options.

**Impact:**
- Configuration options not discoverable
- Developers miss useful features (debug logging, etc.)
- Inconsistent behavior across environments

**Proposed Solution:**
Add environment variables section to README or create dedicated doc:

```markdown
## Environment Variables

Core respects the following environment variables:

### Logging & Debug

| Variable | Default | Description |
|----------|---------|-------------|
| `CORE_LOG_LEVEL` | `info` | Logging level: `debug`, `info`, `warn`, `error` |
| `CORE_LOG_FORMAT` | `text` | Log format: `text`, `json` |
| `CORE_VERBOSE` | `false` | Enable verbose output |
| `CORE_DEBUG` | `false` | Enable debug mode |

### Paths & Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CORE_CONFIG_DIR` | `~/.config/core` | Config directory |
| `CORE_CACHE_DIR` | `~/.cache/core` | Cache directory |
| `CORE_DATA_DIR` | `~/.local/share/core` | Data directory |
| `CORE_WORK_DIR` | Current dir | Working directory |

### Build & CI

| Variable | Default | Description |
|----------|---------|-------------|
| `CORE_BUILD_DIR` | `dist/` | Build output directory |
| `CORE_SKIP_CHECKS` | `false` | Skip environment checks |
| `CI` | (auto) | CI mode detection |
| `GITHUB_TOKEN` | (required) | GitHub API token for releases |

### Examples

```bash
# Enable debug logging
CORE_LOG_LEVEL=debug core build

# Use custom config directory
CORE_CONFIG_DIR=/tmp/core-config core setup

# Skip environment checks (CI)
CORE_SKIP_CHECKS=1 core build
```
```

**Acceptance Criteria:**
- [ ] Environment variables documented
- [ ] Table shows variable, default, description
- [ ] Examples provided for common use cases
- [ ] Linked from main README
- [ ] Added to configuration.md

---

### Issue #9: Create CLI Output Formatting Guide

**Priority:** 🟠 Medium  
**Effort:** ~1 hour  
**Labels:** `documentation`, `dx-improvement`

**Problem:**
Core CLI has rich output formatting (tables, progress bars, trees, colored output) but no guide for service developers on how to use these consistently.

**Impact:**
- Inconsistent output formatting across commands
- Developers reinvent formatting code
- Services don't follow CLI UX patterns

**Proposed Solution:**
Create `docs/cli-output-guide.md`:

```markdown
# CLI Output Formatting Guide

Guide for creating consistent, user-friendly CLI output in Core.

## Core Principles

1. **Be helpful** - Show what's happening and why
2. **Be consistent** - Use same patterns across commands
3. **Be quiet by default** - Only show what's needed
4. **Be verbose on demand** - Support `-v/--verbose` flag

## Output Helpers

### Success/Error Messages

```go
import "github.com/host-uk/core/pkg/cli"

// Success message (green checkmark)
cli.Success("Build completed successfully")

// Error message (red X)
cli.Error("Build failed: %v", err)

// Warning message (yellow !)
cli.Warning("Skipping tests (--skip-tests flag set)")

// Info message (blue i)
cli.Info("Using Go 1.25.5")
```

### Progress Indicators

```go
// Simple progress message
cli.Progress("Building...")
// Output: ⏳ Building...

// Progress with completion
progress := cli.NewProgress("Building")
// ... do work ...
progress.Complete()
// Output: ✓ Building... done
```

### Tables

```go
import "github.com/host-uk/core/pkg/cli"

table := cli.NewTable()
table.SetHeader([]string{"Name", "Status", "Version"})
table.AddRow([]string{"core", "ok", "0.1.0"})
table.AddRow([]string{"core-gui", "ok", "0.1.0"})
table.Render()
```

Output:
```
Name      | Status | Version
----------|--------|--------
core      | ok     | 0.1.0
core-gui  | ok     | 0.1.0
```

### Trees

```go
tree := cli.NewTree("Packages")
tree.AddNode("pkg/core")
tree.AddNode("pkg/cli")
tree.AddNode("pkg/build")
tree.Render()
```

Output:
```
Packages
├── pkg/core
├── pkg/cli
└── pkg/build
```

## Color Guidelines

- **Green** - Success, completion, "go"
- **Red** - Errors, failures, "stop"
- **Yellow** - Warnings, skipped items
- **Blue** - Info, neutral status
- **Gray** - Secondary info, timestamps

## Verbose Output

Support `--verbose` flag for detailed output:

```go
func runCommand(cmd *cobra.Command, args []string) error {
    verbose, _ := cmd.Flags().GetBool("verbose")
    
    if verbose {
        cli.Info("Starting build process")
        cli.Info("Go version: %s", goVersion)
        cli.Info("Build dir: %s", buildDir)
    }
    
    // ... build logic ...
    
    cli.Success("Build completed")
    return nil
}
```

## Error Messages

Good error messages:
- Explain what failed
- Show the actual error
- Suggest how to fix it

```go
// ❌ Bad
return fmt.Errorf("failed")

// ✅ Good
return fmt.Errorf("failed to read config file %s: %w\n  Try: core setup", configPath, err)
```

## Examples

### Build Command Output

```
$ core build
⏳ Building for linux/amd64...
✓ Built: dist/core-linux-amd64
⏳ Generating checksums...
✓ Build completed in 2.3s
```

### Status Table

```
$ core dev work --status
Repository Status Table
======================

Name      | Status | Branch | Ahead | Behind
----------|--------|--------|-------|-------
core      | clean  | dev    | 0     | 0
core-gui  | dirty  | dev    | 2     | 0
```

## See Also

- [Error Handling](error-handling.md)
- [CLI Architecture](architecture.md#cli)
```

**Acceptance Criteria:**
- [ ] CLI output guide created
- [ ] Examples show all formatting types
- [ ] Code samples tested and work
- [ ] Linked from main README and architecture docs
- [ ] Referenced in CONTRIBUTING.md

---

## Low Priority / Nice to Have

### Issue #10: Add Testing Guide

**Priority:** 🟢 Low  
**Effort:** ~1.5 hours  
**Labels:** `documentation`, `testing`, `dx-improvement`

**Problem:**
While test patterns exist (test naming convention documented), there's no comprehensive guide on:
- How to structure tests
- Table-driven test patterns
- Testing services with dependencies
- Mocking patterns
- Testing concurrent code

**Proposed Solution:**
Create `docs/testing.md` with comprehensive testing guide.

---

### Issue #11: CLI Completion Documentation

**Priority:** 🟢 Low  
**Effort:** ~20 minutes  
**Labels:** `documentation`, `good-first-issue`

**Problem:**
`core completion` command exists but no documentation on how to enable shell completions in common shells.

**Proposed Solution:**
Add section to README:

```markdown
## Shell Completions

Enable tab completion for `core` commands:

### Bash
```bash
# Add to ~/.bashrc
eval "$(core completion bash)"
```

### Zsh
```bash
# Add to ~/.zshrc
eval "$(core completion zsh)"
```

### Fish
```bash
# Add to ~/.config/fish/config.fish
core completion fish | source
```

### PowerShell
```powershell
# Add to $PROFILE
core completion powershell | Out-String | Invoke-Expression
```
```

---

### Issue #12: Add Linting Configuration Documentation

**Priority:** 🟢 Low  
**Effort:** ~30 minutes  
**Labels:** `documentation`, `good-first-issue`

**Problem:**
`task lint` runs golangci-lint but developers don't know what rules apply or how to configure linting.

**Proposed Solution:**
- Create `.golangci.yml` if doesn't exist
- Add comments explaining key rules
- Document in README or linting guide

---

### Issue #13: Service Registration Pattern Reconciliation

**Priority:** 🟢 Low  
**Effort:** ~1 hour  
**Labels:** `documentation`, `architecture`

**Problem:**
README mentions `core.WithService()` but actual code uses `framework.WithName()`. This creates confusion about the correct pattern.

**Impact:**
- Developers unsure which pattern to follow
- Inconsistency between docs and code

**Proposed Solution:**
- Audit all documentation for service registration examples
- Update to use consistent pattern
- Add examples showing both approaches if both are valid
- Clarify in architecture docs when to use each

---

### Issue #14: Add Architecture Decision Records (ADRs)

**Priority:** 🟢 Low  
**Effort:** ~2 hours  
**Labels:** `documentation`, `architecture`, `enhancement`

**Problem:**
No documentation of architectural decisions (why IPC pattern chosen, why dual-constructor, etc.). This makes it harder for new developers to understand design rationale.

**Proposed Solution:**
Create `docs/adr/` directory with Architecture Decision Records:

```
docs/adr/
├── 0001-service-registration-pattern.md
├── 0002-ipc-bridge-vs-direct-bindings.md
├── 0003-dual-constructor-pattern.md
└── README.md
```

Each ADR documents:
- Context (what's the problem)
- Decision (what we chose)
- Rationale (why we chose it)
- Consequences (pros/cons)
- Alternatives (what we didn't choose)

---

## Summary Matrix

| # | Title | Priority | Effort | Impact | Labels |
|---|-------|----------|--------|--------|--------|
| 1 | Missing CONTRIBUTING.md | 🔴 Critical | 30m | High | `documentation`, `good-first-issue`, `dx-improvement` |
| 2 | Missing Development Setup Guide | 🔴 Critical | 45m | High | `documentation`, `dx-improvement`, `onboarding` |
| 3 | No CODE_OF_CONDUCT.md | 🟡 High | 15m | Medium | `documentation`, `community`, `good-first-issue` |
| 4 | CLI Help Text Lacks Examples | 🟡 High | 1h | High | `enhancement`, `cli`, `dx-improvement` |
| 5 | Missing Examples Directory | 🟡 High | 2h | High | `documentation`, `examples`, `dx-improvement` |
| 6 | Expand Troubleshooting Guide | 🟠 Medium | 1h | Medium | `documentation`, `dx-improvement` |
| 7 | Add Task Help Command | 🟠 Medium | 30m | Medium | `enhancement`, `dx-improvement`, `good-first-issue` |
| 8 | Document Environment Variables | 🟠 Medium | 45m | Medium | `documentation`, `dx-improvement` |
| 9 | Create CLI Output Formatting Guide | 🟠 Medium | 1h | Medium | `documentation`, `dx-improvement` |
| 10 | Add Testing Guide | 🟢 Low | 1.5h | Low | `documentation`, `testing`, `dx-improvement` |
| 11 | CLI Completion Documentation | 🟢 Low | 20m | Low | `documentation`, `good-first-issue` |
| 12 | Linting Configuration Documentation | 🟢 Low | 30m | Low | `documentation`, `good-first-issue` |
| 13 | Service Registration Pattern Reconciliation | 🟢 Low | 1h | Low | `documentation`, `architecture` |
| 14 | Add Architecture Decision Records | 🟢 Low | 2h | Low | `documentation`, `architecture`, `enhancement` |

## Quick Wins (Do First)

These issues provide maximum impact for minimal effort:

1. **CODE_OF_CONDUCT.md** (15 min) - Makes project more welcoming
2. **CLI Completion Docs** (20 min) - Instant productivity boost for users
3. **CONTRIBUTING.md** (30 min) - Critical for open source health
4. **Task Help Command** (30 min) - Makes development easier immediately

## How to Use This Document

Each issue above can be:
1. Copied directly into a GitHub issue
2. Assigned appropriate labels
3. Assigned to milestone (e.g., "DX Improvements Q1")
4. Broken down further if needed

**Note:** I cannot directly create GitHub issues due to repository permissions, but this document provides all the content needed to create them manually or via GitHub CLI.

## Commands to Create Issues

If you have `gh` CLI installed:

```bash
# Create issue #1
gh issue create --title "Add CONTRIBUTING.md" --body-file issue-1.md --label "documentation,good-first-issue,dx-improvement"

# Or create all at once with a script
for i in {1..14}; do
  gh issue create --title "$(grep "^### Issue #$i:" DX_IMPROVEMENTS.md | cut -d: -f2-)" \
                  --label "dx-improvement" \
                  # ... body content
done
```

## Maintenance

This document should be updated as:
- Issues are created and linked back
- Issues are completed
- New DX issues are discovered
- Priorities change based on feedback

---

**Created:** 2026-02-01  
**Author:** GitHub Copilot (DX Analysis)  
**Status:** Ready for issue creation
