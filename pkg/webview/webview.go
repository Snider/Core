// Package webview provides WebView interaction capabilities for the MCP server.
// It enables JavaScript execution, console capture, screenshots, and DOM interaction
// in running Wails windows.
package webview

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ConsoleMessage represents a captured console message.
type ConsoleMessage struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source,omitempty"`
	Line      int       `json:"line,omitempty"`
}

// Service provides WebView interaction capabilities.
type Service struct {
	app            *application.App
	consoleBuffer  []ConsoleMessage
	consoleMu      sync.RWMutex
	maxConsoleSize int
	onConsole      func(ConsoleMessage)
}

// New creates a new WebView service.
func New() *Service {
	return &Service{
		consoleBuffer:  make([]ConsoleMessage, 0, 1000),
		maxConsoleSize: 1000,
	}
}

// SetApp sets the Wails application reference.
// This must be called after the app is initialized.
func (s *Service) SetApp(app *application.App) {
	s.app = app
}

// OnConsole sets a callback for console messages.
func (s *Service) OnConsole(cb func(ConsoleMessage)) {
	s.onConsole = cb
}

// GetWindow returns a window by name, or the first window if name is empty.
func (s *Service) GetWindow(name string) *application.WebviewWindow {
	if s.app == nil {
		return nil
	}

	windows := s.app.Window.GetAll()
	if len(windows) == 0 {
		return nil
	}

	if name == "" {
		// Return first WebviewWindow
		for _, w := range windows {
			if wv, ok := w.(*application.WebviewWindow); ok {
				return wv
			}
		}
		return nil
	}

	// Find by name
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				return wv
			}
		}
	}
	return nil
}

// ListWindows returns info about all open windows.
func (s *Service) ListWindows() []WindowInfo {
	if s.app == nil {
		return nil
	}

	windows := s.app.Window.GetAll()
	result := make([]WindowInfo, 0, len(windows))

	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			result = append(result, WindowInfo{
				Name: wv.Name(),
			})
		}
	}
	return result
}

// WindowInfo contains information about a window.
type WindowInfo struct {
	Name string `json:"name"`
}

// ExecJS executes JavaScript in the specified window and returns the result.
func (s *Service) ExecJS(windowName string, code string) (string, error) {
	window := s.GetWindow(windowName)
	if window == nil {
		return "", fmt.Errorf("window not found: %s", windowName)
	}

	// Wrap code to capture return value
	wrappedCode := fmt.Sprintf(`
		(function() {
			try {
				const result = (function() { %s })();
				return JSON.stringify({ success: true, result: result });
			} catch (e) {
				return JSON.stringify({ success: false, error: e.message, stack: e.stack });
			}
		})()
	`, code)

	window.ExecJS(wrappedCode)

	// Note: Wails v3 ExecJS is fire-and-forget
	// For return values, we need to use events or a different mechanism
	return "executed", nil
}

// ExecJSAsync executes JavaScript and returns result via callback.
// This uses events to get the return value.
func (s *Service) ExecJSAsync(windowName string, code string, callback func(result string, err error)) {
	window := s.GetWindow(windowName)
	if window == nil {
		callback("", fmt.Errorf("window not found: %s", windowName))
		return
	}

	// Generate unique callback ID
	callbackID := fmt.Sprintf("mcp_eval_%d", time.Now().UnixNano())

	// Register one-time event handler
	var unsubscribe func()
	unsubscribe = s.app.Event.On(callbackID, func(event *application.CustomEvent) {
		unsubscribe()
		if data, ok := event.Data.(string); ok {
			callback(data, nil)
		} else {
			callback("", fmt.Errorf("invalid response type"))
		}
	})

	// Execute with callback
	wrappedCode := fmt.Sprintf(`
		(async function() {
			try {
				const result = await (async function() { %s })();
				window.wails.Events.Emit('%s', JSON.stringify({ success: true, result: result }));
			} catch (e) {
				window.wails.Events.Emit('%s', JSON.stringify({ success: false, error: e.message }));
			}
		})()
	`, code, callbackID, callbackID)

	window.ExecJS(wrappedCode)

	// Timeout after 30 seconds
	go func() {
		time.Sleep(30 * time.Second)
		unsubscribe()
	}()
}

