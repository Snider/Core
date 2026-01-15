# Display API Reference

Complete API reference for the Display service (`pkg/display`).

## Service Creation

```go
func NewService(c *core.Core) (any, error)
```

## Window Management

### CreateWindow

```go
func (s *Service) CreateWindow(opts CreateWindowOptions) (*WindowInfo, error)
```

Creates a new window with the specified options.

```go
type CreateWindowOptions struct {
    Name   string
    Title  string
    URL    string
    X      int
    Y      int
    Width  int
    Height int
}
```

### CloseWindow

```go
func (s *Service) CloseWindow(name string) error
```

### GetWindowInfo

```go
func (s *Service) GetWindowInfo(name string) (*WindowInfo, error)
```

Returns:

```go
type WindowInfo struct {
    Name       string
    Title      string
    X          int
    Y          int
    Width      int
    Height     int
    IsVisible  bool
    IsFocused  bool
    IsMaximized bool
    IsMinimized bool
}
```

### ListWindowInfos

```go
func (s *Service) ListWindowInfos() []*WindowInfo
```

### Window Position & Size

```go
func (s *Service) SetWindowPosition(name string, x, y int) error
func (s *Service) SetWindowSize(name string, width, height int) error
func (s *Service) SetWindowBounds(name string, x, y, width, height int) error
```

### Window State

```go
func (s *Service) MaximizeWindow(name string) error
func (s *Service) MinimizeWindow(name string) error
func (s *Service) RestoreWindow(name string) error
func (s *Service) FocusWindow(name string) error
func (s *Service) SetWindowFullscreen(name string, fullscreen bool) error
func (s *Service) SetWindowAlwaysOnTop(name string, onTop bool) error
func (s *Service) SetWindowVisibility(name string, visible bool) error
```

### Window Title

```go
func (s *Service) SetWindowTitle(name, title string) error
func (s *Service) GetWindowTitle(name string) (string, error)
```

### Window Background

```go
func (s *Service) SetWindowBackgroundColour(name string, r, g, b, a uint8) error
```

### Focus

```go
func (s *Service) GetFocusedWindow() string
```

## Screen Management

### GetScreens

```go
func (s *Service) GetScreens() []*Screen
```

Returns:

```go
type Screen struct {
    ID          string
    Name        string
    X           int
    Y           int
    Width       int
    Height      int
    ScaleFactor float64
    IsPrimary   bool
}
```

### GetScreen

```go
func (s *Service) GetScreen(id string) (*Screen, error)
```

### GetPrimaryScreen

```go
func (s *Service) GetPrimaryScreen() (*Screen, error)
```

### GetScreenAtPoint

```go
func (s *Service) GetScreenAtPoint(x, y int) (*Screen, error)
```

### GetScreenForWindow

```go
func (s *Service) GetScreenForWindow(name string) (*Screen, error)
```

### GetWorkAreas

```go
func (s *Service) GetWorkAreas() []*WorkArea
```

Returns usable screen space (excluding dock/taskbar).

## Layout Management

### SaveLayout / RestoreLayout

```go
func (s *Service) SaveLayout(name string) error
func (s *Service) RestoreLayout(name string) error
func (s *Service) ListLayouts() []string
func (s *Service) DeleteLayout(name string) error
func (s *Service) GetLayout(name string) *Layout
```

### TileWindows

```go
func (s *Service) TileWindows(mode TileMode, windows []string) error
```

Tile modes:

```go
const (
    TileModeLeft      TileMode = "left"
    TileModeRight     TileMode = "right"
    TileModeGrid      TileMode = "grid"
    TileModeQuadrants TileMode = "quadrants"
)
```

### SnapWindow

```go
func (s *Service) SnapWindow(name string, position SnapPosition) error
```

Snap positions:

```go
const (
    SnapPositionLeft        SnapPosition = "left"
    SnapPositionRight       SnapPosition = "right"
    SnapPositionTop         SnapPosition = "top"
    SnapPositionBottom      SnapPosition = "bottom"
    SnapPositionTopLeft     SnapPosition = "top-left"
    SnapPositionTopRight    SnapPosition = "top-right"
    SnapPositionBottomLeft  SnapPosition = "bottom-left"
    SnapPositionBottomRight SnapPosition = "bottom-right"
)
```

### StackWindows

```go
func (s *Service) StackWindows(windows []string, offsetX, offsetY int) error
```

### ApplyWorkflowLayout

```go
func (s *Service) ApplyWorkflowLayout(workflow WorkflowType) error
```

Workflow types:

```go
const (
    WorkflowCoding    WorkflowType = "coding"
    WorkflowDebugging WorkflowType = "debugging"
    WorkflowPresenting WorkflowType = "presenting"
)
```

## Dialogs

### File Dialogs

```go
func (s *Service) OpenSingleFileDialog(opts OpenFileOptions) (string, error)
func (s *Service) OpenFileDialog(opts OpenFileOptions) ([]string, error)
func (s *Service) SaveFileDialog(opts SaveFileOptions) (string, error)
func (s *Service) OpenDirectoryDialog(opts OpenDirectoryOptions) (string, error)
```

Options:

```go
type OpenFileOptions struct {
    Title            string
    DefaultDirectory string
    AllowMultiple    bool
    Filters          []FileFilter
}

type SaveFileOptions struct {
    Title            string
    DefaultDirectory string
    DefaultFilename  string
    Filters          []FileFilter
}

type FileFilter struct {
    DisplayName string
    Pattern     string  // e.g., "*.png;*.jpg"
}
```

### ConfirmDialog

```go
func (s *Service) ConfirmDialog(title, message string) (bool, error)
```

### PromptDialog

```go
func (s *Service) PromptDialog(title, message string) (string, bool, error)
```

## System Tray

```go
func (s *Service) SetTrayIcon(icon []byte) error
func (s *Service) SetTrayTooltip(tooltip string) error
func (s *Service) SetTrayLabel(label string) error
func (s *Service) SetTrayMenu(items []TrayMenuItem) error
func (s *Service) GetTrayInfo() map[string]any
```

Menu item:

```go
type TrayMenuItem struct {
    Label       string
    ActionID    string
    IsSeparator bool
}
```

## Clipboard

```go
func (s *Service) ReadClipboard() (string, error)
func (s *Service) WriteClipboard(text string) error
func (s *Service) HasClipboard() bool
func (s *Service) ClearClipboard() error
```

## Notifications

```go
func (s *Service) ShowNotification(opts NotificationOptions) error
func (s *Service) ShowInfoNotification(title, message string) error
func (s *Service) ShowWarningNotification(title, message string) error
func (s *Service) ShowErrorNotification(title, message string) error
func (s *Service) RequestNotificationPermission() (bool, error)
func (s *Service) CheckNotificationPermission() (bool, error)
```

Options:

```go
type NotificationOptions struct {
    ID       string
    Title    string
    Message  string
    Subtitle string
}
```

## Theme

```go
func (s *Service) GetTheme() *Theme
func (s *Service) GetSystemTheme() string
```

Returns:

```go
type Theme struct {
    IsDark bool
}
```

## Events

```go
func (s *Service) GetEventManager() *EventManager
```

The EventManager handles WebSocket connections for real-time events.
