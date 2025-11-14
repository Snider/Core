package core

import "github.com/wailsapp/wails/v3/pkg/application"

// ActionServiceStartup is a message sent when the application's services are starting up.
// This provides a hook for services to perform initialization tasks.
type ActionServiceStartup struct{}

// ActionServiceShutdown is a message sent when the application is shutting down.
// This allows services to perform cleanup tasks, such as saving state or closing resources.
type ActionServiceShutdown struct{}

// ActionDisplayOpenWindow is a structured message for requesting a new window.
type ActionDisplayOpenWindow struct {
	Name    string
	Options application.WebviewWindowOptions
}
