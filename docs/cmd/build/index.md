# core build

Build Go, Wails, Docker, and LinuxKit projects with automatic project detection.

## Subcommands

| Command | Description |
|---------|-------------|
| [sdk](sdk/) | Generate API SDKs from OpenAPI |
| `from-path` | Build from a local directory |
| `pwa` | Build from a live PWA URL |

## Usage

```bash
core build [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--type` | Project type: `go`, `wails`, `docker`, `linuxkit`, `taskfile` (auto-detected) |
| `--targets` | Build targets: `linux/amd64,darwin/arm64,windows/amd64` |
| `--output` | Output directory (default: `dist`) |
| `--ci` | CI mode - minimal output with JSON artifact list at the end |
| `--image` | Docker image name (for docker builds) |
| `--config` | Config file path (for linuxkit: YAML config, for docker: Dockerfile) |
| `--format` | Output format for linuxkit (iso-bios, qcow2-bios, raw, vmdk) |
| `--push` | Push Docker image after build (default: false) |
| `--archive` | Create archives (tar.gz for linux/darwin, zip for windows) - default: true |
| `--checksum` | Generate SHA256 checksums and CHECKSUMS.txt - default: true |
| `--no-sign` | Skip all code signing |
| `--notarize` | Enable macOS notarization (requires Apple credentials) |

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

# Build and push to registry
core build --type docker --image ghcr.io/myorg/myapp --push
```

### LinuxKit Image

```bash
# Build LinuxKit ISO
core build --type linuxkit

# Build with specific format
core build --type linuxkit --config linuxkit.yml --format qcow2-bios
```

## Project Detection

Core automatically detects project type based on files:

| Files | Type |
|-------|------|
| `wails.json` | Wails |
| `go.mod` | Go |
| `Dockerfile` | Docker |
| `Taskfile.yml` | Taskfile |
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

Optional `.core/build.yaml` - see [Configuration](example.md#configuration) for examples.

## Code Signing

Core supports GPG signing for checksums and native code signing for macOS.

### GPG Signing

Signs `CHECKSUMS.txt` with a detached ASCII signature (`.asc`):

```bash
# Build with GPG signing (default if key configured)
core build

# Skip signing
core build --no-sign
```

Users can verify:

```bash
gpg --verify CHECKSUMS.txt.asc CHECKSUMS.txt
sha256sum -c CHECKSUMS.txt
```

### macOS Code Signing

Signs Darwin binaries with your Developer ID and optionally notarizes with Apple:

```bash
# Build with codesign (automatic if identity configured)
core build

# Build with notarization (takes 1-5 minutes)
core build --notarize
```

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `GPG_KEY_ID` | GPG key ID or fingerprint |
| `CODESIGN_IDENTITY` | macOS Developer ID (fallback) |
| `APPLE_ID` | Apple account email |
| `APPLE_TEAM_ID` | Apple Developer Team ID |
| `APPLE_APP_PASSWORD` | App-specific password for notarization |

## Building from PWAs and Static Sites

### Build from Local Directory

Build a desktop app from static web application files:

```bash
core build from-path --path ./dist
```

### Build from Live PWA

Build a desktop app from a live Progressive Web App URL:

```bash
core build pwa --url https://example.com
```
