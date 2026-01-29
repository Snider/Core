# core dev pull

Pull updates across all repositories.

Pulls updates for all repos. By default only pulls repos that are behind. Use `--all` to pull all repos.

## Usage

```bash
core dev pull [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |
| `--all` | Pull all repos, not just those behind |

## Examples

```bash
# Pull only repos that are behind
core dev pull

# Pull all repos
core dev pull --all

# Use specific registry
core dev pull --registry ~/projects/repos.yaml
```

## Output

```
Pulling 2 repo(s) that are behind:
  ✓ core-php (3 commits)
  ✓ core-tenant (1 commit)

Done: 2 pulled
```

## See Also

- [push command](../push/) - Push local commits
- [health command](../health/) - Check sync status
- [work command](../work/) - Full workflow
