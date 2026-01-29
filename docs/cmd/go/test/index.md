# core go test

Run Go tests with coverage and filtered output.

## Usage

```bash
core go test [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--pkg` | Package to test (default: `./...`) |
| `--run` | Run only tests matching regexp |
| `--short` | Run only short tests |
| `--race` | Enable race detector |
| `--coverage` | Show detailed per-package coverage |
| `--json` | Output JSON results |
| `-v` | Verbose output |

## Examples

```bash
core go test                    # All tests
core go test --pkg ./pkg/core   # Specific package
core go test --run TestHash     # Specific test
core go test --coverage         # With coverage
core go test --race             # Race detection
```
