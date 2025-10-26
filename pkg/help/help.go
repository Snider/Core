package help

import (
	"embed"
	"fmt"

	"github.com/Snider/Core/pkg/core"
)

//go:embed all:public/*
var helpStatic embed.FS

// Options holds configuration for the help service.
type Options struct{}

// Service manages the in-app help system.
type Service struct {
	*core.Runtime[Options]
	assets embed.FS // Corrected type
}

// New is a factory function that creates a new help Service.
func New(c *core.Core) (any, error) {
	return &Service{
		Runtime: core.NewRuntime(c, Options{}),
		assets:  helpStatic,
	}, nil
}

// Show displays the help window.
func (s *Service) Show() error {
	msg := map[string]any{
		"action": "display.open_window",
		"name":   "help", // Changed from "docs" to "help"
		"options": map[string]any{
			"Title":  "Help", // Changed from "Documentation" to "Help"
			"Width":  800,
			"Height": 600,
		},
	}

	// Dispatch the message through the core. The core will route it to the
	// appropriate handler, in this case, the one registered by the display service.
	return s.Core().ACTION(msg)
}

// ShowAt displays a specific section of the help documentation.
func (s *Service) ShowAt(anchor string) error {
	msg := map[string]any{
		"action": "display.open_window",
		"name":   "help",
		"options": map[string]any{
			"Title":  "Help",
			"Width":  800,
			"Height": 600,
			"URL":    fmt.Sprintf("/#%s", anchor), // Assuming MkDocs anchors work this way
		},
	}
	return s.Core().ACTION(msg)
}

// Ensure Service implements the core.Help interface.
var _ core.Help = (*Service)(nil)
