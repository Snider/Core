// Package help provides a flexible, embeddable help system for Go applications.
// It is designed to be easily integrated into various environments, offering
// a standalone help window that can be triggered from within the application.
//
// The core of this package is the `Service`, which manages the help content
// and its presentation. The help content can be sourced from an embedded
// filesystem, a local directory, or the default `mkdocs` source, providing
// flexibility in how documentation is bundled with the application.
//
// A key feature of the `help` package is its ability to operate with or

// without a `Snider/display` module. When a display module is not available,
// the service falls back to a direct `wails3` implementation for displaying
// the help window, ensuring that the help functionality remains available
// in different system configurations.
//
// The package defines several interfaces (`Logger`, `App`, `Core`, `Display`, `Help`)
// to decouple the help service from the main application, promoting a clean
// architecture and making it easier to mock dependencies for testing.
//
// Usage:
// To use the help service, create a new instance with `New()`, providing
// `Options` to configure the source of the help assets. The service can then
// be initialized with a `Core` and `Display` implementation. The `Show()` and
// `ShowAt()` methods can be called to display the help window or a specific
// section of it.
package help

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:public/*
var helpStatic embed.FS

// Logger defines the interface for a basic logger. It includes methods for
// logging informational and error messages, which helps in decoupling the
// help service from a specific logging implementation.
type Logger interface {
	// Info logs an informational message.
	Info(message string, args ...any)
	// Error logs an error message.
	Error(message string, args ...any)
}

// App defines the interface for accessing application-level components,
// such as the logger. This allows the help service to interact with the
// application's infrastructure in a loosely coupled manner.
type App interface {
	// Logger returns the application's logger instance.
	Logger() Logger
}

// Core defines the interface for the core runtime functionalities that the
// help service depends on. This typically includes methods for performing
// actions and accessing the application context.
type Core interface {
	// ACTION dispatches a message to the core runtime for processing.
	// This is used to trigger actions like opening a window.
	ACTION(msg map[string]any) error
	// App returns the application-level context.
	App() App
}

// Display defines the interface for a display service. The help service
// uses this interface to check for the presence of a display module,
// allowing it to function as an optional dependency.
type Display interface{}

// Help defines the public interface of the help service. It exposes methods
// for showing the help window and navigating to specific sections.
type Help interface {
	// Show displays the main help window.
	Show() error
	// ShowAt displays a specific section of the help documentation,
	// identified by an anchor.
	ShowAt(anchor string) error
	// ServiceStartup is a lifecycle method called when the application starts.
	ServiceStartup(ctx context.Context) error
}

// Options holds the configuration for the help service. It allows for
// customization of the help content source.
type Options struct {
	// Source specifies the directory or path to the help content.
	// If empty, it defaults to "mkdocs".
	Source string
	// Assets provides an alternative way to specify the help content
	// using a filesystem interface, which is useful for embedded assets.
	Assets fs.FS
}

// Service manages the in-app help system. It handles the initialization
// of the help content, interaction with the core runtime, and display
// of the help window.
type Service struct {
	core    Core
	display Display
	assets  fs.FS
	opts    Options
}

// New creates a new instance of the help Service. It initializes the service
// with the provided options, setting up the asset filesystem based on the
// specified source. If no source is provided, it defaults to the embedded
// "mkdocs" content.
//
// Example:
//
//	// Create a new help service with default options.
//	helpService, err := help.New(help.Options{})
//	if err != nil {
//		log.Fatal(err)
//	}
func New(opts Options) (*Service, error) {
	if opts.Source == "" {
		opts.Source = "mkdocs"
	}

	s := &Service{
		opts: opts,
	}

	var err error
	if opts.Assets != nil {
		s.assets = opts.Assets
	} else if s.opts.Source != "mkdocs" {
		s.assets = os.DirFS(s.opts.Source)
	} else {
		s.assets, err = fs.Sub(helpStatic, "public")
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Init initializes the service with its core dependencies. This method is
// intended to be called by the dependency injection system of the application
// to provide the necessary `Core` and `Display` implementations.
func (s *Service) Init(c Core, d Display) {
	s.core = c
	s.display = d
}

// ServiceStartup is a lifecycle method that is called by the application when
// it starts. It performs necessary checks to ensure that the service has been
// properly initialized with its dependencies.
func (s *Service) ServiceStartup(context.Context) error {
	if s.core == nil {
		return fmt.Errorf("core runtime not initialized")
	}
	s.core.App().Logger().Info("Help service started")
	return nil
}

// Show displays the main help window. If a `Display` service is available,
// it sends an action to the core runtime to open the window. Otherwise, it
// falls back to using the `wails3` application instance to create a new
// window. This ensures that the help functionality is available even when
// the `Snider/display` module is not in use.
func (s *Service) Show() error {
	if s.display == nil {
		app := application.Get()
		if app == nil {
			return fmt.Errorf("wails application not running")
		}
		app.Window.NewWithOptions(application.WebviewWindowOptions{
			Title:  "Help",
			Width:  800,
			Height: 600,
			URL:    "/",
		})
		return nil
	}
	if s.core == nil {
		return fmt.Errorf("core runtime not initialized")
	}
	msg := map[string]any{
		"action": "display.open_window",
		"name":   "help",
		"options": map[string]any{
			"Title":  "Help",
			"Width":  800,
			"Height": 600,
		},
	}

	return s.core.ACTION(msg)
}

// ShowAt displays a specific section of the help documentation, identified
// by an anchor. Similar to `Show`, it uses the `Display` service if available,
// or falls back to a direct `wails3` implementation. The anchor is appended
// to the URL, allowing the help window to open directly to the relevant
// section.
func (s *Service) ShowAt(anchor string) error {
	if s.display == nil {
		app := application.Get()
		if app == nil {
			return fmt.Errorf("wails application not running")
		}
		url := fmt.Sprintf("/#%s", anchor)
		app.Window.NewWithOptions(application.WebviewWindowOptions{
			Title:  "Help",
			Width:  800,
			Height: 600,
			URL:    url,
		})
		return nil
	}
	if s.core == nil {
		return fmt.Errorf("core runtime not initialized")
	}

	url := fmt.Sprintf("/#%s", anchor)

	msg := map[string]any{
		"action": "display.open_window",
		"name":   "help",
		"options": map[string]any{
			"Title":  "Help",
			"Width":  800,
			"Height": 600,
			"URL":    url,
		},
	}
	return s.core.ACTION(msg)
}

// Ensure Service implements the Help interface.
var _ Help = (*Service)(nil)
