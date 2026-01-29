# Go Test Examples

```bash
# All tests
core go test

# Specific package
core go test --pkg ./pkg/core

# Specific test
core go test --run TestHash

# With coverage
core go test --coverage

# Race detection
core go test --race

# Short tests only
core go test --short

# Verbose
core go test -v

# JSON output (CI)
core go test --json
```
