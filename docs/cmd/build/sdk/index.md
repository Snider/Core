# core build sdk

Generate API SDKs from OpenAPI specifications.

## Usage

```bash
core build sdk [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--spec` | Path to OpenAPI spec file |
| `--lang` | Generate only this language |
| `--version` | Version to embed |
| `--dry-run` | Preview without generating |

## Examples

```bash
core build sdk                      # Generate all
core build sdk --lang typescript    # TypeScript only
core build sdk --spec ./api.yaml    # Custom spec
core build sdk --dry-run            # Preview
```
