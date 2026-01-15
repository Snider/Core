# Display Service

The Display service (`pkg/display`) provides comprehensive window management, dialogs, system tray, clipboard, and notification functionality.

## Features

- Window creation, positioning, and lifecycle
- System tray with menus
- Native dialogs (open, save, confirm)
- Clipboard operations
- System notifications
- Multi-monitor support
- Layout management

## Basic Usage

```go
import "github.com/Snider/Core/pkg/display"

// Register with Core
c, _ := core.New(
    core.WithService(display.NewService),
)

// Get service
svc := core.MustServiceFor[*display.Service](c, "display")
```

## Window Management

### Create Window

```go
info, err := svc.CreateWindow(display.CreateWindowOptions{
    Name:   "settings",
    Title:  "Settings",
    URL:    "/settings",
    Width:  800,
    Height: 600,
    X:      100,
    Y:      100,
})
```

### Position & Size

```go
// Move window
svc.SetWindowPosition("main", 100, 100)

// Resize window
svc.SetWindowSize("main", 1200, 800)

// Both at once
svc.SetWindowBounds("main", 100, 100, 1200, 800)
```

### Window State

```go
svc.MaximizeWindow("main")
svc.MinimizeWindow("main")
svc.RestoreWindow("main")
svc.FocusWindow("main")
svc.SetWindowFullscreen("main", true)
svc.SetWindowAlwaysOnTop("main", true)
```

### List Windows

```go
windows := svc.ListWindowInfos()
for _, w := range windows {
    fmt.Printf("%s: %dx%d at (%d,%d)\n",
        w.Name, w.Width, w.Height, w.X, w.Y)
}
```

## Dialogs

### File Dialogs

```go
// Open file
path, err := svc.OpenSingleFileDialog(display.OpenFileOptions{
    Title: "Select File",
    Filters: []display.FileFilter{
        {DisplayName: "Images", Pattern: "*.png;*.jpg"},
    },
})

// Open multiple files
paths, err := svc.OpenFileDialog(display.OpenFileOptions{
    Title:         "Select Files",
    AllowMultiple: true,
})

// Save file
path, err := svc.SaveFileDialog(display.SaveFileOptions{
    Title:           "Save As",
    DefaultFilename: "document.txt",
})

// Select directory
dir, err := svc.OpenDirectoryDialog(display.OpenDirectoryOptions{
    Title: "Select Folder",
})
```

### Confirm Dialog

```go
confirmed, err := svc.ConfirmDialog("Delete File", "Are you sure?")
if confirmed {
    // User clicked Yes
}
```

## System Tray

```go
// Set tooltip
svc.SetTrayTooltip("My Application")

// Set label
svc.SetTrayLabel("Running")

// Set icon (PNG bytes)
iconData, _ := os.ReadFile("icon.png")
svc.SetTrayIcon(iconData)

// Set menu
svc.SetTrayMenu([]display.TrayMenuItem{
    {Label: "Open", ActionID: "open"},
    {Label: "Settings", ActionID: "settings"},
    {IsSeparator: true},
    {Label: "Quit", ActionID: "quit"},
})
```

## Clipboard

```go
// Write to clipboard
svc.WriteClipboard("Hello, World!")

// Read from clipboard
text, err := svc.ReadClipboard()

// Check if has content
hasContent := svc.HasClipboard()

// Clear clipboard
svc.ClearClipboard()
```

## Notifications

```go
// Basic notification
svc.ShowNotification(display.NotificationOptions{
    Title:   "Download Complete",
    Message: "Your file has been downloaded.",
})

// Convenience methods
svc.ShowInfoNotification("Info", "Operation completed")
svc.ShowWarningNotification("Warning", "Low disk space")
svc.ShowErrorNotification("Error", "Connection failed")
```

## Multi-Monitor Support

```go
// List all screens
screens := svc.GetScreens()

// Get primary screen
primary, _ := svc.GetPrimaryScreen()

// Get screen at point
screen, _ := svc.GetScreenAtPoint(500, 300)

// Get screen for window
screen, _ := svc.GetScreenForWindow("main")

// Get work areas (excluding dock/taskbar)
workAreas := svc.GetWorkAreas()
```

## Layout Management

```go
// Save current layout
svc.SaveLayout("coding")

// Restore layout
svc.RestoreLayout("coding")

// List saved layouts
layouts := svc.ListLayouts()

// Tile windows
svc.TileWindows(display.TileModeGrid, nil)

// Snap to edge
svc.SnapWindow("main", display.SnapPositionLeft)

// Apply workflow preset
svc.ApplyWorkflowLayout(display.WorkflowCoding)
```

## Theme

```go
theme := svc.GetTheme()
fmt.Println(theme.IsDark)
```

## Frontend Usage (TypeScript)

```typescript
import {
    CreateWindow,
    SetWindowPosition,
    ShowNotification,
    OpenFileDialog
} from '@bindings/display/service';

// Create window
await CreateWindow({
    name: "settings",
    title: "Settings",
    width: 800,
    height: 600
});

// Show notification
await ShowNotification({
    title: "Success",
    message: "Settings saved!"
});
```