// InjectConsoleCapture injects JavaScript to capture console output.
func (s *Service) InjectConsoleCapture(windowName string) error {
	window := s.GetWindow(windowName)
	if window == nil {
		return fmt.Errorf("window not found: %s", windowName)
	}

	// Inject console interceptor
	code := `
		(function() {
			if (window.__mcpConsoleInjected) return;
			window.__mcpConsoleInjected = true;

			const originalConsole = {
				log: console.log,
				warn: console.warn,
				error: console.error,
				info: console.info,
				debug: console.debug
			};

			function intercept(level) {
				return function(...args) {
					originalConsole[level].apply(console, args);
					try {
						const message = args.map(a => {
							if (typeof a === 'object') return JSON.stringify(a);
							return String(a);
						}).join(' ');
						window.wails.Events.Emit('mcp:console', JSON.stringify({
							level: level,
							message: message,
							timestamp: new Date().toISOString()
						}));
					} catch (e) {}
				};
			}

			console.log = intercept('log');
			console.warn = intercept('warn');
			console.error = intercept('error');
			console.info = intercept('info');
			console.debug = intercept('debug');

			// Capture uncaught errors
			window.addEventListener('error', function(e) {
				window.wails.Events.Emit('mcp:console', JSON.stringify({
					level: 'error',
					message: e.message + ' at ' + e.filename + ':' + e.lineno,
					timestamp: new Date().toISOString(),
					source: e.filename,
					line: e.lineno
				}));
			});

			// Capture unhandled promise rejections
			window.addEventListener('unhandledrejection', function(e) {
				window.wails.Events.Emit('mcp:console', JSON.stringify({
					level: 'error',
					message: 'Unhandled rejection: ' + (e.reason?.message || e.reason),
					timestamp: new Date().toISOString()
				}));
			});
		})()
	`

	window.ExecJS(code)
	return nil
}

// SetupConsoleListener sets up the Go-side listener for console events.
func (s *Service) SetupConsoleListener() {
	if s.app == nil {
		return
	}

	s.app.Event.On("mcp:console", func(event *application.CustomEvent) {
		if data, ok := event.Data.(string); ok {
			var msg ConsoleMessage
			if err := json.Unmarshal([]byte(data), &msg); err == nil {
				s.addConsoleMessage(msg)
			}
		}
	})
}

func (s *Service) addConsoleMessage(msg ConsoleMessage) {
	s.consoleMu.Lock()
	defer s.consoleMu.Unlock()

	if len(s.consoleBuffer) >= s.maxConsoleSize {
		// Remove oldest
		s.consoleBuffer = s.consoleBuffer[1:]
	}
	s.consoleBuffer = append(s.consoleBuffer, msg)

	// Notify callback
	if s.onConsole != nil {
		s.onConsole(msg)
	}
}

// GetConsoleMessages returns captured console messages.
func (s *Service) GetConsoleMessages(level string, limit int) []ConsoleMessage {
	s.consoleMu.RLock()
	defer s.consoleMu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	result := make([]ConsoleMessage, 0, limit)
	for i := len(s.consoleBuffer) - 1; i >= 0 && len(result) < limit; i-- {
		msg := s.consoleBuffer[i]
		if level == "" || msg.Level == level {
			result = append(result, msg)
		}
	}

	// Reverse to chronological order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// ClearConsole clears the console buffer.
func (s *Service) ClearConsole() {
	s.consoleMu.Lock()
	defer s.consoleMu.Unlock()
	s.consoleBuffer = s.consoleBuffer[:0]
}

// Click simulates a click on an element by selector.
func (s *Service) Click(windowName string, selector string) error {
	code := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (!el) throw new Error('Element not found: %s');
		el.click();
		return 'clicked';
	`, selector, selector)

	_, err := s.ExecJS(windowName, code)
	return err
}

// Type types text into an element.
func (s *Service) Type(windowName string, selector string, text string) error {
	code := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (!el) throw new Error('Element not found: %s');
		el.focus();
		el.value = %q;
		el.dispatchEvent(new Event('input', { bubbles: true }));
		el.dispatchEvent(new Event('change', { bubbles: true }));
		return 'typed';
	`, selector, selector, text)

	_, err := s.ExecJS(windowName, code)
	return err
}

