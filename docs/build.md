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
| `--no-sign` | Skip code signing |
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

sign:
  enabled: true
  gpg:
    key: $GPG_KEY_ID
  macos:
    identity: "Developer ID Application: Your Name (TEAM_ID)"
    notarize: false
    apple_id: $APPLE_ID
    team_id: $APPLE_TEAM_ID
    app_password: $APPLE_APP_PASSWORD
```

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
