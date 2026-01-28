package display

import (
	"context"
	"fmt"

	"github.com/host-uk/core/pkg/core"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

// Options holds configuration for the display service.
// This struct is used to configure the display service at startup.
type Options struct{}

// Service manages windowing, dialogs, and other visual elements.
// It is the primary interface for interacting with the UI.
type Service struct {
	*core.ServiceRuntime[Options]
	app          App
	config       Options
	windowStates *WindowStateManager
	layouts      *LayoutManager
	notifier     *notifications.NotificationService
	events       *WSEventManager
}

// newDisplayService contains the common logic for initializing a Service struct.
// It is called by the New function.
func newDisplayService() (*Service, error) {
	return &Service{}, nil
}

// New is the constructor for the display service.
// It creates a new Service and returns it.
//
// example:
//
//	displayService, err := display.New()
//	if err != nil {
//		log.Fatal(err)
//	}
func New() (*Service, error) {
	s, err := newDisplayService()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Register creates and registers a new display service with the given Core instance.
// This wires up the ServiceRuntime so the service can access other services.
func Register(c *core.Core) (any, error) {
	s, err := New()
	if err != nil {
		return nil, err
	}
	s.ServiceRuntime = core.NewServiceRuntime[Options](c, Options{})
	return s, nil
}

// ServiceName returns the canonical name for this service.
func (s *Service) ServiceName() string {
	return "github.com/host-uk/core/display"
}

// ServiceStartup is called by Wails when the app starts. It initializes the display service
// and sets up the main application window and system tray.
func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return s.Startup(ctx)
}

// Startup is called when the app starts. It initializes the display service
// and sets up the main application window and system tray.
//
//	err := displayService.Startup(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
func (s *Service) Startup(ctx context.Context) error {
	s.app = newWailsApp(application.Get())
	s.windowStates = NewWindowStateManager()
	s.layouts = NewLayoutManager()
	s.events = NewWSEventManager(s)
	s.events.SetupWindowEventListeners()
	s.app.Logger().Info("Display service started")
	s.buildMenu()
	s.systemTray()
	return s.OpenWindow()
}

// handleOpenWindowAction processes a message to configure and create a new window
// using the specified name and options.
func (s *Service) handleOpenWindowAction(msg map[string]any) error {
	opts := parseWindowOptions(msg)
	s.app.Window().NewWithOptions(opts)
	return nil
}

// parseWindowOptions extracts window configuration from a map and returns it
// as a `application.WebviewWindowOptions` struct. This function is used by
// `handleOpenWindowAction` to parse the incoming message.
func parseWindowOptions(msg map[string]any) application.WebviewWindowOptions {
	opts := application.WebviewWindowOptions{}
	if name, ok := msg["name"].(string); ok {
		opts.Name = name
	}
	if optsMap, ok := msg["options"].(map[string]any); ok {
		if title, ok := optsMap["Title"].(string); ok {
			opts.Title = title
		}
		if width, ok := optsMap["Width"].(float64); ok {
			opts.Width = int(width)
		}
		if height, ok := optsMap["Height"].(float64); ok {
			opts.Height = int(height)
		}
	}
	return opts
}

// ShowEnvironmentDialog displays a dialog containing detailed information about
// the application's runtime environment. This is useful for debugging and
// understanding the context in which the application is running.
//
// example:
//
//	displayService.ShowEnvironmentDialog()
func (s *Service) ShowEnvironmentDialog() {
	envInfo := s.app.Env().Info()

	details := "Environment Information:\n\n"
	details += fmt.Sprintf("Operating System: %s\n", envInfo.OS)
	details += fmt.Sprintf("Architecture: %s\n", envInfo.Arch)
	details += fmt.Sprintf("Debug Mode: %t\n\n", envInfo.Debug)
	details += fmt.Sprintf("Dark Mode: %t\n\n", s.app.Env().IsDarkMode())
	details += "Platform Information:"

	// Add platform-specific details
	for key, value := range envInfo.PlatformInfo {
		details += fmt.Sprintf("\n%s: %v", key, value)
	}

	if envInfo.OSInfo != nil {
		details += fmt.Sprintf("\n\nOS Details:\nName: %s\nVersion: %s",
			envInfo.OSInfo.Name,
			envInfo.OSInfo.Version)
	}

	dialog := s.app.Dialog().Info()
	dialog.SetTitle("Environment Information")
	dialog.SetMessage(details)
	dialog.Show()
}

// OpenWindow creates a new window with the given options. If no options are
// provided, it will use the default options.
//
// example:
//
//	err := displayService.OpenWindow(
//		display.WithName("my-window"),
//		display.WithTitle("My Window"),
//		display.WithWidth(800),
//		display.WithHeight(600),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
func (s *Service) OpenWindow(opts ...WindowOption) error {
	wailsOpts := buildWailsWindowOptions(opts...)

	// Apply saved window state (position, size)
	if s.windowStates != nil {
		wailsOpts = s.windowStates.ApplyState(wailsOpts)
	}

	window := s.app.Window().NewWithOptions(wailsOpts)

	// Set up state tracking for this window
	if s.windowStates != nil && window != nil {
		s.trackWindowState(wailsOpts.Name, window)
	}

	return nil
}

// trackWindowState sets up event listeners to track window position/size changes.
func (s *Service) trackWindowState(name string, window *application.WebviewWindow) {
	// Register for window events
	window.OnWindowEvent(events.Common.WindowDidMove, func(event *application.WindowEvent) {
		s.windowStates.CaptureState(name, window)
	})

	window.OnWindowEvent(events.Common.WindowDidResize, func(event *application.WindowEvent) {
		s.windowStates.CaptureState(name, window)
	})

	// Attach event manager listeners for WebSocket broadcasts
	if s.events != nil {
		s.events.AttachWindowListeners(window)
		// Emit window create event
		s.events.EmitWindowEvent(EventWindowCreate, name, map[string]any{
			"name": name,
		})
	}

	// Capture initial state
	s.windowStates.CaptureState(name, window)
}

// buildWailsWindowOptions creates Wails window options from the given
// `WindowOption`s. This function is used by `OpenWindow` to construct the
// options for the new window.
func buildWailsWindowOptions(opts ...WindowOption) application.WebviewWindowOptions {
	// Default options
	winOpts := &Window{
		Name:   "main",
		Title:  "Core",
		Width:  1280,
		Height: 800,
		URL:    "/",
	}

	// Apply functional options
	for _, opt := range opts {
		if opt != nil {
			_ = opt(winOpts)
		}
	}

	return *winOpts
}

// monitorScreenChanges listens for theme change events and logs when the screen
// configuration changes.
func (s *Service) monitorScreenChanges() {
	s.app.Event().OnApplicationEvent(events.Common.ThemeChanged, func(event *application.ApplicationEvent) {
		s.app.Logger().Info("Screen configuration changed")
	})
}

// WindowInfo contains information about a window for MCP.
type WindowInfo struct {
	Name      string `json:"name"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Maximized bool   `json:"maximized"`
}

// ScreenInfo contains information about a display screen.
type ScreenInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Primary bool   `json:"primary"`
}

// GetWindowInfo returns information about a window by name.
func (s *Service) GetWindowInfo(name string) (*WindowInfo, error) {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				x, y := wv.Position()
				width, height := wv.Size()
				return &WindowInfo{
					Name:      name,
					X:         x,
					Y:         y,
					Width:     width,
					Height:    height,
					Maximized: wv.IsMaximised(),
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("window not found: %s", name)
}

// ListWindowInfos returns information about all windows.
func (s *Service) ListWindowInfos() []WindowInfo {
	windows := s.app.Window().GetAll()
	result := make([]WindowInfo, 0, len(windows))
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			x, y := wv.Position()
			width, height := wv.Size()
			result = append(result, WindowInfo{
				Name:      wv.Name(),
				X:         x,
				Y:         y,
				Width:     width,
				Height:    height,
				Maximized: wv.IsMaximised(),
			})
		}
	}
	return result
}

// GetScreens returns information about all available screens.
func (s *Service) GetScreens() []ScreenInfo {
	app := application.Get()
	if app == nil || app.Screen == nil {
		return nil
	}

	screens := app.Screen.GetAll()
	if screens == nil {
		return nil
	}

	result := make([]ScreenInfo, 0, len(screens))
	for _, screen := range screens {
		result = append(result, ScreenInfo{
			ID:      screen.ID,
			Name:    screen.Name,
			X:       screen.Bounds.X,
			Y:       screen.Bounds.Y,
			Width:   screen.Bounds.Width,
			Height:  screen.Bounds.Height,
			Primary: screen.IsPrimary,
		})
	}
	return result
}

// SetWindowPosition moves a window to the specified position.
func (s *Service) SetWindowPosition(name string, x, y int) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.SetPosition(x, y)
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// SetWindowSize resizes a window.
func (s *Service) SetWindowSize(name string, width, height int) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.SetSize(width, height)
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// SetWindowBounds sets both position and size of a window.
func (s *Service) SetWindowBounds(name string, x, y, width, height int) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.SetPosition(x, y)
				wv.SetSize(width, height)
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// MaximizeWindow maximizes a window.
func (s *Service) MaximizeWindow(name string) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.Maximise()
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// RestoreWindow restores a maximized/minimized window.
func (s *Service) RestoreWindow(name string) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.Restore()
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// MinimizeWindow minimizes a window.
func (s *Service) MinimizeWindow(name string) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.Minimise()
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// FocusWindow brings a window to the front.
func (s *Service) FocusWindow(name string) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.Focus()
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// ResetWindowState clears saved window positions.
func (s *Service) ResetWindowState() error {
	if s.windowStates != nil {
		return s.windowStates.Clear()
	}
	return nil
}

// GetSavedWindowStates returns all saved window states.
func (s *Service) GetSavedWindowStates() map[string]*WindowState {
	if s.windowStates == nil {
		return nil
	}

	result := make(map[string]*WindowState)
	for _, name := range s.windowStates.ListStates() {
		result[name] = s.windowStates.GetState(name)
	}
	return result
}

// CreateWindowOptions contains options for creating a new window.
type CreateWindowOptions struct {
	Name   string `json:"name"`
	Title  string `json:"title,omitempty"`
	URL    string `json:"url,omitempty"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// CreateWindow creates a new window with the specified options.
func (s *Service) CreateWindow(opts CreateWindowOptions) (*WindowInfo, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("window name is required")
	}

	// Set defaults
	if opts.Width == 0 {
		opts.Width = 800
	}
	if opts.Height == 0 {
		opts.Height = 600
	}
	if opts.URL == "" {
		opts.URL = "/"
	}
	if opts.Title == "" {
		opts.Title = opts.Name
	}

	wailsOpts := application.WebviewWindowOptions{
		Name:   opts.Name,
		Title:  opts.Title,
		URL:    opts.URL,
		Width:  opts.Width,
		Height: opts.Height,
		X:      opts.X,
		Y:      opts.Y,
	}

	window := s.app.Window().NewWithOptions(wailsOpts)
	if window == nil {
		return nil, fmt.Errorf("failed to create window")
	}

	// Track window state
	if s.windowStates != nil {
		s.trackWindowState(opts.Name, window)
	}

	return &WindowInfo{
		Name:   opts.Name,
		X:      opts.X,
		Y:      opts.Y,
		Width:  opts.Width,
		Height: opts.Height,
	}, nil
}

// CloseWindow closes a window by name.
func (s *Service) CloseWindow(name string) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.Close()
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// SetWindowVisibility shows or hides a window.
func (s *Service) SetWindowVisibility(name string, visible bool) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				if visible {
					wv.Show()
				} else {
					wv.Hide()
				}
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// SetWindowAlwaysOnTop sets whether a window stays on top of other windows.
func (s *Service) SetWindowAlwaysOnTop(name string, alwaysOnTop bool) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.SetAlwaysOnTop(alwaysOnTop)
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// SetWindowTitle changes a window's title.
func (s *Service) SetWindowTitle(name string, title string) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.SetTitle(title)
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// SetWindowFullscreen sets a window to fullscreen mode.
func (s *Service) SetWindowFullscreen(name string, fullscreen bool) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				if fullscreen {
					wv.Fullscreen()
				} else {
					wv.UnFullscreen()
				}
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// WorkArea represents usable screen space (excluding dock, menubar, etc).
type WorkArea struct {
	ScreenID string `json:"screenId"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// GetWorkAreas returns the usable work area for all screens.
func (s *Service) GetWorkAreas() []WorkArea {
	app := application.Get()
	if app == nil || app.Screen == nil {
		return nil
	}

	screens := app.Screen.GetAll()
	if screens == nil {
		return nil
	}

	result := make([]WorkArea, 0, len(screens))
	for _, screen := range screens {
		result = append(result, WorkArea{
			ScreenID: screen.ID,
			X:        screen.WorkArea.X,
			Y:        screen.WorkArea.Y,
			Width:    screen.WorkArea.Width,
			Height:   screen.WorkArea.Height,
		})
	}
	return result
}

// GetFocusedWindow returns the name of the currently focused window, or empty if none.
func (s *Service) GetFocusedWindow() string {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.IsFocused() {
				return wv.Name()
			}
		}
	}
	return ""
}

// SaveLayout saves the current window arrangement as a named layout.
func (s *Service) SaveLayout(name string) error {
	if s.layouts == nil {
		return fmt.Errorf("layout manager not initialized")
	}

	// Capture current window states
	windowStates := make(map[string]WindowState)
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			x, y := wv.Position()
			width, height := wv.Size()
			windowStates[wv.Name()] = WindowState{
				X:         x,
				Y:         y,
				Width:     width,
				Height:    height,
				Maximized: wv.IsMaximised(),
			}
		}
	}

	return s.layouts.SaveLayout(name, windowStates)
}

// RestoreLayout applies a saved layout, positioning all windows.
func (s *Service) RestoreLayout(name string) error {
	if s.layouts == nil {
		return fmt.Errorf("layout manager not initialized")
	}

	layout := s.layouts.GetLayout(name)
	if layout == nil {
		return fmt.Errorf("layout not found: %s", name)
	}

	// Apply saved positions to existing windows
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if state, exists := layout.Windows[wv.Name()]; exists {
				wv.SetPosition(state.X, state.Y)
				wv.SetSize(state.Width, state.Height)
				if state.Maximized {
					wv.Maximise()
				} else {
					wv.Restore()
				}
			}
		}
	}

	return nil
}

// ListLayouts returns all saved layout names with metadata.
func (s *Service) ListLayouts() []LayoutInfo {
	if s.layouts == nil {
		return nil
	}
	return s.layouts.ListLayouts()
}

// DeleteLayout removes a saved layout by name.
func (s *Service) DeleteLayout(name string) error {
	if s.layouts == nil {
		return fmt.Errorf("layout manager not initialized")
	}
	return s.layouts.DeleteLayout(name)
}

// GetLayout returns a specific layout by name.
func (s *Service) GetLayout(name string) *Layout {
	if s.layouts == nil {
		return nil
	}
	return s.layouts.GetLayout(name)
}

// GetEventManager returns the event manager for WebSocket event subscriptions.
func (s *Service) GetEventManager() *WSEventManager {
	return s.events
}

// GetWindowTitle returns the title of a window by name.
// Note: Wails v3 doesn't expose a title getter, so we track it ourselves or return the name.
func (s *Service) GetWindowTitle(name string) (string, error) {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				// Window name as fallback since Wails v3 doesn't have a title getter
				return name, nil
			}
		}
	}
	return "", fmt.Errorf("window not found: %s", name)
}

// GetScreen returns information about a specific screen by ID.
func (s *Service) GetScreen(id string) (*ScreenInfo, error) {
	app := application.Get()
	if app == nil || app.Screen == nil {
		return nil, fmt.Errorf("screen service not available")
	}

	screens := app.Screen.GetAll()
	for _, screen := range screens {
		if screen.ID == id {
			return &ScreenInfo{
				ID:      screen.ID,
				Name:    screen.Name,
				X:       screen.Bounds.X,
				Y:       screen.Bounds.Y,
				Width:   screen.Bounds.Width,
				Height:  screen.Bounds.Height,
				Primary: screen.IsPrimary,
			}, nil
		}
	}
	return nil, fmt.Errorf("screen not found: %s", id)
}

// GetPrimaryScreen returns information about the primary screen.
func (s *Service) GetPrimaryScreen() (*ScreenInfo, error) {
	app := application.Get()
	if app == nil || app.Screen == nil {
		return nil, fmt.Errorf("screen service not available")
	}

	screens := app.Screen.GetAll()
	for _, screen := range screens {
		if screen.IsPrimary {
			return &ScreenInfo{
				ID:      screen.ID,
				Name:    screen.Name,
				X:       screen.Bounds.X,
				Y:       screen.Bounds.Y,
				Width:   screen.Bounds.Width,
				Height:  screen.Bounds.Height,
				Primary: true,
			}, nil
		}
	}
	return nil, fmt.Errorf("no primary screen found")
}

// GetScreenAtPoint returns the screen containing a specific point.
func (s *Service) GetScreenAtPoint(x, y int) (*ScreenInfo, error) {
	app := application.Get()
	if app == nil || app.Screen == nil {
		return nil, fmt.Errorf("screen service not available")
	}

	screens := app.Screen.GetAll()
	for _, screen := range screens {
		bounds := screen.Bounds
		if x >= bounds.X && x < bounds.X+bounds.Width &&
			y >= bounds.Y && y < bounds.Y+bounds.Height {
			return &ScreenInfo{
				ID:      screen.ID,
				Name:    screen.Name,
				X:       bounds.X,
				Y:       bounds.Y,
				Width:   bounds.Width,
				Height:  bounds.Height,
				Primary: screen.IsPrimary,
			}, nil
		}
	}
	return nil, fmt.Errorf("no screen found at point (%d, %d)", x, y)
}

// GetScreenForWindow returns the screen containing a specific window.
func (s *Service) GetScreenForWindow(name string) (*ScreenInfo, error) {
	// Get window position
	info, err := s.GetWindowInfo(name)
	if err != nil {
		return nil, err
	}

	// Find screen at window center
	centerX := info.X + info.Width/2
	centerY := info.Y + info.Height/2

	return s.GetScreenAtPoint(centerX, centerY)
}

// SetWindowBackgroundColour sets the background color of a window with alpha for transparency.
// Note: On Windows, only alpha 0 or 255 are supported. Other values treated as 255.
func (s *Service) SetWindowBackgroundColour(name string, r, g, b, a uint8) error {
	windows := s.app.Window().GetAll()
	for _, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if wv.Name() == name {
				wv.SetBackgroundColour(application.RGBA{Red: r, Green: g, Blue: b, Alpha: a})
				return nil
			}
		}
	}
	return fmt.Errorf("window not found: %s", name)
}

// TileMode represents different tiling arrangements.
type TileMode string

const (
	TileModeLeft        TileMode = "left"
	TileModeRight       TileMode = "right"
	TileModeTop         TileMode = "top"
	TileModeBottom      TileMode = "bottom"
	TileModeTopLeft     TileMode = "top-left"
	TileModeTopRight    TileMode = "top-right"
	TileModeBottomLeft  TileMode = "bottom-left"
	TileModeBottomRight TileMode = "bottom-right"
	TileModeGrid        TileMode = "grid"
)

// TileWindows arranges windows in a tiled layout.
// mode can be: left, right, top, bottom, top-left, top-right, bottom-left, bottom-right, grid
// If windowNames is empty, tiles all windows.
func (s *Service) TileWindows(mode TileMode, windowNames []string) error {
	// Get work area for primary screen
	workAreas := s.GetWorkAreas()
	if len(workAreas) == 0 {
		return fmt.Errorf("no work areas available")
	}
	wa := workAreas[0] // Use primary screen work area

	// Get windows to tile
	allWindows := s.app.Window().GetAll()
	var windowsToTile []*application.WebviewWindow

	if len(windowNames) == 0 {
		// Tile all windows
		for _, w := range allWindows {
			if wv, ok := w.(*application.WebviewWindow); ok {
				windowsToTile = append(windowsToTile, wv)
			}
		}
	} else {
		// Tile specific windows
		nameSet := make(map[string]bool)
		for _, name := range windowNames {
			nameSet[name] = true
		}
		for _, w := range allWindows {
			if wv, ok := w.(*application.WebviewWindow); ok {
				if nameSet[wv.Name()] {
					windowsToTile = append(windowsToTile, wv)
				}
			}
		}
	}

	if len(windowsToTile) == 0 {
		return fmt.Errorf("no windows to tile")
	}

	switch mode {
	case TileModeLeft:
		// All windows on left half
		for _, wv := range windowsToTile {
			wv.SetPosition(wa.X, wa.Y)
			wv.SetSize(wa.Width/2, wa.Height)
		}

	case TileModeRight:
		// All windows on right half
		for _, wv := range windowsToTile {
			wv.SetPosition(wa.X+wa.Width/2, wa.Y)
			wv.SetSize(wa.Width/2, wa.Height)
		}

	case TileModeTop:
		// All windows on top half
		for _, wv := range windowsToTile {
			wv.SetPosition(wa.X, wa.Y)
			wv.SetSize(wa.Width, wa.Height/2)
		}

	case TileModeBottom:
		// All windows on bottom half
		for _, wv := range windowsToTile {
			wv.SetPosition(wa.X, wa.Y+wa.Height/2)
			wv.SetSize(wa.Width, wa.Height/2)
		}

	case TileModeTopLeft:
		for _, wv := range windowsToTile {
			wv.SetPosition(wa.X, wa.Y)
			wv.SetSize(wa.Width/2, wa.Height/2)
		}

	case TileModeTopRight:
		for _, wv := range windowsToTile {
			wv.SetPosition(wa.X+wa.Width/2, wa.Y)
			wv.SetSize(wa.Width/2, wa.Height/2)
		}

	case TileModeBottomLeft:
		for _, wv := range windowsToTile {
			wv.SetPosition(wa.X, wa.Y+wa.Height/2)
			wv.SetSize(wa.Width/2, wa.Height/2)
		}

	case TileModeBottomRight:
		for _, wv := range windowsToTile {
			wv.SetPosition(wa.X+wa.Width/2, wa.Y+wa.Height/2)
			wv.SetSize(wa.Width/2, wa.Height/2)
		}

	case TileModeGrid:
		// Arrange in a grid
		count := len(windowsToTile)
		cols := 1
		rows := 1
		// Calculate optimal grid
		for cols*rows < count {
			if cols <= rows {
				cols++
			} else {
				rows++
			}
		}

		cellWidth := wa.Width / cols
		cellHeight := wa.Height / rows

		for i, wv := range windowsToTile {
			col := i % cols
			row := i / cols
			wv.SetPosition(wa.X+col*cellWidth, wa.Y+row*cellHeight)
			wv.SetSize(cellWidth, cellHeight)
		}

	default:
		return fmt.Errorf("unknown tile mode: %s", mode)
	}

	return nil
}

// SnapPosition represents positions for snapping windows.
type SnapPosition string

const (
	SnapLeft        SnapPosition = "left"
	SnapRight       SnapPosition = "right"
	SnapTop         SnapPosition = "top"
	SnapBottom      SnapPosition = "bottom"
	SnapTopLeft     SnapPosition = "top-left"
	SnapTopRight    SnapPosition = "top-right"
	SnapBottomLeft  SnapPosition = "bottom-left"
	SnapBottomRight SnapPosition = "bottom-right"
	SnapCenter      SnapPosition = "center"
)

// SnapWindow snaps a window to a screen edge or corner.
func (s *Service) SnapWindow(name string, position SnapPosition) error {
	// Get window
	window, err := s.GetWindowInfo(name)
	if err != nil {
		return err
	}

	// Get screen for window
	screen, err := s.GetScreenForWindow(name)
	if err != nil {
		return err
	}

	// Get work area for this screen
	workAreas := s.GetWorkAreas()
	var wa *WorkArea
	for _, area := range workAreas {
		if area.ScreenID == screen.ID {
			wa = &area
			break
		}
	}
	if wa == nil {
		// Fallback to screen bounds
		wa = &WorkArea{
			ScreenID: screen.ID,
			X:        screen.X,
			Y:        screen.Y,
			Width:    screen.Width,
			Height:   screen.Height,
		}
	}

	// Calculate position based on snap position
	var x, y, width, height int

	switch position {
	case SnapLeft:
		x = wa.X
		y = wa.Y
		width = wa.Width / 2
		height = wa.Height

	case SnapRight:
		x = wa.X + wa.Width/2
		y = wa.Y
		width = wa.Width / 2
		height = wa.Height

	case SnapTop:
		x = wa.X
		y = wa.Y
		width = wa.Width
		height = wa.Height / 2

	case SnapBottom:
		x = wa.X
		y = wa.Y + wa.Height/2
		width = wa.Width
		height = wa.Height / 2

	case SnapTopLeft:
		x = wa.X
		y = wa.Y
		width = wa.Width / 2
		height = wa.Height / 2

	case SnapTopRight:
		x = wa.X + wa.Width/2
		y = wa.Y
		width = wa.Width / 2
		height = wa.Height / 2

	case SnapBottomLeft:
		x = wa.X
		y = wa.Y + wa.Height/2
		width = wa.Width / 2
		height = wa.Height / 2

	case SnapBottomRight:
		x = wa.X + wa.Width/2
		y = wa.Y + wa.Height/2
		width = wa.Width / 2
		height = wa.Height / 2

	case SnapCenter:
		// Center the window without resizing
		x = wa.X + (wa.Width-window.Width)/2
		y = wa.Y + (wa.Height-window.Height)/2
		width = window.Width
		height = window.Height

	default:
		return fmt.Errorf("unknown snap position: %s", position)
	}

	return s.SetWindowBounds(name, x, y, width, height)
}

// StackWindows arranges windows in a cascade (stacked) pattern.
// Each window is offset by the given amount from the previous one.
func (s *Service) StackWindows(windowNames []string, offsetX, offsetY int) error {
	if offsetX == 0 {
		offsetX = 30
	}
	if offsetY == 0 {
		offsetY = 30
	}

	// Get work area for primary screen
	workAreas := s.GetWorkAreas()
	if len(workAreas) == 0 {
		return fmt.Errorf("no work areas available")
	}
	wa := workAreas[0]

	// Get windows to stack
	allWindows := s.app.Window().GetAll()
	var windowsToStack []*application.WebviewWindow

	if len(windowNames) == 0 {
		for _, w := range allWindows {
			if wv, ok := w.(*application.WebviewWindow); ok {
				windowsToStack = append(windowsToStack, wv)
			}
		}
	} else {
		nameSet := make(map[string]bool)
		for _, name := range windowNames {
			nameSet[name] = true
		}
		for _, w := range allWindows {
			if wv, ok := w.(*application.WebviewWindow); ok {
				if nameSet[wv.Name()] {
					windowsToStack = append(windowsToStack, wv)
				}
			}
		}
	}

	if len(windowsToStack) == 0 {
		return fmt.Errorf("no windows to stack")
	}

	// Calculate window size (leave room for cascade)
	maxOffset := (len(windowsToStack) - 1) * offsetX
	windowWidth := wa.Width - maxOffset - 50
	maxOffsetY := (len(windowsToStack) - 1) * offsetY
	windowHeight := wa.Height - maxOffsetY - 50

	// Ensure minimum size
	if windowWidth < 400 {
		windowWidth = 400
	}
	if windowHeight < 300 {
		windowHeight = 300
	}

	// Position each window
	for i, wv := range windowsToStack {
		x := wa.X + (i * offsetX)
		y := wa.Y + (i * offsetY)
		wv.SetPosition(x, y)
		wv.SetSize(windowWidth, windowHeight)
		wv.Focus() // Bring to front in order
	}

	return nil
}

// WorkflowType represents predefined workflow layouts.
type WorkflowType string

const (
	WorkflowCoding     WorkflowType = "coding"
	WorkflowDebugging  WorkflowType = "debugging"
	WorkflowPresenting WorkflowType = "presenting"
	WorkflowSideBySide WorkflowType = "side-by-side"
)

// ApplyWorkflowLayout applies a predefined layout for a specific workflow.
func (s *Service) ApplyWorkflowLayout(workflow WorkflowType) error {
	switch workflow {
	case WorkflowCoding:
		// Main editor takes 70% left, tools on right 30%
		return s.applyWorkflowCoding()

	case WorkflowDebugging:
		// Code on top 60%, debug output on bottom 40%
		return s.applyWorkflowDebugging()

	case WorkflowPresenting:
		// Single window maximized
		return s.applyWorkflowPresenting()

	case WorkflowSideBySide:
		// Two windows side by side 50/50
		return s.TileWindows(TileModeGrid, nil)

	default:
		return fmt.Errorf("unknown workflow: %s", workflow)
	}
}

func (s *Service) applyWorkflowCoding() error {
	workAreas := s.GetWorkAreas()
	if len(workAreas) == 0 {
		return fmt.Errorf("no work areas available")
	}
	wa := workAreas[0]

	windows := s.app.Window().GetAll()
	if len(windows) == 0 {
		return fmt.Errorf("no windows to arrange")
	}

	// First window gets 70% width on left
	if len(windows) >= 1 {
		if wv, ok := windows[0].(*application.WebviewWindow); ok {
			wv.SetPosition(wa.X, wa.Y)
			wv.SetSize(wa.Width*70/100, wa.Height)
		}
	}

	// Remaining windows stack on right 30%
	rightX := wa.X + wa.Width*70/100
	rightWidth := wa.Width * 30 / 100
	remainingHeight := wa.Height / max(1, len(windows)-1)

	for i := 1; i < len(windows); i++ {
		if wv, ok := windows[i].(*application.WebviewWindow); ok {
			wv.SetPosition(rightX, wa.Y+(i-1)*remainingHeight)
			wv.SetSize(rightWidth, remainingHeight)
		}
	}

	return nil
}

func (s *Service) applyWorkflowDebugging() error {
	workAreas := s.GetWorkAreas()
	if len(workAreas) == 0 {
		return fmt.Errorf("no work areas available")
	}
	wa := workAreas[0]

	windows := s.app.Window().GetAll()
	if len(windows) == 0 {
		return fmt.Errorf("no windows to arrange")
	}

	// First window gets top 60%
	if len(windows) >= 1 {
		if wv, ok := windows[0].(*application.WebviewWindow); ok {
			wv.SetPosition(wa.X, wa.Y)
			wv.SetSize(wa.Width, wa.Height*60/100)
		}
	}

	// Remaining windows split bottom 40%
	bottomY := wa.Y + wa.Height*60/100
	bottomHeight := wa.Height * 40 / 100
	remainingWidth := wa.Width / max(1, len(windows)-1)

	for i := 1; i < len(windows); i++ {
		if wv, ok := windows[i].(*application.WebviewWindow); ok {
			wv.SetPosition(wa.X+(i-1)*remainingWidth, bottomY)
			wv.SetSize(remainingWidth, bottomHeight)
		}
	}

	return nil
}

func (s *Service) applyWorkflowPresenting() error {
	windows := s.app.Window().GetAll()
	if len(windows) == 0 {
		return fmt.Errorf("no windows to arrange")
	}

	// Maximize first window, minimize others
	for i, w := range windows {
		if wv, ok := w.(*application.WebviewWindow); ok {
			if i == 0 {
				wv.Maximise()
				wv.Focus()
			} else {
				wv.Minimise()
			}
		}
	}

	return nil
}
