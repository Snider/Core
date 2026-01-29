# Migration Guide

Migrating from legacy scripts and tools to Core CLI.

## From push-all.sh

The `push-all.sh` script has been replaced by `core dev` commands.

| Legacy | Core CLI | Notes |
|--------|----------|-------|
| `./push-all.sh --status` | `core dev work --status` | Status table |
| `./push-all.sh --commit` | `core dev commit` | Commit dirty repos |
| `./push-all.sh` | `core dev work` | Full workflow |

### Quick Migration

```bash
# Instead of
./push-all.sh --status

# Use
core dev work --status
```

### New Features

Core CLI adds features not available in the legacy script:

```bash
# Quick health summary
core dev health
# Output: "18 repos │ clean │ synced"

# Pull repos that are behind
core dev pull

# GitHub integration
core dev issues      # List open issues
core dev reviews     # List PRs needing review
core dev ci          # Check CI status

# Dependency analysis
core dev impact core-php  # What depends on core-php?
```

---

## From Raw Go Commands

Core wraps Go commands with enhanced defaults and output.

| Raw Command | Core CLI | Benefits |
|-------------|----------|----------|
| `go test ./...` | `core go test` | Filters warnings, sets CGO_ENABLED=0 |
| `go test -coverprofile=...` | `core go cov` | HTML reports, thresholds |
| `gofmt -w .` | `core go fmt --fix` | Uses goimports if available |
| `golangci-lint run` | `core go lint` | Consistent interface |
| `go build` | `core build` | Cross-compile, sign, archive |

### Why Use Core?

```bash
# Raw go test shows linker warnings on macOS
go test ./...
# ld: warning: -no_pie is deprecated...

# Core filters noise
core go test
# PASS (clean output)
```

### Environment Setup

Core automatically sets:
- `CGO_ENABLED=0` - Static binaries
- `MACOSX_DEPLOYMENT_TARGET=26.0` - Suppress macOS warnings
- Colour output for coverage reports

---

## From Raw PHP Commands

Core orchestrates Laravel development services.

| Raw Command | Core CLI | Benefits |
|-------------|----------|----------|
| `php artisan serve` | `core php dev` | Adds Vite, Horizon, Reverb, Redis |
| `./vendor/bin/pest` | `core php test` | Auto-detects test runner |
| `./vendor/bin/pint` | `core php fmt --fix` | Consistent interface |
| Manual Coolify deploy | `core php deploy` | Tracked, scriptable |

### Development Server Comparison

```bash
# Raw: Start each service manually
php artisan serve &
npm run dev &
php artisan horizon &
php artisan reverb:start &

# Core: One command
core php dev
# Starts all services, shows unified logs
```

---

## From goreleaser

Core's release system is simpler than goreleaser for host-uk projects.

| goreleaser | Core CLI |
|------------|----------|
| `.goreleaser.yaml` | `.core/release.yaml` |
| `goreleaser release --snapshot` | `core ci` (dry-run) |
| `goreleaser release` | `core ci --we-are-go-for-launch` |

### Configuration Migration

**goreleaser:**
```yaml
builds:
  - main: ./cmd/app
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]

archives:
  - format: tar.gz
    files: [LICENSE, README.md]

release:
  github:
    owner: host-uk
    name: app
```

**Core:**
```yaml
version: 1

project:
  name: app
  repository: host-uk/app

targets:
  - os: linux
    arch: amd64
  - os: darwin
    arch: arm64

publishers:
  - type: github
```

### Key Differences

1. **Separate build and release** - Core separates `core build` from `core ci`
2. **Safe by default** - `core ci` is dry-run unless `--we-are-go-for-launch`
3. **Simpler config** - Fewer options, sensible defaults

---

## From Manual Git Operations

Core automates multi-repo git workflows.

| Manual | Core CLI |
|--------|----------|
| `cd repo1 && git status && cd ../repo2 && ...` | `core dev work --status` |
| Check each repo for uncommitted changes | `core dev health` |
| Commit each repo individually | `core dev commit` |
| Push each repo individually | `core dev push` |

### Example: Committing Across Repos

**Manual:**
```bash
cd core-php
git add -A
git commit -m "feat: add feature"
cd ../core-tenant
git add -A
git commit -m "feat: use new feature"
# ... repeat for each repo
```

**Core:**
```bash
core dev commit
# Interactive: reviews changes, suggests messages
# Adds Co-Authored-By automatically
```

---

## Deprecated Commands

These commands have been removed or renamed:

| Deprecated | Replacement | Version |
|------------|-------------|---------|
| `core sdk generate` | `core build sdk` | v0.5.0 |
| `core dev task*` | `core ai task*` | v0.8.0 |
| `core release` | `core ci` | v0.6.0 |

---

## Version Compatibility

| Core Version | Go Version | Breaking Changes |
|--------------|------------|------------------|
| v1.0.0+ | 1.23+ | Stable API |
| v0.8.0 | 1.22+ | Task commands moved to `ai` |
| v0.6.0 | 1.22+ | Release command renamed to `ci` |
| v0.5.0 | 1.21+ | SDK generation moved to `build sdk` |

---

## Getting Help

If you encounter issues during migration:

1. Check [Troubleshooting](troubleshooting.md)
2. Run `core doctor` to verify setup
3. Use `--help` on any command: `core dev work --help`

---

## See Also

- [Getting Started](getting-started.md) - Fresh installation
- [Workflows](workflows.md) - Common task sequences
- [Configuration](configuration.md) - Config file reference
