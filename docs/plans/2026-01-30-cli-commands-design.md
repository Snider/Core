# CLI Commands Registration Design

## Overview

Move CLI commands from `cmd/` into self-registering packages in `pkg/`. This enables build variants with reduced attack surface - only compiled code exists in the binary.

## Pattern

Same pattern as `i18n.RegisterLocales()`:
- Packages register themselves during `init()`
- Registration is stored until `cli.Init()` runs
- Build tags control which packages are imported

## Registration API

```go
// pkg/cli/commands.go

type CommandRegistration func(root *cobra.Command)

var (
    registeredCommands   []CommandRegistration
    registeredCommandsMu sync.Mutex
)

// RegisterCommands registers a function that adds commands to the CLI.
func RegisterCommands(fn CommandRegistration) {
    registeredCommandsMu.Lock()
    defer registeredCommandsMu.Unlock()
    registeredCommands = append(registeredCommands, fn)
}

func attachRegisteredCommands(root *cobra.Command) {
    registeredCommandsMu.Lock()
    defer registeredCommandsMu.Unlock()
    for _, fn := range registeredCommands {
        fn(root)
    }
}
```

## Integration with Core.App

The CLI stores `rootCmd` in `core.App`, unifying GUI and CLI under the same pattern:

```go
// pkg/cli/runtime.go

func Init(opts Options) error {
    once.Do(func() {
        rootCmd := &cobra.Command{
            Use:     opts.AppName,
            Version: opts.Version,
        }

        attachRegisteredCommands(rootCmd)

        c, err := framework.New(
            framework.WithApp(rootCmd),
            // ... services ...
        )
        // ...
    })
    return initErr
}

func RootCmd() *cobra.Command {
    return framework.App().(*cobra.Command)
}

func Execute() error {
    return RootCmd().Execute()
}
```

## Package Structure

Commands move from `cmd/` to `pkg/` with a `cmd.go` file:

```
pkg/
├── php/
│   ├── i18n.go          # registers locales
│   ├── cmd.go           # registers commands
│   ├── locales/
│   └── ...
├── dev/
│   ├── cmd.go           # registers commands
│   └── ...
├── cli/
│   ├── commands.go      # RegisterCommands API
│   ├── runtime.go       # Init, Execute
│   └── ...
```

Each `cmd.go`:

```go
// pkg/php/cmd.go
package php

import "github.com/host-uk/core/pkg/cli"

func init() {
    cli.RegisterCommands(AddCommands)
}

func AddCommands(root *cobra.Command) {
    // ... existing command setup ...
}
```

## Build Variants

Import files with build tags in `cmd/variants/`:

```
cmd/
├── main.go
└── variants/
    ├── full.go      # default: all packages
    ├── ci.go        # CI/release only
    ├── php.go       # PHP tooling only
    └── minimal.go   # core only
```

```go
// cmd/variants/full.go
//go:build !ci && !php && !minimal

package variants

import (
    _ "github.com/host-uk/core/pkg/ai"
    _ "github.com/host-uk/core/pkg/build"
    _ "github.com/host-uk/core/pkg/ci"
    _ "github.com/host-uk/core/pkg/dev"
    _ "github.com/host-uk/core/pkg/docs"
    _ "github.com/host-uk/core/pkg/doctor"
    _ "github.com/host-uk/core/pkg/go"
    _ "github.com/host-uk/core/pkg/php"
    _ "github.com/host-uk/core/pkg/pkg"
    _ "github.com/host-uk/core/pkg/sdk"
    _ "github.com/host-uk/core/pkg/setup"
    _ "github.com/host-uk/core/pkg/test"
    _ "github.com/host-uk/core/pkg/vm"
)
```

```go
// cmd/variants/ci.go
//go:build ci

package variants

import (
    _ "github.com/host-uk/core/pkg/build"
    _ "github.com/host-uk/core/pkg/ci"
    _ "github.com/host-uk/core/pkg/doctor"
    _ "github.com/host-uk/core/pkg/sdk"
)
```

## Build Commands

- `go build` → full variant (default)
- `go build -tags ci` → CI variant
- `go build -tags php` → PHP-only variant

## Benefits

1. **Smaller attack surface** - only compiled code exists in binary
2. **Self-registering packages** - same pattern as `i18n.RegisterLocales()`
3. **Uses existing `core.App`** - no new framework concepts
4. **Simple build variants** - just add `-tags` flag
5. **Defence in depth** - no code = no vulnerabilities

## Migration Steps

1. Add `RegisterCommands()` to `pkg/cli/commands.go`
2. Update `pkg/cli/runtime.go` to use `core.App` for rootCmd
3. Move each `cmd/*` package to `pkg/*/cmd.go`
4. Create `cmd/variants/` with build tag files
5. Simplify `cmd/main.go` to minimal entry point
6. Remove old `cmd/core_dev.go` and `cmd/core_ci.go`
