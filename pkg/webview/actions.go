package webview

import (
	"context"
	"fmt"
	"time"
)

// Action represents a browser action that can be performed.
type Action interface {
	Execute(ctx context.Context, wv *Webview) error
}

// ClickAction represents a click action.
type ClickAction struct {
	Selector string
}

// Execute performs the click action.
func (a ClickAction) Execute(ctx context.Context, wv *Webview) error {
	return wv.click(ctx, a.Selector)
}

// TypeAction represents a typing action.
type TypeAction struct {
	Selector string
	Text     string
}

// Execute performs the type action.
func (a TypeAction) Execute(ctx context.Context, wv *Webview) error {
	return wv.typeText(ctx, a.Selector, a.Text)
}

// NavigateAction represents a navigation action.
type NavigateAction struct {
	URL string
}

// Execute performs the navigate action.
func (a NavigateAction) Execute(ctx context.Context, wv *Webview) error {
	_, err := wv.client.Call(ctx, "Page.navigate", map[string]any{
		"url": a.URL,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	return wv.waitForLoad(ctx)
}

// WaitAction represents a wait action.
type WaitAction struct {
	Duration time.Duration
}

// Execute performs the wait action.
func (a WaitAction) Execute(ctx context.Context, wv *Webview) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(a.Duration):
		return nil
	}
}

// WaitForSelectorAction represents waiting for a selector.
type WaitForSelectorAction struct {
	Selector string
}

// Execute waits for the selector to appear.
func (a WaitForSelectorAction) Execute(ctx context.Context, wv *Webview) error {
	return wv.waitForSelector(ctx, a.Selector)
}

// ScrollAction represents a scroll action.
type ScrollAction struct {
	X int
	Y int
}

// Execute performs the scroll action.
func (a ScrollAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf("window.scrollTo(%d, %d)", a.X, a.Y)
	_, err := wv.evaluate(ctx, script)
	return err
}

// ScrollIntoViewAction scrolls an element into view.
type ScrollIntoViewAction struct {
	Selector string
}

// Execute scrolls the element into view.
func (a ScrollIntoViewAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf("document.querySelector(%q)?.scrollIntoView({behavior: 'smooth', block: 'center'})", a.Selector)
	_, err := wv.evaluate(ctx, script)
	return err
}

// FocusAction focuses an element.
type FocusAction struct {
	Selector string
}

// Execute focuses the element.
func (a FocusAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf("document.querySelector(%q)?.focus()", a.Selector)
	_, err := wv.evaluate(ctx, script)
	return err
}

// BlurAction removes focus from an element.
type BlurAction struct {
	Selector string
}

// Execute removes focus from the element.
func (a BlurAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf("document.querySelector(%q)?.blur()", a.Selector)
	_, err := wv.evaluate(ctx, script)
	return err
}

// ClearAction clears the value of an input element.
type ClearAction struct {
	Selector string
}

// Execute clears the input value.
func (a ClearAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (el) {
			el.value = '';
			el.dispatchEvent(new Event('input', {bubbles: true}));
			el.dispatchEvent(new Event('change', {bubbles: true}));
		}
	`, a.Selector)
	_, err := wv.evaluate(ctx, script)
	return err
}

// SelectAction selects an option in a select element.
type SelectAction struct {
	Selector string
	Value    string
}

// Execute selects the option.
func (a SelectAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (el) {
			el.value = %q;
			el.dispatchEvent(new Event('change', {bubbles: true}));
		}
	`, a.Selector, a.Value)
	_, err := wv.evaluate(ctx, script)
	return err
}

// CheckAction checks or unchecks a checkbox.
type CheckAction struct {
	Selector string
	Checked  bool
}

