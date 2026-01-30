// Package i18n provides internationalization for the CLI.
package i18n

import (
	"runtime"
)

var missingKeyHandler MissingKeyHandler

// OnMissingKey registers a handler for missing translation keys.
// Called when T() can't find a key in ModeCollect.
//
//	i18n.SetMode(i18n.ModeCollect)
//	i18n.OnMissingKey(func(m i18n.MissingKey) {
//	    log.Printf("MISSING: %s at %s:%d", m.Key, m.CallerFile, m.CallerLine)
//	})
func OnMissingKey(h MissingKeyHandler) {
	missingKeyHandler = h
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
