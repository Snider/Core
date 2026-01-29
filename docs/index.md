# Core CLI

Core is a unified CLI for the host-uk ecosystem - build, release, and deploy Go, Wails, PHP, and container workloads.

## Installation

```bash
# Go install
go install github.com/host-uk/core/cmd/core@latest

# Or download from releases
curl -fsSL https://github.com/host-uk/core/releases/latest/download/core-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/').tar.gz | tar -xzf - -C /usr/local/bin
```

## Commands

### Build & Release

| Command | Description |
|---------|-------------|
| [`core build`](cmd/build.md) | Build Go, Wails, Docker, and LinuxKit projects |
| [`core release`](cmd/release.md) | Build and publish to GitHub, npm, Homebrew, etc. |
| [`core sdk`](cmd/sdk.md) | Generate and manage API SDKs |

### Containers

| Command | Description |
|---------|-------------|
| [`core run`](cmd/run.md) | Run LinuxKit images with qemu/hyperkit |
| `core ps` | List running containers |
| `core stop` | Stop running containers |
| `core logs` | View container logs |
| `core exec` | Execute commands in containers |
| [`core templates`](cmd/templates.md) | Manage LinuxKit templates |

### Development

| Command | Description |
|---------|-------------|
| [`core dev`](cmd/dev.md) | Portable development environment (100+ tools) |
| [`core php`](cmd/php.md) | Laravel/PHP development tools |
| [`core doctor`](cmd/doctor.md) | Check development environment |

### GitHub & Multi-Repo

| Command | Description |
|---------|-------------|
| [`core search`](cmd/search.md) | Search GitHub for repositories |
| [`core install`](cmd/search.md) | Clone a repository from GitHub |
| [`core setup`](cmd/setup.md) | Clone all repos from registry |
| [`core work`](cmd/work.md) | Multi-repo git operations |
| [`core health`](cmd/work.md) | Quick health check across repos |
| [`core issues`](cmd/work.md) | List open issues across repos |
| [`core reviews`](cmd/work.md) | List PRs needing review |
| [`core ci`](cmd/work.md) | Check CI status across repos |

### Documentation

| Command | Description |
|---------|-------------|
| [`core docs`](cmd/docs.md) | Documentation management |

## Quick Start

```bash
# Build a Go project
core build

# Build for specific targets
core build --targets linux/amd64,darwin/arm64

# Release to GitHub
core release

# Release to multiple package managers
core release  # Publishes to all configured targets

# Start PHP dev environment
core php dev

# Run a LinuxKit image
core run server.iso
```

## Configuration

Core uses `.core/` directory for project configuration:

```
.core/
├── release.yaml    # Release targets and settings
├── build.yaml      # Build configuration (optional)
└── linuxkit/       # LinuxKit templates
    └── server.yml
```

## Documentation

### Command Reference
- [Build](cmd/build.md) - Cross-platform builds with code signing
- [Release](cmd/release.md) - Publishing to package managers
- [SDK](cmd/sdk.md) - Generate API clients from OpenAPI
- [Run](cmd/run.md) - Container management
- [Templates](cmd/templates.md) - LinuxKit templates
- [Dev](cmd/dev.md) - Portable development environment
- [PHP](cmd/php.md) - Laravel development
- [Doctor](cmd/doctor.md) - Environment check
- [Search & Install](cmd/search.md) - GitHub integration
- [Setup](cmd/setup.md) - Clone repos from registry
- [Work](cmd/work.md) - Multi-repo operations
- [Docs](cmd/docs.md) - Documentation management

### Reference
- [Configuration](configuration.md) - All config options
- [Examples](examples/) - Sample configurations

## Framework

Core also provides a Go framework for building desktop applications:

- [Framework Overview](framework/overview.md)
- [Services](framework/services.md)
- [Lifecycle](framework/lifecycle.md)
