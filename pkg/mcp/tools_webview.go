package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/log"
	"github.com/host-uk/core/pkg/webview"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// webviewInstance holds the current webview connection.
// This is managed by the MCP service.
var webviewInstance *webview.Webview

// WebviewConnectInput contains parameters for connecting to Chrome DevTools.
type WebviewConnectInput struct {
	DebugURL string `json:"debug_url"`         // Chrome DevTools URL (e.g., http://localhost:9222)
	Timeout  int    `json:"timeout,omitempty"` // Default timeout in seconds (default: 30)
}

// WebviewConnectOutput contains the result of connecting to Chrome.
type WebviewConnectOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// WebviewNavigateInput contains parameters for navigating to a URL.
type WebviewNavigateInput struct {
	URL string `json:"url"` // URL to navigate to
}

// WebviewNavigateOutput contains the result of navigation.
type WebviewNavigateOutput struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
}

// WebviewClickInput contains parameters for clicking an element.
type WebviewClickInput struct {
	Selector string `json:"selector"` // CSS selector
}

// WebviewClickOutput contains the result of a click action.
type WebviewClickOutput struct {
	Success bool `json:"success"`
}

// WebviewTypeInput contains parameters for typing text.
type WebviewTypeInput struct {
	Selector string `json:"selector"` // CSS selector
	Text     string `json:"text"`     // Text to type
}

// WebviewTypeOutput contains the result of a type action.
type WebviewTypeOutput struct {
	Success bool `json:"success"`
}

// WebviewQueryInput contains parameters for querying an element.
type WebviewQueryInput struct {
	Selector string `json:"selector"`      // CSS selector
	All      bool   `json:"all,omitempty"` // If true, return all matching elements
}

// WebviewQueryOutput contains the result of a query.
type WebviewQueryOutput struct {
	Found    bool                 `json:"found"`
	Count    int                  `json:"count"`
	Elements []WebviewElementInfo `json:"elements,omitempty"`
}

// WebviewElementInfo represents information about a DOM element.
type WebviewElementInfo struct {
	NodeID      int                  `json:"nodeId"`
	TagName     string               `json:"tagName"`
	Attributes  map[string]string    `json:"attributes,omitempty"`
	BoundingBox *webview.BoundingBox `json:"boundingBox,omitempty"`
}

// WebviewConsoleInput contains parameters for getting console output.
type WebviewConsoleInput struct {
	Clear bool `json:"clear,omitempty"` // If true, clear console after getting messages
}

// WebviewConsoleOutput contains console messages.
type WebviewConsoleOutput struct {
	Messages []WebviewConsoleMessage `json:"messages"`
	Count    int                     `json:"count"`
}

// WebviewConsoleMessage represents a console message.
type WebviewConsoleMessage struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url,omitempty"`
	Line      int    `json:"line,omitempty"`
}

// WebviewEvalInput contains parameters for evaluating JavaScript.
type WebviewEvalInput struct {
	Script string `json:"script"` // JavaScript to evaluate
}