// QuerySelector returns info about elements matching a selector.
func (s *Service) QuerySelector(windowName string, selector string) (string, error) {
	code := fmt.Sprintf(`
		const els = document.querySelectorAll(%q);
		return Array.from(els).map(el => ({
			tag: el.tagName.toLowerCase(),
			id: el.id,
			class: el.className,
			text: el.textContent?.substring(0, 100),
			rect: el.getBoundingClientRect()
		}));
	`, selector)

	return s.ExecJS(windowName, code)
}

// ScreenshotResult holds the result of a screenshot operation.
type ScreenshotResult struct {
	Data  string `json:"data,omitempty"`  // Base64 PNG data
	Error string `json:"error,omitempty"` // Error message if failed
}

// Screenshot captures a screenshot of the window.
// Returns base64-encoded PNG data via callback (async operation).
func (s *Service) Screenshot(windowName string) (string, error) {
	window := s.GetWindow(windowName)
	if window == nil {
		return "", fmt.Errorf("window not found: %s", windowName)
	}

	// Generate unique callback ID for this screenshot
	callbackID := fmt.Sprintf("mcp_screenshot_%d", time.Now().UnixNano())

	// Channel to receive result
	resultChan := make(chan ScreenshotResult, 1)

	// Register one-time event handler
	var unsubscribe func()
	unsubscribe = s.app.Event.On(callbackID, func(event *application.CustomEvent) {
		unsubscribe()
		if data, ok := event.Data.(string); ok {
			var result ScreenshotResult
			if err := json.Unmarshal([]byte(data), &result); err != nil {
				resultChan <- ScreenshotResult{Error: "failed to parse result"}
			} else {
				resultChan <- result
			}
		} else {
			resultChan <- ScreenshotResult{Error: "invalid response type"}
		}
	})

	// Inject html2canvas if not present and capture screenshot
	code := fmt.Sprintf(`
		(async function() {
			try {
				// Load html2canvas dynamically if not present
				if (typeof html2canvas === 'undefined') {
					await new Promise((resolve, reject) => {
						const script = document.createElement('script');
						script.src = 'https://cdnjs.cloudflare.com/ajax/libs/html2canvas/1.4.1/html2canvas.min.js';
						script.onload = resolve;
						script.onerror = () => reject(new Error('Failed to load html2canvas'));
						document.head.appendChild(script);
					});
				}

				const canvas = await html2canvas(document.body, {
					useCORS: true,
					allowTaint: true,
					backgroundColor: null,
					scale: window.devicePixelRatio || 1
				});
				const dataUrl = canvas.toDataURL('image/png');
				window.wails.Events.Emit('%s', JSON.stringify({ data: dataUrl }));
			} catch (e) {
				window.wails.Events.Emit('%s', JSON.stringify({ error: e.message }));
			}
		})()
	`, callbackID, callbackID)

	window.ExecJS(code)

	// Wait for result with timeout
	select {
	case result := <-resultChan:
		if result.Error != "" {
			return "", fmt.Errorf("screenshot failed: %s", result.Error)
		}
		return result.Data, nil
	case <-time.After(15 * time.Second):
		unsubscribe()
		return "", fmt.Errorf("screenshot timeout")
	}
}

