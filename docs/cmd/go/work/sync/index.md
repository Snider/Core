# core go work sync

Sync go.work with modules.

Wrapper around `go work sync`. Synchronises the workspace's build list back to the workspace modules.

## Usage

```bash
core go work sync
```

## What It Does

- Updates each module's go.mod to match the workspace build list
- Ensures all modules use compatible dependency versions
- Run after adding new modules or updating dependencies

## Examples

```bash
# Sync workspace
core go work sync
```

## When To Use

- After running `go get` to update a dependency
- After adding a new module with `core go work use`
- When modules have conflicting dependency versions

## See Also

- [init](../init/) - Initialize workspace
- [use](../use/) - Add module to workspace
