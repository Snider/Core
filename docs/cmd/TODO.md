# Documentation TODO

Commands and flags found in CLI but missing from documentation.

## Missing Flags

### core pkg search

- `--refresh` - Bypass cache and fetch fresh data
- `--type` - Filter by type in name (mod, services, plug, website)

### core vm run

- `--ssh-port` - SSH port for exec commands (default: 2222)

## Discrepancies

### core sdk

- Docs describe `core sdk generate` command but CLI only has `core sdk diff` and `core sdk validate`
- SDK generation is actually at `core build sdk`, not `core sdk generate`
