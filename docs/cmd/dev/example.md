# Dev Examples

## Multi-Repo Workflow

```bash
# Quick status
core dev health

# Detailed breakdown
core dev health --verbose

# Full workflow
core dev work

# Status only
core dev work --status

# Commit and push
core dev work --commit

# Commit dirty repos
core dev commit

# Commit all without prompting
core dev commit --all

# Push unpushed
core dev push

# Push without confirmation
core dev push --force

# Pull behind repos
core dev pull

# Pull all repos
core dev pull --all
```

## GitHub Integration

```bash
# Open issues
core dev issues

# Filter by assignee
core dev issues --assignee @me

# Limit results
core dev issues --limit 5

# PRs needing review
core dev reviews

# All PRs including drafts
core dev reviews --all

# Filter by author
core dev reviews --author username

# CI status
core dev ci

# Only failed runs
core dev ci --failed

# Specific branch
core dev ci --branch develop
```

## Dependency Analysis

```bash
# What depends on core-php?
core dev impact core-php
```

## Task Management

```bash
# List tasks
core ai tasks

# Filter by status and priority
core ai tasks --status pending --priority high

# Filter by labels
core ai tasks --labels bug,urgent

# Show task details
core ai task abc123

# Auto-select highest priority task
core ai task --auto

# Claim a task
core ai task abc123 --claim

# Update task status
core ai task:update abc123 --status in_progress

# Add progress notes
core ai task:update abc123 --progress 50 --notes 'Halfway done'

# Complete a task
core ai task:complete abc123 --output 'Feature implemented'

# Mark as failed
core ai task:complete abc123 --failed --error 'Build failed'

# Commit with task reference
core ai task:commit abc123 -m 'add user authentication'

# Commit with scope and push
core ai task:commit abc123 -m 'fix login bug' --scope auth --push

# Create PR for task
core ai task:pr abc123

# Create draft PR with labels
core ai task:pr abc123 --draft --labels 'enhancement,needs-review'
```

## Service API Management

```bash
# Synchronize public service APIs
core dev sync

# Or using the api command
core dev api sync
```

## Dev Environment

```bash
# First time setup
core dev install
core dev boot

# Open shell
core dev shell

# Mount and serve
core dev serve

# Run tests
core dev test

# Sandboxed Claude
core dev claude
```

## Configuration

### repos.yaml

```yaml
org: host-uk
repos:
  core-php:
    type: package
    description: Foundation framework
  core-tenant:
    type: package
    depends: [core-php]
```

### ~/.core/config.yaml

```yaml
version: 1

images:
  source: auto  # auto | github | registry | cdn

  cdn:
    url: https://images.example.com/core-devops

  github:
    repo: host-uk/core-images

  registry:
    image: ghcr.io/host-uk/core-devops
```

### .core/test.yaml

```yaml
version: 1

commands:
  - name: unit
    run: vendor/bin/pest --parallel
  - name: types
    run: vendor/bin/phpstan analyse
  - name: lint
    run: vendor/bin/pint --test

env:
  APP_ENV: testing
  DB_CONNECTION: sqlite
```
