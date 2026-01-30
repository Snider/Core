package cli

import (
	"context"

	"github.com/host-uk/core/pkg/framework"
	"github.com/host-uk/core/pkg/i18n"
)

// I18nService wraps i18n as a Core service.
type I18nService struct {
	*framework.ServiceRuntime[I18nOptions]
	svc *i18n.Service
}

// I18nOptions configures the i18n service.
type I18nOptions struct {
	// Language overrides auto-detection (e.g., "en-GB", "de")
	Language string
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

		return &I18nService{
			ServiceRuntime: framework.NewServiceRuntime(c, opts),
			svc:            svc,
		}, nil
	}
}

// OnStartup initialises the i18n service.
func (s *I18nService) OnStartup(ctx context.Context) error {
	s.Core().RegisterQuery(s.handleQuery)
	return nil
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
