# core pkg

Package management for host-uk repositories.

## Usage

```bash
core pkg <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| [`search`](#pkg-search) | Search GitHub for packages |
| [`install`](#pkg-install) | Clone a package from GitHub |
| [`list`](#pkg-list) | List installed packages |
| [`update`](#pkg-update) | Update installed packages |
| [`outdated`](#pkg-outdated) | Check for outdated packages |

---

## pkg search

Search GitHub for host-uk packages.

```bash
core pkg search [flags]
```

Results are cached for 1 hour in `.core/cache/`.

### Flags

| Flag | Description |
|------|-------------|
| `--org` | GitHub organisation (default: host-uk) |
| `--pattern` | Repo name pattern (* for wildcard) |
| `--type` | Filter by type in name (mod, services, plug, website) |
| `--limit` | Max results (default: 50) |
| `--refresh` | Bypass cache and fetch fresh data |

### Examples

```bash
# List all repos in org
core pkg search

# Search for core-* repos
core pkg search --pattern 'core-*'

# Search different org
core pkg search --org mycompany

# Bypass cache
core pkg search --refresh
```

---

## pkg install

Clone a package from GitHub.

```bash
core pkg install <org/repo> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--dir` | Target directory (default: ./packages or current dir) |
| `--add` | Add to repos.yaml registry |

### Examples

```bash
# Clone to packages/
core pkg install host-uk/core-php

# Clone to custom directory
core pkg install host-uk/core-tenant --dir ./packages

# Clone and add to registry
core pkg install host-uk/core-admin --add
```

---

## pkg list

List installed packages from repos.yaml.

```bash
core pkg list
```

Shows installed status (✓) and description for each package.

---

## pkg update

Pull latest changes for installed packages.

```bash
core pkg update [<name>...] [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--all` | Update all packages |

### Examples

```bash
# Update specific package
core pkg update core-php

# Update all packages
core pkg update --all
```

---

## pkg outdated

Check which packages have unpulled commits.

```bash
core pkg outdated
```

Fetches from remote and shows packages that are behind.

---

## See Also

- [setup](../setup/) - Clone all repos from registry
- [dev work](../dev/work/) - Multi-repo workflow
