# core build sdk

Generate typed API clients from OpenAPI specifications. Supports TypeScript, Python, Go, and PHP.

## Usage

```bash
core build sdk [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--spec` | Path to OpenAPI spec file |
| `--lang` | Generate only this language (typescript, python, go, php) |
| `--version` | Version to embed in generated SDKs |
| `--dry-run` | Show what would be generated without writing files |

## Examples

```bash
core build sdk                      # Generate all
core build sdk --lang typescript    # TypeScript only
core build sdk --spec ./api.yaml    # Custom spec
core build sdk --dry-run            # Preview
```
