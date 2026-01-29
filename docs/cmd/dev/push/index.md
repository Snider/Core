# core dev push

Push commits across all repositories.

Pushes unpushed commits for all repos. Shows repos with commits to push and confirms before pushing.

## Usage

```bash
core dev push [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |
| `--force` | Skip confirmation prompt |

## Examples

```bash
# Push with confirmation
core dev push

# Push without confirmation
core dev push --force

# Use specific registry
core dev push --registry ~/projects/repos.yaml
```

## Output

```
3 repo(s) with unpushed commits:
  core-php: 2 commit(s)
  core-admin: 1 commit(s)
  core-tenant: 1 commit(s)

Push all? [y/N] y

  ✓ core-php
  ✓ core-admin
  ✓ core-tenant
```

## See Also

- [commit command](../commit/) - Create commits before pushing
- [pull command](../pull/) - Pull updates from remote
- [work command](../work/) - Full workflow (status + commit + push)