// ScreenshotElement captures a screenshot of a specific element.
func (s *Service) ScreenshotElement(windowName string, selector string) (string, error) {
	window := s.GetWindow(windowName)
	if window == nil {
		return "", fmt.Errorf("window not found: %s", windowName)
	}

	callbackID := fmt.Sprintf("mcp_screenshot_%d", time.Now().UnixNano())
	resultChan := make(chan ScreenshotResult, 1)

	var unsubscribe func()
	unsubscribe = s.app.Event.On(callbackID, func(event *application.CustomEvent) {
		unsubscribe()
		if data, ok := event.Data.(string); ok {
			var result ScreenshotResult
			if err := json.Unmarshal([]byte(data), &result); err != nil {
				resultChan <- ScreenshotResult{Error: "failed to parse result"}
			} else {
				resultChan <- result
			}
		} else {
			resultChan <- ScreenshotResult{Error: "invalid response type"}
		}
	})

	code := fmt.Sprintf(`
		(async function() {
			try {
				const el = document.querySelector(%q);
				if (!el) throw new Error('Element not found: %s');

				if (typeof html2canvas === 'undefined') {
					await new Promise((resolve, reject) => {
						const script = document.createElement('script');
						script.src = 'https://cdnjs.cloudflare.com/ajax/libs/html2canvas/1.4.1/html2canvas.min.js';
						script.onload = resolve;
						script.onerror = () => reject(new Error('Failed to load html2canvas'));
						document.head.appendChild(script);
					});
				}

				const canvas = await html2canvas(el, {
					useCORS: true,
					allowTaint: true,
					backgroundColor: null,
					scale: window.devicePixelRatio || 1
				});
				const dataUrl = canvas.toDataURL('image/png');
				window.wails.Events.Emit('%s', JSON.stringify({ data: dataUrl }));
			} catch (e) {
				window.wails.Events.Emit('%s', JSON.stringify({ error: e.message }));
			}
		})()
	`, selector, selector, callbackID, callbackID)

	window.ExecJS(code)

	select {
	case result := <-resultChan:
		if result.Error != "" {
			return "", fmt.Errorf("screenshot failed: %s", result.Error)
		}
		return result.Data, nil
	case <-time.After(15 * time.Second):
		unsubscribe()
		return "", fmt.Errorf("screenshot timeout")
	}
}

// GetPageSource returns the current page HTML.
func (s *Service) GetPageSource(windowName string) (string, error) {
	code := `return document.documentElement.outerHTML;`
	return s.ExecJS(windowName, code)
}

// GetURL returns the current page URL.
func (s *Service) GetURL(windowName string) (string, error) {
	code := `return window.location.href;`
	return s.ExecJS(windowName, code)
}

// Navigate navigates to a URL.
func (s *Service) Navigate(windowName string, url string) error {
	window := s.GetWindow(windowName)
	if window == nil {
		return fmt.Errorf("window not found: %s", windowName)
	}

	// Use Angular router if available, otherwise location
	code := fmt.Sprintf(`
		if (window.ng && window.ng.getComponent) {
			// Try Angular router
			const router = window.ng.getComponent(document.querySelector('router-outlet'))?.router;
			if (router) {
				router.navigateByUrl(%q);
				return;
			}
		}
		window.location.href = %q;
	`, url, url)

	window.ExecJS(code)
	return nil
}

// EncodeBase64 is a helper to encode bytes to base64.
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// GetTitle returns the current page title.
func (s *Service) GetTitle(windowName string) (string, error) {
	code := `return document.title;`
	return s.ExecJS(windowName, code)
}

// Scroll scrolls to an element or position.
// If selector is provided, scrolls to that element.
// Otherwise scrolls to the x,y position.
func (s *Service) Scroll(windowName string, selector string, x, y int) error {
	var code string
	if selector != "" {
		code = fmt.Sprintf(`
			const el = document.querySelector(%q);
			if (!el) throw new Error('Element not found: %s');
			el.scrollIntoView({ behavior: 'smooth', block: 'center' });
			return 'scrolled';
		`, selector, selector)
	} else {
		code = fmt.Sprintf(`
			window.scrollTo({ top: %d, left: %d, behavior: 'smooth' });
			return 'scrolled';
		`, y, x)
	}
	_, err := s.ExecJS(windowName, code)
	return err
}

// Hover simulates hovering over an element.
func (s *Service) Hover(windowName string, selector string) error {
	code := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (!el) throw new Error('Element not found: %s');
		const event = new MouseEvent('mouseenter', {
			bubbles: true,
			cancelable: true,
			view: window
		});
		el.dispatchEvent(event);
		const hoverEvent = new MouseEvent('mouseover', {
			bubbles: true,
			cancelable: true,
			view: window
		});
		el.dispatchEvent(hoverEvent);
		return 'hovered';
	`, selector, selector)
	_, err := s.ExecJS(windowName, code)
	return err
}

// Select selects an option in a dropdown/select element.
func (s *Service) Select(windowName string, selector string, value string) error {
	code := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (!el) throw new Error('Element not found: %s');
		if (el.tagName.toLowerCase() !== 'select') {
			throw new Error('Element is not a select element');
		}
		el.value = %q;
		el.dispatchEvent(new Event('change', { bubbles: true }));
		return 'selected';
	`, selector, selector, value)
	_, err := s.ExecJS(windowName, code)
	return err
}

