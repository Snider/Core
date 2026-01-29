# core setup

Clone repositories from registry or bootstrap a new workspace.

## Overview

The `setup` command operates in three modes:

1. **Registry mode** - When `repos.yaml` exists nearby, clones repositories into packages/
2. **Bootstrap mode** - When no registry exists, clones `core-devops` first, then presents an interactive wizard to select packages
3. **Repo setup mode** - When run in a git repo root, offers to create `.core/build.yaml` configuration

## Usage

```bash
core setup [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |
| `--dry-run` | Show what would be cloned without cloning |
| `--only` | Only clone repos of these types (comma-separated: foundation,module,product) |
| `--all` | Skip wizard, clone all packages (non-interactive) |
| `--name` | Project directory name for bootstrap mode |
| `--build` | Run build after cloning |

---

## Registry Mode

When `repos.yaml` is found nearby (current directory or parents), setup clones all defined repositories:

```bash
# In a directory with repos.yaml
core setup

# Preview what would be cloned
core setup --dry-run

# Only clone foundation packages
core setup --only foundation

# Multiple types
core setup --only foundation,module
```

In registry mode with a TTY, an interactive wizard allows you to select which packages to clone. Use `--all` to skip the wizard and clone everything.

---

## Bootstrap Mode

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
1. Detects if current directory is empty
2. If not empty, prompts for project name (or uses `--name`)
3. Clones `core-devops` (contains `repos.yaml`)
4. Loads the registry from core-devops
5. Shows interactive package selection wizard (unless `--all`)
6. Clones selected packages
7. Optionally runs build (with `--build`)

---

## Repo Setup Mode

When run in a git repository root (without `repos.yaml`), setup offers two choices:

1. **Setup Working Directory** - Creates `.core/build.yaml` based on detected project type
2. **Create Package** - Creates a subdirectory and clones packages there

```bash
cd ~/Code/my-go-project
core setup

# Output:
# >> This directory is a git repository
# > Setup Working Directory
#   Create Package (clone repos into subdirectory)
```

Choosing "Setup Working Directory" detects the project type and generates configuration:

| Detected File | Project Type |
|---------------|--------------|
| `wails.json` | Wails |
| `go.mod` | Go |
| `composer.json` | PHP |
| `package.json` | Node.js |

Creates three config files in `.core/`:

| File | Purpose |
|------|---------|
| `build.yaml` | Build targets, flags, output settings |
| `release.yaml` | Changelog format, GitHub release config |
| `test.yaml` | Test commands, environment variables |

Also auto-detects GitHub repo from git remote for release config.

See [Configuration Files](example.md#configuration-files) for generated config examples.

---

## Interactive Wizard

When running in a terminal (TTY), the setup command presents an interactive multi-select wizard:

- Packages are grouped by type (foundation, module, product, template)
- Use arrow keys to navigate
- Press space to select/deselect packages
- Type to filter the list
- Press enter to confirm selection

The wizard is skipped when:
- `--all` flag is specified
- Not running in a TTY (e.g., CI pipelines)
- `--dry-run` is specified

---

## Examples

### Clone from Registry

```bash
# Clone all repos (interactive wizard)
core setup

# Clone all repos (non-interactive)
core setup --all

# Preview without cloning
core setup --dry-run

# Only foundation packages
core setup --only foundation
```

### Bootstrap New Workspace

```bash
# Interactive bootstrap in empty directory
mkdir workspace && cd workspace
core setup

# Non-interactive with all packages
core setup --all --name my-project

# Bootstrap and run build
core setup --all --name my-project --build
```

---

## Registry Format

The registry file (`repos.yaml`) defines repositories. See [Configuration Files](example.md#configuration-files) for format.

---

## Finding Registry

Core looks for `repos.yaml` in:

1. Current directory
2. Parent directories (walking up to root)
3. `~/Code/host-uk/repos.yaml`
4. `~/.config/core/repos.yaml`

---

## After Setup

```bash
# Check workspace health
core dev health

# Full workflow (status + commit + push)
core dev work

# Build the project
core build

# Run tests
core go test    # Go projects
core php test   # PHP projects
```

---

## See Also

- [dev work](../dev/work/) - Multi-repo operations
- [build](../build/) - Build projects
- [doctor](../doctor/) - Check environment
