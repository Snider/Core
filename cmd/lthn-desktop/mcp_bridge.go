package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/host-uk/core/pkg/display"
	"github.com/host-uk/core/pkg/mcp"
	"github.com/host-uk/core/pkg/webview"
	"github.com/host-uk/core/pkg/ws"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MCPBridge wires together MCP, WebView, Display and WebSocket services
// and starts the MCP HTTP server after Wails initializes.
type MCPBridge struct {
	mcpService   *mcp.Service
	webview      *webview.Service
	display      *display.Service
	wsHub        *ws.Hub
	claudeBridge *ClaudeBridge
	app          *application.App
	port         int
	running      bool
	mu           sync.Mutex
}

// NewMCPBridge creates a new MCP bridge with all services wired up.
func NewMCPBridge(port int, displaySvc *display.Service) *MCPBridge {
	wv := webview.New()
	hub := ws.NewHub()
	mcpSvc := mcp.NewStandaloneWithPort(port)
	mcpSvc.SetWebView(wv)
	mcpSvc.SetDisplay(displaySvc)

	// Create Claude bridge to forward messages to MCP core on port 9876
	claudeBridge := NewClaudeBridge("ws://localhost:9876/ws")

	return &MCPBridge{
		mcpService:   mcpSvc,
		webview:      wv,
		display:      displaySvc,
		wsHub:        hub,
		claudeBridge: claudeBridge,
		port:         port,
	}
}

// ServiceStartup is called by Wails when the app starts.
// This wires up the app reference and starts the HTTP server.
func (b *MCPBridge) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get the Wails app reference
	b.app = application.Get()
	if b.app == nil {
		return fmt.Errorf("failed to get Wails app reference")
	}

	// Wire up the WebView service with the app
	b.webview.SetApp(b.app)

	// Set up console listener
	b.webview.SetupConsoleListener()

	// Inject console capture into all windows after a short delay
	// (windows may not be created yet)
	go b.injectConsoleCapture()

	// Start the HTTP server for MCP
	go b.startHTTPServer()

	log.Printf("MCP Bridge started on port %d", b.port)
	return nil
}

// injectConsoleCapture injects the console capture script into windows.
func (b *MCPBridge) injectConsoleCapture() {
	// Wait a bit for windows to be created
	// In production, you'd use events to detect window creation
	windows := b.webview.ListWindows()
	for _, w := range windows {
		if err := b.webview.InjectConsoleCapture(w.Name); err != nil {
			log.Printf("Failed to inject console capture in %s: %v", w.Name, err)
		}
	}
}

// startHTTPServer starts the HTTP server for MCP and WebSocket.
func (b *MCPBridge) startHTTPServer() {
	b.running = true

	// Start the WebSocket hub
	hubCtx := context.Background()
	go b.wsHub.Run(hubCtx)

	// Claude bridge disabled - port 9876 is not an MCP WebSocket server
	// b.claudeBridge.Start()

	mux := http.NewServeMux()

	// WebSocket endpoint for GUI clients
	mux.HandleFunc("/ws", b.wsHub.HandleWebSocket)

	// WebSocket endpoint for real-time display events
	mux.HandleFunc("/events", b.handleEventsWebSocket)

	// MCP info endpoint
	mux.HandleFunc("/mcp", b.handleMCPInfo)

	// MCP tools endpoint (simple HTTP for now, SSE later)
	mux.HandleFunc("/mcp/tools", b.handleMCPTools)
	mux.HandleFunc("/mcp/call", b.handleMCPCall)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"mcp":     true,
			"webview": b.webview != nil,
			"display": b.display != nil,
		})
	})

	addr := fmt.Sprintf(":%d", b.port)
	log.Printf("MCP HTTP server listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("MCP HTTP server error: %v", err)
	}
}

// handleMCPInfo returns MCP server information.
func (b *MCPBridge) handleMCPInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	info := map[string]any{
		"name":    "core",
		"version": "0.1.0",
		"capabilities": map[string]any{
			"webview":       true,
			"display":       b.display != nil,
			"windowControl": b.display != nil,
			"screenControl": b.display != nil,
			"websocket":     fmt.Sprintf("ws://localhost:%d/ws", b.port),
			"events":        fmt.Sprintf("ws://localhost:%d/events", b.port),
		},
	}
	json.NewEncoder(w).Encode(info)
}