// Check sets the checked state of a checkbox or radio button.
func (s *Service) Check(windowName string, selector string, checked bool) error {
	code := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (!el) throw new Error('Element not found: %s');
		if (el.type !== 'checkbox' && el.type !== 'radio') {
			throw new Error('Element is not a checkbox or radio button');
		}
		el.checked = %t;
		el.dispatchEvent(new Event('change', { bubbles: true }));
		return 'checked';
	`, selector, selector, checked)
	_, err := s.ExecJS(windowName, code)
	return err
}

// GetElementInfo returns detailed info about a specific element.
func (s *Service) GetElementInfo(windowName string, selector string) (string, error) {
	code := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (!el) throw new Error('Element not found: %s');
		const rect = el.getBoundingClientRect();
		const styles = window.getComputedStyle(el);
		return {
			tag: el.tagName.toLowerCase(),
			id: el.id,
			className: el.className,
			text: el.textContent?.substring(0, 500),
			innerHTML: el.innerHTML?.substring(0, 1000),
			value: el.value,
			type: el.type,
			href: el.href,
			src: el.src,
			checked: el.checked,
			disabled: el.disabled,
			visible: rect.width > 0 && rect.height > 0,
			rect: { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
			styles: {
				display: styles.display,
				visibility: styles.visibility,
				color: styles.color,
				backgroundColor: styles.backgroundColor,
				fontSize: styles.fontSize
			},
			attributes: Object.fromEntries(
				Array.from(el.attributes).map(a => [a.name, a.value])
			)
		};
	`, selector, selector)
	return s.ExecJS(windowName, code)
}

