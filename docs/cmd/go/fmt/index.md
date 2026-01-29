# core go fmt

Format Go code using goimports or gofmt.

## Usage

```bash
core go fmt [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--fix` | Apply fixes (default: check only) |
| `--diff` | Show diff |

## Examples

```bash
core go fmt           # Check
core go fmt --fix     # Apply fixes
core go fmt --diff    # Show diff
```
