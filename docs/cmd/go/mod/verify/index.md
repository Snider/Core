# core go mod verify

Verify dependencies have not been modified.

Wrapper around `go mod verify`. Checks that dependencies in the module cache match their checksums in go.sum.

## Usage

```bash
core go mod verify
```

## What It Does

- Verifies each module in the cache
- Compares against go.sum checksums
- Reports any tampering or corruption

## Examples

```bash
# Verify all dependencies
core go mod verify
```

## Output

```
all modules verified
```

Or if verification fails:

```
github.com/example/pkg v1.2.3: dir has been modified
```

## See Also

- [download](../download/) - Download modules
- [tidy](../tidy/) - Clean up go.mod
