# Getting Started

This guide walks you through installing Core and running your first build.

## Prerequisites

Before installing Core, ensure you have:

| Tool | Minimum Version | Check Command |
|------|-----------------|---------------|
| Go | 1.23+ | `go version` |
| Git | 2.30+ | `git --version` |

Optional (for specific features):

| Tool | Required For | Install |
|------|--------------|---------|
| `gh` | GitHub integration (`core dev issues`, `core dev reviews`) | [cli.github.com](https://cli.github.com) |
| Docker | Container builds | [docker.com](https://docker.com) |
| `task` | Task automation | `go install github.com/go-task/task/v3/cmd/task@latest` |

## Installation

### Option 1: Go Install (Recommended)

```bash
# Install latest release
go install github.com/host-uk/core/cmd/core@latest

# Verify installation
core doctor
```

If `core: command not found`, add Go's bin directory to your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Option 2: Download Binary

Download pre-built binaries from [GitHub Releases](https://github.com/host-uk/core/releases):

```bash
# macOS (Apple Silicon)
curl -Lo core https://github.com/host-uk/core/releases/latest/download/core-darwin-arm64
chmod +x core
sudo mv core /usr/local/bin/

# macOS (Intel)
curl -Lo core https://github.com/host-uk/core/releases/latest/download/core-darwin-amd64
chmod +x core
sudo mv core /usr/local/bin/

# Linux (x86_64)
curl -Lo core https://github.com/host-uk/core/releases/latest/download/core-linux-amd64
chmod +x core
sudo mv core /usr/local/bin/
```

### Option 3: Build from Source

```bash
# Clone repository
git clone https://github.com/host-uk/core.git
cd core

# Build with Task (recommended)
task cli:build
# Binary at ./bin/core

# Or build with Go directly
CGO_ENABLED=0 go build -o core ./cmd/core/
sudo mv core /usr/local/bin/
```

## Your First Build

### 1. Navigate to a Go Project

```bash
cd ~/Code/my-go-project
```

### 2. Initialise Configuration

```bash
core setup
```

This detects your project type and creates configuration files in `.core/`:
- `build.yaml` - Build settings
- `release.yaml` - Release configuration
- `test.yaml` - Test commands

### 3. Build

```bash
core build
```

Output appears in `dist/`:

```
dist/
├── my-project-darwin-arm64.tar.gz
├── my-project-linux-amd64.tar.gz
└── CHECKSUMS.txt
```

### 4. Cross-Compile (Optional)

```bash
core build --targets linux/amd64,linux/arm64,darwin/arm64,windows/amd64
```

## Your First Release

Releases are **safe by default** - Core runs in dry-run mode unless you explicitly confirm.

### 1. Preview

```bash
core ci
```

This shows what would be published without actually publishing.

### 2. Publish

```bash
core ci --we-are-go-for-launch
```

This creates a GitHub release with your built artifacts.

## Multi-Repo Workflow

If you work with multiple repositories (like the host-uk ecosystem):

### 1. Clone All Repositories

```bash
mkdir host-uk && cd host-uk
core setup
```

Select packages in the interactive wizard.

### 2. Check Status

```bash
core dev health
# Output: "18 repos │ clean │ synced"
```

### 3. Work Across Repos

```bash
core dev work --status    # See status table
core dev work             # Commit and push all dirty repos
```

## Next Steps

| Task | Command | Documentation |
|------|---------|---------------|
| Run tests | `core go test` | [go/test](cmd/go/test/) |
| Format code | `core go fmt --fix` | [go/fmt](cmd/go/fmt/) |
| Lint code | `core go lint` | [go/lint](cmd/go/lint/) |
| PHP development | `core php dev` | [php](cmd/php/) |
| View all commands | `core --help` | [cmd](cmd/) |

## Getting Help

```bash
# Check environment
core doctor

# Command help
core <command> --help

# Full documentation
https://github.com/host-uk/core/tree/main/docs
```

## See Also

- [Configuration](configuration.md) - All config options
- [Workflows](workflows.md) - Common task sequences
- [Troubleshooting](troubleshooting.md) - When things go wrong
