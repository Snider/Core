package core

import "github.com/wailsapp/wails/v3/pkg/application"

type ActionServiceStartup struct{}

// ActionDisplayOpenWindow is a structured message for requesting a new window.
type ActionDisplayOpenWindow struct {
	Name    string
	Options application.WebviewWindowOptions
}
