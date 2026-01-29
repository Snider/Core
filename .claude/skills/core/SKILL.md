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
| Run tests | `core test` | Sets macOS deployment target, filters warnings |
| Run tests with coverage | `core test --coverage` | Per-package breakdown |
| Run specific test | `core test --run TestName` | Regex filter |
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

# Search GitHub repos
core search <query>

# Clone a specific repo
core install <repo>
```

## PHP Development

```bash
# Start PHP dev environment
core php dev

# Run artisan commands
core php artisan <command>

# Run composer
core php composer <command>
```

## Container Management

```bash
# Run LinuxKit image
core run server.iso

# List running containers
core ps

# Stop container
core stop <name>

# View logs
core logs <name>

# Execute command in container
core exec <name> <command>
```

## Decision Tree

```
Need to run tests?
  └── core test [--coverage] [--pkg <pattern>]

Need to build?
  └── core build [--targets <os/arch>]

Need to release?
  └── core release

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
```

## Common Mistakes

| Wrong | Right | Why |
|-------|-------|-----|
| `go test ./...` | `core test` | Missing deployment target, noisy output |
| `go build` | `core build` | Missing cross-compile, signing, checksums |
| `git status` in each repo | `core health` | Slow, manual |
| `gh pr list` per repo | `core reviews` | Aggregated view |
| Manual commits across repos | `core commit` | Consistent messages, Co-Authored-By |

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