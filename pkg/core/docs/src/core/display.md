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
  core "github.com/Snider/Core"
  display "github.com/Snider/Core/display"
)

app := core.New(
  core.WithService(display.Register),
  core.WithServiceLock(),
)
```

## Use
- Open a window: `OpenWindow(OptName("main"), ...)`
- Get a window: `Window("main")`
- Save/restore state automatically when `Core.Config` is present

## API
- `Register(c *core.Core) error`
- `OpenWindow(opts ...Option) *Window`
- `Window(name string) *Window`
- Options: `OptName`, `OptWidth`, `OptHeight`, `OptURL`, `OptTitle`

## Example
```go
func (d *API) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
  d.OpenWindow(
    OptName("main"), OptWidth(1280), OptHeight(900), OptURL("/"), OptTitle("Core"),
  )
  return nil
}
```


