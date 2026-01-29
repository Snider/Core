# core dev reviews

List PRs needing review across all repositories.

Fetches open PRs from GitHub for all repos in the registry. Shows review status (approved, changes requested, pending). Requires the `gh` CLI to be installed and authenticated.

## Usage

```bash
core dev reviews [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |
| `--all` | Show all PRs including drafts |
| `--author` | Filter by PR author |

## Examples

```bash
# List PRs needing review
core dev reviews

# Include draft PRs
core dev reviews --all

# Filter by author
core dev reviews --author username
```

## Output

```
core-php (2 PRs)
  #45  feat: Add caching layer          ✓ approved      @alice
  #43  fix: Memory leak in worker       ⏳ pending       @bob

core-admin (1 PR)
  #28  refactor: Extract components     ✗ changes       @charlie
```

## Review Status

| Symbol | Meaning |
|--------|---------|
| `✓` | Approved |
| `⏳` | Pending review |
| `✗` | Changes requested |

## Requirements

- GitHub CLI (`gh`) must be installed
- Must be authenticated: `gh auth login`

## See Also

- [issues command](../issues/) - List open issues
- [ci command](../ci/) - Check CI status
