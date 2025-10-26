package docs

import (
	"embed"

	"github.com/Snider/Core/pkg/core"
)

//go:embed all:static/*
var docsStatic embed.FS

// Options holds configuration for the docs service.
type Options struct{}

// Service manages documentation.
type Service struct {
	*core.Runtime[Options]
	config  core.Config
	display core.Display
	assets  embed.FS
}

// New is a factory function that creates a new docs Service.
func New(c *core.Core) (any, error) {
	configSvc := core.ServiceFor[core.Config](c, "config")
	displaySvc := core.ServiceFor[core.Display](c, "display")

	return &Service{
		Runtime: core.NewRuntime(c, Options{}),
		config:  configSvc,
		display: displaySvc,
		assets:  docsStatic,
	}, nil
}

// Show displays the documentation window.
func (s *Service) Show() error {
	// The display service is not available during static analysis, so we must
	// check for nil before using it.
	if s.display == nil {
		return nil
	}
	return s.display.OpenWindow(
	// Options would be passed here, perhaps fetched from config
	)
}

// ShowAt displays a specific section of the documentation.
func (s *Service) ShowAt(anchor string) error {
	if s.display == nil {
		return nil
	}
	// Implementation for opening at a specific anchor
	return nil
}