// Execute checks/unchecks the checkbox.
func (a CheckAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (el && el.checked !== %t) {
			el.click();
		}
	`, a.Selector, a.Checked)
	_, err := wv.evaluate(ctx, script)
	return err
}

// HoverAction hovers over an element.
type HoverAction struct {
	Selector string
}

// Execute hovers over the element.
func (a HoverAction) Execute(ctx context.Context, wv *Webview) error {
	elem, err := wv.querySelector(ctx, a.Selector)
	if err != nil {
		return err
	}

	if elem.BoundingBox == nil {
		return fmt.Errorf("element has no bounding box")
	}

	x := elem.BoundingBox.X + elem.BoundingBox.Width/2
	y := elem.BoundingBox.Y + elem.BoundingBox.Height/2

	_, err = wv.client.Call(ctx, "Input.dispatchMouseEvent", map[string]any{
		"type": "mouseMoved",
		"x":    x,
		"y":    y,
	})
	return err
}

// DoubleClickAction double-clicks an element.
type DoubleClickAction struct {
	Selector string
}

// Execute double-clicks the element.
func (a DoubleClickAction) Execute(ctx context.Context, wv *Webview) error {
	elem, err := wv.querySelector(ctx, a.Selector)
	if err != nil {
		return err
	}

	if elem.BoundingBox == nil {
		// Fallback to JavaScript
		script := fmt.Sprintf(`
			const el = document.querySelector(%q);
			if (el) {
				const event = new MouseEvent('dblclick', {bubbles: true, cancelable: true, view: window});
				el.dispatchEvent(event);
			}
		`, a.Selector)
		_, err := wv.evaluate(ctx, script)
		return err
	}

	x := elem.BoundingBox.X + elem.BoundingBox.Width/2
	y := elem.BoundingBox.Y + elem.BoundingBox.Height/2

	// Double click sequence
	for i := 0; i < 2; i++ {
		for _, eventType := range []string{"mousePressed", "mouseReleased"} {
			_, err := wv.client.Call(ctx, "Input.dispatchMouseEvent", map[string]any{
				"type":       eventType,
				"x":          x,
				"y":          y,
				"button":     "left",
				"clickCount": i + 1,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// RightClickAction right-clicks an element.
type RightClickAction struct {
	Selector string
}

// Execute right-clicks the element.
func (a RightClickAction) Execute(ctx context.Context, wv *Webview) error {
	elem, err := wv.querySelector(ctx, a.Selector)
	if err != nil {
		return err
	}

	if elem.BoundingBox == nil {
		// Fallback to JavaScript
		script := fmt.Sprintf(`
			const el = document.querySelector(%q);
			if (el) {
				const event = new MouseEvent('contextmenu', {bubbles: true, cancelable: true, view: window});
				el.dispatchEvent(event);
			}
		`, a.Selector)
		_, err := wv.evaluate(ctx, script)
		return err
	}

	x := elem.BoundingBox.X + elem.BoundingBox.Width/2
	y := elem.BoundingBox.Y + elem.BoundingBox.Height/2

	for _, eventType := range []string{"mousePressed", "mouseReleased"} {
		_, err := wv.client.Call(ctx, "Input.dispatchMouseEvent", map[string]any{
			"type":       eventType,
			"x":          x,
			"y":          y,
			"button":     "right",
			"clickCount": 1,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// PressKeyAction presses a key.
type PressKeyAction struct {
	Key string // e.g., "Enter", "Tab", "Escape"
}

// Execute presses the key.
func (a PressKeyAction) Execute(ctx context.Context, wv *Webview) error {
	// Map common key names to CDP key codes
	keyMap := map[string]struct {
		code       string
		keyCode    int
		text       string
		unmodified string
	}{
		"Enter":      {"Enter", 13, "\r", "\r"},
		"Tab":        {"Tab", 9, "", ""},
		"Escape":     {"Escape", 27, "", ""},
		"Backspace":  {"Backspace", 8, "", ""},
		"Delete":     {"Delete", 46, "", ""},
		"ArrowUp":    {"ArrowUp", 38, "", ""},
		"ArrowDown":  {"ArrowDown", 40, "", ""},
		"ArrowLeft":  {"ArrowLeft", 37, "", ""},
		"ArrowRight": {"ArrowRight", 39, "", ""},
		"Home":       {"Home", 36, "", ""},
		"End":        {"End", 35, "", ""},
		"PageUp":     {"PageUp", 33, "", ""},
		"PageDown":   {"PageDown", 34, "", ""},
	}

	keyInfo, ok := keyMap[a.Key]
	if !ok {
		// For simple characters, just send key events
		_, err := wv.client.Call(ctx, "Input.dispatchKeyEvent", map[string]any{
			"type": "keyDown",
			"text": a.Key,
		})
		if err != nil {
			return err
		}
		_, err = wv.client.Call(ctx, "Input.dispatchKeyEvent", map[string]any{
			"type": "keyUp",
		})
		return err
	}

	params := map[string]any{
		"type":                  "keyDown",
		"code":                  keyInfo.code,
		"key":                   a.Key,
		"windowsVirtualKeyCode": keyInfo.keyCode,
		"nativeVirtualKeyCode":  keyInfo.keyCode,
	}
	if keyInfo.text != "" {
		params["text"] = keyInfo.text
		params["unmodifiedText"] = keyInfo.unmodified
	}

	_, err := wv.client.Call(ctx, "Input.dispatchKeyEvent", params)
	if err != nil {
		return err
	}

	params["type"] = "keyUp"
	delete(params, "text")
	delete(params, "unmodifiedText")
	_, err = wv.client.Call(ctx, "Input.dispatchKeyEvent", params)
	return err
}

// SetAttributeAction sets an attribute on an element.
type SetAttributeAction struct {
	Selector  string
	Attribute string
	Value     string
}

// Execute sets the attribute.
func (a SetAttributeAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf("document.querySelector(%q)?.setAttribute(%q, %q)", a.Selector, a.Attribute, a.Value)
	_, err := wv.evaluate(ctx, script)
	return err
}

// RemoveAttributeAction removes an attribute from an element.
type RemoveAttributeAction struct {
	Selector  string
	Attribute string
}

// Execute removes the attribute.
func (a RemoveAttributeAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf("document.querySelector(%q)?.removeAttribute(%q)", a.Selector, a.Attribute)
	_, err := wv.evaluate(ctx, script)
	return err
}

// SetValueAction sets the value of an input element.
type SetValueAction struct {
	Selector string
	Value    string
}

// Execute sets the value.
func (a SetValueAction) Execute(ctx context.Context, wv *Webview) error {
	script := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (el) {
			el.value = %q;
			el.dispatchEvent(new Event('input', {bubbles: true}));
			el.dispatchEvent(new Event('change', {bubbles: true}));
		}
	`, a.Selector, a.Value)
	_, err := wv.evaluate(ctx, script)
	return err
}

