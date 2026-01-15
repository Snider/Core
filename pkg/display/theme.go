package display

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// ThemeInfo contains information about the current theme.
type ThemeInfo struct {
	IsDark bool   `json:"isDark"`
	Theme  string `json:"theme"`  // "dark" or "light"
	System bool   `json:"system"` // Whether following system theme
}

// GetTheme returns the current application theme.
func (s *Service) GetTheme() ThemeInfo {
	app := application.Get()
	if app == nil {
		return ThemeInfo{Theme: "unknown"}
	}

	isDark := app.Env.IsDarkMode()
	theme := "light"
	if isDark {
		theme = "dark"
	}

	return ThemeInfo{
		IsDark: isDark,
		Theme:  theme,
		System: true, // Wails follows system theme by default
	}
}

// GetSystemTheme returns the system's theme preference.
// This is the same as GetTheme since Wails follows the system theme.
func (s *Service) GetSystemTheme() ThemeInfo {
	return s.GetTheme()
}
