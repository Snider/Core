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

## repos.yaml

Package registry for multi-repo workspaces:

```yaml
# Organisation name (used for GitHub URLs)
org: host-uk

# Base path for cloning (default: current directory)
base_path: .

# Default settings for all repos
defaults:
  ci: github
  license: EUPL-1.2
  branch: main

# Repository definitions
repos:
  # Foundation packages (no dependencies)
  core-php:
    type: foundation
    description: Foundation framework

  core-devops:
    type: foundation
    description: Development environment
    clone: false  # Skip during setup (already exists)

  # Module packages (depend on foundation)
  core-tenant:
    type: module
    depends_on: [core-php]
    description: Multi-tenancy module

  core-admin:
    type: module
    depends_on: [core-php, core-tenant]
    description: Admin panel

  core-api:
    type: module
    depends_on: [core-php]
    description: REST API framework

  # Product packages (user-facing applications)
  core-bio:
    type: product
    depends_on: [core-php, core-tenant]
    description: Link-in-bio product
    domain: bio.host.uk.com

  core-social:
    type: product
    depends_on: [core-php, core-tenant]
    description: Social scheduling
    domain: social.host.uk.com

  # Templates
  core-template:
    type: template
    description: Starter template for new projects
```

### repos.yaml Fields

| Field | Required | Description |
|-------|----------|-------------|
| `org` | Yes | GitHub organisation name |
| `base_path` | No | Directory for cloning (default: `.`) |
| `defaults` | No | Default settings applied to all repos |
| `repos` | Yes | Map of repository definitions |

### Repository Fields

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `foundation`, `module`, `product`, or `template` |
| `description` | No | Human-readable description |
| `depends_on` | No | List of package dependencies |
| `clone` | No | Set `false` to skip during setup |
| `domain` | No | Production domain (for products) |
| `branch` | No | Override default branch |

### Package Types

| Type | Description | Dependencies |
|------|-------------|--------------|
| `foundation` | Core framework packages | None |
| `module` | Reusable modules | Foundation packages |
| `product` | User-facing applications | Foundation + modules |
| `template` | Starter templates | Any |

---

## Environment Variables

Complete reference of environment variables used by Core CLI.

### Authentication

| Variable | Used By | Description |
|----------|---------|-------------|
| `GITHUB_TOKEN` | `core ci`, `core dev` | GitHub API authentication |
| `ANTHROPIC_API_KEY` | `core ai`, `core dev claude` | Claude API key |
| `AGENTIC_TOKEN` | `core ai task*` | Agentic API authentication |
| `AGENTIC_BASE_URL` | `core ai task*` | Agentic API endpoint |

### Publishing

| Variable | Used By | Description |
|----------|---------|-------------|
| `NPM_TOKEN` | `core ci` (npm publisher) | npm registry auth token |
| `CHOCOLATEY_API_KEY` | `core ci` (chocolatey publisher) | Chocolatey API key |
| `DOCKER_USERNAME` | `core ci` (docker publisher) | Docker registry username |
| `DOCKER_PASSWORD` | `core ci` (docker publisher) | Docker registry password |

### Deployment

| Variable | Used By | Description |
|----------|---------|-------------|
| `COOLIFY_URL` | `core php deploy` | Coolify server URL |
| `COOLIFY_TOKEN` | `core php deploy` | Coolify API token |
| `COOLIFY_APP_ID` | `core php deploy` | Production application ID |
| `COOLIFY_STAGING_APP_ID` | `core php deploy --staging` | Staging application ID |

### Build

| Variable | Used By | Description |
|----------|---------|-------------|
| `CGO_ENABLED` | `core build`, `core go *` | Enable/disable CGO (default: 0) |
| `GOOS` | `core build` | Target operating system |
| `GOARCH` | `core build` | Target architecture |

### Configuration Paths

| Variable | Description |
|----------|-------------|
| `CORE_CONFIG` | Override config directory (default: `~/.core/`) |
| `CORE_REGISTRY` | Override repos.yaml path |

---

## Defaults

If no configuration exists, sensible defaults are used:

- **Targets**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- **Publishers**: GitHub only
- **Changelog**: feat, fix, perf, refactor included
