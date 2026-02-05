# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Core is a Web3 Framework written in Go using Wails v3 to replace Electron for desktop applications. It provides a dependency injection framework for managing services with lifecycle support.

## Build & Development Commands

This project uses [Task](https://taskfile.dev/) for automation. Key commands:

```bash
# Run all tests
task test

# Generate test coverage
task cov
task cov-view          # Opens coverage HTML report

# GUI application (Wails)
task gui:dev           # Development mode with hot-reload
task gui:build         # Production build

# CLI application
task cli:build         # Build CLI
task cli:run           # Build and run CLI

# Code review
task review            # Submit for CodeRabbit review
task check             # Run mod tidy + tests + review
```

Run a single test: `go test -run TestName ./...`

## Architecture

### Core Framework (`core.go`, `interfaces.go`)

The `Core` struct is the central application container managing:
- **Services**: Named service registry with type-safe retrieval via `ServiceFor[T]()`
- **Actions/IPC**: Message-passing system where services communicate via `ACTION(msg Message)` and register handlers via `RegisterAction()`
- **Lifecycle**: Services implementing `Startable` (OnStartup) and/or `Stoppable` (OnShutdown) interfaces are automatically called during app lifecycle

Creating a Core instance:
```go
core, err := core.New(
    core.WithService(myServiceFactory),
    core.WithAssets(assets),
    core.WithServiceLock(),  // Prevents late service registration
)
```

### Service Registration Pattern

Services are registered via factory functions that receive the Core instance:
```go
func NewMyService(c *core.Core) (any, error) {
    return &MyService{runtime: core.NewServiceRuntime(c, opts)}, nil
}

core.New(core.WithService(NewMyService))
```

- `WithService`: Auto-discovers service name from package path, registers IPC handler if service has `HandleIPCEvents` method
- `WithName`: Explicitly names a service

### Runtime (`runtime_pkg.go`)

`Runtime` is the Wails service wrapper that bootstraps the Core and its services. Use `NewWithFactories()` for custom service registration or `NewRuntime()` for basic setup.

### ServiceRuntime Generic Helper (`runtime.go`)

Embed `ServiceRuntime[T]` in services to get access to Core and typed options:
```go
type MyService struct {
    *core.ServiceRuntime[MyServiceOptions]
}
```

### Error Handling (`e.go`)

Use the `E()` helper for contextual errors:
```go
return core.E("service.Method", "what failed", underlyingErr)
```

### Test Naming Convention

Tests use `_Good`, `_Bad`, `_Ugly` suffix pattern:
- `_Good`: Happy path tests
- `_Bad`: Expected error conditions
- `_Ugly`: Panic/edge cases

## Go Workspace

Uses Go 1.25 workspaces. The workspace includes:
- Root module (Core framework)
- `cmd/core-gui` (Wails GUI application)
- `cmd/bugseti` (BugSETI system tray app - distributed bug fixing)
- `cmd/examples/*` (Example applications)

After adding modules: `go work sync`

## Additional Packages

### pkg/ws (WebSocket Hub)

Real-time streaming via WebSocket connections. Implements a hub pattern for managing connections and channel-based subscriptions.

```go
hub := ws.NewHub()
go hub.Run(ctx)

// Register HTTP handler
http.HandleFunc("/ws", hub.Handler())

// Send process output to subscribers
hub.SendProcessOutput(processID, "output line")
```

Message types: `process_output`, `process_status`, `event`, `error`, `ping/pong`, `subscribe/unsubscribe`

### pkg/webview (Browser Automation)

Chrome DevTools Protocol (CDP) client for browser automation, testing, and scraping.

```go
wv, err := webview.New(webview.WithDebugURL("http://localhost:9222"))
defer wv.Close()

wv.Navigate("https://example.com")
wv.Click("#submit-button")
wv.Type("#input", "text")
screenshot, _ := wv.Screenshot()
```

Features: Navigation, DOM queries, console capture, screenshots, JavaScript evaluation, Angular helpers

### pkg/mcp (MCP Server)

Model Context Protocol server with tools for:
- **File operations**: file_read, file_write, file_edit, file_delete, file_rename, file_exists, dir_list, dir_create
- **RAG**: rag_query, rag_ingest, rag_collections (Qdrant + Ollama)
- **Metrics**: metrics_record, metrics_query (JSONL storage)
- **Language detection**: lang_detect, lang_list
- **Process management**: process_start, process_stop, process_kill, process_list, process_output, process_input
- **WebSocket**: ws_start, ws_info
- **Webview/CDP**: webview_connect, webview_navigate, webview_click, webview_type, webview_query, webview_console, webview_eval, webview_screenshot, webview_wait, webview_disconnect

Run server: `core mcp serve` (stdio) or `MCP_ADDR=:9000 core mcp serve` (TCP)

## BugSETI Application

System tray application for distributed bug fixing - "like SETI@home but for code".

Features:
- Fetches OSS issues from GitHub
- AI-powered context preparation via seeder
- Issue queue management
- Automated PR submission
- Stats tracking and leaderboard

Build: `task bugseti:build`
Run: `task bugseti:dev`