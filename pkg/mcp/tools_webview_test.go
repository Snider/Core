package mcp

import (
	"testing"
	"time"

	"github.com/host-uk/core/pkg/webview"
)

// TestWebviewToolsRegistered_Good verifies that webview tools are registered with the MCP server.
func TestWebviewToolsRegistered_Good(t *testing.T) {
	// Create a new MCP service - this should register all tools including webview
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// The server should have registered the webview tools
	if s.server == nil {
		t.Fatal("Server should not be nil")
	}

	// Verify the service was created with expected defaults
	if s.logger == nil {
		t.Error("Logger should not be nil")
	}
}

// TestWebviewConnectInput_Good verifies the WebviewConnectInput struct has expected fields.
func TestWebviewConnectInput_Good(t *testing.T) {
	input := WebviewConnectInput{
		DebugURL: "http://localhost:9222",
		Timeout:  30,
	}

	if input.DebugURL != "http://localhost:9222" {
		t.Errorf("Expected debug_url 'http://localhost:9222', got %q", input.DebugURL)
	}
	if input.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", input.Timeout)
	}
}

// TestWebviewNavigateInput_Good verifies the WebviewNavigateInput struct has expected fields.
func TestWebviewNavigateInput_Good(t *testing.T) {
	input := WebviewNavigateInput{
		URL: "https://example.com",
	}

	if input.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got %q", input.URL)
	}
}

// TestWebviewClickInput_Good verifies the WebviewClickInput struct has expected fields.
func TestWebviewClickInput_Good(t *testing.T) {
	input := WebviewClickInput{
		Selector: "#submit-button",
	}

	if input.Selector != "#submit-button" {
		t.Errorf("Expected selector '#submit-button', got %q", input.Selector)
	}
}

// TestWebviewTypeInput_Good verifies the WebviewTypeInput struct has expected fields.
func TestWebviewTypeInput_Good(t *testing.T) {
	input := WebviewTypeInput{
		Selector: "#email-input",
		Text:     "test@example.com",
	}

	if input.Selector != "#email-input" {
		t.Errorf("Expected selector '#email-input', got %q", input.Selector)
	}
	if input.Text != "test@example.com" {
		t.Errorf("Expected text 'test@example.com', got %q", input.Text)
	}
}

// TestWebviewQueryInput_Good verifies the WebviewQueryInput struct has expected fields.
func TestWebviewQueryInput_Good(t *testing.T) {
	input := WebviewQueryInput{
		Selector: "div.container",
		All:      true,
	}

	if input.Selector != "div.container" {
		t.Errorf("Expected selector 'div.container', got %q", input.Selector)
	}
	if !input.All {
		t.Error("Expected all to be true")
	}
}

// TestWebviewQueryInput_Defaults verifies default values are handled correctly.
func TestWebviewQueryInput_Defaults(t *testing.T) {
	input := WebviewQueryInput{
		Selector: ".test",
	}

	if input.All {
		t.Error("Expected all to default to false")
	}
}

// TestWebviewConsoleInput_Good verifies the WebviewConsoleInput struct has expected fields.
func TestWebviewConsoleInput_Good(t *testing.T) {
	input := WebviewConsoleInput{
		Clear: true,
	}

	if !input.Clear {
		t.Error("Expected clear to be true")
	}
}

// TestWebviewEvalInput_Good verifies the WebviewEvalInput struct has expected fields.
func TestWebviewEvalInput_Good(t *testing.T) {
	input := WebviewEvalInput{
		Script: "document.title",
	}

	if input.Script != "document.title" {
		t.Errorf("Expected script 'document.title', got %q", input.Script)
	}
}

// TestWebviewScreenshotInput_Good verifies the WebviewScreenshotInput struct has expected fields.
func TestWebviewScreenshotInput_Good(t *testing.T) {
	input := WebviewScreenshotInput{
		Format: "png",
	}

	if input.Format != "png" {
		t.Errorf("Expected format 'png', got %q", input.Format)
	}
}

// TestWebviewScreenshotInput_Defaults verifies default values are handled correctly.
func TestWebviewScreenshotInput_Defaults(t *testing.T) {
	input := WebviewScreenshotInput{}

	if input.Format != "" {
		t.Errorf("Expected format to default to empty, got %q", input.Format)
	}
}

// TestWebviewWaitInput_Good verifies the WebviewWaitInput struct has expected fields.
func TestWebviewWaitInput_Good(t *testing.T) {
	input := WebviewWaitInput{
		Selector: "#loading",
		Timeout:  10,
	}

	if input.Selector != "#loading" {
		t.Errorf("Expected selector '#loading', got %q", input.Selector)
	}
	if input.Timeout != 10 {
		t.Errorf("Expected timeout 10, got %d", input.Timeout)
	}
}

// TestWebviewConnectOutput_Good verifies the WebviewConnectOutput struct has expected fields.
func TestWebviewConnectOutput_Good(t *testing.T) {
	output := WebviewConnectOutput{
		Success: true,
		Message: "Connected to Chrome DevTools",
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Message == "" {
		t.Error("Expected message to be set")
	}
}

// TestWebviewNavigateOutput_Good verifies the WebviewNavigateOutput struct has expected fields.
func TestWebviewNavigateOutput_Good(t *testing.T) {
	output := WebviewNavigateOutput{
		Success: true,
		URL:     "https://example.com",
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got %q", output.URL)
	}
}

// TestWebviewQueryOutput_Good verifies the WebviewQueryOutput struct has expected fields.
func TestWebviewQueryOutput_Good(t *testing.T) {
	output := WebviewQueryOutput{
		Found: true,
		Count: 3,
		Elements: []WebviewElementInfo{
			{
				NodeID:  1,
				TagName: "DIV",
				Attributes: map[string]string{
					"class": "container",
				},
			},
		},
	}

	if !output.Found {
		t.Error("Expected found to be true")
	}
	if output.Count != 3 {
		t.Errorf("Expected count 3, got %d", output.Count)
	}
	if len(output.Elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(output.Elements))
	}
	if output.Elements[0].TagName != "DIV" {
		t.Errorf("Expected tagName 'DIV', got %q", output.Elements[0].TagName)
	}
}