// WebviewEvalOutput contains the result of JavaScript evaluation.
type WebviewEvalOutput struct {
	Success bool   `json:"success"`
	Result  any    `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

// WebviewScreenshotInput contains parameters for taking a screenshot.
type WebviewScreenshotInput struct {
	Format string `json:"format,omitempty"` // "png" or "jpeg" (default: png)
}

// WebviewScreenshotOutput contains the screenshot data.
type WebviewScreenshotOutput struct {
	Success bool   `json:"success"`
	Data    string `json:"data"` // Base64 encoded image
	Format  string `json:"format"`
}

// WebviewWaitInput contains parameters for waiting operations.
type WebviewWaitInput struct {
	Selector string `json:"selector,omitempty"` // Wait for selector
	Timeout  int    `json:"timeout,omitempty"`  // Timeout in seconds
}

// WebviewWaitOutput contains the result of waiting.
type WebviewWaitOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// WebviewDisconnectInput contains parameters for disconnecting.
type WebviewDisconnectInput struct{}

// WebviewDisconnectOutput contains the result of disconnecting.
type WebviewDisconnectOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// registerWebviewTools adds webview tools to the MCP server.
func (s *Service) registerWebviewTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_connect",
		Description: "Connect to Chrome DevTools Protocol. Start Chrome with --remote-debugging-port=9222 first.",
	}, s.webviewConnect)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_disconnect",
		Description: "Disconnect from Chrome DevTools.",
	}, s.webviewDisconnect)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_navigate",
		Description: "Navigate the browser to a URL.",
	}, s.webviewNavigate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_click",
		Description: "Click on an element by CSS selector.",
	}, s.webviewClick)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_type",
		Description: "Type text into an element by CSS selector.",
	}, s.webviewType)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_query",
		Description: "Query DOM elements by CSS selector.",
	}, s.webviewQuery)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_console",
		Description: "Get browser console output.",
	}, s.webviewConsole)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_eval",
		Description: "Evaluate JavaScript in the browser context.",
	}, s.webviewEval)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_screenshot",
		Description: "Capture a screenshot of the browser window.",
	}, s.webviewScreenshot)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "webview_wait",
		Description: "Wait for an element to appear by CSS selector.",
	}, s.webviewWait)
}

// webviewConnect handles the webview_connect tool call.
func (s *Service) webviewConnect(ctx context.Context, req *mcp.CallToolRequest, input WebviewConnectInput) (*mcp.CallToolResult, WebviewConnectOutput, error) {
	s.logger.Security("MCP tool execution", "tool", "webview_connect", "debug_url", input.DebugURL, "user", log.Username())

	if input.DebugURL == "" {
		return nil, WebviewConnectOutput{}, fmt.Errorf("debug_url is required")
	}

	// Close existing connection if any
	if webviewInstance != nil {
		_ = webviewInstance.Close()
		webviewInstance = nil
	}

	// Set up options
	opts := []webview.Option{
		webview.WithDebugURL(input.DebugURL),
	}

	if input.Timeout > 0 {
		opts = append(opts, webview.WithTimeout(time.Duration(input.Timeout)*time.Second))
	}

	// Create new webview instance
	wv, err := webview.New(opts...)
	if err != nil {
		log.Error("mcp: webview connect failed", "debug_url", input.DebugURL, "err", err)
		return nil, WebviewConnectOutput{}, fmt.Errorf("failed to connect: %w", err)
	}

	webviewInstance = wv

	return nil, WebviewConnectOutput{
		Success: true,
		Message: fmt.Sprintf("Connected to Chrome DevTools at %s", input.DebugURL),
	}, nil
}

// webviewDisconnect handles the webview_disconnect tool call.
func (s *Service) webviewDisconnect(ctx context.Context, req *mcp.CallToolRequest, input WebviewDisconnectInput) (*mcp.CallToolResult, WebviewDisconnectOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "webview_disconnect", "user", log.Username())

	if webviewInstance == nil {
		return nil, WebviewDisconnectOutput{
			Success: true,
			Message: "No active connection",
		}, nil
	}

	if err := webviewInstance.Close(); err != nil {
		log.Error("mcp: webview disconnect failed", "err", err)
		return nil, WebviewDisconnectOutput{}, fmt.Errorf("failed to disconnect: %w", err)
	}

	webviewInstance = nil

	return nil, WebviewDisconnectOutput{
		Success: true,
		Message: "Disconnected from Chrome DevTools",
	}, nil
}

// webviewNavigate handles the webview_navigate tool call.
func (s *Service) webviewNavigate(ctx context.Context, req *mcp.CallToolRequest, input WebviewNavigateInput) (*mcp.CallToolResult, WebviewNavigateOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "webview_navigate", "url", input.URL, "user", log.Username())

	if webviewInstance == nil {
		return nil, WebviewNavigateOutput{}, fmt.Errorf("not connected; use webview_connect first")
	}

	if input.URL == "" {
		return nil, WebviewNavigateOutput{}, fmt.Errorf("url is required")
	}

	if err := webviewInstance.Navigate(input.URL); err != nil {
		log.Error("mcp: webview navigate failed", "url", input.URL, "err", err)
		return nil, WebviewNavigateOutput{}, fmt.Errorf("failed to navigate: %w", err)
	}

	return nil, WebviewNavigateOutput{
		Success: true,
		URL:     input.URL,
	}, nil
}

// webviewClick handles the webview_click tool call.
func (s *Service) webviewClick(ctx context.Context, req *mcp.CallToolRequest, input WebviewClickInput) (*mcp.CallToolResult, WebviewClickOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "webview_click", "selector", input.Selector, "user", log.Username())

	if webviewInstance == nil {
		return nil, WebviewClickOutput{}, fmt.Errorf("not connected; use webview_connect first")
	}

	if input.Selector == "" {
		return nil, WebviewClickOutput{}, fmt.Errorf("selector is required")
	}

	if err := webviewInstance.Click(input.Selector); err != nil {
		log.Error("mcp: webview click failed", "selector", input.Selector, "err", err)
		return nil, WebviewClickOutput{}, fmt.Errorf("failed to click: %w", err)
	}

	return nil, WebviewClickOutput{Success: true}, nil
}

// webviewType handles the webview_type tool call.
func (s *Service) webviewType(ctx context.Context, req *mcp.CallToolRequest, input WebviewTypeInput) (*mcp.CallToolResult, WebviewTypeOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "webview_type", "selector", input.Selector, "user", log.Username())

	if webviewInstance == nil {
		return nil, WebviewTypeOutput{}, fmt.Errorf("not connected; use webview_connect first")
	}

	if input.Selector == "" {
		return nil, WebviewTypeOutput{}, fmt.Errorf("selector is required")
	}

	if err := webviewInstance.Type(input.Selector, input.Text); err != nil {
		log.Error("mcp: webview type failed", "selector", input.Selector, "err", err)
		return nil, WebviewTypeOutput{}, fmt.Errorf("failed to type: %w", err)
	}

	return nil, WebviewTypeOutput{Success: true}, nil
}

// webviewQuery handles the webview_query tool call.
func (s *Service) webviewQuery(ctx context.Context, req *mcp.CallToolRequest, input WebviewQueryInput) (*mcp.CallToolResult, WebviewQueryOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "webview_query", "selector", input.Selector, "all", input.All, "user", log.Username())

	if webviewInstance == nil {
		return nil, WebviewQueryOutput{}, fmt.Errorf("not connected; use webview_connect first")
	}

	if input.Selector == "" {
		return nil, WebviewQueryOutput{}, fmt.Errorf("selector is required")
	}

	if input.All {
		elements, err := webviewInstance.QuerySelectorAll(input.Selector)
		if err != nil {
			log.Error("mcp: webview query all failed", "selector", input.Selector, "err", err)
			return nil, WebviewQueryOutput{}, fmt.Errorf("failed to query: %w", err)
		}

		output := WebviewQueryOutput{
			Found:    len(elements) > 0,
			Count:    len(elements),
			Elements: make([]WebviewElementInfo, len(elements)),
		}

		for i, elem := range elements {
			output.Elements[i] = WebviewElementInfo{
				NodeID:      elem.NodeID,
				TagName:     elem.TagName,
				Attributes:  elem.Attributes,
				BoundingBox: elem.BoundingBox,
			}
		}

		return nil, output, nil
	}

	elem, err := webviewInstance.QuerySelector(input.Selector)
	if err != nil {
		// Element not found is not necessarily an error
		return nil, WebviewQueryOutput{
			Found: false,
			Count: 0,
		}, nil
	}

	return nil, WebviewQueryOutput{
		Found: true,
		Count: 1,
		Elements: []WebviewElementInfo{{
			NodeID:      elem.NodeID,
			TagName:     elem.TagName,
			Attributes:  elem.Attributes,
			BoundingBox: elem.BoundingBox,
		}},
	}, nil
}

// webviewConsole handles the webview_console tool call.
func (s *Service) webviewConsole(ctx context.Context, req *mcp.CallToolRequest, input WebviewConsoleInput) (*mcp.CallToolResult, WebviewConsoleOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "webview_console", "clear", input.Clear, "user", log.Username())

	if webviewInstance == nil {
		return nil, WebviewConsoleOutput{}, fmt.Errorf("not connected; use webview_connect first")
	}

	messages := webviewInstance.GetConsole()

	output := WebviewConsoleOutput{
		Messages: make([]WebviewConsoleMessage, len(messages)),
		Count:    len(messages),
	}

	for i, msg := range messages {
		output.Messages[i] = WebviewConsoleMessage{
			Type:      msg.Type,
			Text:      msg.Text,
			Timestamp: msg.Timestamp.Format(time.RFC3339),
			URL:       msg.URL,
			Line:      msg.Line,
		}
	}

	if input.Clear {
		webviewInstance.ClearConsole()
	}

	return nil, output, nil
}

// webviewEval handles the webview_eval tool call.
func (s *Service) webviewEval(ctx context.Context, req *mcp.CallToolRequest, input WebviewEvalInput) (*mcp.CallToolResult, WebviewEvalOutput, error) {
	s.logger.Security("MCP tool execution", "tool", "webview_eval", "user", log.Username())

	if webviewInstance == nil {
		return nil, WebviewEvalOutput{}, fmt.Errorf("not connected; use webview_connect first")
	}

	if input.Script == "" {
		return nil, WebviewEvalOutput{}, fmt.Errorf("script is required")
	}

	result, err := webviewInstance.Evaluate(input.Script)
	if err != nil {
		log.Error("mcp: webview eval failed", "err", err)
		return nil, WebviewEvalOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return nil, WebviewEvalOutput{
		Success: true,
		Result:  result,
	}, nil
}

// webviewScreenshot handles the webview_screenshot tool call.
func (s *Service) webviewScreenshot(ctx context.Context, req *mcp.CallToolRequest, input WebviewScreenshotInput) (*mcp.CallToolResult, WebviewScreenshotOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "webview_screenshot", "format", input.Format, "user", log.Username())

	if webviewInstance == nil {
		return nil, WebviewScreenshotOutput{}, fmt.Errorf("not connected; use webview_connect first")
	}

	format := input.Format
	if format == "" {
		format = "png"
	}

	data, err := webviewInstance.Screenshot()
	if err != nil {
		log.Error("mcp: webview screenshot failed", "err", err)
		return nil, WebviewScreenshotOutput{}, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return nil, WebviewScreenshotOutput{
		Success: true,
		Data:    base64.StdEncoding.EncodeToString(data),
		Format:  format,
	}, nil
}

// webviewWait handles the webview_wait tool call.
func (s *Service) webviewWait(ctx context.Context, req *mcp.CallToolRequest, input WebviewWaitInput) (*mcp.CallToolResult, WebviewWaitOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "webview_wait", "selector", input.Selector, "timeout", input.Timeout, "user", log.Username())

	if webviewInstance == nil {
		return nil, WebviewWaitOutput{}, fmt.Errorf("not connected; use webview_connect first")
	}

	if input.Selector == "" {
		return nil, WebviewWaitOutput{}, fmt.Errorf("selector is required")
	}

	if err := webviewInstance.WaitForSelector(input.Selector); err != nil {
		log.Error("mcp: webview wait failed", "selector", input.Selector, "err", err)
		return nil, WebviewWaitOutput{}, fmt.Errorf("failed to wait for selector: %w", err)
	}

	return nil, WebviewWaitOutput{
		Success: true,
		Message: fmt.Sprintf("Element found: %s", input.Selector),
	}, nil
}
