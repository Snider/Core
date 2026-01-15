package display

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ClipboardContentType represents the type of content in the clipboard.
type ClipboardContentType string

const (
	ClipboardText  ClipboardContentType = "text"
	ClipboardImage ClipboardContentType = "image"
	ClipboardHTML  ClipboardContentType = "html"
)

// ClipboardContent holds clipboard data.
type ClipboardContent struct {
	Type ClipboardContentType `json:"type"`
	Text string               `json:"text,omitempty"`
	HTML string               `json:"html,omitempty"`
}

// ReadClipboard reads text content from the system clipboard.
func (s *Service) ReadClipboard() (string, error) {
	app := application.Get()
	if app == nil || app.Clipboard == nil {
		return "", fmt.Errorf("application or clipboard not available")
	}

	text, ok := app.Clipboard.Text()
	if !ok {
		return "", fmt.Errorf("failed to read clipboard")
	}
	return text, nil
}

// WriteClipboard writes text content to the system clipboard.
func (s *Service) WriteClipboard(text string) error {
	app := application.Get()
	if app == nil || app.Clipboard == nil {
		return fmt.Errorf("application or clipboard not available")
	}

	if !app.Clipboard.SetText(text) {
		return fmt.Errorf("failed to write to clipboard")
	}
	return nil
}

// HasClipboard checks if the clipboard has content.
func (s *Service) HasClipboard() bool {
	text, err := s.ReadClipboard()
	return err == nil && text != ""
}

// ClearClipboard clears the clipboard by setting empty text.
func (s *Service) ClearClipboard() error {
	return s.WriteClipboard("")
}
