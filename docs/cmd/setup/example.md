# Setup Examples

## Clone from Registry

```bash
# Clone all repos defined in repos.yaml
core setup

# Preview what would be cloned
core setup --dry-run

# Only foundation packages
core setup --only foundation

# Multiple types
core setup --only foundation,module

# Use specific registry file
core setup --registry ~/projects/repos.yaml
```

## Bootstrap New Workspace

```bash
# In an empty directory - bootstraps in place
mkdir my-workspace && cd my-workspace
core setup

# Shows interactive wizard to select packages:
# ┌─────────────────────────────────────────────┐
# │ Select packages to clone                     │
# │ Use space to select, enter to confirm        │
# │                                              │
# │ ── Foundation (core framework) ──            │
# │ ☑ core-php         Foundation framework      │
# │ ☑ core-tenant      Multi-tenancy module      │
# │                                              │
# │ ── Products (applications) ──                │
# │ ☐ core-bio         Link-in-bio product       │
# │ ☐ core-social      Social scheduling         │
# └─────────────────────────────────────────────┘

# Non-interactive: clone all packages
core setup --all

# Create workspace in subdirectory
cd ~/Code
core setup --name my-project

# CI mode: fully non-interactive
core setup --all --name ci-test
```

## Setup Single Repository

```bash
# In a git repo without .core/ configuration
cd ~/Code/my-go-project
core setup

# Shows choice dialog:
# ┌─────────────────────────────────────────────┐
# │ Setup options                                │
# │ You're in a git repository. What would you  │
# │ like to do?                                  │
# │                                              │
# │ ● Setup this repo (create .core/ config)    │
# │ ○ Create a new workspace (clone repos)      │
# └─────────────────────────────────────────────┘

# Preview generated configuration
core setup --dry-run

# Output:
# → Setting up repository configuration
#
# ✓ Detected project type: go
#   → Also found: (none)
#
# → Would create:
#   /Users/you/Code/my-go-project/.core/build.yaml
#
# Configuration preview:
#   version: 1
#   project:
#     name: my-go-project
#     description: Go application
#     main: ./cmd/my-go-project
#     binary: my-go-project
#   ...
```

## Configuration Files

### repos.yaml (Workspace Registry)

```yaml
org: host-uk
base_path: .
defaults:
  ci: github
  license: EUPL-1.2
  branch: main
repos:
  core-php:
    type: foundation
    description: Foundation framework
  core-tenant:
    type: module
    depends_on: [core-php]
    description: Multi-tenancy module
  core-admin:
    type: module
    depends_on: [core-php, core-tenant]
    description: Admin panel
  core-bio:
    type: product
    depends_on: [core-php, core-tenant]
    description: Link-in-bio product
    domain: bio.host.uk.com
  core-devops:
    type: foundation
    clone: false  # Already exists, skip cloning
```

### .core/build.yaml (Repository Config)

Generated for Go projects:

```yaml
version: 1
project:
  name: my-project
  description: Go application
  main: ./cmd/my-project
  binary: my-project
build:
  cgo: false
  flags:
    - -trimpath
  ldflags:
    - -s
    - -w
  env: []
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
sign:
  enabled: false
```

Generated for Wails projects:

```yaml
version: 1
project:
  name: my-app
  description: Wails desktop application
  main: .
  binary: my-app
targets:
  - os: darwin
    arch: amd64
  - os: darwin
    arch: arm64
  - os: windows
    arch: amd64
  - os: linux
    arch: amd64
```

### .core/release.yaml (Release Config)

Generated for Go projects:

```yaml
version: 1
project:
  name: my-project
  repository: owner/my-project

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

publishers:
  - type: github
    draft: false
    prerelease: false
```

### .core/test.yaml (Test Config)

Generated for Go projects:

```yaml
version: 1

commands:
  - name: unit
    run: go test ./...
  - name: coverage
    run: go test -coverprofile=coverage.out ./...
  - name: race
    run: go test -race ./...

env:
  CGO_ENABLED: "0"
```

Generated for PHP projects:

```yaml
version: 1

commands:
  - name: unit
    run: vendor/bin/pest --parallel
  - name: types
    run: vendor/bin/phpstan analyse
  - name: lint
    run: vendor/bin/pint --test

env:
  APP_ENV: testing
  DB_CONNECTION: sqlite
```

Generated for Node.js projects:

```yaml
version: 1

commands:
  - name: unit
    run: npm test
  - name: lint
    run: npm run lint
  - name: typecheck
    run: npm run typecheck

env:
  NODE_ENV: test
```

## Workflow Examples

### New Developer Setup

```bash
# Clone the workspace
mkdir host-uk && cd host-uk
core setup

# Select packages in wizard, then:
core health        # Check all repos are healthy
core doctor        # Verify environment
```

### CI Pipeline Setup

```bash
# Non-interactive full clone
core setup --all --name workspace

# Or with specific packages
core setup --only foundation,module --name workspace
```

### Adding Build Config to Existing Repo

```bash
cd my-existing-project
core setup              # Choose "Setup this repo"
# Edit .core/build.yaml as needed
core build              # Build the project
```
