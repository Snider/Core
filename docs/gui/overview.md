# GUI Application

The Core GUI (`cmd/core-gui`) is a Wails v3 desktop application that demonstrates the Core framework capabilities with integrated MCP support.

## Features

- Angular frontend with Wails bindings
- MCP HTTP server for AI tool integration
- WebView automation capabilities
- Real-time WebSocket communication
- System tray support
- Multi-window management

## Architecture

```
┌─────────────────────────────────────────┐
│              Angular Frontend            │
│         (TypeScript + Wails Bindings)   │
└─────────────────┬───────────────────────┘
                  │ IPC
┌─────────────────┴───────────────────────┐
│              Wails Runtime              │
│         (Window, Events, Bindings)      │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────┴───────────────────────┐
│              MCP Bridge                  │
│   ┌─────────┬──────────┬─────────────┐  │
│   │ Display │ WebView  │ WebSocket   │  │
│   │ Service │ Service  │ Hub         │  │
│   └─────────┴──────────┴─────────────┘  │
└─────────────────────────────────────────┘
```

## Running the GUI

### Development Mode

```bash
# From project root
task gui:dev

# Or directly
cd cmd/core-gui
wails3 dev
```

### Production Build

```bash
task gui:build
```

## Directory Structure

```
cmd/core-gui/
├── main.go           # Application entry point
├── mcp_bridge.go     # MCP HTTP server and tool handler
├── claude_bridge.go  # Claude MCP client (optional)
├── frontend/         # Angular application
│   ├── src/
│   │   ├── app/      # Angular components
│   │   └── lib/      # Shared utilities
│   └── bindings/     # Generated Wails bindings
└── public/           # Static assets
```

## Services Integrated

The GUI integrates several Core services:

| Service | Purpose |
|---------|---------|
| Display | Window management, dialogs, tray |
| WebView | JavaScript execution, DOM interaction |
| MCP | AI tool protocol server |
| WebSocket | Real-time communication |

## Configuration

The application uses the Config service for settings:

```go
// Default settings
DefaultRoute: "/"
Language: "en"
Features: []
```

## Frontend Bindings

Wails generates TypeScript bindings for Go services:

```typescript
import { CreateWindow, ShowNotification } from '@bindings/display/service';
import { Translate, SetLanguage } from '@bindings/i18n/service';

// Create a new window
await CreateWindow({
    name: "settings",
    title: "Settings",
    width: 800,
    height: 600
});

// Show notification
await ShowNotification({
    title: "Success",
    message: "Operation completed!"
});
```

## WebSocket Communication

Connect to the WebSocket endpoint for real-time updates:

```typescript
const ws = new WebSocket('ws://localhost:9877/ws');

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};

ws.send(JSON.stringify({
    type: 'ping',
    data: {}
}));
```

## System Tray

The application includes system tray support:

```go
// Set tray menu
display.SetTrayMenu([]display.TrayMenuItem{
    {Label: "Open", ActionID: "open"},
    {Label: "Settings", ActionID: "settings"},
    {IsSeparator: true},
    {Label: "Quit", ActionID: "quit"},
})
```

## Building for Distribution

### macOS

```bash
task gui:build
# Creates: build/bin/core-gui.app
```

### Windows

```bash
task gui:build
# Creates: build/bin/core-gui.exe
```

### Linux

```bash
task gui:build
# Creates: build/bin/core-gui
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `MCP_PORT` | MCP server port (default: 9877) |
| `DEBUG` | Enable debug logging |
