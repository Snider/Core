# core go work init

Initialize a Go workspace.

Wrapper around `go work init`. Creates a new go.work file in the current directory.

## Usage

```bash
core go work init
```

## What It Does

- Creates a go.work file
- Automatically adds current module if go.mod exists
- Enables multi-module development

## Examples

```bash
# Initialize workspace
core go work init

# Then add more modules
core go work use ./pkg/mymodule
```

## Generated File

```go
go 1.25

use .
```

## See Also

- [use](../use/) - Add module to workspace
- [sync](../sync/) - Sync workspace
