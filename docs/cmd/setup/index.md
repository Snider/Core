# core setup

Clone all repositories from the registry.

Clones all repositories defined in repos.yaml into packages/. Skips repos that already exist.

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

## Examples

```bash
# Clone all repos from registry
core setup

# Preview what would be cloned
core setup --dry-run

# Only clone foundation packages
core setup --only foundation

# Clone specific types
core setup --only foundation,module
```

## Registry Format

The registry file (`repos.yaml`) defines repositories:

```yaml
repos:
  core:
    type: foundation
    url: https://github.com/host-uk/core
    description: Go CLI for the host-uk ecosystem

  core-php:
    type: foundation
    url: https://github.com/host-uk/core-php
    description: PHP/Laravel packages

  core-tenant:
    type: module
    url: https://github.com/host-uk/core-tenant
    description: Multi-tenancy module
```

## Output

```
Setting up host-uk workspace...

Cloning repositories:
  [1/4] core............... ✓
  [2/4] core-php........... ✓
  [3/4] core-tenant........ ✓
  [4/4] core-admin......... ✓

Done! 4 repositories cloned to packages/
```

## Finding Registry

Core looks for `repos.yaml` in:

1. Current directory
2. Parent directories (up to 5 levels)
3. `~/.core/repos.yaml`

## After Setup

```bash
# Check health of all repos
core dev health

# Pull latest changes
core dev pull --all

# Full workflow (status + commit + push)
core dev work
```

## See Also

- [work commands](../dev/work/) - Multi-repo operations
- [search command](../pkg/search/) - Find and install repos
