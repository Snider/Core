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
| `--run` | Run specific tests |
| `--short` | Skip long-running tests |
| `--race` | Enable race detection |
| `--coverage` | Show coverage summary |
| `--json` | JSON output for CI |
| `-v` | Verbose output |

## Examples

```bash
core go test                    # All tests
core go test --pkg ./pkg/core   # Specific package
core go test --run TestHash     # Specific test
core go test --coverage         # With coverage
core go test --race             # Race detection
```
