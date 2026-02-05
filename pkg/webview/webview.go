// Package webview provides browser automation via Chrome DevTools Protocol (CDP).
//
// The package allows controlling Chrome/Chromium browsers for automated testing,
// web scraping, and GUI automation. Start Chrome with --remote-debugging-port=9222
// to enable the DevTools protocol.
//
// Example usage:
//
//	wv, err := webview.New(webview.WithDebugURL("http://localhost:9222"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer wv.Close()
//
//	if err := wv.Navigate("https://example.com"); err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := wv.Click("#submit-button"); err != nil {
//	    log.Fatal(err)
//	}
package webview

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// Webview represents a connection to a Chrome DevTools Protocol endpoint.
type Webview struct {
	mu           sync.RWMutex
	client       *CDPClient
	ctx          context.Context
	cancel       context.CancelFunc
	timeout      time.Duration
	consoleLogs  []ConsoleMessage
	consoleLimit int
}

// ConsoleMessage represents a captured console log message.
type ConsoleMessage struct {
	Type      string    `json:"type"`      // log, warn, error, info, debug
	Text      string    `json:"text"`      // Message text
	Timestamp time.Time `json:"timestamp"` // When the message was logged
	URL       string    `json:"url"`       // Source URL
	Line      int       `json:"line"`      // Source line number
	Column    int       `json:"column"`    // Source column number
}

// ElementInfo represents information about a DOM element.
type ElementInfo struct {
	NodeID      int               `json:"nodeId"`
	TagName     string            `json:"tagName"`
	Attributes  map[string]string `json:"attributes"`
	InnerHTML   string            `json:"innerHTML,omitempty"`
	InnerText   string            `json:"innerText,omitempty"`
	BoundingBox *BoundingBox      `json:"boundingBox,omitempty"`
}

// BoundingBox represents the bounding rectangle of an element.
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Option configures a Webview instance.
type Option func(*Webview) error

// WithDebugURL sets the Chrome DevTools debugging URL.
// Example: http://localhost:9222
func WithDebugURL(url string) Option {
	return func(wv *Webview) error {
		client, err := NewCDPClient(url)
		if err != nil {
			return fmt.Errorf("failed to connect to Chrome DevTools: %w", err)
		}
		wv.client = client
		return nil
	}
}

// WithTimeout sets the default timeout for operations.
func WithTimeout(d time.Duration) Option {
	return func(wv *Webview) error {
		wv.timeout = d
		return nil
	}
}

// WithConsoleLimit sets the maximum number of console messages to retain.
// Default is 1000.
func WithConsoleLimit(limit int) Option {
	return func(wv *Webview) error {
		wv.consoleLimit = limit
		return nil
	}
}

// New creates a new Webview instance with the given options.
func New(opts ...Option) (*Webview, error) {
	ctx, cancel := context.WithCancel(context.Background())

	wv := &Webview{
		ctx:          ctx,
		cancel:       cancel,
		timeout:      30 * time.Second,
		consoleLogs:  make([]ConsoleMessage, 0, 100),
		consoleLimit: 1000,
	}

	for _, opt := range opts {
		if err := opt(wv); err != nil {
			cancel()
			return nil, err
		}
	}

	if wv.client == nil {
		cancel()
		return nil, fmt.Errorf("no debug URL provided; use WithDebugURL option")
	}

	// Enable console capture
	if err := wv.enableConsole(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to enable console capture: %w", err)
	}

	return wv, nil
}

// Close closes the Webview connection.
func (wv *Webview) Close() error {
	wv.cancel()
	if wv.client != nil {
		return wv.client.Close()
	}
	return nil
}

// Navigate navigates to the specified URL.
func (wv *Webview) Navigate(url string) error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	_, err := wv.client.Call(ctx, "Page.navigate", map[string]any{
		"url": url,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	// Wait for page load
	return wv.waitForLoad(ctx)
}

// Click clicks on an element matching the selector.
func (wv *Webview) Click(selector string) error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	return wv.click(ctx, selector)
}

// Type types text into an element matching the selector.
func (wv *Webview) Type(selector, text string) error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	return wv.typeText(ctx, selector, text)
}

// QuerySelector finds an element by CSS selector and returns its information.
func (wv *Webview) QuerySelector(selector string) (*ElementInfo, error) {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	return wv.querySelector(ctx, selector)
}

