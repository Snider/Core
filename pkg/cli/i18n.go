package cli

import (
	"context"
	"sync"

	"github.com/host-uk/core/pkg/framework"
	"github.com/host-uk/core/pkg/i18n"
)

// I18nService wraps i18n as a Core service.
type I18nService struct {
	*framework.ServiceRuntime[I18nOptions]
	svc *i18n.Service

	// Collect mode state
	missingKeys   []i18n.MissingKeyAction
	missingKeysMu sync.Mutex
}

// I18nOptions configures the i18n service.
type I18nOptions struct {
	// Language overrides auto-detection (e.g., "en-GB", "de")
	Language string
	// Mode sets the translation mode (Normal, Strict, Collect)
	Mode i18n.Mode
}

// NewI18nService creates an i18n service factory.
func NewI18nService(opts I18nOptions) func(*framework.Core) (any, error) {
	return func(c *framework.Core) (any, error) {
		svc, err := i18n.New()
		if err != nil {
			return nil, err
		}

		if opts.Language != "" {
			svc.SetLanguage(opts.Language)
		}

		// Set mode if specified
		svc.SetMode(opts.Mode)

		return &I18nService{
			ServiceRuntime: framework.NewServiceRuntime(c, opts),
			svc:            svc,
			missingKeys:    make([]i18n.MissingKeyAction, 0),
		}, nil
	}
}

// OnStartup initialises the i18n service.
func (s *I18nService) OnStartup(ctx context.Context) error {
	s.Core().RegisterQuery(s.handleQuery)

	// Register action handler for collect mode
	if s.svc.Mode() == i18n.ModeCollect {
		i18n.SetActionHandler(s.handleMissingKey)
	}

	return nil
}

// handleMissingKey accumulates missing keys in collect mode.
func (s *I18nService) handleMissingKey(action i18n.MissingKeyAction) {
	s.missingKeysMu.Lock()
	defer s.missingKeysMu.Unlock()
	s.missingKeys = append(s.missingKeys, action)
}

// MissingKeys returns all missing keys collected in collect mode.
// Call this at the end of a QA session to report missing translations.
func (s *I18nService) MissingKeys() []i18n.MissingKeyAction {
	s.missingKeysMu.Lock()
	defer s.missingKeysMu.Unlock()
	result := make([]i18n.MissingKeyAction, len(s.missingKeys))
	copy(result, s.missingKeys)
	return result
}

// ClearMissingKeys resets the collected missing keys.
func (s *I18nService) ClearMissingKeys() {
	s.missingKeysMu.Lock()
	defer s.missingKeysMu.Unlock()
	s.missingKeys = s.missingKeys[:0]
}

// SetMode changes the translation mode.
func (s *I18nService) SetMode(mode i18n.Mode) {
	s.svc.SetMode(mode)

	// Update action handler registration
	if mode == i18n.ModeCollect {
		i18n.SetActionHandler(s.handleMissingKey)
	} else {
		i18n.SetActionHandler(nil)
	}
}

// Mode returns the current translation mode.
func (s *I18nService) Mode() i18n.Mode {
	return s.svc.Mode()
}

// Queries for i18n service

// QueryTranslate requests a translation.
type QueryTranslate struct {
	Key  string
	Args map[string]any
}

func (s *I18nService) handleQuery(c *framework.Core, q framework.Query) (any, bool, error) {
	switch m := q.(type) {
	case QueryTranslate:
		return s.svc.T(m.Key, m.Args), true, nil
	}
	return nil, false, nil
}

// T translates a key with optional arguments.
func (s *I18nService) T(key string, args ...map[string]any) string {
	if len(args) > 0 {
		return s.svc.T(key, args[0])
	}
	return s.svc.T(key)
}

// SetLanguage changes the current language.
func (s *I18nService) SetLanguage(lang string) {
	s.svc.SetLanguage(lang)
}

// Language returns the current language.
func (s *I18nService) Language() string {
	return s.svc.Language()
}

// AvailableLanguages returns all available languages.
func (s *I18nService) AvailableLanguages() []string {
	return s.svc.AvailableLanguages()
}

// --- Package-level convenience ---

// T translates a key using the CLI's i18n service.
// Falls back to the global i18n.T if CLI not initialised.
func T(key string, args ...map[string]any) string {
	if instance == nil {
		// CLI not initialised, use global i18n
		if len(args) > 0 {
			return i18n.T(key, args[0])
		}
		return i18n.T(key)
	}

	svc, err := framework.ServiceFor[*I18nService](instance.core, "i18n")
	if err != nil {
		// i18n service not registered, use global
		if len(args) > 0 {
			return i18n.T(key, args[0])
		}
		return i18n.T(key)
	}

	return svc.T(key, args...)
}
