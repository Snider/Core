# core search & install

Search GitHub for repositories and install them locally.

## core search

Search GitHub for repositories matching a pattern.

```bash
core search <pattern> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--org` | Search within a specific organization |
| `--limit` | Maximum results (default: 10) |
| `--language` | Filter by programming language |

### Examples

```bash
# Search by pattern
core search "cli tool"

# Search within organization
core search --org host-uk

# Search with language filter
core search --org host-uk --language go

# Search all core-* repos
core search "core-" --org host-uk
```

### Output

```
Found 5 repositories:

  host-uk/core
    Go CLI for the host-uk ecosystem
    ★ 42  Go  Updated 2 hours ago

  host-uk/core-php
    PHP/Laravel packages for Core
    ★ 18  PHP  Updated 1 day ago

  host-uk/core-images
    Docker and LinuxKit images
    ★ 8  Dockerfile  Updated 3 days ago
```

## core install

Clone a repository from GitHub.

```bash
core install <repo> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--path` | Destination directory (default: current dir) |
| `--branch` | Clone specific branch |
| `--depth` | Shallow clone depth |

### Examples

```bash
# Install by full name
core install host-uk/core

# Install to specific path
core install host-uk/core --path ~/Code/host-uk

# Install specific branch
core install host-uk/core --branch dev

# Shallow clone
core install host-uk/core --depth 1
```

### Authentication

Uses GitHub CLI (`gh`) authentication. Ensure you're logged in:

```bash
gh auth status
gh auth login  # if not authenticated
```

## Workflow Example

```bash
# Find repositories
core search --org host-uk

# Install one
core install host-uk/core-php --path ~/Code/host-uk

# Check setup
core doctor
```

## See Also

- [setup command](setup.md) - Clone all repos from registry
- [doctor command](doctor.md) - Check environment
