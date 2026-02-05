package webview

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConsoleWatcher provides advanced console message watching capabilities.
type ConsoleWatcher struct {
	mu       sync.RWMutex
	wv       *Webview
	messages []ConsoleMessage
	filters  []ConsoleFilter
	limit    int
	handlers []ConsoleHandler
}

// ConsoleFilter filters console messages.
type ConsoleFilter struct {
	Type    string // Filter by type (log, warn, error, info, debug), empty for all
	Pattern string // Filter by text pattern (substring match)
}

// ConsoleHandler is called when a matching console message is received.
type ConsoleHandler func(msg ConsoleMessage)

// NewConsoleWatcher creates a new console watcher for the webview.
func NewConsoleWatcher(wv *Webview) *ConsoleWatcher {
	cw := &ConsoleWatcher{
		wv:       wv,
		messages: make([]ConsoleMessage, 0, 100),
		filters:  make([]ConsoleFilter, 0),
		limit:    1000,
		handlers: make([]ConsoleHandler, 0),
	}

	// Subscribe to console events from the webview's client
	wv.client.OnEvent("Runtime.consoleAPICalled", func(params map[string]any) {
		cw.handleConsoleEvent(params)
	})

	return cw
}

// AddFilter adds a filter to the watcher.
func (cw *ConsoleWatcher) AddFilter(filter ConsoleFilter) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.filters = append(cw.filters, filter)
}

// ClearFilters removes all filters.
func (cw *ConsoleWatcher) ClearFilters() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.filters = cw.filters[:0]
}

// AddHandler adds a handler for console messages.
func (cw *ConsoleWatcher) AddHandler(handler ConsoleHandler) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.handlers = append(cw.handlers, handler)
}

// SetLimit sets the maximum number of messages to retain.
func (cw *ConsoleWatcher) SetLimit(limit int) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.limit = limit
}

// Messages returns all captured messages.
func (cw *ConsoleWatcher) Messages() []ConsoleMessage {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	result := make([]ConsoleMessage, len(cw.messages))
	copy(result, cw.messages)
	return result
}

// FilteredMessages returns messages matching the current filters.
func (cw *ConsoleWatcher) FilteredMessages() []ConsoleMessage {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	if len(cw.filters) == 0 {
		result := make([]ConsoleMessage, len(cw.messages))
		copy(result, cw.messages)
		return result
	}

	result := make([]ConsoleMessage, 0)
	for _, msg := range cw.messages {
		if cw.matchesFilter(msg) {
			result = append(result, msg)
		}
	}
	return result
}

// Errors returns all error messages.
func (cw *ConsoleWatcher) Errors() []ConsoleMessage {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	result := make([]ConsoleMessage, 0)
	for _, msg := range cw.messages {
		if msg.Type == "error" {
			result = append(result, msg)
		}
	}
	return result
}

// Warnings returns all warning messages.
func (cw *ConsoleWatcher) Warnings() []ConsoleMessage {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	result := make([]ConsoleMessage, 0)
	for _, msg := range cw.messages {
		if msg.Type == "warning" {
			result = append(result, msg)
		}
	}
	return result
}

// Clear clears all captured messages.
func (cw *ConsoleWatcher) Clear() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.messages = cw.messages[:0]
}

// WaitForMessage waits for a message matching the filter.
func (cw *ConsoleWatcher) WaitForMessage(ctx context.Context, filter ConsoleFilter) (*ConsoleMessage, error) {
	// First check existing messages
	cw.mu.RLock()
	for _, msg := range cw.messages {
		if cw.matchesSingleFilter(msg, filter) {
			cw.mu.RUnlock()
			return &msg, nil
		}
	}
	cw.mu.RUnlock()

	// Set up a channel for new messages
	msgCh := make(chan ConsoleMessage, 1)
	handler := func(msg ConsoleMessage) {
		if cw.matchesSingleFilter(msg, filter) {
			select {
			case msgCh <- msg:
			default:
			}
		}
	}

	cw.AddHandler(handler)
	defer func() {
		cw.mu.Lock()
		// Remove handler (simple implementation - in production you'd want a handle-based removal)
		cw.handlers = cw.handlers[:len(cw.handlers)-1]
		cw.mu.Unlock()
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-msgCh:
		return &msg, nil
	}
}

