---
title: Core.Display
---

# Core.Display

Short: Windows, tray, and window state.

## Overview
Manages Wails windows, remembers positions/sizes, exposes JS bindings, and integrates with `Core.Config` for persistence.

## Setup
```go
import (
  "github.com/Snider/Core"
  "github.com/Snider/Core/display"
  "github.com/Snider/Core/config"
  "github.com/wailsapp/wails/v3/pkg/application"
)

configService, err := config.New()
if err != nil {
  // handle error
}
displayService, err := display.New(configService)
if err != nil {
  // handle error
}
app := core.New(
  core.WithService(displayService),
  core.WithServiceLock(),
)

// Example of dynamic dependency injection (used with core.WithService)
// app := core.New(
//   core.WithService(display.Register),
//   core.WithServiceLock(),
// )
```

## Use
- The main application window is typically managed by the `Display` service's `ServiceStartup`.
- To open additional windows or control existing ones, send `core.Message` actions.
- Save/restore state automatically when `Core.Config` is present.

## Notes
- The `Display` service integrates with `Core.Config` to persist window states.
- Window management is primarily done through `core.ACTION` messages.