// QuerySelectorAll finds all elements matching the selector.
func (wv *Webview) QuerySelectorAll(selector string) ([]*ElementInfo, error) {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	return wv.querySelectorAll(ctx, selector)
}

// GetConsole returns captured console messages.
func (wv *Webview) GetConsole() []ConsoleMessage {
	wv.mu.RLock()
	defer wv.mu.RUnlock()

	result := make([]ConsoleMessage, len(wv.consoleLogs))
	copy(result, wv.consoleLogs)
	return result
}

// ClearConsole clears captured console messages.
func (wv *Webview) ClearConsole() {
	wv.mu.Lock()
	defer wv.mu.Unlock()
	wv.consoleLogs = wv.consoleLogs[:0]
}

// Screenshot captures a screenshot and returns it as PNG bytes.
func (wv *Webview) Screenshot() ([]byte, error) {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	result, err := wv.client.Call(ctx, "Page.captureScreenshot", map[string]any{
		"format": "png",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	dataStr, ok := result["data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid screenshot data")
	}

	data, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot: %w", err)
	}

	return data, nil
}

// Evaluate executes JavaScript and returns the result.
// Note: This intentionally executes arbitrary JavaScript in the browser context
// for browser automation purposes. The script runs in the sandboxed browser environment.
func (wv *Webview) Evaluate(script string) (any, error) {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	return wv.evaluate(ctx, script)
}

// WaitForSelector waits for an element matching the selector to appear.
func (wv *Webview) WaitForSelector(selector string) error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	return wv.waitForSelector(ctx, selector)
}

// GetURL returns the current page URL.
func (wv *Webview) GetURL() (string, error) {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	result, err := wv.evaluate(ctx, "window.location.href")
	if err != nil {
		return "", err
	}

	url, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("invalid URL result")
	}

	return url, nil
}

// GetTitle returns the current page title.
func (wv *Webview) GetTitle() (string, error) {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	result, err := wv.evaluate(ctx, "document.title")
	if err != nil {
		return "", err
	}

	title, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("invalid title result")
	}

	return title, nil
}

// GetHTML returns the outer HTML of an element or the whole document.
func (wv *Webview) GetHTML(selector string) (string, error) {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	var script string
	if selector == "" {
		script = "document.documentElement.outerHTML"
	} else {
		script = fmt.Sprintf("document.querySelector(%q)?.outerHTML || ''", selector)
	}

	result, err := wv.evaluate(ctx, script)
	if err != nil {
		return "", err
	}

	html, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("invalid HTML result")
	}

	return html, nil
}

// SetViewport sets the viewport size.
func (wv *Webview) SetViewport(width, height int) error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	_, err := wv.client.Call(ctx, "Emulation.setDeviceMetricsOverride", map[string]any{
		"width":             width,
		"height":            height,
		"deviceScaleFactor": 1,
		"mobile":            false,
	})
	return err
}

// SetUserAgent sets the user agent string.
func (wv *Webview) SetUserAgent(userAgent string) error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	_, err := wv.client.Call(ctx, "Emulation.setUserAgentOverride", map[string]any{
		"userAgent": userAgent,
	})
	return err
}

// Reload reloads the current page.
func (wv *Webview) Reload() error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	_, err := wv.client.Call(ctx, "Page.reload", nil)
	if err != nil {
		return fmt.Errorf("failed to reload: %w", err)
	}

	return wv.waitForLoad(ctx)
}

// GoBack navigates back in history.
func (wv *Webview) GoBack() error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	_, err := wv.client.Call(ctx, "Page.goBackOrForward", map[string]any{
		"delta": -1,
	})
	return err
}

// GoForward navigates forward in history.
func (wv *Webview) GoForward() error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	_, err := wv.client.Call(ctx, "Page.goBackOrForward", map[string]any{
		"delta": 1,
	})
	return err
}

// addConsoleMessage adds a console message to the log.
func (wv *Webview) addConsoleMessage(msg ConsoleMessage) {
	wv.mu.Lock()
	defer wv.mu.Unlock()

	if len(wv.consoleLogs) >= wv.consoleLimit {
		// Remove oldest messages
		wv.consoleLogs = wv.consoleLogs[len(wv.consoleLogs)-wv.consoleLimit+100:]
	}
	wv.consoleLogs = append(wv.consoleLogs, msg)
}

