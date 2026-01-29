# core build

Build Go, Wails, Docker, and LinuxKit projects with automatic project detection.

## Usage

```bash
core build [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--type` | Project type: `go`, `wails`, `docker`, `linuxkit` (auto-detected) |
| `--targets` | Build targets: `linux/amd64,darwin/arm64,windows/amd64` |
| `--output` | Output directory (default: `dist`) |
| `--ci` | CI mode - non-interactive, fail fast |
| `--image` | Docker image name (for docker builds) |

## Examples

### Go Project

```bash
# Auto-detect and build
core build

# Build for specific platforms
core build --targets linux/amd64,linux/arm64,darwin/arm64

# CI mode
core build --ci
```

### Wails Project

```bash
# Build Wails desktop app
core build --type wails

# Build for all desktop platforms
core build --type wails --targets darwin/amd64,darwin/arm64,windows/amd64,linux/amd64
```

### Docker Image

```bash
# Build Docker image
core build --type docker

# With custom image name
core build --type docker --image ghcr.io/myorg/myapp
```

### LinuxKit Image

```bash
# Build LinuxKit ISO
core build --type linuxkit
```

## Project Detection

Core automatically detects project type based on files:

| Files | Type |
|-------|------|
| `wails.json` | Wails |
| `go.mod` | Go |
| `Dockerfile` | Docker |
| `composer.json` | PHP |
| `package.json` | Node |

## Output

Build artifacts are placed in `dist/` by default:

```
dist/
├── myapp-linux-amd64.tar.gz
├── myapp-linux-arm64.tar.gz
├── myapp-darwin-amd64.tar.gz
├── myapp-darwin-arm64.tar.gz
├── myapp-windows-amd64.zip
└── CHECKSUMS.txt
```

## Configuration

Optional `.core/build.yaml`:

```yaml
version: 1

project:
  name: myapp
  binary: myapp

build:
  main: ./cmd/myapp
  ldflags:
    - -s -w
    - -X main.version={{.Version}}

targets:
  - os: linux
    arch: amd64
  - os: linux
    arch: arm64
  - os: darwin
    arch: arm64
```