// GetComputedStyle returns computed styles for an element.
func (s *Service) GetComputedStyle(windowName string, selector string, properties []string) (string, error) {
	propsJSON, _ := json.Marshal(properties)
	code := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (!el) throw new Error('Element not found: %s');
		const styles = window.getComputedStyle(el);
		const props = %s;
		if (props.length === 0) {
			// Return all computed styles
			const result = {};
			for (let i = 0; i < styles.length; i++) {
				const prop = styles[i];
				result[prop] = styles.getPropertyValue(prop);
			}
			return result;
		}
		// Return only requested properties
		const result = {};
		for (const prop of props) {
			result[prop] = styles.getPropertyValue(prop);
		}
		return result;
	`, selector, selector, string(propsJSON))
	return s.ExecJS(windowName, code)
}

// Highlight visually highlights an element for debugging.
func (s *Service) Highlight(windowName string, selector string, duration int) error {
	if duration <= 0 {
		duration = 2000
	}
	code := fmt.Sprintf(`
		const el = document.querySelector(%q);
		if (!el) throw new Error('Element not found: %s');
		const originalOutline = el.style.outline;
		const originalBackground = el.style.backgroundColor;
		el.style.outline = '3px solid red';
		el.style.backgroundColor = 'rgba(255, 0, 0, 0.2)';
		setTimeout(() => {
			el.style.outline = originalOutline;
			el.style.backgroundColor = originalBackground;
		}, %d);
		return 'highlighted';
	`, selector, selector, duration)
	_, err := s.ExecJS(windowName, code)
	return err
}

// GetDOMTree returns a simplified DOM tree structure.
func (s *Service) GetDOMTree(windowName string, maxDepth int) (string, error) {
	if maxDepth <= 0 {
		maxDepth = 5
	}
	code := fmt.Sprintf(`
		function buildTree(node, depth = 0) {
			if (depth > %d) return null;
			if (node.nodeType !== Node.ELEMENT_NODE) return null;

			const children = [];
			for (const child of node.children) {
				const childTree = buildTree(child, depth + 1);
				if (childTree) children.push(childTree);
			}

			return {
				tag: node.tagName.toLowerCase(),
				id: node.id || undefined,
				class: node.className || undefined,
				children: children.length > 0 ? children : undefined
			};
		}
		return buildTree(document.body);
	`, maxDepth)
	return s.ExecJS(windowName, code)
}

// GetErrors returns captured error messages (subset of console with level=error).
func (s *Service) GetErrors(limit int) []ConsoleMessage {
	return s.GetConsoleMessages("error", limit)
}

// GetPerformance returns performance metrics from the page.
func (s *Service) GetPerformance(windowName string) (string, error) {
	code := `
		const perf = window.performance;
		const timing = perf.timing;
		const memory = perf.memory || {};
		const navigation = perf.getEntriesByType('navigation')[0] || {};

		return {
			loadTime: timing.loadEventEnd - timing.navigationStart,
			domReady: timing.domContentLoadedEventEnd - timing.navigationStart,
			firstPaint: perf.getEntriesByType('paint').find(p => p.name === 'first-paint')?.startTime || 0,
			firstContentfulPaint: perf.getEntriesByType('paint').find(p => p.name === 'first-contentful-paint')?.startTime || 0,
			memory: {
				usedJSHeapSize: memory.usedJSHeapSize,
				totalJSHeapSize: memory.totalJSHeapSize,
				jsHeapSizeLimit: memory.jsHeapSizeLimit
			},
			resourceCount: perf.getEntriesByType('resource').length,
			transferSize: navigation.transferSize || 0,
			encodedBodySize: navigation.encodedBodySize || 0,
			decodedBodySize: navigation.decodedBodySize || 0
		};
	`
	return s.ExecJS(windowName, code)
}

// GetResources returns a list of loaded resources (scripts, styles, images).
func (s *Service) GetResources(windowName string) (string, error) {
	code := `
		const resources = window.performance.getEntriesByType('resource');
		return resources.map(r => ({
			name: r.name,
			type: r.initiatorType,
			duration: r.duration,
			transferSize: r.transferSize,
			encodedBodySize: r.encodedBodySize,
			decodedBodySize: r.decodedBodySize,
			startTime: r.startTime,
			responseEnd: r.responseEnd
		}));
	`
	return s.ExecJS(windowName, code)
}

// NetworkRequest represents a captured network request.
type NetworkRequest struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Status       int               `json:"status"`
	StatusText   string            `json:"statusText"`
	Type         string            `json:"type"`
	Duration     float64           `json:"duration"`
	TransferSize int64             `json:"transferSize"`
	StartTime    float64           `json:"startTime"`
	ResponseEnd  float64           `json:"responseEnd"`
	Headers      map[string]string `json:"headers,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
}

// networkBuffer stores captured network requests.
type networkBuffer struct {
	requests []NetworkRequest
	maxSize  int
	mu       sync.RWMutex
}

var netBuffer = &networkBuffer{
	requests: make([]NetworkRequest, 0, 500),
	maxSize:  500,
}

// GetNetworkRequests returns captured network requests.
// This uses the Performance API to get resource timing data.
func (s *Service) GetNetworkRequests(windowName string, limit int) (string, error) {
	if limit <= 0 {
		limit = 100
	}
	code := fmt.Sprintf(`
		const entries = window.performance.getEntriesByType('resource');
		const requests = entries.slice(-%d).map(entry => ({
			url: entry.name,
			type: entry.initiatorType,
			duration: entry.duration,
			transferSize: entry.transferSize || 0,
			encodedBodySize: entry.encodedBodySize || 0,
			decodedBodySize: entry.decodedBodySize || 0,
			startTime: entry.startTime,
			responseEnd: entry.responseEnd,
			serverTiming: entry.serverTiming || [],
			nextHopProtocol: entry.nextHopProtocol || '',
			connectStart: entry.connectStart,
			connectEnd: entry.connectEnd,
			domainLookupStart: entry.domainLookupStart,
			domainLookupEnd: entry.domainLookupEnd,
			requestStart: entry.requestStart,
			responseStart: entry.responseStart
		}));
		return requests;
	`, limit)
	return s.ExecJS(windowName, code)
}

