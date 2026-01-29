# Documentation TODO

Commands and flags found in CLI but missing from documentation.

## Missing Commands

### core dev

- `core dev api` - Tools for managing service APIs
- `core dev api sync` - Synchronizes the public service APIs with their internal implementations
- `core dev sync` - Synchronizes the public service APIs (duplicate of api sync)
- `core dev ci` - Check CI status across all repos
- `core dev tasks` - List available tasks from core-agentic
- `core dev task` - Show task details or auto-select a task
- `core dev task:update` - Update task status or progress
- `core dev task:complete` - Mark a task as completed
- `core dev task:commit` - Auto-commit changes with task reference
- `core dev task:pr` - Create a pull request for a task

### core go

- `core go work` - Workspace management (init, sync, use)
- `core go work sync` - Sync workspace
- `core go work init` - Initialize workspace
- `core go work use` - Add module to workspace

### core build

- `core build from-path` - Build from a local directory
- `core build pwa` - Build from a live PWA URL

### core php

- `core php packages link` - Link local packages (subcommand documentation exists but not detailed)
- `core php packages unlink` - Unlink packages
- `core php packages update` - Update linked packages
- `core php packages list` - List linked packages

### core vm

- `core vm templates show` - Display template content
- `core vm templates vars` - Show template variables

## Missing Flags

### core dev boot

- `--fresh` - Stop existing and start fresh

### core dev claude

- `--model` - Model to use (opus, sonnet)

### core dev install

- Docs mention `--source` and `--force` flags that don't appear in CLI help

### core dev tasks

- `--status` - Filter by status (pending, in_progress, completed, blocked)
- `--priority` - Filter by priority (critical, high, medium, low)
- `--labels` - Filter by labels (comma-separated)
- `--project` - Filter by project
- `--limit` - Max number of tasks to return (default 20)

### core dev task

- `--auto` - Auto-select highest priority pending task
- `--claim` - Claim the task after showing details
- `--context` - Show gathered context for AI collaboration

### core dev task:update

- `--status` - New status (pending, in_progress, completed, blocked)
- `--progress` - Progress percentage (0-100)
- `--notes` - Notes about the update

### core dev task:complete

- `--output` - Summary of the completed work
- `--failed` - Mark the task as failed
- `--error` - Error message if failed

### core dev task:commit

- `--message` / `-m` - Commit message
- `--scope` - Scope for the commit type
- `--push` - Push changes after committing

### core dev task:pr

- `--title` - PR title
- `--draft` - Create as draft PR
- `--labels` - Labels to add (comma-separated)
- `--base` - Base branch (defaults to main)

### core dev health

- `--verbose` - Show detailed breakdown

### core dev issues

- `--assignee` - Filter by assignee (use @me for yourself)
- `--limit` - Max issues per repo (default 10)
- Docs mention `--label` which is not in CLI; CLI has `--assignee` instead

### core dev reviews

- `--all` - Show all PRs including drafts
- `--author` - Filter by PR author

### core dev ci

- `--branch` - Filter by branch (default: main)
- `--failed` - Show only failed runs

### core dev update

- `--apply` - Download and apply the update (docs mention `--force` instead)

### core dev test

- `--name` - Run named test command from .core/test.yaml
- Docs mention `--unit` which is not in CLI

### core go test

- `--json` - Output JSON results

### core go cov

- `--open` - Generate and open HTML report in browser
- `--threshold` - Minimum coverage percentage (exit 1 if below)

### core go fmt

- `--check` - Check only, exit 1 if not formatted

### core build

- `--archive` - Create archives (tar.gz for linux/darwin, zip for windows)
- `--checksum` - Generate SHA256 checksums and CHECKSUMS.txt
- `--config` - Config file path
- `--format` - Output format for linuxkit (iso-bios, qcow2-bios, raw, vmdk)
- `--push` - Push Docker image after build

### core build sdk

- `--dry-run` - Show what would be generated without writing files
- `--version` - Version to embed in generated SDKs

### core build from-path

- `--path` - The path to the static web application files

### core build pwa

- `--url` - The URL of the PWA to build

### core setup

- `--dry-run` - Show what would be cloned without cloning
- `--only` - Only clone repos of these types (comma-separated: foundation,module,product)
- Docs mention `--path` and `--ssh` which are not in CLI

### core doctor

- `--verbose` - Show detailed version information

### core test

- All flags are missing from the minimal docs page:
  - `--coverage` - Show detailed per-package coverage
  - `--json` - Output JSON for CI/agents
  - `--pkg` - Package pattern to test
  - `--race` - Enable race detector
  - `--run` - Run only tests matching this regex
  - `--short` - Skip long-running tests
  - `--verbose` - Show test output as it runs

### core pkg search

- `--refresh` - Bypass cache and fetch fresh data
- `--type` - Filter by type in name (mod, services, plug, website)

### core pkg install

- `--add` - Add to repos.yaml registry

### core vm run

- `--ssh-port` - SSH port for exec commands (default: 2222)

## Discrepancies

### core sdk

- Docs describe `core sdk generate` command but CLI only has `core sdk diff` and `core sdk validate`
- SDK generation is actually at `core build sdk`, not `core sdk generate`

### core dev install

- Docs mention `--source` and `--force` flags that are not shown in CLI help

### core dev update

- Docs mention `--force` flag but CLI has `--apply` instead

### core dev test

- Docs mention `--unit` flag but CLI has `--name` flag

### core dev issues

- Docs mention `--label` flag but CLI has `--assignee` flag and no `--label`

### core dev push

- Docs mention `--all` flag but CLI only has `--force` flag

### core dev pull

- Docs mention `--rebase` flag but CLI only has `--all` flag

### core setup

- Docs mention `--path` and `--ssh` flags but CLI has `--dry-run` and `--only` flags instead

### core pkg

- Docs describe package management for "Go modules" but CLI help says it's for "core-* repos" (GitHub repos)
- `core pkg install` works differently: docs show Go module paths, CLI shows GitHub repo format

### core php serve

- Docs mention `--production` flag but CLI has different flags: `--name`, `--tag`, `--port`, `--https-port`, `-d`, `--env-file`, `--container`

### core dev work/commit/push flags

- Documentation flags don't match CLI flags in several places (e.g., `--message` vs no such flag in CLI)
