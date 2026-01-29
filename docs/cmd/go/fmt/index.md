# core go fmt

Format Go code using goimports or gofmt.

## Usage

```bash
core go fmt [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--fix` | Fix formatting in place |
| `--diff` | Show diff of changes |
| `--check` | Check only, exit 1 if not formatted |

## Examples

```bash
core go fmt           # Check formatting
core go fmt --fix     # Fix formatting
core go fmt --diff    # Show diff
```
