// Package i18n provides internationalization for the CLI.
package i18n

import (
	"io/fs"
	"runtime"
	"sync"
	"sync/atomic"
)

var missingKeyHandler atomic.Value // stores MissingKeyHandler

// localeRegistration holds a filesystem and directory for locale loading.
type localeRegistration struct {
	fsys fs.FS
	dir  string
}

var (
	registeredLocales   []localeRegistration
	registeredLocalesMu sync.Mutex
	localesLoaded       bool
)

// RegisterLocales registers a filesystem containing locale files to be loaded.
// Call this in your package's init() to register translations.
// Locales are loaded when the i18n service initialises.
//
//	//go:embed locales/*.json
//	var localeFS embed.FS
//
//	func init() {
//	    i18n.RegisterLocales(localeFS, "locales")
//	}
func RegisterLocales(fsys fs.FS, dir string) {
	registeredLocalesMu.Lock()
	defer registeredLocalesMu.Unlock()
	registeredLocales = append(registeredLocales, localeRegistration{fsys: fsys, dir: dir})

	// If locales already loaded (service already running), load immediately
	if localesLoaded {
		if svc := Default(); svc != nil {
			_ = svc.LoadFS(fsys, dir)
		}
	}
}

// loadRegisteredLocales loads all registered locale filesystems into the service.
// Called by the service during initialisation.
func loadRegisteredLocales(svc *Service) {
	registeredLocalesMu.Lock()
	defer registeredLocalesMu.Unlock()

	for _, reg := range registeredLocales {
		_ = svc.LoadFS(reg.fsys, reg.dir)
	}
	localesLoaded = true
}

// OnMissingKey registers a handler for missing translation keys.
// Called when T() can't find a key in ModeCollect.
// Thread-safe: can be called concurrently with translations.
//
//	i18n.SetMode(i18n.ModeCollect)
//	i18n.OnMissingKey(func(m i18n.MissingKey) {
//	    log.Printf("MISSING: %s at %s:%d", m.Key, m.CallerFile, m.CallerLine)
//	})
func OnMissingKey(h MissingKeyHandler) {
	missingKeyHandler.Store(h)
}

// dispatchMissingKey creates and dispatches a MissingKey event.
// Called internally when a key is missing in ModeCollect.
func dispatchMissingKey(key string, args map[string]any) {
	v := missingKeyHandler.Load()
	if v == nil {
		return
	}
	h, ok := v.(MissingKeyHandler)
	if !ok || h == nil {
		return
	}

	_, file, line, ok := runtime.Caller(2) // Skip dispatchMissingKey and handleMissingKey
	if !ok {
		file = "unknown"
		line = 0
	}

	h(MissingKey{
		Key:        key,
		Args:       args,
		CallerFile: file,
		CallerLine: line,
	})
}
