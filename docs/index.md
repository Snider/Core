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

| Command | Description |
|---------|-------------|
| `core build` | Build Go, Wails, Docker, and LinuxKit projects |
| `core release` | Build and publish to GitHub, npm, Homebrew, etc. |
| `core run` | Run LinuxKit images with qemu/hyperkit |
| `core php` | Laravel/PHP development environment |
| `core ps` | List running containers |
| `core stop` | Stop running containers |
| `core logs` | View container logs |
| `core exec` | Execute commands in containers |

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

- [Build Command](build.md) - Cross-platform builds
- [Release Command](release.md) - Publishing to package managers
- [PHP Commands](php.md) - Laravel development
- [Run Command](run.md) - Container management
- [Configuration](configuration.md) - All config options
- [Examples](examples/) - Sample configurations

## Framework

Core also provides a Go framework for building desktop applications:

- [Framework Overview](framework/overview.md)
- [Services](framework/services.md)
- [Lifecycle](framework/lifecycle.md)
