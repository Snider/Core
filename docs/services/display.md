---
title: display
---
# Service: `display`

The `display` service manages the application's GUI windows, dialogs, and system tray interactions.

## Types

### `type ActionOpenWindow`

`ActionOpenWindow` is an IPC message used to request the creation of a new window.

```go
type ActionOpenWindow struct {
    application.WebviewWindowOptions
}
```

### `type WindowOption`

`WindowOption` is a functional option type for configuring a window during creation.

```go
type WindowOption func(*application.WebviewWindowOptions) error
```

## Methods

### `func OpenWindow(opts ...WindowOption) error`

`OpenWindow` creates and shows a new window with the specified options.

### `func NewWithURL(url string) (*application.WebviewWindow, error)`

`NewWithURL` creates a new window pointing to the specified URL using default settings.

### `func NewWithOptions(opts ...WindowOption) (*application.WebviewWindow, error)`

`NewWithOptions` creates a new window by applying a series of `WindowOption` functions.

### `func SelectDirectory() (string, error)`

`SelectDirectory` opens a native directory selection dialog and returns the selected path.

### `func ShowEnvironmentDialog()`

`ShowEnvironmentDialog` displays a dialog containing detailed information about the application's runtime environment (OS, version, etc.).

### `func ServiceStartup(ctx context.Context, options application.ServiceOptions) error`

`ServiceStartup` initializes the display service, setting up the main window, menu, and system tray.

### `func HandleIPCEvents(c *core.Core, msg core.Message) error`

`HandleIPCEvents` processes display-related IPC messages, such as `ActionOpenWindow`.
