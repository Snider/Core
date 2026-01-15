package display

import "github.com/wailsapp/wails/v3/pkg/application"

// ActionOpenWindow is an IPC message used to request a new window. It contains
// the options for the new window.
//
// example:
//
//	action := display.ActionOpenWindow{
//		WebviewWindowOptions: application.WebviewWindowOptions{
//			Name: "my-window",
//			Title: "My Window",
//			Width: 800,
//			Height: 600,
//		},
//	}
type ActionOpenWindow struct {
	application.WebviewWindowOptions
}
