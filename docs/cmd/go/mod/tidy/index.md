# core go mod tidy

Add missing and remove unused modules.

Wrapper around `go mod tidy`. Ensures go.mod and go.sum are in sync with the source code.

## Usage

```bash
core go mod tidy
```

## What It Does

- Adds missing module requirements
- Removes unused module requirements
- Updates go.sum with checksums

## Examples

```bash
# Tidy the current module
core go mod tidy
```

## See Also

- [download](../download/) - Download modules
- [verify](../verify/) - Verify dependencies