// enableConsole enables console message capture.
func (wv *Webview) enableConsole() error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	// Enable Runtime domain for console events
	_, err := wv.client.Call(ctx, "Runtime.enable", nil)
	if err != nil {
		return err
	}

	// Enable Page domain for navigation events
	_, err = wv.client.Call(ctx, "Page.enable", nil)
	if err != nil {
		return err
	}

	// Enable DOM domain
	_, err = wv.client.Call(ctx, "DOM.enable", nil)
	if err != nil {
		return err
	}

	// Subscribe to console events
	wv.client.OnEvent("Runtime.consoleAPICalled", func(params map[string]any) {
		wv.handleConsoleEvent(params)
	})

	return nil
}

// handleConsoleEvent processes console API events.
func (wv *Webview) handleConsoleEvent(params map[string]any) {
	msgType, _ := params["type"].(string)

	// Extract args
	args, _ := params["args"].([]any)
	var text string
	for i, arg := range args {
		if argMap, ok := arg.(map[string]any); ok {
			if val, ok := argMap["value"]; ok {
				if i > 0 {
					text += " "
				}
				text += fmt.Sprint(val)
			}
		}
	}

	// Extract stack trace info
	stackTrace, _ := params["stackTrace"].(map[string]any)
	var url string
	var line, column int
	if callFrames, ok := stackTrace["callFrames"].([]any); ok && len(callFrames) > 0 {
		if frame, ok := callFrames[0].(map[string]any); ok {
			url, _ = frame["url"].(string)
			lineFloat, _ := frame["lineNumber"].(float64)
			colFloat, _ := frame["columnNumber"].(float64)
			line = int(lineFloat)
			column = int(colFloat)
		}
	}

	wv.addConsoleMessage(ConsoleMessage{
		Type:      msgType,
		Text:      text,
		Timestamp: time.Now(),
		URL:       url,
		Line:      line,
		Column:    column,
	})
}

// waitForLoad waits for the page to finish loading.
func (wv *Webview) waitForLoad(ctx context.Context) error {
	// Use Page.loadEventFired event or poll document.readyState
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			result, err := wv.evaluate(ctx, "document.readyState")
			if err != nil {
				continue
			}
			if state, ok := result.(string); ok && state == "complete" {
				return nil
			}
		}
	}
}

// waitForSelector waits for an element to appear.
func (wv *Webview) waitForSelector(ctx context.Context, selector string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	script := fmt.Sprintf("!!document.querySelector(%q)", selector)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			result, err := wv.evaluate(ctx, script)
			if err != nil {
				continue
			}
			if found, ok := result.(bool); ok && found {
				return nil
			}
		}
	}
}

