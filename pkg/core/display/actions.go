package display

import "github.com/wailsapp/wails/v3/pkg/application"

// ActionOpenWindow is an IPC message used to request a new window.
type ActionOpenWindow struct {
	application.WebviewWindowOptions
}
