# core pkg search

Search GitHub for repositories matching a pattern.

Uses `gh` CLI for authenticated search. Results are cached for 1 hour.

## Usage

```bash
core pkg search [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--pattern` | Repo name pattern (* for wildcard) |
| `--org` | GitHub organization (default: host-uk) |
| `--type` | Filter by type in name (mod, services, plug, website) |
| `--limit` | Max results (default: 50) |
| `--refresh` | Bypass cache and fetch fresh data |

## Examples

```bash
# List all host-uk repos
core pkg search

# Search for core-* repos
core pkg search --pattern "core-*"

# Search different org
core pkg search --org mycompany

# Filter by type
core pkg search --type services

# Bypass cache
core pkg search --refresh

# Combine filters
core pkg search --pattern "core-*" --type mod --limit 20
```

## Output

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

## Authentication

Uses GitHub CLI (`gh`) authentication. Ensure you're logged in:

```bash
gh auth status
gh auth login  # if not authenticated
```

## See Also

- [pkg install](../) - Clone a package from GitHub
- [setup command](../../setup/) - Clone all repos from registry
