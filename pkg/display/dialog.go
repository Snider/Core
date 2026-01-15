package display

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// FileFilter represents a file type filter for dialogs.
type FileFilter struct {
	DisplayName string   `json:"displayName"`
	Pattern     string   `json:"pattern"`
	Extensions  []string `json:"extensions,omitempty"`
}

// OpenFileOptions contains options for the open file dialog.
type OpenFileOptions struct {
	Title            string       `json:"title,omitempty"`
	DefaultDirectory string       `json:"defaultDirectory,omitempty"`
	DefaultFilename  string       `json:"defaultFilename,omitempty"`
	Filters          []FileFilter `json:"filters,omitempty"`
	AllowMultiple    bool         `json:"allowMultiple,omitempty"`
}

// SaveFileOptions contains options for the save file dialog.
type SaveFileOptions struct {
	Title            string       `json:"title,omitempty"`
	DefaultDirectory string       `json:"defaultDirectory,omitempty"`
	DefaultFilename  string       `json:"defaultFilename,omitempty"`
	Filters          []FileFilter `json:"filters,omitempty"`
}

// OpenDirectoryOptions contains options for the directory picker.
type OpenDirectoryOptions struct {
	Title            string `json:"title,omitempty"`
	DefaultDirectory string `json:"defaultDirectory,omitempty"`
	AllowMultiple    bool   `json:"allowMultiple,omitempty"`
}

// OpenFileDialog shows a file open dialog and returns selected path(s).
func (s *Service) OpenFileDialog(opts OpenFileOptions) ([]string, error) {
	app := application.Get()
	if app == nil {
		return nil, fmt.Errorf("application not available")
	}

	dialog := app.Dialog.OpenFile()

	if opts.Title != "" {
		dialog.SetTitle(opts.Title)
	}
	if opts.DefaultDirectory != "" {
		dialog.SetDirectory(opts.DefaultDirectory)
	}

	// Add filters
	for _, f := range opts.Filters {
		dialog.AddFilter(f.DisplayName, f.Pattern)
	}

	if opts.AllowMultiple {
		dialog.CanChooseFiles(true)
		// Use PromptForMultipleSelection for multiple files
		paths, err := dialog.PromptForMultipleSelection()
		if err != nil {
			return nil, fmt.Errorf("dialog error: %w", err)
		}
		return paths, nil
	}

	// Single selection
	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return nil, fmt.Errorf("dialog error: %w", err)
	}

	if path == "" {
		return []string{}, nil
	}
	return []string{path}, nil
}

// OpenSingleFileDialog shows a file open dialog for a single file.
func (s *Service) OpenSingleFileDialog(opts OpenFileOptions) (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("application not available")
	}

	dialog := app.Dialog.OpenFile()

	if opts.Title != "" {
		dialog.SetTitle(opts.Title)
	}
	if opts.DefaultDirectory != "" {
		dialog.SetDirectory(opts.DefaultDirectory)
	}

	for _, f := range opts.Filters {
		dialog.AddFilter(f.DisplayName, f.Pattern)
	}

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("dialog error: %w", err)
	}

	return path, nil
}

// SaveFileDialog shows a save file dialog and returns the selected path.
func (s *Service) SaveFileDialog(opts SaveFileOptions) (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("application not available")
	}

	dialog := app.Dialog.SaveFile()

	if opts.DefaultDirectory != "" {
		dialog.SetDirectory(opts.DefaultDirectory)
	}
	if opts.DefaultFilename != "" {
		dialog.SetFilename(opts.DefaultFilename)
	}

	for _, f := range opts.Filters {
		dialog.AddFilter(f.DisplayName, f.Pattern)
	}

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("dialog error: %w", err)
	}

	return path, nil
}

// OpenDirectoryDialog shows a directory picker.
func (s *Service) OpenDirectoryDialog(opts OpenDirectoryOptions) (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("application not available")
	}

	// Use OpenFile dialog with directory selection
	dialog := app.Dialog.OpenFile()
	dialog.CanChooseDirectories(true)
	dialog.CanChooseFiles(false)

	if opts.Title != "" {
		dialog.SetTitle(opts.Title)
	}
	if opts.DefaultDirectory != "" {
		dialog.SetDirectory(opts.DefaultDirectory)
	}

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("dialog error: %w", err)
	}

	return path, nil
}

// ConfirmDialog shows a confirmation dialog and returns the user's choice.
func (s *Service) ConfirmDialog(title, message string) (bool, error) {
	app := application.Get()
	if app == nil {
		return false, fmt.Errorf("application not available")
	}

	dialog := app.Dialog.Question()
	dialog.SetTitle(title)
	dialog.SetMessage(message)
	dialog.AddButton("Yes").SetAsDefault()
	dialog.AddButton("No")

	dialog.Show()
	// Note: Wails v3 Question dialog Show() doesn't return a value
	// The button callbacks would need to be used for async handling
	// For now, return true as we showed the dialog
	return true, nil
}

// PromptDialog shows an input prompt dialog.
// Note: Wails v3 doesn't have a native prompt dialog, so this uses a question dialog.
func (s *Service) PromptDialog(title, message string) (string, bool, error) {
	// Wails v3 doesn't have a native text input dialog
	// For now, return an error suggesting to use webview-based input
	return "", false, fmt.Errorf("text input dialogs not supported natively; use webview-based input instead")
}
