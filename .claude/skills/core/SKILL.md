---
name: core
description: Use when working in host-uk repositories, running tests, building, releasing, or managing multi-repo workflows. Provides the core CLI command reference.
---

# Core CLI

The `core` command provides a unified interface for Go/Wails development, multi-repo management, and deployment.

**Rule:** Always prefer `core <command>` over raw commands. It handles environment setup, output formatting, and cross-platform concerns.

## Command Quick Reference

| Task | Command | Notes |
|------|---------|-------|
| Run Go tests | `core test` | Sets macOS deployment target, filters warnings |
| Run Go tests with coverage | `core test --coverage` | Per-package breakdown |
| Run PHP tests | `core php test` | Auto-detects Pest/PHPUnit |
| Start PHP dev server | `core php dev` | FrankenPHP + Vite + Horizon + Reverb |
| Format PHP code | `core php fmt --fix` | Laravel Pint |
| Deploy PHP app | `core php deploy` | Coolify deployment |
| Build project | `core build` | Auto-detects project type |
| Build for targets | `core build --targets linux/amd64,darwin/arm64` | Cross-compile |
| Release | `core release` | Build + publish to GitHub/npm/Homebrew |
| Check environment | `core doctor` | Verify tools installed |
| Multi-repo status | `core health` | Quick summary across repos |
| Multi-repo workflow | `core work` | Status + commit + push |
| Commit dirty repos | `core commit` | Claude-assisted commit messages |
| Push repos | `core push` | Push repos with unpushed commits |
| Pull repos | `core pull` | Pull repos that are behind |
| List issues | `core issues` | Open issues across repos |
| List PRs | `core reviews` | PRs needing review |
| Check CI | `core ci` | GitHub Actions status |
| Generate SDK | `core sdk` | Generate API clients from OpenAPI |
| Sync docs | `core docs sync` | Sync docs across repos |
| Search packages | `core pkg search <query>` | GitHub search for core-* repos |
| Install package | `core pkg install <name>` | Clone and register package |
| Update packages | `core pkg update` | Pull latest for all packages |
| Run VM | `core vm run <image>` | Run LinuxKit VM |

## Testing

**Always use `core test` instead of `go test`.**

```bash
# Run all tests with coverage summary
core test

# Detailed per-package coverage
core test --coverage

# Test specific packages
core test --pkg ./pkg/crypt

# Run specific tests
core test --run TestHash
core test --run "Test.*Good"

# Skip integration tests
core test --short

# Race detection
core test --race

# JSON output for CI/parsing
core test --json
```

**Why:** Sets `MACOSX_DEPLOYMENT_TARGET=26.0` to suppress linker warnings, filters noise from output, provides colour-coded coverage.

### JSON Output

For programmatic use:

```json
{
  "passed": 14,
  "failed": 0,
  "skipped": 0,
  "coverage": 75.1,
  "exit_code": 0,
  "failed_packages": []
}
```

## Building

**Always use `core build` instead of `go build`.**

```bash
# Auto-detect and build
core build

# Build for specific targets
core build --targets linux/amd64,darwin/arm64

# Build Docker image
core build --type docker

# Build LinuxKit image
core build --type linuxkit --format qcow2-bios

# CI mode (JSON output)
core build --ci
```

**Why:** Handles cross-compilation, code signing, archiving, checksums, and CI output formatting.

## Multi-Repo Workflow

When working across host-uk repositories:

```bash
# Quick health check
core health
# Output: "18 repos │ clean │ synced"

# Full status table
core work --status

# Commit + push workflow
core work

# Commit dirty repos with Claude
core commit

# Push repos with unpushed commits
core push

# Pull repos that are behind
core pull
```

### Dependency Analysis

```bash
# What depends on core-php?
core impact core-php
```

## GitHub Integration

Requires `gh` CLI authenticated.

```bash
# Open issues across all repos
core issues

# Include closed issues
core issues --all

# PRs needing review
core reviews

# CI status
core ci
```

## SDK Generation

Generate API clients from OpenAPI specs:

```bash
# Generate all configured SDKs
core sdk

# Generate specific language
core sdk --lang typescript
core sdk --lang php

# Specify OpenAPI spec
core sdk --spec ./openapi.yaml
```

## Documentation

```bash
# List docs across repos
core docs list

# Sync docs to central location
core docs sync
```

## Environment Setup

```bash
# Check development environment
core doctor

# Clone all repos from registry
core setup
```

## Package Management

Manage host-uk/core-* packages and repositories.

```bash
# Search GitHub for packages
core pkg search <query>
core pkg search core-           # Find all core-* packages
core pkg search --org host-uk   # Search specific org

# Install/clone a package
core pkg install core-api
core pkg install host-uk/core-api  # Full name

# List installed packages
core pkg list
core pkg list --format json     # JSON output

# Update installed packages
core pkg update                 # Update all
core pkg update core-api        # Update specific package

# Check for outdated packages
core pkg outdated
```

## PHP Development

**Always use `core php` commands instead of raw artisan/composer/phpunit.**

### Quick Reference

| Task | Command | Notes |
|------|---------|-------|
| Start dev environment | `core php dev` | FrankenPHP + Vite + Horizon + Reverb + Redis |
| Run PHP tests | `core php test` | Auto-detects Pest/PHPUnit |
| Format code | `core php fmt --fix` | Laravel Pint |
| Static analysis | `core php analyse` | PHPStan/Larastan |
| Build Docker image | `core php build` | Production-ready FrankenPHP |
| Deploy to Coolify | `core php deploy` | With status tracking |