// evaluate evaluates JavaScript in the page context via CDP Runtime.evaluate.
// This is the core method for executing JavaScript in the browser.
func (wv *Webview) evaluate(ctx context.Context, script string) (any, error) {
	result, err := wv.client.Call(ctx, "Runtime.evaluate", map[string]any{
		"expression":    script,
		"returnByValue": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate script: %w", err)
	}

	// Check for exception
	if exceptionDetails, ok := result["exceptionDetails"].(map[string]any); ok {
		if exception, ok := exceptionDetails["exception"].(map[string]any); ok {
			if description, ok := exception["description"].(string); ok {
				return nil, fmt.Errorf("JavaScript error: %s", description)
			}
		}
		return nil, fmt.Errorf("JavaScript error")
	}

	// Extract result value
	if resultObj, ok := result["result"].(map[string]any); ok {
		return resultObj["value"], nil
	}

	return nil, nil
}

// querySelector finds an element by selector.
func (wv *Webview) querySelector(ctx context.Context, selector string) (*ElementInfo, error) {
	// Get document root
	docResult, err := wv.client.Call(ctx, "DOM.getDocument", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	root, ok := docResult["root"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid document root")
	}

	rootID, ok := root["nodeId"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid root node ID")
	}

	// Query selector
	queryResult, err := wv.client.Call(ctx, "DOM.querySelector", map[string]any{
		"nodeId":   int(rootID),
		"selector": selector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query selector: %w", err)
	}

	nodeID, ok := queryResult["nodeId"].(float64)
	if !ok || nodeID == 0 {
		return nil, fmt.Errorf("element not found: %s", selector)
	}

	return wv.getElementInfo(ctx, int(nodeID))
}

// querySelectorAll finds all elements matching the selector.
func (wv *Webview) querySelectorAll(ctx context.Context, selector string) ([]*ElementInfo, error) {
	// Get document root
	docResult, err := wv.client.Call(ctx, "DOM.getDocument", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	root, ok := docResult["root"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid document root")
	}

	rootID, ok := root["nodeId"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid root node ID")
	}

	// Query selector all
	queryResult, err := wv.client.Call(ctx, "DOM.querySelectorAll", map[string]any{
		"nodeId":   int(rootID),
		"selector": selector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query selector all: %w", err)
	}

	nodeIDs, ok := queryResult["nodeIds"].([]any)
	if !ok {
		return nil, fmt.Errorf("invalid node IDs")
	}

	elements := make([]*ElementInfo, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		if nodeID, ok := id.(float64); ok {
			if elem, err := wv.getElementInfo(ctx, int(nodeID)); err == nil {
				elements = append(elements, elem)
			}
		}
	}

	return elements, nil
}

// getElementInfo retrieves information about a DOM node.
func (wv *Webview) getElementInfo(ctx context.Context, nodeID int) (*ElementInfo, error) {
	// Describe node to get attributes
	descResult, err := wv.client.Call(ctx, "DOM.describeNode", map[string]any{
		"nodeId": nodeID,
	})
	if err != nil {
		return nil, err
	}

	node, ok := descResult["node"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid node description")
	}

	tagName, _ := node["nodeName"].(string)

	// Parse attributes
	attrs := make(map[string]string)
	if attrList, ok := node["attributes"].([]any); ok {
		for i := 0; i < len(attrList)-1; i += 2 {
			key, _ := attrList[i].(string)
			val, _ := attrList[i+1].(string)
			attrs[key] = val
		}
	}

	// Get bounding box
	var box *BoundingBox
	if boxResult, err := wv.client.Call(ctx, "DOM.getBoxModel", map[string]any{
		"nodeId": nodeID,
	}); err == nil {
		if model, ok := boxResult["model"].(map[string]any); ok {
			if content, ok := model["content"].([]any); ok && len(content) >= 8 {
				x, _ := content[0].(float64)
				y, _ := content[1].(float64)
				x2, _ := content[2].(float64)
				y2, _ := content[5].(float64)
				box = &BoundingBox{
					X:      x,
					Y:      y,
					Width:  x2 - x,
					Height: y2 - y,
				}
			}
		}
	}

	return &ElementInfo{
		NodeID:      nodeID,
		TagName:     tagName,
		Attributes:  attrs,
		BoundingBox: box,
	}, nil
}

// click performs a click on an element.
func (wv *Webview) click(ctx context.Context, selector string) error {
	// Find element and get its center coordinates
	elem, err := wv.querySelector(ctx, selector)
	if err != nil {
		return err
	}

	if elem.BoundingBox == nil {
		// Fallback to JavaScript click
		script := fmt.Sprintf("document.querySelector(%q)?.click()", selector)
		_, err := wv.evaluate(ctx, script)
		return err
	}

	// Calculate center point
	x := elem.BoundingBox.X + elem.BoundingBox.Width/2
	y := elem.BoundingBox.Y + elem.BoundingBox.Height/2

	// Dispatch mouse events
	for _, eventType := range []string{"mousePressed", "mouseReleased"} {
		_, err := wv.client.Call(ctx, "Input.dispatchMouseEvent", map[string]any{
			"type":       eventType,
			"x":          x,
			"y":          y,
			"button":     "left",
			"clickCount": 1,
		})
		if err != nil {
			return fmt.Errorf("failed to dispatch %s: %w", eventType, err)
		}
	}

	return nil
}

// typeText types text into an element.
func (wv *Webview) typeText(ctx context.Context, selector, text string) error {
	// Focus the element first
	script := fmt.Sprintf("document.querySelector(%q)?.focus()", selector)
	_, err := wv.evaluate(ctx, script)
	if err != nil {
		return fmt.Errorf("failed to focus element: %w", err)
	}

	// Type each character
	for _, char := range text {
		_, err := wv.client.Call(ctx, "Input.dispatchKeyEvent", map[string]any{
			"type": "keyDown",
			"text": string(char),
		})
		if err != nil {
			return fmt.Errorf("failed to dispatch keyDown: %w", err)
		}

		_, err = wv.client.Call(ctx, "Input.dispatchKeyEvent", map[string]any{
			"type": "keyUp",
		})
		if err != nil {
			return fmt.Errorf("failed to dispatch keyUp: %w", err)
		}
	}

	return nil
}
