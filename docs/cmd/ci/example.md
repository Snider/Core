# CI Examples

## Quick Start

```bash
# Build first
core build

# Preview release
core ci

# Publish
core ci --we-are-go-for-launch
```

## Configuration

`.core/release.yaml`:

```yaml
version: 1

project:
  name: myapp
  repository: host-uk/myapp

publishers:
  - type: github
```

## Publisher Examples

### GitHub + Docker

```yaml
publishers:
  - type: github

  - type: docker
    registry: ghcr.io
    image: host-uk/myapp
    platforms:
      - linux/amd64
      - linux/arm64
    tags:
      - latest
      - "{{.Version}}"
```

### Full Stack (GitHub + npm + Homebrew)

```yaml
publishers:
  - type: github

  - type: npm
    package: "@host-uk/myapp"
    access: public

  - type: homebrew
    tap: host-uk/homebrew-tap
```

### LinuxKit Image

```yaml
publishers:
  - type: linuxkit
    config: .core/linuxkit/server.yml
    formats:
      - iso
      - qcow2
    platforms:
      - linux/amd64
      - linux/arm64
```

## Changelog Configuration

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
