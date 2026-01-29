# core dev health

Quick health check across all repositories.

Shows a summary of repository health: total repos, dirty repos, unpushed commits, etc.

## Usage

```bash
core dev health [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |
| `--verbose` | Show detailed breakdown |

## Examples

```bash
# Quick health summary
core dev health

# Detailed breakdown
core dev health --verbose

# Use specific registry
core dev health --registry ~/projects/repos.yaml
```

## Output

```
18 repos │ 2 dirty │ 1 ahead │ all synced
```

With `--verbose`:

```
Repos:     18
Dirty:     2 (core-php, core-admin)
Ahead:     1 (core-tenant)
Behind:    0
Synced:    ✓
```

## See Also

- [work command](../work/) - Full workflow (status + commit + push)
- [commit command](../commit/) - Claude-assisted commits
