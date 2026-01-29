# core dev impact

Show impact of changing a repository.

Analyses the dependency graph to show which repos would be affected by changes to the specified repo.

## Usage

```bash
core dev impact <repo-name> [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |

## Examples

```bash
# Show what depends on core-php
core dev impact core-php

# Show what depends on core-tenant
core dev impact core-tenant
```

## Output

```
Impact of changes to core-php:

Direct dependents (5):
  core-tenant
  core-admin
  core-api
  core-mcp
  core-commerce

Indirect dependents (12):
  core-bio (via core-tenant)
  core-social (via core-tenant)
  core-analytics (via core-tenant)
  core-notify (via core-tenant)
  core-trust (via core-tenant)
  core-support (via core-tenant)
  core-content (via core-tenant)
  core-developer (via core-tenant)
  core-agentic (via core-mcp)
  ...

Total: 17 repos affected
```

## Use Cases

- Before making breaking changes, see what needs updating
- Plan release order based on dependency graph
- Understand the ripple effect of changes

## See Also

- [health command](../health/) - Quick repo status
- [setup command](../../setup/) - Clone repos with dependencies
