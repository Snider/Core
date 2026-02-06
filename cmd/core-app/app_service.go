package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppService provides native desktop capabilities to the Wails frontend.
// These methods are callable via window.go.main.AppService.{Method}()
// from any JavaScript/webview context.
type AppService struct {
	app *application.App
	env *AppEnvironment
}

func NewAppService(env *AppEnvironment) *AppService {
	return &AppService{env: env}
}

// ServiceStartup is called by Wails when the application starts.
func (s *AppService) ServiceStartup(app *application.App) {
	s.app = app
}

// GetVersion returns the application version.
func (s *AppService) GetVersion() string {
	return "0.1.0"
}

// GetDataDir returns the persistent data directory path.
func (s *AppService) GetDataDir() string {
	return s.env.DataDir
}

// GetDatabasePath returns the SQLite database file path.
func (s *AppService) GetDatabasePath() string {
	return s.env.DatabasePath
}

// ShowWindow shows and focuses the main application window.
func (s *AppService) ShowWindow(name string) {
	if s.app == nil {
		return
	}
	if w, ok := s.app.Window.Get(name); ok {
		w.Show()
		w.Focus()
	}
}
