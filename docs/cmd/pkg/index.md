# core pkg

Package management for host-uk/core-* repos.

## Usage

```bash
core pkg <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| [search](search/) | Search GitHub for packages |
| `install` | Clone a package from GitHub |
| `list` | List installed packages |
| `update` | Update installed packages |
| `outdated` | Check for outdated packages |

## pkg search

Search GitHub for host-uk packages.

```bash
core pkg search <query> [flags]
```

See [search](search/) for details.

## pkg install

Clone a package from GitHub.

```bash
core pkg install <repo> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--add` | Add to repos.yaml registry |

### Examples

```bash
core pkg install core-php
core pkg install core-tenant --add
```

## pkg list

List installed packages.

```bash
core pkg list
```

## pkg update

Update installed packages.

```bash
core pkg update
```

## pkg outdated

Check for outdated packages.

```bash
core pkg outdated
```
