package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Snider/Core/pkg/core"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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

// New is the factory function for the core.WithService pattern.
func New(c *core.Core) (any, error) {
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

	// --- Determine initial language ---
	configSvc := core.ServiceFor[core.Config](c, "config")
	initialLang := ""
	if configSvc != nil {
		var lang string
		_ = configSvc.Get("language", &lang)
		if lang != "" {
			initialLang = lang
		}
	}

	if initialLang == "" {
		detectedLang, _ := detectLanguage(availableLangs)
		if detectedLang != "" {
			initialLang = detectedLang
		}
	}

	if initialLang == "" {
		initialLang = "en"
	}

	s := &Service{
		Runtime:        core.NewRuntime(c, Options{}),
		bundle:         bundle,
		availableLangs: availableLangs,
	}
	s.SetLanguage(initialLang)

	return s, nil
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
	s.localizer = i18n.NewLocalizer(s.bundle, lang)
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
