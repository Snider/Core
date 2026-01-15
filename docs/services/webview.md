# WebView Service

The WebView service (`pkg/webview`) provides programmatic interaction with web content in your application windows.

## Features

- JavaScript execution
- DOM manipulation
- Element interaction (click, type, select)
- Console message capture
- Screenshots
- Network request monitoring
- Performance metrics

## Basic Usage

```go
import "github.com/Snider/Core/pkg/webview"

// Create service
wv := webview.New()

// Set Wails app reference
wv.SetApp(app)
```

## JavaScript Execution

```go
// Execute JavaScript and get result
result, err := wv.ExecJS("main", `
    document.title
`)

// Execute complex scripts
result, err := wv.ExecJS("main", `
    const items = document.querySelectorAll('.item');
    Array.from(items).map(el => el.textContent);
`)
```

## DOM Interaction

### Click Element

```go
err := wv.Click("main", "#submit-button")
err := wv.Click("main", ".nav-link:first-child")
```

### Type Text

```go
err := wv.Type("main", "#search-input", "hello world")
```

### Select Option

```go
err := wv.Select("main", "#country-select", "US")
```

### Check/Uncheck

```go
err := wv.Check("main", "#agree-checkbox", true)
```

### Hover

```go
err := wv.Hover("main", ".dropdown-trigger")
```

### Scroll

```go
// Scroll to element
err := wv.Scroll("main", "#section-3", 0, 0)

// Scroll by coordinates
err := wv.Scroll("main", "", 0, 500)
```

## Element Information

### Query Selector

```go
elements, err := wv.QuerySelector("main", ".list-item")
```

### Element Info

```go
info, err := wv.GetElementInfo("main", "#user-card")
// Returns: tag, id, classes, text, attributes, bounds
```

### Computed Styles

```go
styles, err := wv.GetComputedStyle("main", ".button",
    []string{"color", "background-color", "font-size"})
```

### DOM Tree

```go
tree, err := wv.GetDOMTree("main", 5) // max depth 5
```

## Console Messages

```go
// Setup console listener
wv.SetupConsoleListener()

// Inject capture script
wv.InjectConsoleCapture("main")

// Get messages
messages := wv.GetConsoleMessages("all", 100)
messages := wv.GetConsoleMessages("error", 50)

// Clear buffer
wv.ClearConsole()

// Get errors only
errors := wv.GetErrors(50)
```

## Screenshots

```go
// Full page screenshot (base64 PNG)
data, err := wv.Screenshot("main")

// Element screenshot
data, err := wv.ScreenshotElement("main", "#chart")

// Export as PDF
pdfData, err := wv.ExportToPDF("main", map[string]any{
    "margin": 20,
})
```

## Page Information

```go
// Get current URL
url, err := wv.GetURL("main")

// Get page title
title, err := wv.GetTitle("main")

// Get page source
source, err := wv.GetPageSource("main")

// Navigate
err := wv.Navigate("main", "https://example.com")
```

## Network Monitoring

```go
// Inject network interceptor
wv.InjectNetworkInterceptor("main")

// Get captured requests
requests, err := wv.GetNetworkRequests("main", 100)

// Clear request log
wv.ClearNetworkRequests("main")
```

## Performance Metrics

```go
metrics, err := wv.GetPerformance("main")
// Returns: loadTime, domContentLoaded, firstPaint, etc.
```

## Resource Listing

```go
resources, err := wv.GetResources("main")
// Returns: scripts, stylesheets, images, fonts, etc.
```

## Visual Debugging

```go
// Highlight element temporarily
err := wv.Highlight("main", "#target-element", 2000) // 2 seconds
```

## Window Listing

```go
windows := wv.ListWindows()
for _, w := range windows {
    fmt.Println(w.Name)
}
```

## Frontend Usage

The WebView service is primarily used server-side for:

- Automated testing
- AI assistant interactions (via MCP)
- Scripted UI interactions

For normal frontend development, use standard DOM APIs directly.