### Development Server

```bash
# Start full Laravel dev environment
core php dev

# Start with HTTPS (uses mkcert)
core php dev --https

# Skip specific services
core php dev --no-vite --no-horizon

# Custom port
core php dev --port 9000
```

**Services orchestrated:**
- FrankenPHP/Octane (port 8000, HTTPS on 443)
- Vite dev server (port 5173)
- Laravel Horizon (queue workers)
- Laravel Reverb (WebSocket, port 8080)
- Redis (port 6379)

```bash
# View logs
core php logs
core php logs --service frankenphp

# Check status
core php status

# Stop all services
core php stop

# Setup SSL certificates
core php ssl
core php ssl --domain myapp.test
```

### Testing

```bash
# Run all tests (auto-detects Pest/PHPUnit)
core php test

# Run in parallel
core php test --parallel

# With coverage
core php test --coverage

# Filter tests
core php test --filter UserTest
core php test --group api
```

### Code Quality

```bash
# Check formatting (dry-run)
core php fmt

# Auto-fix formatting
core php fmt --fix

# Show diff
core php fmt --diff

# Run static analysis
core php analyse

# Max strictness
core php analyse --level 9
```

### Building & Deployment

```bash
# Build Docker image
core php build
core php build --name myapp --tag v1.0

# Build for specific platform
core php build --platform linux/amd64

# Build LinuxKit image
core php build --type linuxkit --format iso

# Run production container
core php serve --name myapp
core php serve --name myapp -d  # Detached

# Open shell in container
core php shell myapp
```

### Coolify Deployment

```bash
# Deploy to production
core php deploy

# Deploy to staging
core php deploy --staging

# Wait for completion
core php deploy --wait

# Check deployment status
core php deploy:status

# List recent deployments
core php deploy:list

# Rollback
core php deploy:rollback
core php deploy:rollback --id abc123
```

**Required .env configuration:**
```env
COOLIFY_URL=https://coolify.example.com
COOLIFY_TOKEN=your-api-token
COOLIFY_APP_ID=production-app-id
COOLIFY_STAGING_APP_ID=staging-app-id
```

### Package Management

```bash
# Link local packages for development
core php packages link ../my-package
core php packages link ../pkg-a ../pkg-b

# List linked packages
core php packages list

# Update linked packages
core php packages update

# Unlink packages
core php packages unlink vendor/my-package
```

## VM Management

LinuxKit VMs are lightweight, immutable VMs built from YAML templates.

```bash
# Run LinuxKit image
core vm run server.iso

# Run with options
core vm run -d --memory 2048 --cpus 4 image.iso

# Run from template
core vm run --template core-dev --var SSH_KEY="ssh-rsa AAAA..."

# List running VMs
core vm ps
core vm ps -a  # Include stopped

# Stop VM
core vm stop <id>

# View logs
core vm logs <id>
core vm logs -f <id>  # Follow

# Execute command in VM
core vm exec <id> ls -la
core vm exec <id> /bin/sh

# Manage templates
core vm templates              # List templates
core vm templates show <name>  # Show template content
core vm templates vars <name>  # Show template variables
```

## Decision Tree

```
Go project?
  └── Run tests: core test [--coverage]
  └── Build: core build [--targets <os/arch>]
  └── Release: core release

PHP/Laravel project?
  └── Start dev: core php dev [--https]
  └── Run tests: core php test [--parallel]
  └── Format: core php fmt --fix
  └── Analyse: core php analyse
  └── Build image: core php build
  └── Deploy: core php deploy [--staging]

Working across multiple repos?
  └── Quick check: core health
  └── Full workflow: core work
  └── Just commit: core commit
  └── Just push: core push

Need GitHub info?
  └── Issues: core issues
  └── PRs: core reviews
  └── CI: core ci

Setting up environment?
  └── Check: core doctor
  └── Clone all: core setup

Managing packages?
  └── Search: core pkg search <query>
  └── Install: core pkg install <name>
  └── Update: core pkg update
  └── Check outdated: core pkg outdated
```

## Common Mistakes

| Wrong | Right | Why |
|-------|-------|-----|
| `go test ./...` | `core test` | Missing deployment target, noisy output |
| `go build` | `core build` | Missing cross-compile, signing, checksums |
| `php artisan serve` | `core php dev` | Missing Vite, Horizon, Reverb, Redis |
| `./vendor/bin/pest` | `core php test` | Inconsistent invocation |
| `./vendor/bin/pint` | `core php fmt --fix` | Consistent interface |
| `git status` in each repo | `core health` | Slow, manual |
| `gh pr list` per repo | `core reviews` | Aggregated view |
| Manual commits across repos | `core commit` | Consistent messages, Co-Authored-By |
| Manual Coolify deploys | `core php deploy` | Tracked, scriptable |
| Raw `linuxkit run` | `core vm run` | Unified interface, templates |
| `gh repo clone` | `core pkg install` | Auto-detects org, adds to registry |
| Manual GitHub search | `core pkg search` | Filtered to org, formatted output |

## Configuration

Core reads from `.core/` directory:

```
.core/
├── release.yaml    # Release targets
├── build.yaml      # Build settings
└── linuxkit/       # LinuxKit templates
```

And `repos.yaml` in workspace root for multi-repo management.

## Installation

```bash
# Go install
go install github.com/host-uk/core/cmd/core@latest

# Or from source
cd /path/to/core
go install ./cmd/core/
```

Verify: `core doctor`