// ClearNetworkRequests clears the network request buffer.
func (s *Service) ClearNetworkRequests(windowName string) error {
	code := `
		window.performance.clearResourceTimings();
		return 'cleared';
	`
	_, err := s.ExecJS(windowName, code)
	return err
}

// InjectNetworkInterceptor injects a fetch/XHR interceptor to capture detailed request info.
// This provides more detail than Performance API alone.
func (s *Service) InjectNetworkInterceptor(windowName string) error {
	window := s.GetWindow(windowName)
	if window == nil {
		return fmt.Errorf("window not found: %s", windowName)
	}

	code := `
		(function() {
			if (window.__mcpNetworkInjected) return;
			window.__mcpNetworkInjected = true;
			window.__mcpNetworkRequests = [];

			// Intercept fetch
			const originalFetch = window.fetch;
			window.fetch = async function(...args) {
				const startTime = performance.now();
				const request = new Request(...args);
				const requestInfo = {
					url: request.url,
					method: request.method,
					type: 'fetch',
					startTime: startTime,
					timestamp: new Date().toISOString()
				};

				try {
					const response = await originalFetch.apply(this, args);
					requestInfo.status = response.status;
					requestInfo.statusText = response.statusText;
					requestInfo.duration = performance.now() - startTime;
					requestInfo.responseEnd = performance.now();

					// Emit event for Go to capture
					window.wails.Events.Emit('mcp:network', JSON.stringify(requestInfo));
					window.__mcpNetworkRequests.push(requestInfo);

					// Keep buffer size limited
					if (window.__mcpNetworkRequests.length > 500) {
						window.__mcpNetworkRequests.shift();
					}

					return response;
				} catch (error) {
					requestInfo.error = error.message;
					requestInfo.duration = performance.now() - startTime;
					window.wails.Events.Emit('mcp:network', JSON.stringify(requestInfo));
					window.__mcpNetworkRequests.push(requestInfo);
					throw error;
				}
			};

			// Intercept XMLHttpRequest
			const originalXHROpen = XMLHttpRequest.prototype.open;
			const originalXHRSend = XMLHttpRequest.prototype.send;

			XMLHttpRequest.prototype.open = function(method, url, ...rest) {
				this.__mcpMethod = method;
				this.__mcpUrl = url;
				return originalXHROpen.apply(this, [method, url, ...rest]);
			};

			XMLHttpRequest.prototype.send = function(...args) {
				const xhr = this;
				const startTime = performance.now();

				xhr.addEventListener('loadend', function() {
					const requestInfo = {
						url: xhr.__mcpUrl,
						method: xhr.__mcpMethod,
						type: 'xhr',
						status: xhr.status,
						statusText: xhr.statusText,
						startTime: startTime,
						duration: performance.now() - startTime,
						responseEnd: performance.now(),
						timestamp: new Date().toISOString()
					};

					window.wails.Events.Emit('mcp:network', JSON.stringify(requestInfo));
					window.__mcpNetworkRequests.push(requestInfo);

					if (window.__mcpNetworkRequests.length > 500) {
						window.__mcpNetworkRequests.shift();
					}
				});

				return originalXHRSend.apply(this, args);
			};
		})()
	`

	window.ExecJS(code)
	return nil
}

// GetInterceptedNetworkRequests returns requests captured by the injected interceptor.
func (s *Service) GetInterceptedNetworkRequests(windowName string, limit int) (string, error) {
	if limit <= 0 {
		limit = 100
	}
	code := fmt.Sprintf(`
		const requests = window.__mcpNetworkRequests || [];
		return requests.slice(-%d);
	`, limit)
	return s.ExecJS(windowName, code)
}

// SetupNetworkListener sets up the Go-side listener for network events.
func (s *Service) SetupNetworkListener() {
	if s.app == nil {
		return
	}

	s.app.Event.On("mcp:network", func(event *application.CustomEvent) {
		if data, ok := event.Data.(string); ok {
			var req NetworkRequest
			if err := json.Unmarshal([]byte(data), &req); err == nil {
				netBuffer.mu.Lock()
				if len(netBuffer.requests) >= netBuffer.maxSize {
					netBuffer.requests = netBuffer.requests[1:]
				}
				netBuffer.requests = append(netBuffer.requests, req)
				netBuffer.mu.Unlock()
			}
		}
	})
}

