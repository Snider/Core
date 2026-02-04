package icons

import _ "embed"

// AppTray is the main application tray icon.
//
//go:embed apptray.png
var AppTray []byte

// SystrayMacTemplate is the template icon for macOS systray (22x22 PNG, black on transparent).
// Template icons automatically adapt to light/dark mode on macOS.
//
//go:embed systray-mac-template.png
var SystrayMacTemplate []byte

// SystrayDefault is the default icon for Windows/Linux systray.
//
//go:embed systray-default.png
var SystrayDefault []byte
