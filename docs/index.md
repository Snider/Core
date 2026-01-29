# Core CLI

Core is a unified CLI for the host-uk ecosystem - build, release, and deploy Go, Wails, PHP, and container workloads.

## Installation

```bash
# Via Go (recommended)
go install github.com/host-uk/core/cmd/core@latest

# Or download binary from releases
curl -Lo core https://github.com/host-uk/core/releases/latest/download/core-$(go env GOOS)-$(go env GOARCH)
chmod +x core && sudo mv core /usr/local/bin/

# Verify
core doctor
```

See [Getting Started](getting-started.md) for all installation options including building from source.

## Command Reference

See [cmd/](cmd/) for full command documentation.

| Command | Description |
|---------|-------------|
| [go](cmd/go/) | Go development (test, fmt, lint, cov) |
| [php](cmd/php/) | Laravel/PHP development |
| [build](cmd/build/) | Build Go, Wails, Docker, LinuxKit projects |
| [ci](cmd/ci/) | Publish releases (dry-run by default) |
| [sdk](cmd/sdk/) | SDK generation and validation |
| [dev](cmd/dev/) | Multi-repo workflow + dev environment |
| [pkg](cmd/pkg/) | Package search and install |
| [vm](cmd/vm/) | LinuxKit VM management |
| [docs](cmd/docs/) | Documentation management |
| [setup](cmd/setup/) | Clone repos from registry |
| [doctor](cmd/doctor/) | Check development environment |

## Quick Start

```bash
# Go development
core go test              # Run tests
core go test --coverage   # With coverage
core go fmt               # Format code
core go lint              # Lint code

# Build
core build                # Auto-detect and build
core build --targets linux/amd64,darwin/arm64

# Release (dry-run by default)
core ci                   # Preview release
core ci --we-are-go-for-launch  # Actually publish

# Multi-repo workflow
core dev work             # Status + commit + push
core dev work --status    # Just show status

# PHP development
core php dev              # Start dev environment
core php test             # Run tests
```

## Configuration

Core uses `.core/` directory for project configuration:

```
.core/
├── release.yaml    # Release targets and settings
├── build.yaml      # Build configuration (optional)
└── linuxkit/       # LinuxKit templates
```

And `repos.yaml` in workspace root for multi-repo management.

## Guides

- [Getting Started](getting-started.md) - Installation and first steps
- [Workflows](workflows.md) - Common task sequences
- [Troubleshooting](troubleshooting.md) - When things go wrong
- [Migration](migration.md) - Moving from legacy tools

## Reference

- [Configuration](configuration.md) - All config options
- [Glossary](glossary.md) - Term definitions

## Claude Code Skill

Install the skill to teach Claude Code how to use the Core CLI:

```bash
curl -fsSL https://raw.githubusercontent.com/host-uk/core/main/.claude/skills/core/install.sh | bash
```

See [skill/](skill/) for details.