// WaitForError waits for an error message.
func (cw *ConsoleWatcher) WaitForError(ctx context.Context) (*ConsoleMessage, error) {
	return cw.WaitForMessage(ctx, ConsoleFilter{Type: "error"})
}

// HasErrors returns true if there are any error messages.
func (cw *ConsoleWatcher) HasErrors() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	for _, msg := range cw.messages {
		if msg.Type == "error" {
			return true
		}
	}
	return false
}

// Count returns the number of captured messages.
func (cw *ConsoleWatcher) Count() int {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return len(cw.messages)
}

// ErrorCount returns the number of error messages.
func (cw *ConsoleWatcher) ErrorCount() int {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	count := 0
	for _, msg := range cw.messages {
		if msg.Type == "error" {
			count++
		}
	}
	return count
}

// handleConsoleEvent processes incoming console events.
func (cw *ConsoleWatcher) handleConsoleEvent(params map[string]any) {
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

	msg := ConsoleMessage{
		Type:      msgType,
		Text:      text,
		Timestamp: time.Now(),
		URL:       url,
		Line:      line,
		Column:    column,
	}

	cw.addMessage(msg)
}

// addMessage adds a message to the store and notifies handlers.
func (cw *ConsoleWatcher) addMessage(msg ConsoleMessage) {
	cw.mu.Lock()

	// Enforce limit
	if len(cw.messages) >= cw.limit {
		cw.messages = cw.messages[len(cw.messages)-cw.limit+100:]
	}
	cw.messages = append(cw.messages, msg)

	// Copy handlers to call outside lock
	handlers := make([]ConsoleHandler, len(cw.handlers))
	copy(handlers, cw.handlers)
	cw.mu.Unlock()

	// Call handlers
	for _, handler := range handlers {
		handler(msg)
	}
}

// matchesFilter checks if a message matches any filter.
func (cw *ConsoleWatcher) matchesFilter(msg ConsoleMessage) bool {
	if len(cw.filters) == 0 {
		return true
	}
	for _, filter := range cw.filters {
		if cw.matchesSingleFilter(msg, filter) {
			return true
		}
	}
	return false
}

// matchesSingleFilter checks if a message matches a specific filter.
func (cw *ConsoleWatcher) matchesSingleFilter(msg ConsoleMessage, filter ConsoleFilter) bool {
	if filter.Type != "" && msg.Type != filter.Type {
		return false
	}
	if filter.Pattern != "" {
		// Simple substring match
		if !containsString(msg.Text, filter.Pattern) {
			return false
		}
	}
	return true
}

// containsString checks if s contains substr (case-sensitive).
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findString(s, substr) >= 0)
}

// findString finds substr in s, returns -1 if not found.
func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ExceptionInfo represents information about a JavaScript exception.
type ExceptionInfo struct {
	Text         string    `json:"text"`
	LineNumber   int       `json:"lineNumber"`
	ColumnNumber int       `json:"columnNumber"`
	URL          string    `json:"url"`
	StackTrace   string    `json:"stackTrace"`
	Timestamp    time.Time `json:"timestamp"`
}

// ExceptionWatcher watches for JavaScript exceptions.
type ExceptionWatcher struct {
	mu         sync.RWMutex
	wv         *Webview
	exceptions []ExceptionInfo
	handlers   []func(ExceptionInfo)
}

// NewExceptionWatcher creates a new exception watcher.
func NewExceptionWatcher(wv *Webview) *ExceptionWatcher {
	ew := &ExceptionWatcher{
		wv:         wv,
		exceptions: make([]ExceptionInfo, 0),
		handlers:   make([]func(ExceptionInfo), 0),
	}

	// Subscribe to exception events
	wv.client.OnEvent("Runtime.exceptionThrown", func(params map[string]any) {
		ew.handleException(params)
	})

	return ew
}

// Exceptions returns all captured exceptions.
func (ew *ExceptionWatcher) Exceptions() []ExceptionInfo {
	ew.mu.RLock()
	defer ew.mu.RUnlock()

	result := make([]ExceptionInfo, len(ew.exceptions))
	copy(result, ew.exceptions)
	return result
}

