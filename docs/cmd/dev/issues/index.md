# core dev issues

List open issues across all repositories.

Fetches open issues from GitHub for all repos in the registry. Requires the `gh` CLI to be installed and authenticated.

## Usage

```bash
core dev issues [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |
| `--assignee` | Filter by assignee (use `@me` for yourself) |
| `--limit` | Max issues per repo (default 10) |

## Examples

```bash
# List all open issues
core dev issues

# Show issues assigned to you
core dev issues --assignee @me

# Limit to 5 issues per repo
core dev issues --limit 5

# Filter by specific assignee
core dev issues --assignee username
```

## Output

```
core-php (3 issues)
  #42  Add retry logic to HTTP client       bug
  #38  Update documentation for v2 API      docs
  #35  Support custom serializers           enhancement

core-tenant (1 issue)
  #12  Workspace isolation bug              bug, critical
```

## Requirements

- GitHub CLI (`gh`) must be installed
- Must be authenticated: `gh auth login`

## See Also

- [reviews command](../reviews/) - List PRs needing review
- [ci command](../ci/) - Check CI status
