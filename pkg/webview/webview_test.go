package webview

import (
	"testing"
	"time"
)

// TestConsoleMessage_Good verifies the ConsoleMessage struct has expected fields.
func TestConsoleMessage_Good(t *testing.T) {
	msg := ConsoleMessage{
		Type:      "error",
		Text:      "Test error message",
		Timestamp: time.Now(),
		URL:       "https://example.com/script.js",
		Line:      42,
		Column:    10,
	}

	if msg.Type != "error" {
		t.Errorf("Expected type 'error', got %q", msg.Type)
	}
	if msg.Text != "Test error message" {
		t.Errorf("Expected text 'Test error message', got %q", msg.Text)
	}
	if msg.Line != 42 {
		t.Errorf("Expected line 42, got %d", msg.Line)
	}
}

// TestElementInfo_Good verifies the ElementInfo struct has expected fields.
func TestElementInfo_Good(t *testing.T) {
	elem := ElementInfo{
		NodeID:  123,
		TagName: "DIV",
		Attributes: map[string]string{
			"id":    "container",
			"class": "main-content",
		},
		InnerHTML: "<span>Hello</span>",
		InnerText: "Hello",
		BoundingBox: &BoundingBox{
			X:      100,
			Y:      200,
			Width:  300,
			Height: 400,
		},
	}

	if elem.NodeID != 123 {
		t.Errorf("Expected nodeId 123, got %d", elem.NodeID)
	}
	if elem.TagName != "DIV" {
		t.Errorf("Expected tagName 'DIV', got %q", elem.TagName)
	}
	if elem.Attributes["id"] != "container" {
		t.Errorf("Expected id 'container', got %q", elem.Attributes["id"])
	}
	if elem.BoundingBox == nil {
		t.Fatal("Expected bounding box to be set")
	}
	if elem.BoundingBox.Width != 300 {
		t.Errorf("Expected width 300, got %f", elem.BoundingBox.Width)
	}
}

// TestBoundingBox_Good verifies the BoundingBox struct has expected fields.
func TestBoundingBox_Good(t *testing.T) {
	box := BoundingBox{
		X:      10.5,
		Y:      20.5,
		Width:  100.0,
		Height: 50.0,
	}

	if box.X != 10.5 {
		t.Errorf("Expected X 10.5, got %f", box.X)
	}
	if box.Y != 20.5 {
		t.Errorf("Expected Y 20.5, got %f", box.Y)
	}
	if box.Width != 100.0 {
		t.Errorf("Expected width 100.0, got %f", box.Width)
	}
	if box.Height != 50.0 {
		t.Errorf("Expected height 50.0, got %f", box.Height)
	}
}

// TestWithTimeout_Good verifies the WithTimeout option sets timeout correctly.
func TestWithTimeout_Good(t *testing.T) {
	// We can't fully test without a real Chrome connection,
	// but we can verify the option function works
	wv := &Webview{}
	opt := WithTimeout(60 * time.Second)

	err := opt(wv)
	if err != nil {
		t.Fatalf("WithTimeout returned error: %v", err)
	}

	if wv.timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", wv.timeout)
	}
}

// TestWithConsoleLimit_Good verifies the WithConsoleLimit option sets limit correctly.
func TestWithConsoleLimit_Good(t *testing.T) {
	wv := &Webview{}
	opt := WithConsoleLimit(500)

	err := opt(wv)
	if err != nil {
		t.Fatalf("WithConsoleLimit returned error: %v", err)
	}

	if wv.consoleLimit != 500 {
		t.Errorf("Expected consoleLimit 500, got %d", wv.consoleLimit)
	}
}

// TestNew_Bad_NoDebugURL verifies New fails without a debug URL.
func TestNew_Bad_NoDebugURL(t *testing.T) {
	_, err := New()
	if err == nil {
		t.Error("Expected error when creating Webview without debug URL")
	}
}

// TestNew_Bad_InvalidDebugURL verifies New fails with invalid debug URL.
func TestNew_Bad_InvalidDebugURL(t *testing.T) {
	_, err := New(WithDebugURL("http://localhost:99999"))
	if err == nil {
		t.Error("Expected error when connecting to invalid debug URL")
	}
}

// TestActionSequence_Good verifies action sequence building works.
func TestActionSequence_Good(t *testing.T) {
	seq := NewActionSequence().
		Navigate("https://example.com").
		WaitForSelector("#main").
		Click("#button").
		Type("#input", "hello").
		Wait(100 * time.Millisecond)

	if len(seq.actions) != 5 {
		t.Errorf("Expected 5 actions, got %d", len(seq.actions))
	}
}

// TestClickAction_Good verifies ClickAction struct.
func TestClickAction_Good(t *testing.T) {
	action := ClickAction{Selector: "#submit"}
	if action.Selector != "#submit" {
		t.Errorf("Expected selector '#submit', got %q", action.Selector)
	}
}

// TestTypeAction_Good verifies TypeAction struct.
func TestTypeAction_Good(t *testing.T) {
	action := TypeAction{Selector: "#email", Text: "test@example.com"}
	if action.Selector != "#email" {
		t.Errorf("Expected selector '#email', got %q", action.Selector)
	}
	if action.Text != "test@example.com" {
		t.Errorf("Expected text 'test@example.com', got %q", action.Text)
	}
}

// TestNavigateAction_Good verifies NavigateAction struct.
func TestNavigateAction_Good(t *testing.T) {
	action := NavigateAction{URL: "https://example.com"}
	if action.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got %q", action.URL)
	}
}

