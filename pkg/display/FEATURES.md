# Display Server Features for Claude Code Integration

This document tracks the implementation of display server features that enable AI-assisted development workflows.

## Status Legend
- [ ] Not started
- [x] Complete
- [~] In progress

---

## Window Management

### Core Window Operations
- [x] `window_list` - List all windows with positions/sizes
- [x] `window_get` - Get specific window info
- [x] `window_position` - Move window to coordinates
- [x] `window_size` - Resize window
- [x] `window_bounds` - Set position + size in one call
- [x] `window_maximize` - Maximize window
- [x] `window_minimize` - Minimize window
- [x] `window_restore` - Restore from maximized/minimized
- [x] `window_focus` - Bring window to front
- [x] Window state persistence (remembers position across restarts)

### Extended Window Operations
- [x] `window_create` - Create new window at specific position with URL
- [x] `window_close` - Close a window by name
- [x] `window_visibility` - Show/hide window without closing
- [x] `window_title_set` - Change window title dynamically
- [x] `window_title_get` - Get current window title (returns window name)
- [x] `window_always_on_top` - Pin window above others
- [x] `window_background_colour` - Set window background color with alpha (transparency)
- [x] `window_fullscreen` - Enter/exit fullscreen mode

---

## Screen/Monitor Management

### Screen Information
- [x] `screen_list` - List all screens/monitors
- [x] `screen_get` - Get specific screen by ID
- [x] `screen_primary` - Get primary screen info
- [x] `screen_work_area` - Get usable area (excluding dock/menubar)
- [x] `screen_at_point` - Get screen containing a point
- [x] `screen_for_window` - Get screen a window is on

---

## Layout Management

### Layout Operations
- [x] `layout_save` - Save current window arrangement with a name
- [x] `layout_restore` - Restore a saved layout by name
- [x] `layout_list` - List saved layouts
- [x] `layout_delete` - Delete a saved layout
- [x] `layout_get` - Get details of a specific layout

### Smart Layout
- [x] `layout_tile` - Auto-tile windows (left/right/top/bottom/quadrants/grid)
- [x] `layout_stack` - Stack windows in cascade pattern
- [ ] `layout_beside_editor` - Position window beside detected IDE window
- [ ] `layout_suggest` - Given screen dimensions, suggest optimal arrangement
- [x] `layout_snap` - Snap window to screen edge/corner/center

### AI-Optimized Layout
- [ ] `screen_find_space` - Find empty screen space for new window
- [ ] `window_arrange_pair` - Put two windows side-by-side optimally
- [x] `layout_workflow` - Preset layouts: "coding", "debugging", "presenting", "side-by-side"

---

## WebView/Browser Features

### JavaScript Execution
- [x] `webview_eval` - Execute JavaScript and return result
- [x] `webview_list` - List all webview windows

### Console & Errors
- [x] `webview_console` - Get console messages (log, warn, error, info)
- [x] `webview_errors` - Get structured JS errors with stack traces
- [x] `webview_clear_console` - Clear console buffer

### DOM Inspection
- [x] `webview_query` - Query elements by CSS selector
- [x] `webview_dom_tree` - Get full DOM tree structure
- [x] `webview_element_info` - Get detailed info about an element
- [x] `webview_highlight` - Visually highlight an element (debugging)
- [x] `webview_computed_style` - Get computed styles for element

### Interaction
- [x] `webview_click` - Click element by selector
- [x] `webview_type` - Type into element
- [x] `webview_navigate` - Navigate to URL/route
- [x] `webview_scroll` - Scroll to element or position
- [x] `webview_hover` - Hover over element
- [x] `webview_select` - Select option in dropdown
- [x] `webview_check` - Check/uncheck checkbox

### Page Information
- [x] `webview_source` - Get page HTML source
- [x] `webview_url` - Get current URL
- [x] `webview_title` - Get page title
- [x] `webview_screenshot` - Capture rendered page as image
- [x] `webview_screenshot_element` - Capture specific element as image
- [x] `webview_pdf` - Export page as PDF (using html2pdf.js)
- [x] `webview_print` - Open native print dialog

### Network & Performance
- [x] `webview_network` - Get network requests log (via Performance API)
- [x] `webview_network_clear` - Clear network log
- [x] `webview_network_inject` - Inject fetch/XHR interceptor for detailed logging
- [x] `webview_performance` - Get performance metrics (load time, memory)
- [x] `webview_resources` - List loaded resources (scripts, styles, images)

### DevTools
- [ ] `webview_devtools_open` - Open DevTools for window
- [ ] `webview_devtools_close` - Close DevTools

---

## System Integration

### Clipboard
- [x] `clipboard_read` - Read clipboard text content
- [x] `clipboard_write` - Write text to clipboard
- [ ] `clipboard_read_image` - Read image from clipboard
- [ ] `clipboard_write_image` - Write image to clipboard
- [x] `clipboard_has` - Check clipboard content type
- [x] `clipboard_clear` - Clear clipboard contents

