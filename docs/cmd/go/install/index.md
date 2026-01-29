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
core go install                 # Install current module
core go install ./cmd/core      # Install specific path
core go install --no-cgo        # Pure Go (no C dependencies)
core go install -v              # Verbose output
```
