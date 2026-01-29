# core release

Build and publish releases to GitHub, npm, Homebrew, Scoop, AUR, Chocolatey, Docker, and LinuxKit.

## Usage

```bash
core release [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview what would be published |
| `--version` | Override version (default: git tag) |
| `--target` | Release target: `sdk` for SDK-only release |
| `--draft` | Create release as draft |
| `--prerelease` | Mark release as prerelease |

## Quick Start

```bash
# Initialize release config
core release init

# Preview release
core release --dry-run

# Release
core release

# SDK-only release
core release --target sdk
```

## SDK Release

Generate SDKs without building binaries:

```bash
# Generate SDKs with version from git tag
core release --target sdk

# Explicit version
core release --target sdk --version v1.2.3

# Preview
core release --target sdk --dry-run
```

This will:
1. Determine version from git tags (or `--version` flag)
2. Run breaking change detection if configured
3. Generate SDKs for all configured languages
4. Output to `sdk/` directory

See [SDK commands](sdk.md) for more details.

## Publishers

### GitHub Releases

Uploads artifacts and changelog to GitHub Releases.

```yaml
publishers:
  - type: github
    prerelease: false
    draft: false
```

### npm

Publishes binary wrapper that downloads correct platform binary.

```yaml
publishers:
  - type: npm
    package: "@myorg/myapp"
    access: public
```

Requires `NPM_TOKEN` environment variable.

### Homebrew

Generates formula and commits to tap repository.

```yaml
publishers:
  - type: homebrew
    tap: myorg/homebrew-tap
    formula: myapp  # optional, defaults to project name
```

For official Homebrew PR:

```yaml
publishers:
  - type: homebrew
    tap: myorg/homebrew-tap
    official:
      enabled: true
      output: dist/homebrew
```

### Scoop

Generates manifest and commits to bucket repository.

```yaml
publishers:
  - type: scoop
    bucket: myorg/scoop-bucket
```

### AUR

Generates PKGBUILD and pushes to AUR.

```yaml
publishers:
  - type: aur
    maintainer: "Your Name <email@example.com>"
```

### Chocolatey

Generates NuSpec and optionally pushes to Chocolatey.

```yaml
publishers:
  - type: chocolatey
    push: false  # generate only
```

Set `push: true` and `CHOCOLATEY_API_KEY` to publish.

### Docker

Builds and pushes multi-arch Docker images.

```yaml
publishers:
  - type: docker
    registry: ghcr.io
    image: myorg/myapp
    platforms:
      - linux/amd64
      - linux/arm64
    tags:
      - latest
      - "{{.Version}}"
```

### LinuxKit

Builds immutable LinuxKit images.

```yaml
publishers:
  - type: linuxkit
    config: .core/linuxkit/server.yml
    formats:
      - iso
      - qcow2
      - docker  # Immutable container
    platforms:
      - linux/amd64
      - linux/arm64
```

## Full Example

See [examples/full-release.yaml](examples/full-release.yaml) for a complete configuration.

## Changelog

Changelog is auto-generated from conventional commits:

```
feat: Add new feature     → Features
fix: Fix bug              → Bug Fixes
perf: Improve performance → Performance
refactor: Refactor code   → Refactoring
```

Configure in `.core/release.yaml`:

```yaml
changelog:
  include:
    - feat
    - fix
    - perf
  exclude:
    - chore
    - docs
    - test
```
