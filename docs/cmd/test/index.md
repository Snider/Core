# core test

Run Go tests with coverage reporting.

Sets `MACOSX_DEPLOYMENT_TARGET=26.0` to suppress linker warnings on macOS.

## Usage

```bash
core test [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--coverage` | Show detailed per-package coverage |
| `--json` | Output JSON for CI/agents |
| `--pkg` | Package pattern to test (default: ./...) |
| `--race` | Enable race detector |
| `--run` | Run only tests matching this regex |
| `--short` | Skip long-running tests |
| `--verbose` | Show test output as it runs |

## Examples

```bash
# Run all tests with coverage summary
core test

# Show test output as it runs
core test --verbose

# Detailed per-package coverage
core test --coverage

# Test specific packages
core test --pkg ./pkg/...

# Run specific test by name
core test --run TestName

# Run tests matching pattern
core test --run "Test.*Good"

# Skip long-running tests
core test --short

# Enable race detector
core test --race

# Output JSON for CI/agents
core test --json
```

## JSON Output

With `--json`, outputs structured results:

```json
{
  "passed": 14,
  "failed": 0,
  "skipped": 0,
  "coverage": 75.1,
  "exit_code": 0,
  "failed_packages": []
}
```

## See Also

- [go test](../go/test/) - Go-specific test options
- [go cov](../go/cov/) - Coverage reports
