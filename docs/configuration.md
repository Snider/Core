# Configuration

Core uses `.core/` directory for project configuration.

## Directory Structure

```
.core/
├── release.yaml      # Release configuration
├── build.yaml        # Build configuration (optional)
├── php.yaml          # PHP configuration (optional)
└── linuxkit/         # LinuxKit templates
    ├── server.yml
    └── dev.yml
```

## release.yaml

Full release configuration reference:

```yaml
version: 1

project:
  name: myapp
  repository: myorg/myapp

build:
  targets:
    - os: linux
      arch: amd64
    - os: linux
      arch: arm64
    - os: darwin
      arch: amd64
    - os: darwin
      arch: arm64
    - os: windows
      arch: amd64

publishers:
  # GitHub Releases (required - others reference these artifacts)
  - type: github
    prerelease: false
    draft: false

  # npm binary wrapper
  - type: npm
    package: "@myorg/myapp"
    access: public  # or "restricted"

  # Homebrew formula
  - type: homebrew
    tap: myorg/homebrew-tap
    formula: myapp
    official:
      enabled: false
      output: dist/homebrew

  # Scoop manifest (Windows)
  - type: scoop
    bucket: myorg/scoop-bucket
    official:
      enabled: false
      output: dist/scoop

  # AUR (Arch Linux)
  - type: aur
    maintainer: "Name <email>"

  # Chocolatey (Windows)
  - type: chocolatey
    push: false  # true to publish

  # Docker multi-arch
  - type: docker
    registry: ghcr.io
    image: myorg/myapp
    dockerfile: Dockerfile
    platforms:
      - linux/amd64
      - linux/arm64
    tags:
      - latest
      - "{{.Version}}"
    build_args:
      VERSION: "{{.Version}}"

  # LinuxKit images
  - type: linuxkit
    config: .core/linuxkit/server.yml
    formats:
      - iso
      - qcow2
      - docker
    platforms:
      - linux/amd64
      - linux/arm64

changelog:
  include:
    - feat
    - fix
    - perf
    - refactor
  exclude:
    - chore
    - docs
    - style
    - test
    - ci
```

## build.yaml

Optional build configuration:

```yaml
version: 1

project:
  name: myapp
  binary: myapp

build:
  main: ./cmd/myapp
  env:
    CGO_ENABLED: "0"
  flags:
    - -trimpath
  ldflags:
    - -s -w
    - -X main.version={{.Version}}
    - -X main.commit={{.Commit}}

targets:
  - os: linux
    arch: amd64
  - os: darwin
    arch: arm64
```

## php.yaml

PHP/Laravel configuration:

```yaml
version: 1

dev:
  domain: myapp.test
  ssl: true
  port: 8000
  services:
    - frankenphp
    - vite
    - horizon
    - reverb
    - redis

test:
  parallel: true
  coverage: false

deploy:
  coolify:
    server: https://coolify.example.com
    project: my-project
    environment: production
```

## LinuxKit Templates

LinuxKit YAML configuration:

```yaml
kernel:
  image: linuxkit/kernel:6.6
  cmdline: "console=tty0 console=ttyS0"

init:
  - linuxkit/init:latest
  - linuxkit/runc:latest
  - linuxkit/containerd:latest
  - linuxkit/ca-certificates:latest

onboot:
  - name: sysctl
    image: linuxkit/sysctl:latest

services:
  - name: dhcpcd
    image: linuxkit/dhcpcd:latest
  - name: sshd
    image: linuxkit/sshd:latest
  - name: myapp
    image: myorg/myapp:latest
    capabilities:
      - CAP_NET_BIND_SERVICE

files:
  - path: /etc/myapp/config.yaml
    contents: |
      server:
        port: 8080
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub authentication (via gh CLI) |
| `NPM_TOKEN` | npm publish token |
| `CHOCOLATEY_API_KEY` | Chocolatey publish key |
| `COOLIFY_TOKEN` | Coolify deployment token |

## Defaults

If no configuration exists, sensible defaults are used:

- **Targets**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- **Publishers**: GitHub only
- **Changelog**: feat, fix, perf, refactor included