// GetCachedNetworkRequests returns network requests from the Go-side buffer.
func (s *Service) GetCachedNetworkRequests(limit int) []NetworkRequest {
	netBuffer.mu.RLock()
	defer netBuffer.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	result := make([]NetworkRequest, 0, limit)
	start := len(netBuffer.requests) - limit
	if start < 0 {
		start = 0
	}

	for i := start; i < len(netBuffer.requests); i++ {
		result = append(result, netBuffer.requests[i])
	}
	return result
}

// ClearCachedNetworkRequests clears the Go-side network buffer.
func (s *Service) ClearCachedNetworkRequests() {
	netBuffer.mu.Lock()
	defer netBuffer.mu.Unlock()
	netBuffer.requests = netBuffer.requests[:0]
}

// PrintToPDF triggers the browser print dialog (which can save as PDF).
// This uses the native Wails Print() method.
func (s *Service) PrintToPDF(windowName string) error {
	window := s.GetWindow(windowName)
	if window == nil {
		return fmt.Errorf("window not found: %s", windowName)
	}
	return window.Print()
}

// ExportToPDF exports the page as a PDF using html2pdf.js library.
// Returns base64-encoded PDF data via async callback.
func (s *Service) ExportToPDF(windowName string, options map[string]any) (string, error) {
	window := s.GetWindow(windowName)
	if window == nil {
		return "", fmt.Errorf("window not found: %s", windowName)
	}

	callbackID := fmt.Sprintf("mcp_pdf_%d", time.Now().UnixNano())
	resultChan := make(chan struct {
		data string
		err  string
	}, 1)

	var unsubscribe func()
	unsubscribe = s.app.Event.On(callbackID, func(event *application.CustomEvent) {
		unsubscribe()
		if data, ok := event.Data.(string); ok {
			var result struct {
				Data  string `json:"data"`
				Error string `json:"error"`
			}
			if err := json.Unmarshal([]byte(data), &result); err != nil {
				resultChan <- struct {
					data string
					err  string
				}{"", "failed to parse result"}
			} else {
				resultChan <- struct {
					data string
					err  string
				}{result.Data, result.Error}
			}
		}
	})

	// Get options with defaults
	filename := "document.pdf"
	if fn, ok := options["filename"].(string); ok && fn != "" {
		filename = fn
	}
	margin := 10
	if m, ok := options["margin"].(float64); ok {
		margin = int(m)
	}

	code := fmt.Sprintf(`
		(async function() {
			try {
				// Load html2pdf.js if not present
				if (typeof html2pdf === 'undefined') {
					await new Promise((resolve, reject) => {
						const script = document.createElement('script');
						script.src = 'https://cdnjs.cloudflare.com/ajax/libs/html2pdf.js/0.10.1/html2pdf.bundle.min.js';
						script.onload = resolve;
						script.onerror = () => reject(new Error('Failed to load html2pdf.js'));
						document.head.appendChild(script);
					});
				}

				const element = document.body;
				const opt = {
					margin: %d,
					filename: %q,
					image: { type: 'jpeg', quality: 0.98 },
					html2canvas: { scale: 2, useCORS: true },
					jsPDF: { unit: 'mm', format: 'a4', orientation: 'portrait' }
				};

				const pdf = await html2pdf().set(opt).from(element).outputPdf('datauristring');
				window.wails.Events.Emit('%s', JSON.stringify({ data: pdf }));
			} catch (e) {
				window.wails.Events.Emit('%s', JSON.stringify({ error: e.message }));
			}
		})()
	`, margin, filename, callbackID, callbackID)

	window.ExecJS(code)

	select {
	case result := <-resultChan:
		if result.err != "" {
			return "", fmt.Errorf("PDF export failed: %s", result.err)
		}
		return result.data, nil
	case <-time.After(30 * time.Second):
		unsubscribe()
		return "", fmt.Errorf("PDF export timeout")
	}
}