// TestWaitAction_Good verifies WaitAction struct.
func TestWaitAction_Good(t *testing.T) {
	action := WaitAction{Duration: 5 * time.Second}
	if action.Duration != 5*time.Second {
		t.Errorf("Expected duration 5s, got %v", action.Duration)
	}
}

// TestWaitForSelectorAction_Good verifies WaitForSelectorAction struct.
func TestWaitForSelectorAction_Good(t *testing.T) {
	action := WaitForSelectorAction{Selector: ".loading"}
	if action.Selector != ".loading" {
		t.Errorf("Expected selector '.loading', got %q", action.Selector)
	}
}

// TestScrollAction_Good verifies ScrollAction struct.
func TestScrollAction_Good(t *testing.T) {
	action := ScrollAction{X: 0, Y: 500}
	if action.X != 0 {
		t.Errorf("Expected X 0, got %d", action.X)
	}
	if action.Y != 500 {
		t.Errorf("Expected Y 500, got %d", action.Y)
	}
}

// TestFocusAction_Good verifies FocusAction struct.
func TestFocusAction_Good(t *testing.T) {
	action := FocusAction{Selector: "#input"}
	if action.Selector != "#input" {
		t.Errorf("Expected selector '#input', got %q", action.Selector)
	}
}

// TestBlurAction_Good verifies BlurAction struct.
func TestBlurAction_Good(t *testing.T) {
	action := BlurAction{Selector: "#input"}
	if action.Selector != "#input" {
		t.Errorf("Expected selector '#input', got %q", action.Selector)
	}
}

// TestClearAction_Good verifies ClearAction struct.
func TestClearAction_Good(t *testing.T) {
	action := ClearAction{Selector: "#input"}
	if action.Selector != "#input" {
		t.Errorf("Expected selector '#input', got %q", action.Selector)
	}
}

// TestSelectAction_Good verifies SelectAction struct.
func TestSelectAction_Good(t *testing.T) {
	action := SelectAction{Selector: "#dropdown", Value: "option1"}
	if action.Selector != "#dropdown" {
		t.Errorf("Expected selector '#dropdown', got %q", action.Selector)
	}
	if action.Value != "option1" {
		t.Errorf("Expected value 'option1', got %q", action.Value)
	}
}

// TestCheckAction_Good verifies CheckAction struct.
func TestCheckAction_Good(t *testing.T) {
	action := CheckAction{Selector: "#checkbox", Checked: true}
	if action.Selector != "#checkbox" {
		t.Errorf("Expected selector '#checkbox', got %q", action.Selector)
	}
	if !action.Checked {
		t.Error("Expected checked to be true")
	}
}

// TestHoverAction_Good verifies HoverAction struct.
func TestHoverAction_Good(t *testing.T) {
	action := HoverAction{Selector: "#menu-item"}
	if action.Selector != "#menu-item" {
		t.Errorf("Expected selector '#menu-item', got %q", action.Selector)
	}
}

// TestDoubleClickAction_Good verifies DoubleClickAction struct.
func TestDoubleClickAction_Good(t *testing.T) {
	action := DoubleClickAction{Selector: "#editable"}
	if action.Selector != "#editable" {
		t.Errorf("Expected selector '#editable', got %q", action.Selector)
	}
}

// TestRightClickAction_Good verifies RightClickAction struct.
func TestRightClickAction_Good(t *testing.T) {
	action := RightClickAction{Selector: "#context-menu-trigger"}
	if action.Selector != "#context-menu-trigger" {
		t.Errorf("Expected selector '#context-menu-trigger', got %q", action.Selector)
	}
}

// TestPressKeyAction_Good verifies PressKeyAction struct.
func TestPressKeyAction_Good(t *testing.T) {
	action := PressKeyAction{Key: "Enter"}
	if action.Key != "Enter" {
		t.Errorf("Expected key 'Enter', got %q", action.Key)
	}
}

// TestSetAttributeAction_Good verifies SetAttributeAction struct.
func TestSetAttributeAction_Good(t *testing.T) {
	action := SetAttributeAction{
		Selector:  "#element",
		Attribute: "data-value",
		Value:     "test",
	}
	if action.Selector != "#element" {
		t.Errorf("Expected selector '#element', got %q", action.Selector)
	}
	if action.Attribute != "data-value" {
		t.Errorf("Expected attribute 'data-value', got %q", action.Attribute)
	}
	if action.Value != "test" {
		t.Errorf("Expected value 'test', got %q", action.Value)
	}
}

// TestRemoveAttributeAction_Good verifies RemoveAttributeAction struct.
func TestRemoveAttributeAction_Good(t *testing.T) {
	action := RemoveAttributeAction{
		Selector:  "#element",
		Attribute: "disabled",
	}
	if action.Selector != "#element" {
		t.Errorf("Expected selector '#element', got %q", action.Selector)
	}
	if action.Attribute != "disabled" {
		t.Errorf("Expected attribute 'disabled', got %q", action.Attribute)
	}
}

// TestSetValueAction_Good verifies SetValueAction struct.
func TestSetValueAction_Good(t *testing.T) {
	action := SetValueAction{
		Selector: "#input",
		Value:    "new value",
	}
	if action.Selector != "#input" {
		t.Errorf("Expected selector '#input', got %q", action.Selector)
	}
	if action.Value != "new value" {
		t.Errorf("Expected value 'new value', got %q", action.Value)
	}
}

// TestScrollIntoViewAction_Good verifies ScrollIntoViewAction struct.
func TestScrollIntoViewAction_Good(t *testing.T) {
	action := ScrollIntoViewAction{Selector: "#target"}
	if action.Selector != "#target" {
		t.Errorf("Expected selector '#target', got %q", action.Selector)
	}
}
