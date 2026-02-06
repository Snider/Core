// Package icons provides embedded icon assets for the Core App.
package icons

import _ "embed"

// TrayTemplate is the template icon for macOS systray (22x22 PNG, black on transparent).
//
//go:embed tray-template.png
var TrayTemplate []byte

// TrayLight is the light mode icon for Windows/Linux systray.
//
//go:embed tray-light.png
var TrayLight []byte

// TrayDark is the dark mode icon for Windows/Linux systray.
//
//go:embed tray-dark.png
var TrayDark []byte

// AppIcon is the main application icon.
//
//go:embed appicon.png
var AppIcon []byte