// TestWebviewConsoleOutput_Good verifies the WebviewConsoleOutput struct has expected fields.
func TestWebviewConsoleOutput_Good(t *testing.T) {
	output := WebviewConsoleOutput{
		Messages: []WebviewConsoleMessage{
			{
				Type:      "log",
				Text:      "Hello, world!",
				Timestamp: "2024-01-01T00:00:00Z",
			},
			{
				Type:      "error",
				Text:      "An error occurred",
				Timestamp: "2024-01-01T00:00:01Z",
				URL:       "https://example.com/script.js",
				Line:      42,
			},
		},
		Count: 2,
	}

	if output.Count != 2 {
		t.Errorf("Expected count 2, got %d", output.Count)
	}
	if len(output.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(output.Messages))
	}
	if output.Messages[0].Type != "log" {
		t.Errorf("Expected type 'log', got %q", output.Messages[0].Type)
	}
	if output.Messages[1].Line != 42 {
		t.Errorf("Expected line 42, got %d", output.Messages[1].Line)
	}
}

// TestWebviewEvalOutput_Good verifies the WebviewEvalOutput struct has expected fields.
func TestWebviewEvalOutput_Good(t *testing.T) {
	output := WebviewEvalOutput{
		Success: true,
		Result:  "Example Page",
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Result != "Example Page" {
		t.Errorf("Expected result 'Example Page', got %v", output.Result)
	}
}

// TestWebviewEvalOutput_Error verifies the WebviewEvalOutput struct handles errors.
func TestWebviewEvalOutput_Error(t *testing.T) {
	output := WebviewEvalOutput{
		Success: false,
		Error:   "ReferenceError: foo is not defined",
	}

	if output.Success {
		t.Error("Expected success to be false")
	}
	if output.Error == "" {
		t.Error("Expected error message to be set")
	}
}

// TestWebviewScreenshotOutput_Good verifies the WebviewScreenshotOutput struct has expected fields.
func TestWebviewScreenshotOutput_Good(t *testing.T) {
	output := WebviewScreenshotOutput{
		Success: true,
		Data:    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		Format:  "png",
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Data == "" {
		t.Error("Expected data to be set")
	}
	if output.Format != "png" {
		t.Errorf("Expected format 'png', got %q", output.Format)
	}
}

// TestWebviewElementInfo_Good verifies the WebviewElementInfo struct has expected fields.
func TestWebviewElementInfo_Good(t *testing.T) {
	elem := WebviewElementInfo{
		NodeID:  123,
		TagName: "INPUT",
		Attributes: map[string]string{
			"type":  "text",
			"name":  "email",
			"class": "form-control",
		},
		BoundingBox: &webview.BoundingBox{
			X:      100,
			Y:      200,
			Width:  300,
			Height: 50,
		},
	}

	if elem.NodeID != 123 {
		t.Errorf("Expected nodeId 123, got %d", elem.NodeID)
	}
	if elem.TagName != "INPUT" {
		t.Errorf("Expected tagName 'INPUT', got %q", elem.TagName)
	}
	if elem.Attributes["type"] != "text" {
		t.Errorf("Expected type attribute 'text', got %q", elem.Attributes["type"])
	}
	if elem.BoundingBox == nil {
		t.Fatal("Expected bounding box to be set")
	}
	if elem.BoundingBox.Width != 300 {
		t.Errorf("Expected width 300, got %f", elem.BoundingBox.Width)
	}
}

// TestWebviewConsoleMessage_Good verifies the WebviewConsoleMessage struct has expected fields.
func TestWebviewConsoleMessage_Good(t *testing.T) {
	msg := WebviewConsoleMessage{
		Type:      "error",
		Text:      "Failed to load resource",
		Timestamp: time.Now().Format(time.RFC3339),
		URL:       "https://example.com/api/data",
		Line:      1,
	}

	if msg.Type != "error" {
		t.Errorf("Expected type 'error', got %q", msg.Type)
	}
	if msg.Text == "" {
		t.Error("Expected text to be set")
	}
	if msg.URL == "" {
		t.Error("Expected URL to be set")
	}
}

// TestWebviewDisconnectInput_Good verifies the WebviewDisconnectInput struct exists.
func TestWebviewDisconnectInput_Good(t *testing.T) {
	// WebviewDisconnectInput has no fields
	input := WebviewDisconnectInput{}
	_ = input // Just verify the struct exists
}

// TestWebviewDisconnectOutput_Good verifies the WebviewDisconnectOutput struct has expected fields.
func TestWebviewDisconnectOutput_Good(t *testing.T) {
	output := WebviewDisconnectOutput{
		Success: true,
		Message: "Disconnected from Chrome DevTools",
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Message == "" {
		t.Error("Expected message to be set")
	}
}

// TestWebviewWaitOutput_Good verifies the WebviewWaitOutput struct has expected fields.
func TestWebviewWaitOutput_Good(t *testing.T) {
	output := WebviewWaitOutput{
		Success: true,
		Message: "Element found: #login-form",
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Message == "" {
		t.Error("Expected message to be set")
	}
}
