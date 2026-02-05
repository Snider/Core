# Angular Testing with Webview MCP Tools

This guide explains how to use the webview MCP tools to automate testing of Angular applications via Chrome DevTools Protocol (CDP).

## Prerequisites

1. **Chrome/Chromium Browser**: Installed and accessible
2. **Remote Debugging Port**: Chrome must be started with remote debugging enabled

### Starting Chrome with Remote Debugging

```bash
# Linux
google-chrome --remote-debugging-port=9222

# macOS
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222

# Windows
"C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222

# Headless mode (no visible window)
google-chrome --headless --remote-debugging-port=9222
```

## Available MCP Tools

### Connection Management

#### webview_connect
Connect to Chrome DevTools.

```json
{
  "tool": "webview_connect",
  "arguments": {
    "debug_url": "http://localhost:9222",
    "timeout": 30
  }
}
```

#### webview_disconnect
Disconnect from Chrome DevTools.

```json
{
  "tool": "webview_disconnect",
  "arguments": {}
}
```

### Navigation

#### webview_navigate
Navigate to a URL.

```json
{
  "tool": "webview_navigate",
  "arguments": {
    "url": "http://localhost:4200"
  }
}
```

### DOM Interaction

#### webview_click
Click an element by CSS selector.

```json
{
  "tool": "webview_click",
  "arguments": {
    "selector": "#login-button"
  }
}
```

#### webview_type
Type text into an element.

```json
{
  "tool": "webview_type",
  "arguments": {
    "selector": "#email-input",
    "text": "user@example.com"
  }
}
```

#### webview_query
Query DOM elements.

```json
{
  "tool": "webview_query",
  "arguments": {
    "selector": ".error-message",
    "all": true
  }
}
```

#### webview_wait
Wait for an element to appear.

```json
{
  "tool": "webview_wait",
  "arguments": {
    "selector": ".loading-spinner",
    "timeout": 10
  }
}
```

### JavaScript Evaluation

#### webview_eval
Execute JavaScript in the browser context.

```json
{
  "tool": "webview_eval",
  "arguments": {
    "script": "document.title"
  }
}
```

### Console & Debugging

#### webview_console
Get browser console output.

```json
{
  "tool": "webview_console",
  "arguments": {
    "clear": false
  }
}
```

#### webview_screenshot
Capture a screenshot.

```json
{
  "tool": "webview_screenshot",
  "arguments": {
    "format": "png"
  }
}
```

## Angular-Specific Testing Patterns

### 1. Waiting for Angular Zone Stability

Before interacting with Angular components, wait for Zone.js to become stable:

```json
{
  "tool": "webview_eval",
  "arguments": {
    "script": "(function() { const roots = window.getAllAngularRootElements(); if (!roots.length) return true; const injector = window.ng.probe(roots[0]).injector; const zone = injector.get('NgZone'); return zone.isStable; })()"
  }
}
```

### 2. Navigating with Angular Router

Use the Angular Router for client-side navigation:

```json
{
  "tool": "webview_eval",
  "arguments": {
    "script": "(function() { const roots = window.getAllAngularRootElements(); const injector = window.ng.probe(roots[0]).injector; const router = injector.get('Router'); router.navigateByUrl('/dashboard'); return true; })()"
  }
}
```

### 3. Accessing Component Properties

Read or modify component state:

```json
{
  "tool": "webview_eval",
  "arguments": {
    "script": "(function() { const el = document.querySelector('app-user-profile'); const component = window.ng.probe(el).componentInstance; return component.user; })()"
  }
}
```

### 4. Triggering Change Detection

Force Angular to update the view:

```json
{
  "tool": "webview_eval",
  "arguments": {
    "script": "(function() { const roots = window.getAllAngularRootElements(); const injector = window.ng.probe(roots[0]).injector; const appRef = injector.get('ApplicationRef'); appRef.tick(); return true; })()"
  }
}
```

### 5. Testing Form Validation

Check Angular form state:

```json
{
  "tool": "webview_eval",
  "arguments": {
    "script": "(function() { const form = document.querySelector('form'); const component = window.ng.probe(form).componentInstance; return { valid: component.form.valid, errors: component.form.errors }; })()"
  }
}
```

## Complete Test Flow Example

Here's a complete example testing an Angular login flow:

### Step 1: Connect to Chrome

```json
{"tool": "webview_connect", "arguments": {"debug_url": "http://localhost:9222"}}
```

### Step 2: Navigate to the Application

```json
{"tool": "webview_navigate", "arguments": {"url": "http://localhost:4200/login"}}
```

### Step 3: Wait for Angular to Load

```json
{"tool": "webview_wait", "arguments": {"selector": "app-login"}}
```

