package display

import "github.com/wailsapp/wails/v3/pkg/application"

// WindowConfig holds the configuration for a window. This struct is used to
// create a new window with the specified options.
type WindowConfig struct {
	Name                string
	Title               string
	Width               int
	Height              int
	URL                 string
	AlwaysOnTop         bool
	Hidden              bool
	MinimiseButtonState application.ButtonState
	MaximiseButtonState application.ButtonState
	CloseButtonState    application.ButtonState
	Frameless           bool
}

// WindowOption is an interface for applying configuration options to a
// WindowConfig.
type WindowOption interface {
	Apply(*WindowConfig)
}

// WindowOptionFunc is a function that implements the WindowOption interface.
// This allows us to use ordinary functions as window options.
type WindowOptionFunc func(*WindowConfig)

// Apply calls the underlying function to apply the configuration.
func (f WindowOptionFunc) Apply(c *WindowConfig) {
	f(c)
}

// WithName sets the name of the window.
func WithName(name string) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.Name = name
	})
}

// WithTitle sets the title of the window.
func WithTitle(title string) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.Title = title
	})
}

// WithWidth sets the width of the window.
func WithWidth(width int) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.Width = width
	})
}

// WithHeight sets the height of the window.
func WithHeight(height int) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.Height = height
	})
}

// WithURL sets the URL that the window will load.
func WithURL(url string) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.URL = url
	})
}

// WithAlwaysOnTop sets the window to always be on top of other windows.
func WithAlwaysOnTop(alwaysOnTop bool) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.AlwaysOnTop = alwaysOnTop
	})
}

// WithHidden sets the window to be hidden when it is created.
func WithHidden(hidden bool) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.Hidden = hidden
	})
}

// WithMinimiseButtonState sets the state of the minimise button.
func WithMinimiseButtonState(state application.ButtonState) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.MinimiseButtonState = state
	})
}

// WithMaximiseButtonState sets the state of the maximise button.
func WithMaximiseButtonState(state application.ButtonState) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.MaximiseButtonState = state
	})
}

// WithCloseButtonState sets the state of the close button.
func WithCloseButtonState(state application.ButtonState) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.CloseButtonState = state
	})
}

// WithFrameless sets the window to be frameless.
func WithFrameless(frameless bool) WindowOption {
	return WindowOptionFunc(func(c *WindowConfig) {
		c.Frameless = frameless
	})
}
