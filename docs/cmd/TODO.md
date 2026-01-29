# Documentation TODO

Commands and flags found in CLI but missing from documentation.

## Missing Commands

### core php

- `core php packages link` - Link local packages (subcommand documentation exists but not detailed)
- `core php packages unlink` - Unlink packages
- `core php packages update` - Update linked packages
- `core php packages list` - List linked packages

### core vm

- `core vm templates show` - Display template content
- `core vm templates vars` - Show template variables

## Missing Flags

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

### core setup

- Docs mention `--path` and `--ssh` flags but CLI has `--dry-run` and `--only` flags instead

### core pkg

- Docs describe package management for "Go modules" but CLI help says it's for "core-* repos" (GitHub repos)
- `core pkg install` works differently: docs show Go module paths, CLI shows GitHub repo format

### core php serve

- Docs mention `--production` flag but CLI has different flags: `--name`, `--tag`, `--port`, `--https-port`, `-d`, `--env-file`, `--container`
