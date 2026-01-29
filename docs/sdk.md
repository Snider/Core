# core sdk

Generate typed API clients from OpenAPI specifications.

## Usage

```bash
core sdk <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `generate` | Generate SDKs from OpenAPI spec |
| `validate` | Validate OpenAPI spec |
| `diff` | Check for breaking API changes |

## sdk generate

Generate typed API clients for multiple languages.

```bash
core sdk generate [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--spec` | Path to OpenAPI spec file (auto-detected) |
| `--lang` | Generate only this language |

### Examples

```bash
# Generate all configured SDKs
core sdk generate

# Generate only TypeScript SDK
core sdk generate --lang typescript

# Use specific spec file
core sdk generate --spec api/openapi.yaml
```

### Supported Languages

| Language | Generator |
|----------|-----------|
| TypeScript | openapi-generator (typescript-fetch) |
| Python | openapi-generator (python) |
| Go | openapi-generator (go) |
| PHP | openapi-generator (php) |

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

## Release Integration

Generate SDKs as part of the release process:

```bash
# Generate SDKs for release
core release --target sdk

# With explicit version
core release --target sdk --version v1.2.3

# Preview what would be generated
core release --target sdk --dry-run
```

See [release command](release.md) for full details.

## Configuration

Configure SDK generation in `.core/release.yaml`:

```yaml
sdk:
  # OpenAPI spec path (auto-detected if not set)
  spec: api/openapi.yaml

  # Languages to generate
  languages:
    - typescript
    - python
    - go
    - php

  # Output directory
  output: sdk

  # Package naming
  package:
    name: my-api-sdk

  # Breaking change detection
  diff:
    enabled: true
    fail_on_breaking: false  # Warn but continue
```

## Spec Auto-Detection

Core looks for OpenAPI specs in this order:

1. Path specified in config (`sdk.spec`)
2. `openapi.yaml` / `openapi.json`
3. `api/openapi.yaml` / `api/openapi.json`
4. `docs/openapi.yaml` / `docs/openapi.json`
5. Laravel Scramble endpoint (`/docs/api.json`)

## Output Structure

Generated SDKs are placed in language-specific directories:

```
sdk/
├── typescript/
│   ├── src/
│   ├── package.json
│   └── tsconfig.json
├── python/
│   ├── my_api_sdk/
│   ├── setup.py
│   └── requirements.txt
├── go/
│   ├── client.go
│   └── go.mod
└── php/
    ├── src/
    └── composer.json
```
