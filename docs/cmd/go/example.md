# Go Examples

## Testing

```bash
# Run all tests
core go test

# Specific package
core go test --pkg ./pkg/core

# Specific test
core go test --run TestHash

# With coverage
core go test --coverage

# Race detection
core go test --race
```

## Coverage

```bash
# Summary
core go cov

# HTML report
core go cov --html

# Open in browser
core go cov --open

# Fail if below threshold
core go cov --threshold 80
```

## Formatting

```bash
# Check
core go fmt

# Fix
core go fmt --fix

# Show diff
core go fmt --diff
```

## Linting

```bash
# Check
core go lint

# Auto-fix
core go lint --fix
```

## Installing

```bash
# Auto-detect cmd/
core go install

# Specific path
core go install ./cmd/myapp

# Pure Go (no CGO)
core go install --no-cgo
```

## Module Management

```bash
core go mod tidy
core go mod download
core go mod verify
core go mod graph
```

## Workspace

```bash
core go work sync
core go work init
core go work use ./pkg/mymodule
```
