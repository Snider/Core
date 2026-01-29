# core go work use

Add module to workspace.

Wrapper around `go work use`. Adds one or more modules to the go.work file.

## Usage

```bash
core go work use [paths...]
```

## What It Does

- Adds specified module paths to go.work
- Auto-discovers modules if no paths given
- Enables developing multiple modules together

## Examples

```bash
# Add a specific module
core go work use ./pkg/mymodule

# Add multiple modules
core go work use ./pkg/one ./pkg/two

# Auto-discover and add all modules
core go work use
```

## Auto-Discovery

When called without arguments, scans for go.mod files and adds all found modules:

```bash
core go work use
# Added ./pkg/build
# Added ./pkg/repos
# Added ./cmd/core
```

## See Also

- [init](../init/) - Initialize workspace
- [sync](../sync/) - Sync workspace
