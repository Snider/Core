# Core CLI

Multi-repository management tool for the Core Framework ecosystem.

## Installation

```bash
go install github.com/Snider/Core/cmd/core@latest
```

Or build from source:

```bash
cd cmd/core && go build .
```

## Configuration

The CLI looks for `repos.yaml` in:
1. Current directory and parent directories
2. `~/Code/host-uk/repos.yaml`
3. `~/.config/core/repos.yaml`

If no `repos.yaml` is found, it scans the current directory for git repositories.

### repos.yaml format

```yaml
version: 1
org: host-uk
base_path: ~/Code/host-uk

repos:
  core-php:
    type: foundation
    description: Core PHP framework
    depends_on: []
    docs: true

  core-tenant:
    type: module
    depends_on: [core-php]
    description: Multi-tenancy support

defaults:
  ci: github-actions
  branch: main
```

## Commands

### Health Check

Quick summary of repository health:

```bash
core health              # Summary: "18 repos │ clean │ synced │ up to date"
core health --verbose    # Include list of dirty/ahead repos
```

### Work (All-in-One)

Status, commit, and push workflow:

```bash
core work                # Show status table + push repos with commits
core work --status       # Status only, no push
core work --commit       # Use Claude to commit dirty repos first
```

### Commit

Claude-assisted commits across repositories:

```bash
core commit              # Show dirty repos, confirm, Claude commits each
core commit --all        # Skip confirmation
```

Reads `AGENTS.md` from the registry directory for commit context.

### Push

Push commits to remote:

```bash
core push                # Show repos to push, confirm, push all
core push --force        # Skip confirmation
```

### Pull

Pull updates from remote:

```bash
core pull                # Pull only repos that are behind
core pull --all          # Pull all repos
```

### Impact Analysis

Show dependency impact of changing a repository:

```bash
core impact core-php     # Show what depends on core-php
core impact core-tenant  # Show direct and transitive dependents
```

Output:
```
Impact analysis for core-php
Core PHP framework - events, modules, lifecycle

● 16 direct dependent(s):
    core-admin
    core-tenant
    ...

Summary: Changes to core-php affect 16/17 repos
```

### Issues

List open GitHub issues across repositories:

```bash
core issues                    # All open issues
core issues --assignee @me     # Filter by assignee
core issues --limit 5          # Limit per repo
```

### Reviews

List open PRs with review status:

```bash
core reviews              # All open PRs
core reviews --author me  # Filter by author
core reviews --all        # Include draft PRs
```

### CI Status

Check GitHub Actions workflow status:

```bash
core ci                  # Status for all repos (main branch)
core ci --branch dev     # Check specific branch
core ci --failed         # Show only failing runs
```

### Documentation

Manage documentation across repositories:

```bash
core docs list                    # Show docs coverage
core docs sync                    # Sync to ./docs-build
core docs sync --dry-run          # Preview without copying
core docs sync --output ./site    # Custom output directory
```

### Setup

Bootstrap a workspace by cloning all repos from registry:

```bash
core setup                        # Clone all repos into packages/
core setup --dry-run              # Show what would be cloned
core setup --only module,product  # Filter by type
```

### Doctor

Check development environment:

```bash
core doctor                       # Check all requirements
core doctor --verbose             # Show version details
```

Checks: git, gh, php, composer, node, SSH keys, gh auth status.

### Search

Search GitHub for repositories:

```bash
core search --org host-uk                      # List all repos in org
core search --org host-uk --pattern 'core-*'   # Filter by name pattern
core search --org host-uk --type service       # Filter by type in name
core search --org host-uk --refresh            # Bypass cache
```

Results are cached in `.core/cache/` for 1 hour.

### Install

Install individual repositories:

```bash
core install --repo host-uk/core-php           # Clone to packages/
core install --repo host-uk/core-mcp --add     # Clone and add to repos.yaml
core install --repo host-uk/core-api --dir .   # Clone to current directory
```

Uses `gh` with HTTPS (no SSH key required).

## Global Flags

All commands support:

```bash
--registry <path>    # Explicit path to repos.yaml
```

## Requirements

- Go 1.25+
- `gh` CLI for GitHub operations (issues, reviews, ci)
- `claude` CLI for AI-assisted commits

## Cache

API responses are cached in the workspace `.core/` directory:

```
<workspace>/
├── .core/
│   └── cache/
│       └── github/
│           └── <org>/
│               └── repos.json    # Cached repo list (1 hour TTL)
├── repos.yaml
└── packages/
```

The `.core/` directory should be gitignored.

## Architecture

```
cmd/core/
├── cmd/
│   ├── root.go       # CLI setup
│   ├── work.go       # work command
│   ├── health.go     # health command
│   ├── commit.go     # commit command
│   ├── push.go       # push command
│   ├── pull.go       # pull command
│   ├── impact.go     # impact command
│   ├── issues.go     # issues command
│   ├── reviews.go    # reviews command
│   ├── ci.go         # ci command
│   ├── docs.go       # docs command
│   ├── setup.go      # setup command
│   ├── doctor.go     # doctor command
│   ├── search.go     # search command
│   └── install.go    # install command
├── go.mod
└── main.go

pkg/
├── cache/
│   └── cache.go      # File-based cache with TTL
├── git/
│   └── git.go        # Parallel git operations
└── repos/
    └── registry.go   # repos.yaml parser + dependency graph
```
