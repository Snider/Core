# MCP Bridge

The MCP Bridge (`cmd/core-gui/mcp_bridge.go`) connects the Model Context Protocol server with Display, WebView, and WebSocket services.

## Overview

The MCP Bridge provides an HTTP API for AI assistants to interact with the desktop application, enabling:

- Window and screen management
- JavaScript execution in webviews
- DOM interaction (click, type, select)
- Screenshot capture
- File and process management
- Real-time events via WebSocket

## HTTP Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check |
| `GET /mcp` | Server capabilities |
| `GET /mcp/tools` | List available tools |
| `POST /mcp/call` | Execute a tool |
| `WS /ws` | WebSocket for GUI clients |
| `WS /events` | WebSocket for display events |

## Server Capabilities

```bash
curl http://localhost:9877/mcp
```

Response:

```json
{
    "name": "core",
    "version": "0.1.0",
    "capabilities": {
        "webview": true,
        "display": true,
        "windowControl": true,
        "screenControl": true,
        "websocket": "ws://localhost:9877/ws",
        "events": "ws://localhost:9877/events"
    }
}
```

## Tool Categories

### File Operations

| Tool | Description |
|------|-------------|
| `file_read` | Read file contents |
| `file_write` | Write content to file |
| `file_edit` | Edit file by replacing text |
| `file_delete` | Delete a file |
| `file_exists` | Check if file exists |
| `dir_list` | List directory contents |
| `dir_create` | Create directory |

### Window Control

| Tool | Description |
|------|-------------|
| `window_list` | List all windows |
| `window_create` | Create new window |
| `window_close` | Close window |
| `window_position` | Move window |
| `window_size` | Resize window |
| `window_maximize` | Maximize window |
| `window_minimize` | Minimize window |
| `window_focus` | Bring window to front |

### WebView Interaction

| Tool | Description |
|------|-------------|
| `webview_eval` | Execute JavaScript |
| `webview_click` | Click element |
| `webview_type` | Type into element |
| `webview_screenshot` | Capture page |
| `webview_navigate` | Navigate to URL |
| `webview_console` | Get console messages |

### Screen Management

| Tool | Description |
|------|-------------|
| `screen_list` | List all monitors |
| `screen_primary` | Get primary screen |
| `screen_at_point` | Get screen at coordinates |
| `screen_work_areas` | Get usable screen space |

### Layout Management

| Tool | Description |
|------|-------------|
| `layout_save` | Save window arrangement |
| `layout_restore` | Restore saved layout |
| `layout_tile` | Auto-tile windows |
| `layout_snap` | Snap window to edge |

## Calling Tools

```bash
# List windows
curl -X POST http://localhost:9877/mcp/call \
  -H "Content-Type: application/json" \
  -d '{"tool": "window_list", "params": {}}'

# Move window
curl -X POST http://localhost:9877/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "window_position",
    "params": {"name": "main", "x": 100, "y": 100}
  }'

# Execute JavaScript
curl -X POST http://localhost:9877/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "webview_eval",
    "params": {
      "window": "main",
      "code": "document.title"
    }
  }'

# Click element
curl -X POST http://localhost:9877/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "webview_click",
    "params": {
      "window": "main",
      "selector": "#submit-button"
    }
  }'

# Take screenshot
curl -X POST http://localhost:9877/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "webview_screenshot",
    "params": {"window": "main"}
  }'
```

## WebSocket Events

Connect to `/events` for real-time display events:

```javascript
const ws = new WebSocket('ws://localhost:9877/events');

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    switch (data.type) {
        case 'window.focus':
            console.log('Window focused:', data.name);
            break;
        case 'window.move':
            console.log('Window moved:', data.name, data.x, data.y);
            break;
        case 'theme.change':
            console.log('Theme changed:', data.isDark);
            break;
    }
};
```

Event types:

- `window.focus` - Window received focus
- `window.blur` - Window lost focus
- `window.move` - Window position changed
- `window.resize` - Window size changed
- `window.close` - Window was closed
- `window.create` - New window created
- `theme.change` - System theme changed
- `screen.change` - Screen configuration changed

## Go Integration

```go
import "github.com/Snider/Core/cmd/core-gui"

// Create bridge
bridge := NewMCPBridge(9877, displayService)

// Access services
mcpSvc := bridge.GetMCPService()
webview := bridge.GetWebView()
display := bridge.GetDisplay()
```

## Configuration

The bridge starts automatically on Wails app startup via the `ServiceStartup` lifecycle hook:

```go
func (b *MCPBridge) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
    b.app = application.Get()
    b.webview.SetApp(b.app)
    go b.startHTTPServer()
    return nil
}
```

## Security

The MCP server binds to localhost only by default. For production:

- Consider firewall rules
- Add authentication if needed
- Limit exposed tools
