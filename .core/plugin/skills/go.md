---
name: core-go
description: Use when creating Go packages or extending the core CLI.
---

# Go Framework Patterns

Core CLI uses `pkg/` for reusable packages. Use `core go` commands.

## Package Structure

```
core/
├── main.go                 # CLI entry point
├── pkg/
│   ├── cli/               # CLI framework, output, errors
│   ├── {domain}/          # Domain package
│   │   ├── cmd_{name}.go  # Cobra command definitions
│   │   ├── service.go     # Business logic
│   │   └── *_test.go      # Tests
│   └── ...
└── internal/              # Private packages
```

## Adding a CLI Command

1. Create `pkg/{domain}/cmd_{name}.go`:

```go
package domain

import (
    "github.com/host-uk/core/pkg/cli"
    "github.com/spf13/cobra"
)

func NewNameCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "name",
        Short: cli.T("domain.name.short"),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            cli.Success("Done")
            return nil
        },
    }
    return cmd
}
```

2. Register in parent command.

## CLI Output Helpers

```go
import "github.com/host-uk/core/pkg/cli"

cli.Success("Operation completed")      // Green check
cli.Warning("Something to note")        // Yellow warning
cli.Error("Something failed")           // Red error
cli.Info("Informational message")       // Blue info
cli.Fatal(err)                          // Print error and exit 1

// Structured output
cli.Table(headers, rows)
cli.JSON(data)
```

## i18n Pattern

```go
// Use cli.T() for translatable strings
cli.T("domain.action.success")
cli.T("domain.action.error", "details", value)

// Define in pkg/i18n/locales/en.yaml:
domain:
  action:
    success: "Operation completed successfully"
    error: "Failed: {{.details}}"
```

## Test Naming

```go
func TestFeature_Good(t *testing.T) { /* happy path */ }
func TestFeature_Bad(t *testing.T)  { /* expected errors */ }
func TestFeature_Ugly(t *testing.T) { /* panics, edge cases */ }
```

## Commands

| Task | Command |
|------|---------|
| Run tests | `core go test` |
| Coverage | `core go cov` |
| Format | `core go fmt --fix` |
| Lint | `core go lint` |
| Build | `core build` |
| Install | `core go install` |

## Rules

- `CGO_ENABLED=0` for all builds
- UK English in user-facing strings
- All errors via `cli.E("context", "message", err)`
- Table-driven tests preferred
