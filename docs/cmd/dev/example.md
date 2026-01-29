# Dev Examples

## Multi-Repo Workflow

```bash
# Quick status
core dev health

# Full workflow
core dev work

# Status only
core dev work --status

# Commit dirty repos
core dev commit

# Push unpushed
core dev push

# Pull behind
core dev pull
```

## GitHub Integration

```bash
# Open issues
core dev issues

# Include closed
core dev issues --all

# PRs needing review
core dev reviews
```

## Dependency Analysis

```bash
# What depends on core-php?
core dev impact core-php
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

`repos.yaml`:
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

`~/.core/config.yaml`:
```yaml
version: 1
images:
  source: auto
```
