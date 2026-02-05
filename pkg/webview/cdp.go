package webview

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// CDPClient handles communication with Chrome DevTools Protocol via WebSocket.
type CDPClient struct {
	mu       sync.RWMutex
	conn     *websocket.Conn
	debugURL string
	wsURL    string

	// Message tracking
	msgID   atomic.Int64
	pending map[int64]chan *cdpResponse
	pendMu  sync.Mutex

	// Event handlers
	handlers map[string][]func(map[string]any)
	handMu   sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// cdpMessage represents a CDP protocol message.
type cdpMessage struct {
	ID     int64          `json:"id,omitempty"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

// cdpResponse represents a CDP protocol response.
type cdpResponse struct {
	ID     int64          `json:"id"`
	Result map[string]any `json:"result,omitempty"`
	Error  *cdpError      `json:"error,omitempty"`
}

// cdpEvent represents a CDP event.
type cdpEvent struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

// cdpError represents a CDP error.
type cdpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// targetInfo represents Chrome DevTools target information.
type targetInfo struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// NewCDPClient creates a new CDP client connected to the given debug URL.
// The debug URL should be the Chrome DevTools HTTP endpoint (e.g., http://localhost:9222).
func NewCDPClient(debugURL string) (*CDPClient, error) {
	// Get available targets
	resp, err := http.Get(debugURL + "/json")
	if err != nil {
		return nil, fmt.Errorf("failed to get targets: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read targets: %w", err)
	}

	var targets []targetInfo
	if err := json.Unmarshal(body, &targets); err != nil {
		return nil, fmt.Errorf("failed to parse targets: %w", err)
	}

	// Find a page target
	var wsURL string
	for _, t := range targets {
		if t.Type == "page" && t.WebSocketDebuggerURL != "" {
			wsURL = t.WebSocketDebuggerURL
			break
		}
	}

	if wsURL == "" {
		// Try to create a new target
		resp, err := http.Get(debugURL + "/json/new")
		if err != nil {
			return nil, fmt.Errorf("no page targets found and failed to create new: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read new target: %w", err)
		}

		var newTarget targetInfo
		if err := json.Unmarshal(body, &newTarget); err != nil {
			return nil, fmt.Errorf("failed to parse new target: %w", err)
		}

		wsURL = newTarget.WebSocketDebuggerURL
	}

	if wsURL == "" {
		return nil, fmt.Errorf("no WebSocket URL available")
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &CDPClient{
		conn:     conn,
		debugURL: debugURL,
		wsURL:    wsURL,
		pending:  make(map[int64]chan *cdpResponse),
		handlers: make(map[string][]func(map[string]any)),
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
	}

	// Start message reader
	go client.readLoop()

	return client, nil
}

// Close closes the CDP connection.
func (c *CDPClient) Close() error {
	c.cancel()
	<-c.done // Wait for read loop to finish
	return c.conn.Close()
}

// Call sends a CDP method call and waits for the response.
func (c *CDPClient) Call(ctx context.Context, method string, params map[string]any) (map[string]any, error) {
	id := c.msgID.Add(1)

	msg := cdpMessage{
		ID:     id,
		Method: method,
		Params: params,
	}

	// Register response channel
	respCh := make(chan *cdpResponse, 1)
	c.pendMu.Lock()
	c.pending[id] = respCh
	c.pendMu.Unlock()

	defer func() {
		c.pendMu.Lock()
		delete(c.pending, id)
		c.pendMu.Unlock()
	}()

	// Send message
	c.mu.Lock()
	err := c.conn.WriteJSON(msg)
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Wait for response
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, fmt.Errorf("CDP error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

// OnEvent registers a handler for CDP events.
func (c *CDPClient) OnEvent(method string, handler func(map[string]any)) {
	c.handMu.Lock()
	defer c.handMu.Unlock()
	c.handlers[method] = append(c.handlers[method], handler)
}

// readLoop reads messages from the WebSocket connection.
func (c *CDPClient) readLoop() {
	defer close(c.done)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			// Check if context was cancelled
			select {
			case <-c.ctx.Done():
				return
			default:
				// Log error but continue (could be temporary)
				continue
			}
		}

		// Try to parse as response
		var resp cdpResponse
		if err := json.Unmarshal(data, &resp); err == nil && resp.ID > 0 {
			c.pendMu.Lock()
			if ch, ok := c.pending[resp.ID]; ok {
				respCopy := resp
				ch <- &respCopy
			}
			c.pendMu.Unlock()
			continue
		}

		// Try to parse as event
		var event cdpEvent
		if err := json.Unmarshal(data, &event); err == nil && event.Method != "" {
			c.dispatchEvent(event.Method, event.Params)
		}
	}
}

// dispatchEvent dispatches an event to registered handlers.
func (c *CDPClient) dispatchEvent(method string, params map[string]any) {
	c.handMu.RLock()
	handlers := c.handlers[method]
	c.handMu.RUnlock()

	for _, handler := range handlers {
		// Call handler in goroutine to avoid blocking
		go handler(params)
	}
}

// Send sends a fire-and-forget CDP message (no response expected).
func (c *CDPClient) Send(method string, params map[string]any) error {
	msg := cdpMessage{
		Method: method,
		Params: params,
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(msg)
}

// DebugURL returns the debug HTTP URL.
func (c *CDPClient) DebugURL() string {
	return c.debugURL
}

// WebSocketURL returns the WebSocket URL being used.
func (c *CDPClient) WebSocketURL() string {
	return c.wsURL
}

// NewTab creates a new browser tab and returns a new CDPClient connected to it.
func (c *CDPClient) NewTab(url string) (*CDPClient, error) {
	endpoint := c.debugURL + "/json/new"
	if url != "" {
		endpoint += "?" + url
	}

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create new tab: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var target targetInfo
	if err := json.Unmarshal(body, &target); err != nil {
		return nil, fmt.Errorf("failed to parse target: %w", err)
	}

	if target.WebSocketDebuggerURL == "" {
		return nil, fmt.Errorf("no WebSocket URL for new tab")
	}

	// Connect to new tab
	conn, _, err := websocket.DefaultDialer.Dial(target.WebSocketDebuggerURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to new tab: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &CDPClient{
		conn:     conn,
		debugURL: c.debugURL,
		wsURL:    target.WebSocketDebuggerURL,
		pending:  make(map[int64]chan *cdpResponse),
		handlers: make(map[string][]func(map[string]any)),
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
	}

	go client.readLoop()

	return client, nil
}

// CloseTab closes the current tab (target).
func (c *CDPClient) CloseTab() error {
	// Extract target ID from WebSocket URL
	// Format: ws://host:port/devtools/page/TARGET_ID
	// We'll use the Browser.close target API

	ctx := context.Background()
	_, err := c.Call(ctx, "Browser.close", nil)
	return err
}

// ListTargets returns all available targets.
func ListTargets(debugURL string) ([]targetInfo, error) {
	resp, err := http.Get(debugURL + "/json")
	if err != nil {
		return nil, fmt.Errorf("failed to get targets: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read targets: %w", err)
	}

	var targets []targetInfo
	if err := json.Unmarshal(body, &targets); err != nil {
		return nil, fmt.Errorf("failed to parse targets: %w", err)
	}

	return targets, nil
}

// GetVersion returns Chrome version information.
func GetVersion(debugURL string) (map[string]string, error) {
	resp, err := http.Get(debugURL + "/json/version")
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	var version map[string]string
	if err := json.Unmarshal(body, &version); err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	return version, nil
}