// handleMCPTools returns the list of available tools.
func (b *MCPBridge) handleMCPTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Return tool list - grouped by category
	tools := []map[string]string{
		// File operations
		{"name": "file_read", "description": "Read the contents of a file"},
		{"name": "file_write", "description": "Write content to a file"},
		{"name": "file_edit", "description": "Edit a file by replacing text"},
		{"name": "file_delete", "description": "Delete a file"},
		{"name": "file_exists", "description": "Check if file exists"},
		{"name": "file_rename", "description": "Rename or move a file"},
		{"name": "dir_list", "description": "List directory contents"},
		{"name": "dir_create", "description": "Create a directory"},
		{"name": "lang_detect", "description": "Detect file language"},
		{"name": "lang_list", "description": "List supported languages"},
		// Process management
		{"name": "process_start", "description": "Start a process"},
		{"name": "process_stop", "description": "Stop a process"},
		{"name": "process_kill", "description": "Kill a process"},
		{"name": "process_list", "description": "List processes"},
		{"name": "process_output", "description": "Get process output"},
		{"name": "process_input", "description": "Send input to process"},
		// WebSocket streaming
		{"name": "ws_start", "description": "Start WebSocket server"},
		{"name": "ws_info", "description": "Get WebSocket info"},
		// WebView interaction (JS runtime, console, DOM)
		{"name": "webview_list", "description": "List windows"},
		{"name": "webview_eval", "description": "Execute JavaScript"},
		{"name": "webview_console", "description": "Get console messages"},
		{"name": "webview_console_clear", "description": "Clear console buffer"},
		{"name": "webview_click", "description": "Click element"},
		{"name": "webview_type", "description": "Type into element"},
		{"name": "webview_query", "description": "Query DOM elements"},
		{"name": "webview_navigate", "description": "Navigate to URL"},
		{"name": "webview_source", "description": "Get page source"},
		{"name": "webview_url", "description": "Get current page URL"},
		{"name": "webview_title", "description": "Get current page title"},
		{"name": "webview_screenshot", "description": "Capture page as base64 PNG"},
		{"name": "webview_screenshot_element", "description": "Capture specific element as PNG"},
		{"name": "webview_scroll", "description": "Scroll to element or position"},
		{"name": "webview_hover", "description": "Hover over element"},
		{"name": "webview_select", "description": "Select option in dropdown"},
		{"name": "webview_check", "description": "Check/uncheck checkbox or radio"},
		{"name": "webview_element_info", "description": "Get detailed info about element"},
		{"name": "webview_computed_style", "description": "Get computed styles for element"},
		{"name": "webview_highlight", "description": "Visually highlight element"},
		{"name": "webview_dom_tree", "description": "Get DOM tree structure"},
		{"name": "webview_errors", "description": "Get captured error messages"},
		{"name": "webview_performance", "description": "Get performance metrics"},
		{"name": "webview_resources", "description": "List loaded resources"},
		{"name": "webview_network", "description": "Get network requests log"},
		{"name": "webview_network_clear", "description": "Clear network request log"},
		{"name": "webview_network_inject", "description": "Inject network interceptor for detailed logging"},
		{"name": "webview_pdf", "description": "Export page as PDF (base64 data URI)"},
		{"name": "webview_print", "description": "Open print dialog for window"},
		// Window/Display control (native app control)
		{"name": "window_list", "description": "List all windows with positions"},
		{"name": "window_get", "description": "Get info about a specific window"},
		{"name": "window_create", "description": "Create a new window at specific position"},
		{"name": "window_close", "description": "Close a window by name"},
		{"name": "window_position", "description": "Move a window to specific coordinates"},
		{"name": "window_size", "description": "Resize a window"},
		{"name": "window_bounds", "description": "Set position and size in one call"},
		{"name": "window_maximize", "description": "Maximize a window"},
		{"name": "window_minimize", "description": "Minimize a window"},
		{"name": "window_restore", "description": "Restore from maximized/minimized"},
		{"name": "window_focus", "description": "Bring window to front"},
		{"name": "window_focused", "description": "Get currently focused window"},
		{"name": "window_visibility", "description": "Show or hide a window"},
		{"name": "window_always_on_top", "description": "Pin window above others"},
		{"name": "window_title", "description": "Change window title"},
		{"name": "window_title_get", "description": "Get current window title"},
		{"name": "window_fullscreen", "description": "Toggle fullscreen mode"},
		{"name": "screen_list", "description": "List all screens/monitors"},
		{"name": "screen_get", "description": "Get specific screen by ID"},
		{"name": "screen_primary", "description": "Get primary screen info"},
		{"name": "screen_at_point", "description": "Get screen containing a point"},
		{"name": "screen_for_window", "description": "Get screen a window is on"},
		{"name": "screen_work_areas", "description": "Get usable screen space (excluding dock/menubar)"},
		// Layout management
		{"name": "layout_save", "description": "Save current window arrangement with a name"},
		{"name": "layout_restore", "description": "Restore a saved layout by name"},
		{"name": "layout_list", "description": "List all saved layouts"},
		{"name": "layout_delete", "description": "Delete a saved layout"},
		{"name": "layout_get", "description": "Get details of a specific layout"},
		{"name": "layout_tile", "description": "Auto-tile windows (left/right/grid/quadrants)"},
		{"name": "layout_snap", "description": "Snap window to screen edge/corner"},
		{"name": "layout_stack", "description": "Stack windows in cascade pattern"},
		{"name": "layout_workflow", "description": "Apply preset workflow layout (coding/debugging/presenting)"},
		// System tray
		{"name": "tray_set_icon", "description": "Set system tray icon"},
		{"name": "tray_set_tooltip", "description": "Set system tray tooltip"},
		{"name": "tray_set_label", "description": "Set system tray label"},
		{"name": "tray_set_menu", "description": "Set system tray menu items"},
		{"name": "tray_info", "description": "Get system tray info"},
		// Window background colour (for transparency)
		{"name": "window_background_colour", "description": "Set window background colour with alpha"},
		// System integration
		{"name": "clipboard_read", "description": "Read text from system clipboard"},
		{"name": "clipboard_write", "description": "Write text to system clipboard"},
		{"name": "clipboard_has", "description": "Check if clipboard has content"},
		{"name": "clipboard_clear", "description": "Clear the clipboard"},
		{"name": "notification_show", "description": "Show native system notification"},
		{"name": "notification_permission_request", "description": "Request notification permission"},
		{"name": "notification_permission_check", "description": "Check notification permission status"},
		{"name": "theme_get", "description": "Get current system theme (dark/light)"},
		{"name": "theme_system", "description": "Get system theme preference"},
		{"name": "focus_set", "description": "Set focus to specific window"},
		// Dialogs
		{"name": "dialog_open_file", "description": "Show file open dialog"},
		{"name": "dialog_save_file", "description": "Show file save dialog"},
		{"name": "dialog_open_directory", "description": "Show directory picker"},
		{"name": "dialog_confirm", "description": "Show confirmation dialog (yes/no)"},
		{"name": "dialog_prompt", "description": "Show input prompt dialog (not supported natively)"},
		// Event subscriptions (WebSocket)
		{"name": "event_info", "description": "Get WebSocket event server info and connected clients"},
	}
	json.NewEncoder(w).Encode(map[string]any{"tools": tools})
}