### Step 4: Fill in Login Form

```json
{"tool": "webview_type", "arguments": {"selector": "#email", "text": "test@example.com"}}
{"tool": "webview_type", "arguments": {"selector": "#password", "text": "password123"}}
```

### Step 5: Submit the Form

```json
{"tool": "webview_click", "arguments": {"selector": "button[type='submit']"}}
```

### Step 6: Wait for Navigation

```json
{"tool": "webview_wait", "arguments": {"selector": "app-dashboard", "timeout": 10}}
```

### Step 7: Verify Success

```json
{"tool": "webview_eval", "arguments": {"script": "window.location.pathname === '/dashboard'"}}
```

### Step 8: Check Console for Errors

```json
{"tool": "webview_console", "arguments": {"clear": true}}
```

### Step 9: Disconnect

```json
{"tool": "webview_disconnect", "arguments": {}}
```

## Debugging Tips

### 1. Check for JavaScript Errors

Always check the console output after operations:

```json
{"tool": "webview_console", "arguments": {}}
```

### 2. Take Screenshots on Failure

Capture the current state when something unexpected happens:

```json
{"tool": "webview_screenshot", "arguments": {"format": "png"}}
```

### 3. Inspect Element State

Query elements to understand their current state:

```json
{"tool": "webview_query", "arguments": {"selector": ".my-component", "all": false}}
```

### 4. Get Page Source

Retrieve the current HTML for debugging:

```json
{"tool": "webview_eval", "arguments": {"script": "document.documentElement.outerHTML"}}
```

## Common Issues

### Element Not Found

If `webview_click` or `webview_type` fails with "element not found":

1. Check the selector is correct
2. Wait for the element to appear first
3. Verify the element is visible (not hidden)

### Angular Not Detected

If Angular-specific scripts fail:

1. Ensure the Angular app has loaded completely
2. Check that you're using Angular 2+ (not AngularJS)
3. Verify the element has an Angular component attached

### Timeout Errors

If operations timeout:

1. Increase the timeout value
2. Check for loading spinners or blocking operations
3. Verify the network is working correctly

## Best Practices

1. **Always wait for elements** before interacting with them
2. **Check console for errors** after each major step
3. **Use explicit selectors** like IDs or data attributes
4. **Clear console** at the start of each test
5. **Disconnect** when done to free resources
6. **Take screenshots** at key checkpoints
7. **Handle async operations** by waiting for stability

## Go API Usage

For direct Go integration, use the `pkg/webview` package:

```go
package main

import (
    "log"
    "time"

    "github.com/host-uk/core/pkg/webview"
)

func main() {
    // Connect to Chrome
    wv, err := webview.New(
        webview.WithDebugURL("http://localhost:9222"),
        webview.WithTimeout(30*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer wv.Close()

    // Navigate
    if err := wv.Navigate("http://localhost:4200"); err != nil {
        log.Fatal(err)
    }

    // Wait for element
    if err := wv.WaitForSelector("app-root"); err != nil {
        log.Fatal(err)
    }

    // Click button
    if err := wv.Click("#login-button"); err != nil {
        log.Fatal(err)
    }

    // Type text
    if err := wv.Type("#email", "test@example.com"); err != nil {
        log.Fatal(err)
    }

    // Get console output
    messages := wv.GetConsole()
    for _, msg := range messages {
        log.Printf("[%s] %s", msg.Type, msg.Text)
    }

    // Take screenshot
    data, err := wv.Screenshot()
    if err != nil {
        log.Fatal(err)
    }
    // Save data to file...
}
```

### Using Angular Helper

For Angular-specific operations:

```go
package main

import (
    "log"
    "time"

    "github.com/host-uk/core/pkg/webview"
)

func main() {
    wv, err := webview.New(webview.WithDebugURL("http://localhost:9222"))
    if err != nil {
        log.Fatal(err)
    }
    defer wv.Close()

    // Create Angular helper
    angular := webview.NewAngularHelper(wv)

    // Navigate using Angular Router
    if err := angular.NavigateByRouter("/dashboard"); err != nil {
        log.Fatal(err)
    }

    // Wait for Angular to stabilize
    if err := angular.WaitForAngular(); err != nil {
        log.Fatal(err)
    }

    // Get component property
    value, err := angular.GetComponentProperty("app-user-profile", "user")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("User: %v", value)

    // Call component method
    result, err := angular.CallComponentMethod("app-counter", "increment", 5)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Result: %v", result)
}
```

## See Also

- [Chrome DevTools Protocol Documentation](https://chromedevtools.github.io/devtools-protocol/)
- [pkg/webview package documentation](../../pkg/webview/)
- [MCP Tools Reference](../mcp/)
