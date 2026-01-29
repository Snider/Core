# core work

Multi-repo git operations for managing the host-uk organization.

## Overview

The `work` command and related commands (`health`, `issues`, `reviews`, `commit`, `push`, `pull`, `impact`, `ci`) help manage multiple repositories in the host-uk ecosystem simultaneously.

## Commands

| Command | Description |
|---------|-------------|
| `core work` | Multi-repo git operations |
| `core health` | Quick health check across all repos |
| `core issues` | List open issues across all repos |
| `core reviews` | List PRs needing review |
| `core commit` | Claude-assisted commits across repos |
| `core push` | Push commits across all repos |
| `core pull` | Pull updates across all repos |
| `core impact` | Show impact of changing a repo |
| `core ci` | Check CI status across all repos |

## core health

Quick health check showing status of all repos.

```bash
core health
```

Output shows:
- Git status (clean/dirty)
- Current branch
- Commits ahead/behind remote
- CI status

## core issues

List open issues across all repositories.

```bash
core issues [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--assignee` | Filter by assignee |
| `--label` | Filter by label |
| `--limit` | Max issues per repo |

## core reviews

List pull requests needing review.

```bash
core reviews [flags]
```

Shows PRs where:
- You are a requested reviewer
- PR is open and not draft
- CI is passing

## core commit

Create commits across repos with Claude assistance.

```bash
core commit [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--message` | Commit message (auto-generated if not provided) |
| `--all` | Commit in all dirty repos |

Claude analyzes changes and suggests conventional commit messages.

## core push

Push commits across all repos.

```bash
core push [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--all` | Push all repos with unpushed commits |
| `--force` | Force push (use with caution) |

## core pull

Pull updates across all repos.

```bash
core pull [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--all` | Pull all repos |
| `--rebase` | Rebase instead of merge |

## core impact

Show the impact of changing a repository.

```bash
core impact <repo>
```

Shows:
- Dependent repos
- Reverse dependencies
- Potential breaking changes

## core ci

Check CI status across all repos.

```bash
core ci [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--watch` | Watch for status changes |
| `--failing` | Show only failing repos |

## Registry

These commands use `repos.yaml` to know which repos to manage:

```yaml
repos:
  - name: core
    path: ./core
    url: https://github.com/host-uk/core
  - name: core-php
    path: ./core-php
    url: https://github.com/host-uk/core-php
```

Use `core setup` to clone all repos from the registry.

## See Also

- [setup command](setup.md) - Clone repos from registry
- [search command](search.md) - Find and install repos
