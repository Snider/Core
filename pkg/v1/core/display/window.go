package display

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Option func(*Window) error

type Window = application.WebviewWindowOptions

func OptName(s string) Option {
	return func(o *Window) error {
		o.Name = s
		return nil
	}
}
func OptTitle(s string) Option {
	return func(o *Window) error {
		o.Title = s
		return nil
	}
}

func OptURL(s string) Option {
	return func(o *Window) error {
		o.URL = s
		return nil
	}
}

func OptWidth(i int) Option {
	return func(o *Window) error {
		o.Width = i
		return nil
	}
}

func OptHeight(i int) Option {
	return func(o *Window) error {
		o.Height = i
		return nil
	}
}

func applyOptions(opts ...Option) *Window {
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

// NewWithOptions creates a new window using the provided options and returns its handle.
// The caller is responsible for showing the window.
func (d *API) NewWithStruct(options *Window) *application.WebviewWindow {
	return d.core.App.Window.NewWithOptions(*options)
}
func (d *API) NewWithOptions(opts ...Option) *application.WebviewWindow {
	return d.NewWithStruct(applyOptions(opts...))
}

// OpenWindow is a legacy or direct method to open a window using Wails-native options.
// It returns the handle of the created window. The caller is responsible for showing the window.
func (d *API) OpenWindow(opts ...Option) *application.WebviewWindow {
	return d.NewWithOptions(opts...)
}

// SelectDirectory opens a directory selection dialog and returns the selected path.
func (d *API) SelectDirectory() (string, error) {
	dialog := application.OpenFileDialog()
	dialog.SetTitle("Select Project Directory")
	if path, err := dialog.PromptForSingleSelection(); err == nil {
		return path, nil
	}
	return "", nil
}
