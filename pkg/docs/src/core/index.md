---
title: Core
---

# Core

Short: Framework bootstrap and service container.

## What it is
Core wires modules together, provides lifecycle hooks, and locks the service graph for clarity and safety.

## Setup
```go
import "github.com/Snider/Core"

app := core.New(
    core.WithServiceLock(),
)
```

## Use
- Register a module: `core.RegisterModule(name, module)`
- Access a module: `core.Mod[T](c, name)`
- Lock services: `core.WithServiceLock()`

## API
- `New(opts ...) *core.Core`
- `RegisterModule(name string, m any) error`
- `Mod[T any](c *core.Core, name ...string) *T`

