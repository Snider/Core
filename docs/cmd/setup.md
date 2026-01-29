# core setup

Clone all repositories from the registry.

## Usage

```bash
core setup [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml |
| `--path` | Base directory for cloning (default: current dir) |
| `--ssh` | Use SSH URLs instead of HTTPS |

## Examples

```bash
# Clone all repos from registry
core setup

# Clone to specific directory
core setup --path ~/Code/host-uk

# Use SSH for cloning
core setup --ssh
```

## Registry Format

The registry file (`repos.yaml`) defines repositories:

```yaml
repos:
  - name: core
    url: https://github.com/host-uk/core
    description: Go CLI for the host-uk ecosystem

  - name: core-php
    url: https://github.com/host-uk/core-php
    description: PHP/Laravel packages

  - name: core-images
    url: https://github.com/host-uk/core-images
    description: Docker and LinuxKit images

  - name: core-api
    url: https://github.com/host-uk/core-api
    description: API service
```

## Output

```
Setting up host-uk workspace...

Cloning repositories:
  [1/4] core............... ✓
  [2/4] core-php........... ✓
  [3/4] core-images........ ✓
  [4/4] core-api........... ✓

Done! 4 repositories cloned to ~/Code/host-uk
```

## Finding Registry

Core looks for `repos.yaml` in:

1. Current directory
2. Parent directories (up to 5 levels)
3. `~/.core/repos.yaml`

## After Setup

```bash
# Check health of all repos
core health

# Pull latest changes
core pull --all

# Check CI status
core ci
```

## See Also

- [work commands](work.md) - Multi-repo operations
- [search command](search.md) - Find repos on GitHub
- [install command](search.md) - Clone individual repos