// Clear clears all captured exceptions.
func (ew *ExceptionWatcher) Clear() {
	ew.mu.Lock()
	defer ew.mu.Unlock()
	ew.exceptions = ew.exceptions[:0]
}

// HasExceptions returns true if there are any exceptions.
func (ew *ExceptionWatcher) HasExceptions() bool {
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	return len(ew.exceptions) > 0
}

// Count returns the number of exceptions.
func (ew *ExceptionWatcher) Count() int {
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	return len(ew.exceptions)
}

// AddHandler adds a handler for exceptions.
func (ew *ExceptionWatcher) AddHandler(handler func(ExceptionInfo)) {
	ew.mu.Lock()
	defer ew.mu.Unlock()
	ew.handlers = append(ew.handlers, handler)
}

// WaitForException waits for an exception to be thrown.
func (ew *ExceptionWatcher) WaitForException(ctx context.Context) (*ExceptionInfo, error) {
	// Check existing exceptions first
	ew.mu.RLock()
	if len(ew.exceptions) > 0 {
		exc := ew.exceptions[len(ew.exceptions)-1]
		ew.mu.RUnlock()
		return &exc, nil
	}
	ew.mu.RUnlock()

	// Set up a channel for new exceptions
	excCh := make(chan ExceptionInfo, 1)
	handler := func(exc ExceptionInfo) {
		select {
		case excCh <- exc:
		default:
		}
	}

	ew.AddHandler(handler)
	defer func() {
		ew.mu.Lock()
		ew.handlers = ew.handlers[:len(ew.handlers)-1]
		ew.mu.Unlock()
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case exc := <-excCh:
		return &exc, nil
	}
}

// handleException processes exception events.
func (ew *ExceptionWatcher) handleException(params map[string]any) {
	exceptionDetails, ok := params["exceptionDetails"].(map[string]any)
	if !ok {
		return
	}

	text, _ := exceptionDetails["text"].(string)
	lineNum, _ := exceptionDetails["lineNumber"].(float64)
	colNum, _ := exceptionDetails["columnNumber"].(float64)
	url, _ := exceptionDetails["url"].(string)

	// Extract stack trace
	var stackTrace string
	if st, ok := exceptionDetails["stackTrace"].(map[string]any); ok {
		if frames, ok := st["callFrames"].([]any); ok {
			for _, f := range frames {
				if frame, ok := f.(map[string]any); ok {
					funcName, _ := frame["functionName"].(string)
					frameURL, _ := frame["url"].(string)
					frameLine, _ := frame["lineNumber"].(float64)
					frameCol, _ := frame["columnNumber"].(float64)
					stackTrace += fmt.Sprintf("  at %s (%s:%d:%d)\n", funcName, frameURL, int(frameLine), int(frameCol))
				}
			}
		}
	}

	// Try to get exception value description
	if exc, ok := exceptionDetails["exception"].(map[string]any); ok {
		if desc, ok := exc["description"].(string); ok && desc != "" {
			text = desc
		}
	}

	info := ExceptionInfo{
		Text:         text,
		LineNumber:   int(lineNum),
		ColumnNumber: int(colNum),
		URL:          url,
		StackTrace:   stackTrace,
		Timestamp:    time.Now(),
	}

	ew.mu.Lock()
	ew.exceptions = append(ew.exceptions, info)
	handlers := make([]func(ExceptionInfo), len(ew.handlers))
	copy(handlers, ew.handlers)
	ew.mu.Unlock()

	// Call handlers
	for _, handler := range handlers {
		handler(info)
	}
}

// FormatConsoleOutput formats console messages for display.
func FormatConsoleOutput(messages []ConsoleMessage) string {
	var output string
	for _, msg := range messages {
		prefix := ""
		switch msg.Type {
		case "error":
			prefix = "[ERROR]"
		case "warning":
			prefix = "[WARN]"
		case "info":
			prefix = "[INFO]"
		case "debug":
			prefix = "[DEBUG]"
		default:
			prefix = "[LOG]"
		}
		timestamp := msg.Timestamp.Format("15:04:05.000")
		output += fmt.Sprintf("%s %s %s\n", timestamp, prefix, msg.Text)
	}
	return output
}
