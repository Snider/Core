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
	// ModeCollect dispatches MissingKey actions and returns [key] (QA testing).
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

// MissingKey is dispatched when a translation key is not found in ModeCollect.
// Used by QA tools to collect and report missing translations.
type MissingKey struct {
	Key        string         // The missing translation key
	Args       map[string]any // Arguments passed to the translation
	CallerFile string         // Source file where T()/C() was called
	CallerLine int            // Line number where T()/C() was called
}

// MissingKeyAction is an alias for backwards compatibility.
// Deprecated: Use MissingKey instead.
type MissingKeyAction = MissingKey

// MissingKeyHandler receives missing key events for analysis.
type MissingKeyHandler func(missing MissingKey)

var missingKeyHandler MissingKeyHandler

// OnMissingKey registers a handler for missing translation keys.
// Called when T() or C() can't find a key in ModeCollect.
//
//	i18n.SetMode(i18n.ModeCollect)
//	i18n.OnMissingKey(func(m i18n.MissingKey) {
//	    log.Printf("MISSING: %s at %s:%d", m.Key, m.CallerFile, m.CallerLine)
//	})
func OnMissingKey(h MissingKeyHandler) {
	missingKeyHandler = h
}

// SetActionHandler registers a handler for missing key dispatches.
// Deprecated: Use OnMissingKey instead.
func SetActionHandler(h func(action MissingKeyAction)) {
	OnMissingKey(h)
}

// dispatchMissingKey creates and dispatches a MissingKey event.
// Called internally when a key is missing in ModeCollect.
func dispatchMissingKey(key string, args map[string]any) {
	if missingKeyHandler == nil {
		return
	}

	_, file, line, ok := runtime.Caller(2) // Skip dispatchMissingKey and handleMissingKey
	if !ok {
		file = "unknown"
		line = 0
	}

	missingKeyHandler(MissingKey{
		Key:        key,
		Args:       args,
		CallerFile: file,
		CallerLine: line,
	})
}
