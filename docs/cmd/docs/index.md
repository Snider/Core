# core docs

Documentation management across repositories.

## Usage

```bash
core docs <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List documentation across repos |
| `sync` | Sync documentation to output directory |

## docs list

Show documentation coverage across all repos.

```bash
core docs list [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml |

### Output

```
Repo                  README    CLAUDE    CHANGELOG   docs/
──────────────────────────────────────────────────────────────────────
core                  ✓         ✓         —           12 files
core-php              ✓         ✓         ✓           8 files
core-images           ✓         —         —           —

Coverage: 3 with docs, 0 without
```

## docs sync

Sync documentation from all repos to an output directory.

```bash
core docs sync [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml |
| `--output` | Output directory (default: ./docs-build) |
| `--dry-run` | Show what would be synced |

### Output Structure

```
docs-build/
└── packages/
    ├── core/
    │   ├── index.md      # from README.md
    │   ├── claude.md     # from CLAUDE.md
    │   ├── changelog.md  # from CHANGELOG.md
    │   ├── build.md      # from docs/build.md
    │   └── ...
    └── core-php/
        ├── index.md
        └── ...
```

### Example

```bash
# Preview what will be synced
core docs sync --dry-run

# Sync to default output
core docs sync

# Sync to custom directory
core docs sync --output ./site/content
```

## What Gets Synced

For each repo, the following files are collected:

| Source | Destination |
|--------|-------------|
| `README.md` | `index.md` |
| `CLAUDE.md` | `claude.md` |
| `CHANGELOG.md` | `changelog.md` |
| `docs/*.md` | `*.md` |

## Integration with core.help

The synced docs are used to build https://core.help:

1. Run `core docs sync --output ../core-php/docs/packages`
2. VitePress builds the combined documentation
3. Deploy to core.help

## See Also

- [Configuration](../configuration.md) - Project configuration
