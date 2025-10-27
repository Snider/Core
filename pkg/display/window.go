package display

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

type WindowOption func(*application.WebviewWindowOptions) error

type Window = application.WebviewWindowOptions

func WindowName(s string) WindowOption {
	return func(o *Window) error {
		o.Name = s
		return nil
	}
}
func WindowTitle(s string) WindowOption {
	return func(o *Window) error {
		o.Title = s
		return nil
	}
}

func WindowURL(s string) WindowOption {
	return func(o *Window) error {
		o.URL = s
		return nil
	}
}

func WindowWidth(i int) WindowOption {
	return func(o *Window) error {
		o.Width = i
		return nil
	}
}

func WindowHeight(i int) WindowOption {
	return func(o *Window) error {
		o.Height = i
		return nil
	}
}

func applyOptions(opts ...WindowOption) *Window {
	w := &Window{}
	if opts == nil {
		return w
	}
	for _, o := range opts {
		if err := o(w); err != nil {
			return nil
		}
	}
	return w
}

// NewWithStruct creates a new window using the provided options and returns its handle.
func (s *Service) NewWithStruct(options *Window) (*application.WebviewWindow, error) {
	return s.Core().App.Window.NewWithOptions(*options), nil
}

// NewWithOptions creates a new window by applying a series of options.
func (s *Service) NewWithOptions(opts ...WindowOption) (*application.WebviewWindow, error) {
	return s.NewWithStruct(applyOptions(opts...))
}

// NewWithURL creates a new default window pointing to the specified URL.
func (s *Service) NewWithURL(url string) (*application.WebviewWindow, error) {
	return s.NewWithOptions(
		WindowURL(url),
		WindowTitle("Core"),
		WindowHeight(900),
		WindowWidth(1280),
	)
}

// OpenWindow is a convenience method that creates and shows a window from a set of options.
func (s *Service) OpenWindow(opts ...WindowOption) error {
	_, err := s.NewWithOptions(opts...)
	return err
}

// SelectDirectory opens a directory selection dialog and returns the selected path.
func (s *Service) SelectDirectory() (string, error) {
	dialog := application.OpenFileDialog()
	dialog.SetTitle("Select Project Directory")
	return dialog.PromptForSingleSelection()
}

var instance *Window

func (s *Service) Window() *Window { return instance }
