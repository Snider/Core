# Go Lint Examples

```bash
# Check
core go lint

# Auto-fix
core go lint --fix
```

## Configuration

`.golangci.yml`:

```yaml
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
```
