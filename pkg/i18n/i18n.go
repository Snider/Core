package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Snider/Core/pkg/core"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

// Options holds configuration for the i18n service.
type Options struct{}

// Service provides internationalization and localization.
type Service struct {
	*core.Runtime[Options]
	bundle         *i18n.Bundle
	localizer      *i18n.Localizer
	availableLangs []language.Tag
}

// newI18nService contains the common logic for initializing a Service struct.
func newI18nService() (*Service, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	availableLangs, err := getAvailableLanguages()
	if err != nil {
		return nil, err
	}

	for _, lang := range availableLangs {
		filePath := fmt.Sprintf("locales/%s.json", lang.String())
		if _, err := bundle.LoadMessageFileFS(localeFS, filePath); err != nil {
			return nil, fmt.Errorf("failed to load message file %s: %w", filePath, err)
		}
	}

	s := &Service{
		bundle:         bundle,
		availableLangs: availableLangs,
	}
	// Language will be set during ServiceStartup after config is available.
	return s, nil
}

// New is the constructor for static dependency injection.
// It creates a Service instance without initializing the core.Runtime field.
// Dependencies are passed directly here.
func New() (*Service, error) {
	s, err := newI18nService()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Register is the constructor for dynamic dependency injection (used with core.WithService).
// It creates a Service instance and initializes its core.Runtime field.
// Dependencies are injected during ServiceStartup.
func Register(c *core.Core) (any, error) {
	s, err := newI18nService()
	if err != nil {
		return nil, err
	}
	s.Runtime = core.NewRuntime(c, Options{})
	return s, nil
}

// HandleIPCEvents processes IPC messages, including injecting dependencies on startup.
func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
	switch m := msg.(type) {
	case core.ActionServiceStartup:
		return s.ServiceStartup(context.Background(), application.ServiceOptions{})
	default:
		c.App.Logger.Error("Display: Unknown message type", "type", fmt.Sprintf("%T", m))
	}
	return nil
}

// ServiceStartup is called when the app starts, after dependencies are injected.
func (s *Service) ServiceStartup(context.Context, application.ServiceOptions) error {
	// Determine initial language after config is available.
	initialLang := "en"
	var lang string
	_ = s.Config().Get("language", &lang)
	if lang != "" {
		initialLang = lang
	}
	err := s.SetLanguage(initialLang)
	if err != nil {
		return err
	}
	s.Core().App.Logger.Info("I18n service started")
	return nil
}

// --- Language Management ---

func getAvailableLanguages() ([]language.Tag, error) {
	files, err := localeFS.ReadDir("locales")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded locales directory: %w", err)
	}

	var availableLangs []language.Tag
	for _, file := range files {
		lang := strings.TrimSuffix(file.Name(), ".json")
		tag := language.Make(lang)
		availableLangs = append(availableLangs, tag)
	}
	return availableLangs, nil
}

func detectLanguage(supported []language.Tag) (string, error) {
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		return "", nil
	}

	baseLang := strings.Split(langEnv, ".")[0]
	parsedLang, err := language.Parse(baseLang)
	if err != nil {
		return "", fmt.Errorf("failed to parse language tag '%s': %w", baseLang, err)
	}

	if len(supported) == 0 {
		return "", nil
	}

	matcher := language.NewMatcher(supported)
	_, index, confidence := matcher.Match(parsedLang)

	if confidence >= language.Low {
		return supported[index].String(), nil
	}
	return "", nil
}

// --- Public Service Methods ---

func (s *Service) SetLanguage(lang string) error {
	requestedLang, err := language.Parse(lang)
	if err != nil {
		return fmt.Errorf("i18n: failed to parse language tag \"%s\": %w", lang, err)
	}

	if len(s.availableLangs) == 0 {
		return fmt.Errorf("i18n: no available languages loaded in the bundle")
	}

	matcher := language.NewMatcher(s.availableLangs)
	bestMatch, _, confidence := matcher.Match(requestedLang)

	if confidence == language.No {
		return fmt.Errorf("i18n: unsupported language: %s", lang)
	}

	s.localizer = i18n.NewLocalizer(s.bundle, bestMatch.String())
	return nil
}

func (s *Service) Translate(messageID string) string {
	translation, err := s.localizer.Localize(&i18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
		fmt.Fprintf(os.Stderr, "i18n: translation for key \"%s\" not found\n", messageID)
		return messageID
	}
	return translation
}

// Ensure Service implements the core.I18n interface.
var _ core.I18n = (*Service)(nil)
