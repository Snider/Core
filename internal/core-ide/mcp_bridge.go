package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/host-uk/core-gui/pkg/webview"
	"github.com/host-uk/core-gui/pkg/ws"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MCPBridge wires together WebView and WebSocket services
// and starts the MCP HTTP server after Wails initializes.
type MCPBridge struct {
	webview      *webview.Service
	wsHub        *ws.Hub
	claudeBridge *ClaudeBridge
	app          *application.App
	port         int
	running      bool
	mu           sync.Mutex
}

// NewMCPBridge creates a new MCP bridge with all services wired up.
func NewMCPBridge(port int) *MCPBridge {
	wv := webview.New()
	hub := ws.NewHub()

	// Create Claude bridge to forward messages to MCP core on port 9876
	claudeBridge := NewClaudeBridge("ws://localhost:9876/ws")

	return &MCPBridge{
		webview:      wv,
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
	// Wait for windows to be created (poll with timeout)
	var windows []webview.WindowInfo
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		windows = b.webview.ListWindows()
		if len(windows) > 0 {
			break
		}
	}
	if len(windows) == 0 {
		log.Printf("MCP Bridge: no windows found after waiting")
		return
	}
	for _, w := range windows {
		if err := b.webview.InjectConsoleCapture(w.Name); err != nil {
			log.Printf("Failed to inject console capture in %s: %v", w.Name, err)
		}
	}
}

// startHTTPServer starts the HTTP server for MCP and WebSocket.
func (b *MCPBridge) startHTTPServer() {
	b.mu.Lock()
	b.running = true
	b.mu.Unlock()

	// Start the WebSocket hub
	hubCtx := context.Background()
	go b.wsHub.Run(hubCtx)

	// Claude bridge disabled - port 9876 is not an MCP WebSocket server
	// b.claudeBridge.Start()

	mux := http.NewServeMux()

	// WebSocket endpoint for GUI clients
	mux.HandleFunc("/ws", b.wsHub.HandleWebSocket)

	// MCP info endpoint
	mux.HandleFunc("/mcp", b.handleMCPInfo)

	// MCP tools endpoint
	mux.HandleFunc("/mcp/tools", b.handleMCPTools)
	mux.HandleFunc("/mcp/call", b.handleMCPCall)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"mcp":     true,
			"webview": b.webview != nil,
		})
	})

	addr := fmt.Sprintf("127.0.0.1:%d", b.port)
	log.Printf("MCP HTTP server listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("MCP HTTP server error: %v", err)
	}
}

// handleMCPInfo returns MCP server information.
func (b *MCPBridge) handleMCPInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost")

	info := map[string]any{
		"name":    "core-ide",
		"version": "0.1.0",
		"capabilities": map[string]any{
			"webview":   true,
			"websocket": fmt.Sprintf("ws://localhost:%d/ws", b.port),
		},
	}
	json.NewEncoder(w).Encode(info)
}

// handleMCPTools returns the list of available tools.
func (b *MCPBridge) handleMCPTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost")

	tools := []map[string]string{
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
	}
	json.NewEncoder(w).Encode(map[string]any{"tools": tools})
}

// handleMCPCall handles tool calls via HTTP POST.
func (b *MCPBridge) handleMCPCall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost")
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

	// Limit request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result := b.executeWebviewTool(req.Tool, req.Params)
	json.NewEncoder(w).Encode(result)
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
		rawURL := getStringParam(params, "url")
		parsed, err := url.Parse(rawURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return map[string]any{"error": "only http/https URLs are allowed"}
		}
		err = b.webview.Navigate(windowName, rawURL)
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
		return map[string]any{"error": "unknown tool", "tool": tool}
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
