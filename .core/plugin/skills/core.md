---
name: core
description: Use when working in host-uk repositories, running tests, building, releasing, or managing multi-repo workflows. Provides the core CLI command reference.
---

# Core CLI

The `core` command provides a unified interface for Go/PHP development and multi-repo management.

**Rule:** Always prefer `core <command>` over raw commands.

## Quick Reference

| Task | Command |
|------|---------|
| Go tests | `core go test` |
| Go coverage | `core go cov` |
| Go format | `core go fmt --fix` |
| Go lint | `core go lint` |
| PHP dev server | `core php dev` |
| PHP tests | `core php test` |
| PHP format | `core php fmt --fix` |
| Build | `core build` |
| Preview release | `core ci` |
| Publish | `core ci --were-go-for-launch` |
| Multi-repo status | `core dev health` |
| Commit dirty repos | `core dev commit` |
| Push repos | `core dev push` |

## Decision Tree

```
Go project?
  tests: core go test
  format: core go fmt --fix
  build: core build

PHP project?
  dev: core php dev
  tests: core php test
  format: core php fmt --fix
  deploy: core php deploy

Multiple repos?
  status: core dev health
  commit: core dev commit
  push: core dev push
```

## Common Mistakes

| Wrong | Right |
|-------|-------|
| `go test ./...` | `core go test` |
| `go build` | `core build` |
| `php artisan serve` | `core php dev` |
| `./vendor/bin/pest` | `core php test` |
| `git status` per repo | `core dev health` |

Run `core --help` or `core <cmd> --help` for full options.
