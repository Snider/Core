// Package i18n provides internationalization and localization services.
//
// It is designed to be a simple, straightforward i18n solution that should be
// well-suited to most applications.
//
// # Getting Started
//
// To use the i18n service, you first need to create a new instance:
//
//	i18nService, err := i18n.New()
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Once you have a service instance, you can set the language and translate
// messages.
//
// # Locales
//
// The i18n service loads locales from the `locales` directory. Locales are JSON
// files with the language code as the filename (e.g., `en.json`, `es.json`).
// The service uses the `embed` package to bundle the locales into the binary,
// so you don't need to worry about distributing the locale files with your
// application.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

// Options holds configuration for the i18n service.
// This is a placeholder for future configuration options.
type Options struct{}

// Service provides internationalization and localization.
// It is the primary entrypoint for the i18n package.
type Service struct {
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

// New creates a new i18n service.
// The service is initialized with the English language as the default.
func New() (*Service, error) {
	s, err := newI18nService()
	if err != nil {
		return nil, err
	}
	err = s.SetLanguage("en")
	if err != nil {
		return nil, err
	}
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

// SetLanguage sets the language for the i18n service.
// The language tag should be a valid BCP 47 language tag (e.g., "en", "en-US").
// If the language is not supported, an error is returned.
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

// Translate translates a message by its ID.
// It accepts an optional template data argument to interpolate into the translation.
// If the message is not found, the message ID is returned.
func (s *Service) Translate(messageID string, args ...interface{}) string {
	config := &i18n.LocalizeConfig{MessageID: messageID}
	if len(args) > 0 {
		config.TemplateData = args[0]
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "i18n: Translate called with %d arguments, expected at most 1 (template data)\n", len(args))
		}
	}

	translation, err := s.localizer.Localize(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "i18n: translation for key \"%s\" not found\n", messageID)
		return messageID
	}
	return translation
}

// SetBundle is a test helper to inject a bundle.
func (s *Service) SetBundle(bundle *i18n.Bundle) {
	s.bundle = bundle
}
