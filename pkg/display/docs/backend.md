# Backend Documentation

The backend is written in Go and uses the `github.com/Snider/display` package. It utilizes the Wails v3 framework to bridge Go and the web frontend.

## Core Types

### `Service`
The `Service` struct is the main entry point for the display logic.

- **Initialization:**
  - `New() (*Service, error)`: Creates a new instance of the service.
  - `Startup(ctx context.Context) error`: Initializes the Wails application, builds the menu, sets up the system tray, and opens the main window.

- **Window Management:**
  - `OpenWindow(opts ...WindowOption) error`: Opens a new window with the specified options.

- **Dialogs:**
  - `ShowEnvironmentDialog()`: Displays a native dialog containing information about the runtime environment (OS, Arch, Debug mode, etc.).

### `WindowConfig` & `WindowOption`
Window configuration is handled using the Functional Options pattern. The `WindowConfig` struct holds parameters like:
- `Name`, `Title`
- `Width`, `Height`
- `URL`
- `AlwaysOnTop`, `Hidden`, `Frameless`
- Window button states (`MinimiseButtonState`, `MaximiseButtonState`, `CloseButtonState`)

**Available Options:**
- `WithName(name string)`
- `WithTitle(title string)`
- `WithWidth(width int)`
- `WithHeight(height int)`
- `WithURL(url string)`
- `WithAlwaysOnTop(bool)`
- `WithHidden(bool)`
- `WithFrameless(bool)`
- `WithMinimiseButtonState(state)`
- `WithMaximiseButtonState(state)`
- `WithCloseButtonState(state)`

## Subsystems

### Menu (`menu.go`)
The `buildMenu` method constructs the application's main menu, adding standard roles like File, Edit, Window, and Help. It allows for platform-specific adjustments (e.g., AppMenu on macOS).

### System Tray (`tray.go`)
The `systemTray` method initializes the system tray icon and its context menu. It supports:
- Showing/Hiding all windows.
- Displaying environment info.
- Quitting the application.
- Attaching a hidden window for advanced tray interactions.

### Actions (`actions.go`)
Defines structured messages for Inter-Process Communication (IPC) or internal event handling, such as `ActionOpenWindow` which wraps `application.WebviewWindowOptions`.
