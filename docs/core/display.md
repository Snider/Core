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
  display "github.com/Snider/Core/pkg/display"
  "github.com/wailsapp/wails/v3/pkg/application" // For WebviewWindowOptions
  config "github.com/Snider/Core/pkg/config" // Assuming config service is available
)

// Example of static dependency injection
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

## API
- `New(cfg core.Config) (*Service, error)`
- `Register(c *core.Core) (any, error)`
- `(s *Service) ServiceName() string`
- `(s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error`
- `(s *Service) ShowEnvironmentDialog()`
- `(s *Service) ServiceStartup(context.Context, application.ServiceOptions) error`

## Example: Opening a new window via Action
```go
// In another service or component that has access to core.Core
func (myService *MyService) OpenNewWindow() error {
    // Option 1: Using ActionOpenWindow struct
    action := display.ActionOpenWindow{
        WebviewWindowOptions: application.WebviewWindowOptions{
            Name:   "myNewWindow",
            Title:  "My New Window",
            Width:  800,
            Height: 600,
            URL:    "/some-path",
        },
    }
    return myService.Core().ACTION(action)

    // Option 2: Using a map[string]any (less type-safe but flexible)
    // msg := map[string]any{
    //     "action": "display.open_window",
    //     "name":   "myNewWindow",
    //     "options": map[string]any{
    //         "Title":  "My New Window",
    //         "Width":  800,
    //         "Height": 600,
    //         "URL":    "/some-path",
    //     },
    // }
    // return myService.Core().ACTION(msg)
}
```

## Notes
- The `Display` service integrates with `Core.Config` to persist window states.
- Window management is primarily done through `core.ACTION` messages.
