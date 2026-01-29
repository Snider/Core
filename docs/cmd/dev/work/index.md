# core dev work

Multi-repo git operations for managing the host-uk organization.

## Overview

The `core dev work` command and related subcommands help manage multiple repositories in the host-uk ecosystem simultaneously.

## Commands

| Command | Description |
|---------|-------------|
| `core dev work` | Full workflow: status + commit + push |
| `core dev work --status` | Status table only |
| `core dev health` | Quick health check across all repos |
| `core dev issues` | List open issues across all repos |
| `core dev reviews` | List PRs needing review |
| `core dev commit` | Claude-assisted commits across repos |
| `core dev push` | Push commits across all repos |
| `core dev pull` | Pull updates across all repos |
| `core dev impact` | Show impact of changing a repo |

## core dev health

Quick health check showing status of all repos.

```bash
core dev health
```

Output shows:
- Git status (clean/dirty)
- Current branch
- Commits ahead/behind remote

## core dev issues

List open issues across all repositories.

```bash
core dev issues [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--assignee` | Filter by assignee |
| `--label` | Filter by label |
| `--limit` | Max issues per repo |

## core dev reviews

List pull requests needing review.

```bash
core dev reviews [flags]
```

Shows PRs where:
- You are a requested reviewer
- PR is open and not draft
- CI is passing

## core dev commit

Create commits across repos with Claude assistance.

```bash
core dev commit [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--message` | Commit message (auto-generated if not provided) |
| `--all` | Commit in all dirty repos |

Claude analyzes changes and suggests conventional commit messages.

## core dev push

Push commits across all repos.

```bash
core dev push [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--all` | Push all repos with unpushed commits |
| `--force` | Force push (use with caution) |

## core dev pull

Pull updates across all repos.

```bash
core dev pull [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--all` | Pull all repos |
| `--rebase` | Rebase instead of merge |

## core dev impact

Show the impact of changing a repository.

```bash
core dev impact <repo>
```

Shows:
- Dependent repos
- Reverse dependencies
- Potential breaking changes

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

- [setup command](../../setup/) - Clone repos from registry
- [search command](../../pkg/search/) - Find and install repos
