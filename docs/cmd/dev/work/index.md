# core dev work

Multi-repo git operations for managing the host-uk organization.

## Overview

The `core dev work` command and related subcommands help manage multiple repositories in the host-uk ecosystem simultaneously.

## Commands

| Command | Description |
|---------|-------------|
| `core dev work` | Full workflow: status + commit + push |
| `core dev work --status` | Status table only |
| `core dev work --commit` | Use Claude to commit dirty repos |
| `core dev health` | Quick health check across all repos |
| `core dev commit` | Claude-assisted commits across repos |
| `core dev push` | Push commits across all repos |
| `core dev pull` | Pull updates across all repos |
| `core dev issues` | List open issues across all repos |
| `core dev reviews` | List PRs needing review |
| `core dev ci` | Check CI status across all repos |
| `core dev impact` | Show impact of changing a repo |

## core dev work

Manage git status, commits, and pushes across multiple repositories.

```bash
core dev work [flags]
```

Reads `repos.yaml` to discover repositories and their relationships. Shows status, optionally commits with Claude, and pushes changes.

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |
| `--status` | Show status only, don't push |
| `--commit` | Use Claude to commit dirty repos before pushing |

### Examples

```bash
# Full workflow
core dev work

# Status only
core dev work --status

# Commit and push
core dev work --commit
```

## core dev health

Quick health check showing summary of repository health across all repos.

```bash
core dev health [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |
| `--verbose` | Show detailed breakdown |

Output shows:
- Total repos
- Dirty repos
- Unpushed commits
- Repos behind remote

### Examples

```bash
# Quick summary
core dev health

# Detailed breakdown
core dev health --verbose
```

## core dev issues

List open issues across all repositories.

```bash
core dev issues [flags]
```

Fetches open issues from GitHub for all repos in the registry. Requires the `gh` CLI to be installed and authenticated.

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |
| `--assignee` | Filter by assignee (use `@me` for yourself) |
| `--limit` | Max issues per repo (default: 10) |

### Examples

```bash
# List all open issues
core dev issues

# Filter by assignee
core dev issues --assignee @me

# Limit results
core dev issues --limit 5
```

## core dev reviews

List pull requests needing review across all repos.

```bash
core dev reviews [flags]
```

Fetches open PRs from GitHub for all repos in the registry. Shows review status (approved, changes requested, pending). Requires the `gh` CLI to be installed and authenticated.

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |
| `--all` | Show all PRs including drafts |
| `--author` | Filter by PR author |

### Examples

```bash
# List PRs needing review
core dev reviews

# Show all PRs including drafts
core dev reviews --all

# Filter by author
core dev reviews --author username
```

## core dev commit

Create commits across repos with Claude assistance.

```bash
core dev commit [flags]
```

Uses Claude to create commits for dirty repos. Shows uncommitted changes and invokes Claude to generate commit messages.

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |
| `--all` | Commit all dirty repos without prompting |

### Examples

```bash
# Commit with prompts
core dev commit

# Commit all automatically
core dev commit --all
```

## core dev push

Push commits across all repos.

```bash
core dev push [flags]
```

Pushes unpushed commits for all repos. Shows repos with commits to push and confirms before pushing.

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |
| `--force` | Skip confirmation prompt |

### Examples

```bash
# Push with confirmation
core dev push

# Skip confirmation
core dev push --force
```

## core dev pull

Pull updates across all repos.

```bash
core dev pull [flags]
```

Pulls updates for all repos. By default only pulls repos that are behind. Use `--all` to pull all repos.

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |
| `--all` | Pull all repos, not just those behind |

### Examples

```bash
# Pull repos that are behind
core dev pull

# Pull all repos
core dev pull --all
```

## core dev ci

Check GitHub Actions workflow status across all repos.

```bash
core dev ci [flags]
```

Fetches GitHub Actions workflow status for all repos. Shows latest run status for each repo. Requires the `gh` CLI to be installed and authenticated.

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |
| `--branch` | Filter by branch (default: main) |
| `--failed` | Show only failed runs |

### Examples

```bash
# Show CI status for all repos
core dev ci

# Show only failed runs
core dev ci --failed

# Check specific branch
core dev ci --branch develop
```

## core dev impact

Show the impact of changing a repository.

```bash
core dev impact <repo> [flags]
```

Analyzes the dependency graph to show which repos would be affected by changes to the specified repo.

### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |

### Examples

```bash
# Show impact of changing core-php
core dev impact core-php
```

## Registry

These commands use `repos.yaml` to know which repos to manage. See [repos.yaml](../../../configuration.md#reposyaml) for format.

Use `core setup` to clone all repos from the registry.

## See Also

- [setup command](../../setup/) - Clone repos from registry
- [search command](../../pkg/search/) - Find and install repos