### Notifications
- [x] `notification_show` - Show native system notification (macOS/Windows/Linux)
- [x] `notification_permission_request` - Request notification permission
- [x] `notification_permission_check` - Check notification authorization status
- [ ] `notification_clear` - Clear notifications
- [ ] `notification_with_actions` - Interactive notifications with buttons

### Dialogs
- [x] `dialog_open_file` - Show file open dialog
- [x] `dialog_save_file` - Show file save dialog
- [x] `dialog_open_directory` - Show directory picker
- [x] `dialog_message` - Show message dialog (info/warning/error) (via notification_show)
- [x] `dialog_confirm` - Show confirmation dialog
- [~] `dialog_prompt` - Show input prompt dialog (not supported natively in Wails v3)

### Theme & Appearance
- [x] `theme_get` - Get current theme (dark/light)
- [ ] `theme_set` - Set application theme
- [x] `theme_system` - Get system theme preference
- [x] `theme_on_change` - Subscribe to theme changes (via WebSocket events)

---

## Focus & Events

### Focus Management
- [x] `window_focused` - Get currently focused window
- [x] `focus_set` - Set focus to specific window (alias for window_focus)

### Event Subscriptions (WebSocket)
- [x] `event_subscribe` - Subscribe to events (via WebSocket /events endpoint)
- [x] `event_unsubscribe` - Unsubscribe from events
- [x] `event_info` - Get WebSocket event server info
- [x] Events: `window.focus`, `window.blur`, `window.move`, `window.resize`, `window.close`, `window.create`, `theme.change`

---

## System Tray

- [x] `tray_set_icon` - Set tray icon (base64 PNG)
- [x] `tray_set_tooltip` - Set tray tooltip
- [x] `tray_set_label` - Set tray label text
- [x] `tray_set_menu` - Set tray menu items (with nested submenus)
- [x] `tray_info` - Get tray status info
- [ ] `tray_show_message` - Show tray balloon notification

---

## Implementation Priority

### Phase 1 - Core Display Server (DONE)
- [x] Window list/get/position/size/bounds
- [x] Window maximize/minimize/restore/focus
- [x] Window state persistence
- [x] HTTP REST bridge for tools

### Phase 2 - Enhanced Windows (DONE)
- [x] window_create, window_close
- [x] window_visibility, window_always_on_top
- [x] screen_work_area, window_fullscreen, window_title

### Phase 3 - Layouts (DONE)
- [x] layout_save, layout_restore, layout_list
- [x] layout_delete, layout_get
- [ ] layout_tile, layout_beside_editor (future)

### Phase 4 - WebView Debug (DONE)
- [x] webview_screenshot, webview_screenshot_element
- [x] webview_url, webview_source, webview_title
- [x] webview_dom_tree, webview_element_info, webview_computed_style
- [x] webview_scroll, webview_hover, webview_select, webview_check
- [x] webview_highlight, webview_errors
- [x] webview_performance, webview_resources
- [ ] webview_network, webview_devtools (future)

### Phase 5 - System Integration (DONE)
- [x] clipboard_read, clipboard_write, clipboard_has, clipboard_clear
- [x] notification_show (native + dialog fallback)
- [x] notification_permission_request, notification_permission_check
- [x] dialog_open_file, dialog_save_file, dialog_open_directory, dialog_confirm
- [x] theme_get, theme_system

### Phase 6 - Events & Real-time (DONE)
- [x] WebSocket event subscriptions (/events endpoint)
- [x] Real-time window tracking (focus, blur, move, resize, close, create)
- [x] Theme change events
- [x] focus_set, screen_get, screen_primary, screen_at_point, screen_for_window

### Phase 7 - Advanced Features (DONE)
- [x] `window_background_colour` - Window transparency via RGBA alpha
- [x] `layout_tile` - Auto-tile windows in grid/halves/quadrants
- [x] `layout_snap` - Snap windows to edges/corners/center
- [x] `layout_stack` - Cascade windows in stacked pattern
- [x] `layout_workflow` - Preset layouts (coding/debugging/presenting)
- [x] `webview_network` - Network request logging
- [x] `webview_network_clear` - Clear network log
- [x] `webview_network_inject` - Detailed fetch/XHR interceptor
- [x] `webview_pdf` - Export page as PDF
- [x] `webview_print` - Native print dialog
- [x] `tray_set_icon` - Set tray icon dynamically
- [x] `tray_set_tooltip` - Set tray tooltip
- [x] `tray_set_label` - Set tray label
- [x] `tray_set_menu` - Set tray menu items
- [x] `tray_info` - Get tray status

### Phase 8 - Remaining Features (Future)
- [ ] window_opacity (true opacity if Wails adds support)
- [ ] layout_beside_editor, layout_suggest
- [ ] webview_devtools_open, webview_devtools_close
- [ ] clipboard_read_image, clipboard_write_image
- [ ] notification_with_actions, notification_clear
- [ ] tray_show_message - Balloon notifications
