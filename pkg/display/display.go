package display

import (
	"context"
	"fmt"

	"github.com/Snider/Core/pkg/core"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Options holds configuration for the display service.
type Options struct{}

// Service manages windowing, dialogs, and other visual elements.
type Service struct {
	*core.Runtime[Options]
	config core.Config
}

// newDisplayService contains the common logic for initializing a Service struct.
func newDisplayService() (*Service, error) {
	return &Service{}, nil
}

// New is the constructor for static dependency injection.
// It creates a Service instance without initializing the core.Runtime field.
func New(cfg core.Config) (*Service, error) {
	s, err := newDisplayService()
	if err != nil {
		return nil, err
	}
	s.config = cfg
	return s, nil
}

// Register is the constructor for dynamic dependency injection (used with core.WithService).
// It creates a Service instance and initializes its core.Runtime field.
func Register(c *core.Core) (any, error) {
	s, err := newDisplayService()
	if err != nil {
		return nil, err
	}
	s.Runtime = core.NewRuntime(c, Options{})
	return s, nil
}

func (s *Service) ServiceName() string { return "github.com/Snider/Core/display" }

// HandleIPCEvents processes IPC messages and performs actions such as opening windows or initializing services based on message types.
func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
	switch m := msg.(type) {
	case map[string]any:
		if action, ok := m["action"].(string); ok && action == "display.open_window" {
			return s.handleOpenWindowAction(m)
		}
	case ActionOpenWindow:
		_, err := s.NewWithStruct(&m.WebviewWindowOptions)
		return err
	case core.ActionServiceStartup:
		return s.ServiceStartup(context.Background(), application.ServiceOptions{})
	default:
		c.App.Logger.Error("Display: Unknown message type", "type", fmt.Sprintf("%T", m))
	}
	return nil
}

// handleOpenWindowAction processes a message to configure and create a new window using specified name and options.
func (s *Service) handleOpenWindowAction(msg map[string]any) error {
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
	s.Core().App.Window.NewWithOptions(opts)
	return nil
}

// ShowEnvironmentDialog displays a dialog containing detailed information about the application's runtime environment.
func (s *Service) ShowEnvironmentDialog() {
	envInfo := s.Core().App.Env.Info()

	details := fmt.Sprintf(`Environment Information:

Operating System: %s
Architecture: %s
Debug Mode: %t

Dark Mode: %t

Platform Information:`,
		envInfo.OS,
		envInfo.Arch,
		envInfo.Debug,
		s.Core().App.Env.IsDarkMode()) // Use d.core.App

	// Add platform-specific details
	for key, value := range envInfo.PlatformInfo {
		details += fmt.Sprintf("\n%s: %v", key, value)
	}

	if envInfo.OSInfo != nil {
		details += fmt.Sprintf("\n\nOS Details:\nName: %s\nVersion: %s",
			envInfo.OSInfo.Name,
			envInfo.OSInfo.Version)
	}

	dialog := s.Core().App.Dialog.Info()
	dialog.SetTitle("Environment Information")
	dialog.SetMessage(details)
	dialog.Show()
}

// ServiceStartup initializes the display service and sets up the main application window and system tray.
func (s *Service) ServiceStartup(context.Context, application.ServiceOptions) error {
	s.Core().App.Logger.Info("Display service started")
	s.buildMenu()
	s.systemTray()

	// This will be updated to use the restored OpenWindow method
	mainOpts := application.WebviewWindowOptions{
		Name:   "main",
		Title:  "Core",
		Height: 900,
		Width:  1280,
		URL:    "/",
	}
	s.Core().App.Window.NewWithOptions(mainOpts)

	return nil
}

// monitorScreenChanges listens for theme change events and logs when screen configuration changes occur.
func (s *Service) monitorScreenChanges() {
	s.Core().App.Event.OnApplicationEvent(events.Common.ThemeChanged, func(event *application.ApplicationEvent) {
		s.Core().App.Logger.Info("Screen configuration changed")
	})
}
