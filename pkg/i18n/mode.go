// Package i18n provides internationalization for the CLI.
package i18n

import (
	"runtime"
)

// Mode determines how the i18n service handles missing translation keys.
type Mode int

const (
	// ModeNormal returns the key as-is when a translation is missing (production).
	ModeNormal Mode = iota
	// ModeStrict panics immediately when a translation is missing (dev/CI).
	ModeStrict
	// ModeCollect dispatches a MissingKeyAction and returns [key] (QA testing).
	ModeCollect
)

// String returns the string representation of the mode.
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeStrict:
		return "strict"
	case ModeCollect:
		return "collect"
	default:
		return "unknown"
	}
}

// MissingKeyAction is dispatched when a translation key is not found in collect mode.
// It contains caller information for debugging and QA purposes.
type MissingKeyAction struct {
	Key        string         // The missing translation key
	Args       map[string]any // Arguments passed to the translation
	CallerFile string         // Source file where T() was called
	CallerLine int            // Line number where T() was called
}

// ActionHandler is a function that handles MissingKeyAction dispatches.
// Register handlers via SetActionHandler to receive missing key notifications.
type ActionHandler func(action MissingKeyAction)

var actionHandler ActionHandler

// SetActionHandler registers a handler for MissingKeyAction dispatches.
// Only one handler can be active at a time; subsequent calls replace the previous handler.
func SetActionHandler(h ActionHandler) {
	actionHandler = h
}

// dispatchMissingKey creates and dispatches a MissingKeyAction.
// Called internally when a key is missing in collect mode.
func dispatchMissingKey(key string, args map[string]any) {
	if actionHandler == nil {
		return
	}

	_, file, line, ok := runtime.Caller(2) // Skip dispatchMissingKey and handleMissingKey
	if !ok {
		file = "unknown"
		line = 0
	}

	actionHandler(MissingKeyAction{
		Key:        key,
		Args:       args,
		CallerFile: file,
		CallerLine: line,
	})
}