// handleMCPCall handles tool calls via HTTP POST.
// This provides a REST bridge for display/window tools.
func (b *MCPBridge) handleMCPCall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Tool   string         `json:"tool"`
		Params map[string]any `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Execute tools based on prefix
	var result map[string]any
	if len(req.Tool) > 8 && req.Tool[:8] == "webview_" {
		result = b.executeWebviewTool(req.Tool, req.Params)
	} else {
		result = b.executeDisplayTool(req.Tool, req.Params)
	}
	json.NewEncoder(w).Encode(result)
}

// executeDisplayTool handles window and screen tool execution.
func (b *MCPBridge) executeDisplayTool(tool string, params map[string]any) map[string]any {
	if b.display == nil {
		return map[string]any{"error": "display service not available"}
	}

	switch tool {
	case "window_list":
		windows := b.display.ListWindowInfos()
		return map[string]any{"windows": windows}

	case "window_get":
		name, _ := params["name"].(string)
		info, err := b.display.GetWindowInfo(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"window": info}

	case "window_position":
		name, _ := params["name"].(string)
		x, _ := params["x"].(float64)
		y, _ := params["y"].(float64)
		err := b.display.SetWindowPosition(name, int(x), int(y))
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "name": name, "x": int(x), "y": int(y)}

	case "window_size":
		name, _ := params["name"].(string)
		width, _ := params["width"].(float64)
		height, _ := params["height"].(float64)
		err := b.display.SetWindowSize(name, int(width), int(height))
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "name": name, "width": int(width), "height": int(height)}

	case "window_bounds":
		name, _ := params["name"].(string)
		x, _ := params["x"].(float64)
		y, _ := params["y"].(float64)
		width, _ := params["width"].(float64)
		height, _ := params["height"].(float64)
		err := b.display.SetWindowBounds(name, int(x), int(y), int(width), int(height))
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "name": name, "x": int(x), "y": int(y), "width": int(width), "height": int(height)}

	case "window_maximize":
		name, _ := params["name"].(string)
		err := b.display.MaximizeWindow(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "action": "maximize"}

	case "window_minimize":
		name, _ := params["name"].(string)
		err := b.display.MinimizeWindow(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "action": "minimize"}

	case "window_restore":
		name, _ := params["name"].(string)
		err := b.display.RestoreWindow(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "action": "restore"}

	case "window_focus":
		name, _ := params["name"].(string)
		err := b.display.FocusWindow(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "action": "focus"}

	case "screen_list":
		screens := b.display.GetScreens()
		return map[string]any{"screens": screens}

	case "screen_get":
		id := getStringParam(params, "id")
		screen, err := b.display.GetScreen(id)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"screen": screen}

	case "screen_primary":
		screen, err := b.display.GetPrimaryScreen()
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"screen": screen}

	case "screen_at_point":
		x := getIntParam(params, "x")
		y := getIntParam(params, "y")
		screen, err := b.display.GetScreenAtPoint(x, y)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"screen": screen}

	case "screen_for_window":
		name := getStringParam(params, "name")
		screen, err := b.display.GetScreenForWindow(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"screen": screen}

	case "window_create":
		opts := display.CreateWindowOptions{
			Name:   getStringParam(params, "name"),
			Title:  getStringParam(params, "title"),
			URL:    getStringParam(params, "url"),
			X:      getIntParam(params, "x"),
			Y:      getIntParam(params, "y"),
			Width:  getIntParam(params, "width"),
			Height: getIntParam(params, "height"),
		}
		info, err := b.display.CreateWindow(opts)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "window": info}

	case "window_close":
		name, _ := params["name"].(string)
		err := b.display.CloseWindow(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "action": "close"}

	case "window_visibility":
		name, _ := params["name"].(string)
		visible, _ := params["visible"].(bool)
		err := b.display.SetWindowVisibility(name, visible)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "visible": visible}

	case "window_always_on_top":
		name, _ := params["name"].(string)
		onTop, _ := params["onTop"].(bool)
		err := b.display.SetWindowAlwaysOnTop(name, onTop)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "alwaysOnTop": onTop}

	case "window_title":
		name, _ := params["name"].(string)
		title, _ := params["title"].(string)
		err := b.display.SetWindowTitle(name, title)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "title": title}

	case "window_title_get":
		name := getStringParam(params, "name")
		title, err := b.display.GetWindowTitle(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"title": title}

	case "window_fullscreen":
		name, _ := params["name"].(string)
		fullscreen, _ := params["fullscreen"].(bool)
		err := b.display.SetWindowFullscreen(name, fullscreen)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "fullscreen": fullscreen}

	case "screen_work_areas":
		areas := b.display.GetWorkAreas()
		return map[string]any{"workAreas": areas}

	case "window_focused":
		name := b.display.GetFocusedWindow()
		return map[string]any{"focused": name}

	case "layout_save":
		name, _ := params["name"].(string)
		err := b.display.SaveLayout(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "name": name}

	case "layout_restore":
		name, _ := params["name"].(string)
		err := b.display.RestoreLayout(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "name": name}

	case "layout_list":
		layouts := b.display.ListLayouts()
		return map[string]any{"layouts": layouts}

	case "layout_delete":
		name, _ := params["name"].(string)
		err := b.display.DeleteLayout(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "name": name}

	case "layout_get":
		name, _ := params["name"].(string)
		layout := b.display.GetLayout(name)
		if layout == nil {
			return map[string]any{"error": "layout not found", "name": name}
		}
		return map[string]any{"layout": layout}

	case "layout_tile":
		mode := getStringParam(params, "mode")
		var windowNames []string
		if names, ok := params["windows"].([]any); ok {
			for _, n := range names {
				if s, ok := n.(string); ok {
					windowNames = append(windowNames, s)
				}
			}
		}
		err := b.display.TileWindows(display.TileMode(mode), windowNames)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "mode": mode}

	case "layout_snap":
		name := getStringParam(params, "name")
		position := getStringParam(params, "position")
		err := b.display.SnapWindow(name, display.SnapPosition(position))
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "position": position}

	case "layout_stack":
		var windowNames []string
		if names, ok := params["windows"].([]any); ok {
			for _, n := range names {
				if s, ok := n.(string); ok {
					windowNames = append(windowNames, s)
				}
			}
		}
		offsetX := getIntParam(params, "offsetX")
		offsetY := getIntParam(params, "offsetY")
		err := b.display.StackWindows(windowNames, offsetX, offsetY)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "layout_workflow":
		workflow := getStringParam(params, "workflow")
		err := b.display.ApplyWorkflowLayout(display.WorkflowType(workflow))
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true, "workflow": workflow}

	case "tray_set_tooltip":
		tooltip := getStringParam(params, "tooltip")
		err := b.display.SetTrayTooltip(tooltip)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "tray_set_label":
		label := getStringParam(params, "label")
		err := b.display.SetTrayLabel(label)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "tray_set_icon":
		// Icon data as base64 encoded PNG
		iconBase64 := getStringParam(params, "icon")
		if iconBase64 == "" {
			return map[string]any{"error": "icon data required"}
		}
		// Decode base64
		iconData, err := base64.StdEncoding.DecodeString(iconBase64)
		if err != nil {
			return map[string]any{"error": "invalid base64 icon data: " + err.Error()}
		}
		err = b.display.SetTrayIcon(iconData)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "tray_set_menu":
		// Menu items as JSON array
		var items []display.TrayMenuItem
		if menuData, ok := params["menu"].([]any); ok {
			menuJSON, _ := json.Marshal(menuData)
			json.Unmarshal(menuJSON, &items)
		}
		err := b.display.SetTrayMenu(items)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "tray_info":
		info := b.display.GetTrayInfo()
		return info

	case "window_background_colour":
		name := getStringParam(params, "name")
		r := uint8(getIntParam(params, "r"))
		g := uint8(getIntParam(params, "g"))
		b_val := uint8(getIntParam(params, "b"))
		a := uint8(getIntParam(params, "a"))
		if a == 0 {
			a = 255 // Default to opaque
		}
		err := b.display.SetWindowBackgroundColour(name, r, g, b_val, a)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "clipboard_read":
		text, err := b.display.ReadClipboard()
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"text": text}

	case "clipboard_write":
		text, _ := params["text"].(string)
		err := b.display.WriteClipboard(text)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "clipboard_has":
		has := b.display.HasClipboard()
		return map[string]any{"hasContent": has}

	case "clipboard_clear":
		err := b.display.ClearClipboard()
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "notification_show":
		title := getStringParam(params, "title")
		message := getStringParam(params, "message")
		subtitle := getStringParam(params, "subtitle")
		id := getStringParam(params, "id")
		err := b.display.ShowNotification(display.NotificationOptions{
			ID:       id,
			Title:    title,
			Message:  message,
			Subtitle: subtitle,
		})
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "notification_permission_request":
		granted, err := b.display.RequestNotificationPermission()
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"granted": granted}

	case "notification_permission_check":
		authorized, err := b.display.CheckNotificationPermission()
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"authorized": authorized}

	case "theme_get":
		theme := b.display.GetTheme()
		return map[string]any{"theme": theme}

	case "theme_system":
		theme := b.display.GetSystemTheme()
		return map[string]any{"theme": theme}

	case "focus_set":
		name := getStringParam(params, "name")
		err := b.display.FocusWindow(name)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "dialog_open_file":
		title := getStringParam(params, "title")
		defaultDir := getStringParam(params, "defaultDirectory")
		multiple, _ := params["allowMultiple"].(bool)
		opts := display.OpenFileOptions{
			Title:            title,
			DefaultDirectory: defaultDir,
			AllowMultiple:    multiple,
		}
		if multiple {
			paths, err := b.display.OpenFileDialog(opts)
			if err != nil {
				return map[string]any{"error": err.Error()}
			}
			return map[string]any{"paths": paths}
		}
		path, err := b.display.OpenSingleFileDialog(opts)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"path": path}

	case "dialog_save_file":
		title := getStringParam(params, "title")
		defaultDir := getStringParam(params, "defaultDirectory")
		defaultFilename := getStringParam(params, "defaultFilename")
		path, err := b.display.SaveFileDialog(display.SaveFileOptions{
			Title:            title,
			DefaultDirectory: defaultDir,
			DefaultFilename:  defaultFilename,
		})
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"path": path}

	case "dialog_open_directory":
		title := getStringParam(params, "title")
		defaultDir := getStringParam(params, "defaultDirectory")
		path, err := b.display.OpenDirectoryDialog(display.OpenDirectoryOptions{
			Title:            title,
			DefaultDirectory: defaultDir,
		})
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"path": path}

	case "dialog_confirm":
		title := getStringParam(params, "title")
		message := getStringParam(params, "message")
		confirmed, err := b.display.ConfirmDialog(title, message)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"confirmed": confirmed}

	case "dialog_prompt":
		title := getStringParam(params, "title")
		message := getStringParam(params, "message")
		result, ok, err := b.display.PromptDialog(title, message)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"result": result, "ok": ok}

	case "event_info":
		eventMgr := b.display.GetEventManager()
		if eventMgr == nil {
			return map[string]any{"error": "event manager not available"}
		}
		return map[string]any{
			"endpoint":         fmt.Sprintf("ws://localhost:%d/events", b.port),
			"connectedClients": eventMgr.ConnectedClients(),
			"eventTypes": []string{
				"window.focus", "window.blur", "window.move", "window.resize",
				"window.close", "window.create", "theme.change", "screen.change",
			},
		}

	default:
		return map[string]any{"error": "unknown tool", "tool": tool}
	}
}

// executeWebviewTool handles webview/JS tool execution.
func (b *MCPBridge) executeWebviewTool(tool string, params map[string]any) map[string]any {
	if b.webview == nil {
		return map[string]any{"error": "webview service not available"}
	}

	switch tool {
	case "webview_list":
		windows := b.webview.ListWindows()
		return map[string]any{"windows": windows}

	case "webview_eval":
		windowName := getStringParam(params, "window")
		code := getStringParam(params, "code")
		result, err := b.webview.ExecJS(windowName, code)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"result": result}

	case "webview_console":
		level := getStringParam(params, "level")
		limit := getIntParam(params, "limit")
		if limit == 0 {
			limit = 100
		}
		messages := b.webview.GetConsoleMessages(level, limit)
		return map[string]any{"messages": messages}

	case "webview_console_clear":
		b.webview.ClearConsole()
		return map[string]any{"success": true}

	case "webview_click":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		err := b.webview.Click(windowName, selector)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_type":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		text := getStringParam(params, "text")
		err := b.webview.Type(windowName, selector, text)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_query":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		result, err := b.webview.QuerySelector(windowName, selector)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"elements": result}

	case "webview_navigate":
		windowName := getStringParam(params, "window")
		url := getStringParam(params, "url")
		err := b.webview.Navigate(windowName, url)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_source":
		windowName := getStringParam(params, "window")
		result, err := b.webview.GetPageSource(windowName)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"source": result}

	case "webview_url":
		windowName := getStringParam(params, "window")
		result, err := b.webview.GetURL(windowName)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"url": result}

	case "webview_title":
		windowName := getStringParam(params, "window")
		result, err := b.webview.GetTitle(windowName)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"title": result}

	case "webview_screenshot":
		windowName := getStringParam(params, "window")
		data, err := b.webview.Screenshot(windowName)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"data": data}

	case "webview_screenshot_element":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		data, err := b.webview.ScreenshotElement(windowName, selector)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"data": data}

	case "webview_scroll":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		x := getIntParam(params, "x")
		y := getIntParam(params, "y")
		err := b.webview.Scroll(windowName, selector, x, y)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_hover":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		err := b.webview.Hover(windowName, selector)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_select":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		value := getStringParam(params, "value")
		err := b.webview.Select(windowName, selector, value)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_check":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		checked, _ := params["checked"].(bool)
		err := b.webview.Check(windowName, selector, checked)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_element_info":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		result, err := b.webview.GetElementInfo(windowName, selector)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"element": result}

	case "webview_computed_style":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		var properties []string
		if props, ok := params["properties"].([]any); ok {
			for _, p := range props {
				if s, ok := p.(string); ok {
					properties = append(properties, s)
				}
			}
		}
		result, err := b.webview.GetComputedStyle(windowName, selector, properties)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"styles": result}

	case "webview_highlight":
		windowName := getStringParam(params, "window")
		selector := getStringParam(params, "selector")
		duration := getIntParam(params, "duration")
		err := b.webview.Highlight(windowName, selector, duration)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_dom_tree":
		windowName := getStringParam(params, "window")
		maxDepth := getIntParam(params, "maxDepth")
		result, err := b.webview.GetDOMTree(windowName, maxDepth)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"tree": result}

	case "webview_errors":
		limit := getIntParam(params, "limit")
		if limit == 0 {
			limit = 50
		}
		errors := b.webview.GetErrors(limit)
		return map[string]any{"errors": errors}

	case "webview_performance":
		windowName := getStringParam(params, "window")
		result, err := b.webview.GetPerformance(windowName)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"performance": result}

	case "webview_resources":
		windowName := getStringParam(params, "window")
		result, err := b.webview.GetResources(windowName)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"resources": result}

	case "webview_network":
		windowName := getStringParam(params, "window")
		limit := getIntParam(params, "limit")
		result, err := b.webview.GetNetworkRequests(windowName, limit)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"requests": result}

	case "webview_network_clear":
		windowName := getStringParam(params, "window")
		err := b.webview.ClearNetworkRequests(windowName)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_network_inject":
		windowName := getStringParam(params, "window")
		err := b.webview.InjectNetworkInterceptor(windowName)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	case "webview_pdf":
		windowName := getStringParam(params, "window")
		options := make(map[string]any)
		if filename := getStringParam(params, "filename"); filename != "" {
			options["filename"] = filename
		}
		if margin, ok := params["margin"].(float64); ok {
			options["margin"] = margin
		}
		data, err := b.webview.ExportToPDF(windowName, options)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"data": data}

	case "webview_print":
		windowName := getStringParam(params, "window")
		err := b.webview.PrintToPDF(windowName)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"success": true}

	default:
		return map[string]any{"error": "unknown webview tool", "tool": tool}
	}
}

// Helper functions for parameter extraction
func getStringParam(params map[string]any, key string) string {
	if v, ok := params[key].(string); ok {
		return v
	}
	return ""
}

func getIntParam(params map[string]any, key string) int {
	if v, ok := params[key].(float64); ok {
		return int(v)
	}
	return 0
}

// GetMCPService returns the MCP service for direct access.
func (b *MCPBridge) GetMCPService() *mcp.Service {
	return b.mcpService
}

// GetWebView returns the WebView service.
func (b *MCPBridge) GetWebView() *webview.Service {
	return b.webview
}

// GetDisplay returns the Display service.
func (b *MCPBridge) GetDisplay() *display.Service {
	return b.display
}

// handleEventsWebSocket handles WebSocket connections for real-time display events.
func (b *MCPBridge) handleEventsWebSocket(w http.ResponseWriter, r *http.Request) {
	eventMgr := b.display.GetEventManager()
	if eventMgr == nil {
		http.Error(w, "event manager not available", http.StatusServiceUnavailable)
		return
	}
	eventMgr.HandleWebSocket(w, r)
}
