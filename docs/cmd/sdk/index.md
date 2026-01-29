# core sdk

SDK validation and API compatibility tools.

To generate SDKs, use: `core build sdk`

## Usage

```bash
core sdk <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `diff` | Check for breaking API changes |
| `validate` | Validate OpenAPI spec |

## sdk validate

Validate an OpenAPI specification file.

```bash
core sdk validate [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--spec` | Path to OpenAPI spec file (auto-detected) |

### Examples

```bash
# Validate detected spec
core sdk validate

# Validate specific file
core sdk validate --spec api/openapi.yaml
```

## sdk diff

Check for breaking changes between API versions.

```bash
core sdk diff [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--base` | Base spec version (git tag or file path) |
| `--spec` | Current spec file (auto-detected) |

### Examples

```bash
# Compare against previous release
core sdk diff --base v1.0.0

# Compare two files
core sdk diff --base old-api.yaml --spec new-api.yaml
```

### Breaking Changes Detected

- Removed endpoints
- Changed parameter types
- Removed required fields
- Changed response types

## SDK Generation

SDK generation is handled by `core build sdk`, not this command.

```bash
# Generate SDKs
core build sdk

# Generate specific language
core build sdk --lang typescript

# Preview without writing
core build sdk --dry-run
```

See [build sdk](../build/sdk/) for generation details.

## Spec Auto-Detection

Core looks for OpenAPI specs in this order:

1. Path specified in config (`sdk.spec`)
2. `openapi.yaml` / `openapi.json`
3. `api/openapi.yaml` / `api/openapi.json`
4. `docs/openapi.yaml` / `docs/openapi.json`
5. Laravel Scramble endpoint (`/docs/api.json`)

## See Also

- [build sdk](../build/sdk/) - Generate SDKs from OpenAPI
- [ci command](../ci/) - Release workflow
