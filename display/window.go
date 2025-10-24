package display

import "github.com/wailsapp/wails/v3/pkg/application"

// OpenWindow creates and shows a new webview window.
// This function is callable from the frontend.
func (d *API) OpenWindow(name string, options application.WebviewWindowOptions) {
	// Check if a window with that name already exists
	if window, exists := d.core.App.Window.GetByName(name); exists {
		window.Focus()
		return
	}

	window := d.core.App.Window.NewWithOptions(options)
	d.windowHandles[name] = window
	window.Show()
}

// SelectDirectory opens a directory selection dialog and returns the selected path.
func (d *API) SelectDirectory() (string, error) {
	dialog := application.OpenFileDialog()
	dialog.SetTitle("Select Project Directory")
	if path, err := dialog.PromptForSingleSelection(); err == nil {
		// Use selected directory path
		return path, nil
	}
	return "", nil
}
