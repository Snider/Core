// Package docs provides documentation window management.
package docs

import (
	"github.com/Snider/Core/pkg/core"
	"github.com/Snider/Core/pkg/display"
)

// Service manages documentation windows.
type Service struct {
	core    *core.Core
	baseURL string
}

// Options configures the docs service.
type Options struct {
	BaseURL string
}

// New creates a new docs service.
func New(opts Options) (*Service, error) {
	return &Service{
		baseURL: opts.BaseURL,
	}, nil
}

// SetCore sets the core reference for accessing other services.
func (s *Service) SetCore(c *core.Core) {
	s.core = c
}

// SetBaseURL sets the base URL for documentation.
func (s *Service) SetBaseURL(url string) {
	s.baseURL = url
}

// OpenDocsWindow opens a documentation window at the specified path.
// The path is appended to the base URL to form the full documentation URL.
func (s *Service) OpenDocsWindow(path string) error {
	url := s.baseURL
	if path != "" {
		if url != "" && url[len(url)-1] != '/' && path[0] != '/' {
			url += "/"
		}
		url += path
	}

	// Fire an ACTION to request a window from the display service
	return s.core.ACTION(display.ActionOpenWindow{
		WebviewWindowOptions: display.Window{
			Name:   "docs-window",
			Title:  "Documentation",
			URL:    url,
			Width:  1200,
			Height: 800,
		},
	})
}
