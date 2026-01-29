# core setup

Clone repositories from registry or bootstrap a new workspace.

## Overview

The `setup` command operates in three modes:

1. **Registry mode** - When `repos.yaml` exists nearby, clones repositories into packages/
2. **Bootstrap mode** - When no registry exists, clones `core-devops` first, then presents an interactive wizard to select packages
3. **Repo setup mode** - When run in a git repository root, offers to create `.core/build.yaml` configuration

## Usage

```bash
core setup [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |
| `--dry-run` | Show what would happen without making changes |
| `--only` | Only clone repos of these types (comma-separated: foundation,module,product) |
| `--all` | Skip wizard, clone all packages (non-interactive) |
| `--name` | Project directory name for bootstrap mode |
| `--build` | Run build after cloning |

## Modes

### Registry Mode

When `repos.yaml` is found nearby (current directory or parents), setup clones all defined repositories:

```bash
# In a directory with repos.yaml
core setup

# Preview what would be cloned
core setup --dry-run

# Only clone foundation packages
core setup --only foundation
```

### Bootstrap Mode

When no `repos.yaml` exists, setup enters bootstrap mode:

```bash
# In an empty directory - bootstraps workspace in place
mkdir my-project && cd my-project
core setup

# In a non-empty directory - creates subdirectory
cd ~/Code
core setup --name my-workspace

# Non-interactive: clone all packages
core setup --all --name ci-test
```

Bootstrap mode:
1. Clones `core-devops` (contains `repos.yaml`)
2. Shows interactive package selection wizard
3. Clones selected packages

### Repo Setup Mode

When run in a git repository root, offers to set up the repo with `.core/` configuration:

```bash
# In a git repo without .core/
cd ~/Code/my-go-project
core setup

# Choose "Setup this repo" when prompted
# Creates .core/build.yaml based on detected project type
```

Supported project types:
- **Go** - Detected via `go.mod`
- **Wails** - Detected via `wails.json`
- **Node.js** - Detected via `package.json`
- **PHP** - Detected via `composer.json`

## Examples

### Clone from Registry

```bash
# Clone all repos
core setup

# Preview without cloning
core setup --dry-run

# Only foundation packages
core setup --only foundation

# Multiple types
core setup --only foundation,module
```

### Bootstrap New Workspace

```bash
# Interactive bootstrap
mkdir workspace && cd workspace
core setup

# Non-interactive with all packages
core setup --all --name my-project

# Bootstrap in current directory
core setup --name .
```

### Setup Single Repository

```bash
# In a Go project
cd my-go-project
core setup --dry-run

# Output:
# → Setting up repository configuration
# ✓ Detected project type: go
# → Would create:
#   /path/to/my-go-project/.core/build.yaml
```

## Generated Configuration

When setting up a repository, `core setup` generates `.core/build.yaml`:

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
```

## Registry Format

The registry file (`repos.yaml`) defines repositories:

```yaml
org: host-uk
base_path: .
repos:
  core-php:
    type: foundation
    description: Foundation framework
  core-tenant:
    type: module
    depends_on: [core-php]
    description: Multi-tenancy module
  core-bio:
    type: product
    depends_on: [core-php, core-tenant]
    description: Link-in-bio product
```

## Finding Registry

Core looks for `repos.yaml` in:

1. Current directory
2. Parent directories (walking up to root)
3. `~/Code/host-uk/repos.yaml`
4. `~/.config/core/repos.yaml`

## After Setup

```bash
# Check workspace health
core health

# Full workflow (status + commit + push)
core work

# Build the project
core build

# Run tests
core test
```

## See Also

- [work command](../dev/work/) - Multi-repo operations
- [build command](../build/) - Build projects
- [doctor command](../doctor/) - Check environment
