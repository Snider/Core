# MCP Service

The MCP service (`pkg/mcp`) implements the [Model Context Protocol](https://modelcontextprotocol.io/), enabling AI assistants like Claude to interact with your application.

## Overview

MCP provides a standardized way for AI tools to:

- Execute operations in your application
- Query application state
- Interact with the UI
- Manage files and processes

## Basic Setup

```go
import "github.com/Snider/Core/pkg/mcp"

// Create standalone MCP server
mcpService := mcp.NewStandaloneWithPort(9877)

// Or integrate with Core
c, _ := core.New(
    core.WithService(mcp.NewService),
)
```

## Available Tools

The MCP service exposes numerous tools organized by category:

### File Operations

| Tool | Description |
|------|-------------|
| `file_read` | Read file contents |
| `file_write` | Write content to file |
| `file_edit` | Replace text in file |
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
| `window_focus` | Bring to front |

### WebView Interaction

| Tool | Description |
|------|-------------|
| `webview_eval` | Execute JavaScript |
| `webview_click` | Click element |
| `webview_type` | Type into element |
| `webview_screenshot` | Capture page |
| `webview_navigate` | Navigate to URL |
| `webview_console` | Get console logs |

### Process Management

| Tool | Description |
|------|-------------|
| `process_start` | Start a process |
| `process_stop` | Stop a process |
| `process_list` | List running processes |
| `process_output` | Get process output |

## HTTP API

The MCP service exposes an HTTP API:

```bash
# Health check
curl http://localhost:9877/health

# List available tools
curl http://localhost:9877/mcp/tools

# Call a tool
curl -X POST http://localhost:9877/mcp/call \
  -H "Content-Type: application/json" \
  -d '{"tool": "window_list", "params": {}}'
```

## WebSocket Events

Connect to `/events` for real-time updates:

```javascript
const ws = new WebSocket('ws://localhost:9877/events');
ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Event:', data.type, data.data);
};
```

## Integration with Display Service

```go
mcpService := mcp.NewStandaloneWithPort(9877)
mcpService.SetDisplay(displayService)
mcpService.SetWebView(webviewService)
```

## Example: Claude Integration

When Claude connects via MCP, it can:

```
User: "Move the settings window to the left side of the screen"

Claude uses: window_position("settings", 0, 100)
```

```
User: "Take a screenshot of the app"

Claude uses: webview_screenshot("main")
```

```
User: "Click the submit button"

Claude uses: webview_click("main", "#submit-btn")
```

## Security Considerations

- MCP server binds to localhost by default
- No authentication (designed for local AI assistants)
- Consider firewall rules for production

## Configuration

```go
// Custom port
mcp.NewStandaloneWithPort(8080)

// With all services
bridge := NewMCPBridge(9877, displayService)
bridge.SetWebView(webviewService)
```
