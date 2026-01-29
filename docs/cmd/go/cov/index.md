# core go cov

Generate coverage report with thresholds.

## Usage

```bash
core go cov [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--pkg` | Package to test (default: `./...`) |
| `--html` | Generate HTML report |
| `--open` | Open in browser |
| `--threshold` | Fail if coverage below % |

## Examples

```bash
core go cov                     # Summary
core go cov --html              # HTML report
core go cov --open              # Open in browser
core go cov --threshold 80      # Fail if < 80%
core go cov --pkg ./pkg/release # Specific package
```
