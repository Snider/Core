package docs

import (
	"embed"

	"github.com/Snider/Core"
)

//go:embed all:static/*
var docsStatic embed.FS

// Options holds configuration for the doc service.
type Options struct{}

// Service manages the documentation assets and display requests.
// It embeds the core.Runtime to get access to the core instance and its functions.
type Service struct {
	*core.Runtime[Options]
	assets embed.FS
}

// New is a factory function that creates a new docs Service.
// It is self-contained and only depends on the Core, with no knowledge
// of other services at compile time.
func New(c *core.Core) (any, error) {
	s := &Service{
		Runtime: core.NewRuntime(c, Options{}),
		assets:  docsStatic,
	}
	return s, nil
}

// Show triggers the display of the documentation window.
func (s *Service) Show() error {
	// The message is a generic map, which any service can create. The 'display'
	// service will register a handler that knows how to interpret this structure.
	msg := map[string]any{
		"action": "display.open_window",
		"name":   "docs",
		"options": map[string]any{
			"Title":  "Documentation",
			"Width":  800,
			"Height": 600,
		},
	}

	// Dispatch the message through the core. The core will route it to the
	// appropriate handler, in this case, the one registered by the display service.
	return s.Core().ACTION(msg)
}
