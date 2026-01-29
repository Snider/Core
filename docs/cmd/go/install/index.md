# core go install

Install Go binary with auto-detection.

## Usage

```bash
core go install [path] [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--no-cgo` | Disable CGO |
| `-v` | Verbose |

## Examples

```bash
core go install                 # Auto-detect cmd/
core go install ./cmd/core      # Specific path
core go install --no-cgo        # Pure Go
```
