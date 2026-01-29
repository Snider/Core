# core dev ci

Check CI status across all repositories.

Fetches GitHub Actions workflow status for all repos. Shows latest run status for each repo. Requires the `gh` CLI to be installed and authenticated.

## Usage

```bash
core dev ci [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |
| `--branch` | Filter by branch (default: main) |
| `--failed` | Show only failed runs |

## Examples

```bash
# Check CI status for all repos
core dev ci

# Check specific branch
core dev ci --branch develop

# Show only failures
core dev ci --failed
```

## Output

```
core-php         ✓ passing    2m ago
core-tenant      ✓ passing    5m ago
core-admin       ✗ failed     12m ago
core-api         ⏳ running    now
core-bio         ✓ passing    1h ago
```

## Status Icons

| Symbol | Meaning |
|--------|---------|
| `✓` | Passing |
| `✗` | Failed |
| `⏳` | Running |
| `-` | No runs |

## Requirements

- GitHub CLI (`gh`) must be installed
- Must be authenticated: `gh auth login`

## See Also

- [issues command](../issues/) - List open issues
- [reviews command](../reviews/) - List PRs needing review
