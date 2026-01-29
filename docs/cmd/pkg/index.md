# core pkg

Package management for Go modules.

## Usage

```bash
core pkg <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| [search](search/) | Search packages on GitHub |
| `install` | Install a package |
| `list` | List installed packages |
| `update` | Update packages |
| `outdated` | Check for outdated packages |

## pkg install

Install a Go module.

```bash
core pkg install <module> [flags]
```

### Examples

```bash
core pkg install github.com/host-uk/core-php
core pkg install github.com/spf13/cobra@latest
```

## pkg list

List installed packages.

```bash
core pkg list
```

## pkg update

Update packages to latest versions.

```bash
core pkg update [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--all` | Update all packages |

## pkg outdated

Check for outdated packages.

```bash
core pkg outdated
```
