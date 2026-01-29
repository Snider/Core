# SDK Generation Design

## Summary

Generate typed API clients from OpenAPI specs for TypeScript, Python, Go, and PHP. Includes breaking change detection via semantic diff.

## Design Decisions

- **Generator approach**: Hybrid - native generators where available, openapi-generator fallback
- **Languages**: TypeScript, Python, Go, PHP (Core 4)
- **Detection**: Config → common paths → Laravel Scramble
- **Output**: Local `sdk/` + optional monorepo publish
- **Diff**: Semantic with oasdiff, CI-friendly exit codes
- **Priority**: DX (developer experience)

## Package Structure

```
pkg/sdk/
├── sdk.go              # Main SDK type, orchestration
├── detect.go           # OpenAPI spec detection
├── diff.go             # Breaking change detection (oasdiff)
├── generators/
│   ├── generator.go    # Generator interface
│   ├── typescript.go   # openapi-typescript-codegen
│   ├── python.go       # openapi-python-client
│   ├── go.go           # oapi-codegen
│   └── php.go          # openapi-generator (Docker)
└── templates/          # Package scaffolding templates
    ├── typescript/
    │   └── package.json.tmpl
    ├── python/
    │   └── setup.py.tmpl
    ├── go/
    │   └── go.mod.tmpl
    └── php/
        └── composer.json.tmpl
```

## OpenAPI Detection Flow

```
1. Check config: sdk.spec in .core/release.yaml
   ↓ not found
2. Check common paths:
   - api/openapi.yaml
   - api/openapi.json
   - openapi.yaml
   - openapi.json
   - docs/api.yaml
   - swagger.yaml
   ↓ not found
3. Laravel Scramble detection:
   - Check for scramble/scramble in composer.json
   - Run: php artisan scramble:export --path=api/openapi.json
   - Use generated spec
   ↓ not found
4. Error: No OpenAPI spec found
```

## Generator Interface

```go
type Generator interface {
    // Language returns the generator's target language
    Language() string

    // Generate creates SDK from OpenAPI spec
    Generate(ctx context.Context, opts GenerateOptions) error

    // Available checks if generator dependencies are installed
    Available() bool

    // Install provides installation instructions
    Install() string
}

type GenerateOptions struct {
    SpecPath    string   // OpenAPI spec file
    OutputDir   string   // Where to write SDK
    PackageName string   // Package/module name
    Version     string   // SDK version
}
```

### Native Generators

| Language   | Tool                       | Install                        |
|------------|----------------------------|--------------------------------|
| TypeScript | openapi-typescript-codegen | `npm i -g openapi-typescript-codegen` |
| Python     | openapi-python-client      | `pip install openapi-python-client`   |
| Go         | oapi-codegen               | `go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest` |
| PHP        | openapi-generator (Docker) | Requires Docker                |

### Fallback Strategy

```go
func (g *TypeScriptGenerator) Generate(ctx context.Context, opts GenerateOptions) error {
    if g.Available() {
        return g.generateNative(ctx, opts)
    }
    return g.generateDocker(ctx, opts)  // openapi-generator in Docker
}
```

## Breaking Change Detection

Using [oasdiff](https://github.com/Tufin/oasdiff) for semantic OpenAPI comparison:

```go
import "github.com/tufin/oasdiff/diff"
import "github.com/tufin/oasdiff/checker"

func (s *SDK) Diff(base, revision string) (*DiffResult, error) {
    // Load specs
    baseSpec, _ := load.From(loader, base)
    revSpec, _ := load.From(loader, revision)

    // Compute diff
    d, _ := diff.Get(diff.NewConfig(), baseSpec, revSpec)

    // Check for breaking changes
    breaks := checker.CheckBackwardCompatibility(
        checker.GetDefaultChecks(),
        d,
        baseSpec,
        revSpec,
    )

    return &DiffResult{
        Breaking:    len(breaks) > 0,
        Changes:     breaks,
        Summary:     formatSummary(d),
    }, nil
}
```

### Exit Codes for CI

| Exit Code | Meaning |
|-----------|---------|
| 0         | No breaking changes |
| 1         | Breaking changes detected |
| 2         | Error (invalid spec, etc.) |

### Breaking Change Categories

- Removed endpoints
- Changed required parameters
- Modified response schemas
- Changed authentication requirements

## CLI Commands

```bash
# Generate SDKs from OpenAPI spec
core sdk generate                    # Uses .core/release.yaml config
core sdk generate --spec api.yaml    # Explicit spec file
core sdk generate --lang typescript  # Single language

# Check for breaking changes
core sdk diff                        # Compare current vs last release
core sdk diff --spec api.yaml --base v1.0.0

# Validate spec before generation
core sdk validate
core sdk validate --spec api.yaml
```

## Config Schema

In `.core/release.yaml`:

```yaml
sdk:
  # OpenAPI spec source (auto-detected if omitted)
  spec: api/openapi.yaml

  # Languages to generate
  languages:
    - typescript
    - python
    - go
    - php

  # Output directory (default: sdk/)
  output: sdk/

  # Package naming
  package:
    name: myapi        # Base name
    version: "{{.Version}}"

  # Breaking change detection
  diff:
    enabled: true
    fail_on_breaking: true  # CI fails on breaking changes

  # Optional: publish to monorepo
  publish:
    repo: myorg/sdks
    path: packages/myapi
```

## Output Structure

Each generator outputs to `sdk/{lang}/`:

```
sdk/
├── typescript/
│   ├── package.json
│   ├── src/
│   │   ├── index.ts
│   │   ├── client.ts
│   │   └── models/
│   └── tsconfig.json
├── python/
│   ├── setup.py
│   ├── myapi/
│   │   ├── __init__.py
│   │   ├── client.py
│   │   └── models/
│   └── requirements.txt
├── go/
│   ├── go.mod
│   ├── client.go
│   └── models.go
└── php/
    ├── composer.json
    ├── src/
    │   ├── Client.php
    │   └── Models/
    └── README.md
```

## Publishing Workflow

SDK publishing integrates with the existing release pipeline:

```
core release
  → build artifacts
  → generate SDKs (if sdk: configured)
  → run diff check (warns or fails on breaking)
  → publish to GitHub release
  → publish SDKs (optional)
```

### Monorepo Publishing

For projects using a shared SDK monorepo:

1. Clone target repo (shallow)
2. Update `packages/{name}/{lang}/`
3. Commit with version tag
4. Push (triggers downstream CI)

The SDK tarball is also attached to GitHub releases for direct download.

## Implementation Steps

1. Create `pkg/sdk/` package structure
2. Implement OpenAPI detection (`detect.go`)
3. Define Generator interface (`generators/generator.go`)
4. Implement TypeScript generator (native + fallback)
5. Implement Python generator (native + fallback)
6. Implement Go generator (native)
7. Implement PHP generator (Docker-based)
8. Add package templates (`templates/`)
9. Implement diff with oasdiff (`diff.go`)
10. Add CLI commands (`cmd/core/sdk.go`)
11. Integrate with release pipeline
12. Add monorepo publish support

## Dependencies

```go
// go.mod additions
require (
    github.com/tufin/oasdiff v1.x.x
    github.com/getkin/kin-openapi v0.x.x
)
```

## Testing

- Unit tests for each generator
- Integration tests with sample OpenAPI specs
- Diff tests with known breaking/non-breaking changes
- E2E test generating SDKs for a real API
