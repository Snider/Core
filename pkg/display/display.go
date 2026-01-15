package display

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Options holds configuration for the display service.
// This struct is used to configure the display service at startup.
type Options struct{}

// Service manages windowing, dialogs, and other visual elements.
// It is the primary interface for interacting with the UI.
type Service struct {
	app    *application.App
	config Options
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

// Startup is called when the app starts. It initializes the display service
// and sets up the main application window and system tray.
//
//	err := displayService.Startup(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
func (s *Service) Startup(ctx context.Context) error {
	s.app = application.Get()
	s.app.Logger.Info("Display service started")
	s.buildMenu()
	s.systemTray()
	return s.OpenWindow()
}

// handleOpenWindowAction processes a message to configure and create a new window
// using the specified name and options.
func (s *Service) handleOpenWindowAction(msg map[string]any) error {
	opts := parseWindowOptions(msg)
	s.app.Window.NewWithOptions(opts)
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
	envInfo := s.app.Env.Info()

	details := "Environment Information:\n\n"
	details += fmt.Sprintf("Operating System: %s\n", envInfo.OS)
	details += fmt.Sprintf("Architecture: %s\n", envInfo.Arch)
	details += fmt.Sprintf("Debug Mode: %t\n\n", envInfo.Debug)
	details += fmt.Sprintf("Dark Mode: %t\n\n", s.app.Env.IsDarkMode())
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

	dialog := s.app.Dialog.Info()
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
	s.app.Window.NewWithOptions(wailsOpts)
	return nil
}

// buildWailsWindowOptions creates Wails window options from the given
// `WindowOption`s. This function is used by `OpenWindow` to construct the
// options for the new window.
func buildWailsWindowOptions(opts ...WindowOption) application.WebviewWindowOptions {
	// Default options
	winOpts := &WindowConfig{
		Name:   "main",
		Title:  "Core",
		Width:  1280,
		Height: 800,
		URL:    "/",
	}

	// Apply options
	for _, opt := range opts {
		opt.Apply(winOpts)
	}

	// Create Wails window options
	return application.WebviewWindowOptions{
		Name:                winOpts.Name,
		Title:               winOpts.Title,
		Width:               winOpts.Width,
		Height:              winOpts.Height,
		URL:                 winOpts.URL,
		AlwaysOnTop:         winOpts.AlwaysOnTop,
		Hidden:              winOpts.Hidden,
		MinimiseButtonState: winOpts.MinimiseButtonState,
		MaximiseButtonState: winOpts.MaximiseButtonState,
		CloseButtonState:    winOpts.CloseButtonState,
		Frameless:           winOpts.Frameless,
	}
}

// monitorScreenChanges listens for theme change events and logs when the screen
// configuration changes.
func (s *Service) monitorScreenChanges() {
	s.app.Event.OnApplicationEvent(events.Common.ThemeChanged, func(event *application.ApplicationEvent) {
		s.app.Logger.Info("Screen configuration changed")
	})
}
