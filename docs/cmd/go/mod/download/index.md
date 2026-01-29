# core go mod download

Download modules to local cache.

Wrapper around `go mod download`. Downloads all dependencies to the module cache.

## Usage

```bash
core go mod download
```

## What It Does

- Downloads all modules in go.mod to `$GOPATH/pkg/mod`
- Useful for pre-populating cache for offline builds
- Validates checksums against go.sum

## Examples

```bash
# Download all dependencies
core go mod download
```

## See Also

- [tidy](../tidy/) - Clean up go.mod
- [verify](../verify/) - Verify checksums
