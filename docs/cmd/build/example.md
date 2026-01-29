# Build Examples

## Quick Start

```bash
# Auto-detect and build
core build

# Build for specific platforms
core build --targets linux/amd64,darwin/arm64

# CI mode
core build --ci
```

## Configuration

`.core/build.yaml`:

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

## Cross-Platform Build

```bash
core build --targets linux/amd64,linux/arm64,darwin/arm64,windows/amd64
```

Output:
```
dist/
├── myapp-linux-amd64.tar.gz
├── myapp-linux-arm64.tar.gz
├── myapp-darwin-arm64.tar.gz
├── myapp-windows-amd64.zip
└── CHECKSUMS.txt
```

## Code Signing

```yaml
sign:
  enabled: true
  gpg:
    key: $GPG_KEY_ID
  macos:
    identity: "Developer ID Application: Your Name (TEAM_ID)"
    notarize: true
    apple_id: $APPLE_ID
    team_id: $APPLE_TEAM_ID
    app_password: $APPLE_APP_PASSWORD
```

## Docker Build

```bash
core build --type docker --image ghcr.io/myorg/myapp
```

## Wails Desktop App

```bash
core build --type wails --targets darwin/arm64,windows/amd64
```