// ActionSequence represents a sequence of actions to execute.
type ActionSequence struct {
	actions []Action
}

// NewActionSequence creates a new action sequence.
func NewActionSequence() *ActionSequence {
	return &ActionSequence{
		actions: make([]Action, 0),
	}
}

// Add adds an action to the sequence.
func (s *ActionSequence) Add(action Action) *ActionSequence {
	s.actions = append(s.actions, action)
	return s
}

// Click adds a click action.
func (s *ActionSequence) Click(selector string) *ActionSequence {
	return s.Add(ClickAction{Selector: selector})
}

// Type adds a type action.
func (s *ActionSequence) Type(selector, text string) *ActionSequence {
	return s.Add(TypeAction{Selector: selector, Text: text})
}

// Navigate adds a navigate action.
func (s *ActionSequence) Navigate(url string) *ActionSequence {
	return s.Add(NavigateAction{URL: url})
}

// Wait adds a wait action.
func (s *ActionSequence) Wait(d time.Duration) *ActionSequence {
	return s.Add(WaitAction{Duration: d})
}

// WaitForSelector adds a wait for selector action.
func (s *ActionSequence) WaitForSelector(selector string) *ActionSequence {
	return s.Add(WaitForSelectorAction{Selector: selector})
}

// Execute executes all actions in the sequence.
func (s *ActionSequence) Execute(ctx context.Context, wv *Webview) error {
	for i, action := range s.actions {
		if err := action.Execute(ctx, wv); err != nil {
			return fmt.Errorf("action %d failed: %w", i, err)
		}
	}
	return nil
}

// UploadFile uploads a file to a file input element.
func (wv *Webview) UploadFile(selector string, filePaths []string) error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	// Get the element's node ID
	elem, err := wv.querySelector(ctx, selector)
	if err != nil {
		return err
	}

	// Use DOM.setFileInputFiles to set the files
	_, err = wv.client.Call(ctx, "DOM.setFileInputFiles", map[string]any{
		"nodeId": elem.NodeID,
		"files":  filePaths,
	})
	return err
}

// DragAndDrop performs a drag and drop operation.
func (wv *Webview) DragAndDrop(sourceSelector, targetSelector string) error {
	ctx, cancel := context.WithTimeout(wv.ctx, wv.timeout)
	defer cancel()

	// Get source and target elements
	source, err := wv.querySelector(ctx, sourceSelector)
	if err != nil {
		return fmt.Errorf("source element not found: %w", err)
	}
	if source.BoundingBox == nil {
		return fmt.Errorf("source element has no bounding box")
	}

	target, err := wv.querySelector(ctx, targetSelector)
	if err != nil {
		return fmt.Errorf("target element not found: %w", err)
	}
	if target.BoundingBox == nil {
		return fmt.Errorf("target element has no bounding box")
	}

	// Calculate center points
	sourceX := source.BoundingBox.X + source.BoundingBox.Width/2
	sourceY := source.BoundingBox.Y + source.BoundingBox.Height/2
	targetX := target.BoundingBox.X + target.BoundingBox.Width/2
	targetY := target.BoundingBox.Y + target.BoundingBox.Height/2

	// Mouse down on source
	_, err = wv.client.Call(ctx, "Input.dispatchMouseEvent", map[string]any{
		"type":       "mousePressed",
		"x":          sourceX,
		"y":          sourceY,
		"button":     "left",
		"clickCount": 1,
	})
	if err != nil {
		return err
	}

	// Move to target
	_, err = wv.client.Call(ctx, "Input.dispatchMouseEvent", map[string]any{
		"type":   "mouseMoved",
		"x":      targetX,
		"y":      targetY,
		"button": "left",
	})
	if err != nil {
		return err
	}

	// Mouse up on target
	_, err = wv.client.Call(ctx, "Input.dispatchMouseEvent", map[string]any{
		"type":       "mouseReleased",
		"x":          targetX,
		"y":          targetY,
		"button":     "left",
		"clickCount": 1,
	})
	return err
}
