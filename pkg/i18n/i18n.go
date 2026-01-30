// Package i18n provides internationalization for the CLI.
//
// It is designed to be extended by the GUI version, which can import this
// package and add additional translations for GUI-specific strings.
//
// # Getting Started
//
//	svc, err := i18n.New()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(svc.T("cli.success"))
//
// # Extending for GUI
//
// The GUI can extend this package by creating its own Service that embeds
// this one and loads additional locale files:
//
//	guiService, err := i18n.NewWithFS(guiLocaleFS, "locales")
//
// # Locale Files
//
// Locale files are JSON with message IDs as keys. Supports both simple strings
// and go-i18n format with pluralization:
//
//	{
//	    "cli.success": "Operation completed successfully",
//	    "cli.items_found": {
//	        "one": "{{.Count}} item found",
//	        "other": "{{.Count}} items found"
//	    }
//	}
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

// Default is the global i18n service instance.
// Initialized lazily on first use or via Init().
var (
	defaultService *Service
	defaultOnce    sync.Once
	defaultErr     error
)

// Service provides internationalization and localization.
type Service struct {
	bundle         *i18n.Bundle
	localizer      *i18n.Localizer
	currentLang    string
	availableLangs []language.Tag
	mu             sync.RWMutex
}

// New creates a new i18n service with embedded locales.
// The service is initialized with the system language or English as fallback.
func New() (*Service, error) {
	return NewWithFS(localeFS, "locales")
}

// NewWithFS creates a new i18n service loading locales from the given filesystem.
// This allows the GUI to provide its own locale files.
func NewWithFS(fsys fs.FS, dir string) (*Service, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	availableLangs, err := loadLocalesFromFS(bundle, fsys, dir)
	if err != nil {
		return nil, err
	}

	s := &Service{
		bundle:         bundle,
		availableLangs: availableLangs,
		currentLang:    "en",
	}

	// Try to detect system language
	if detected, err := detectLanguage(availableLangs); err == nil && detected != "" {
		_ = s.SetLanguage(detected)
	} else {
		_ = s.SetLanguage("en")
	}

	return s, nil
}

// NewWithBundle creates a service from an existing bundle.
// Useful for extending the CLI i18n with GUI-specific translations.
func NewWithBundle(bundle *i18n.Bundle, langs []language.Tag) *Service {
	s := &Service{
		bundle:         bundle,
		availableLangs: langs,
		currentLang:    "en",
	}
	_ = s.SetLanguage("en")
	return s
}

// Init initializes the default global service.
// Safe to call multiple times; only the first call has effect.
func Init() error {
	defaultOnce.Do(func() {
		defaultService, defaultErr = New()
	})
	return defaultErr
}

// Default returns the global i18n service, initializing if needed.
func Default() *Service {
	if defaultService == nil {
		_ = Init()
	}
	return defaultService
}

// SetDefault sets the global i18n service.
// Useful for GUI to replace with an extended service.
func SetDefault(s *Service) {
	defaultService = s
}

// T translates a message using the default service.
// Shorthand for Default().T(messageID, args...).
func T(messageID string, args ...interface{}) string {
	return Default().T(messageID, args...)
}

// --- Language Management ---

func loadLocalesFromFS(bundle *i18n.Bundle, fsys fs.FS, dir string) ([]language.Tag, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}

	var langs []language.Tag
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		if _, err := bundle.LoadMessageFileFS(fsys, filePath); err != nil {
			return nil, fmt.Errorf("failed to load locale %s: %w", entry.Name(), err)
		}

		lang := strings.TrimSuffix(entry.Name(), ".json")
		tag := language.Make(lang)
		langs = append(langs, tag)
	}

	if len(langs) == 0 {
		return nil, fmt.Errorf("no locale files found in %s", dir)
	}

	return langs, nil
}

func detectLanguage(supported []language.Tag) (string, error) {
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		// Try LC_ALL, LC_MESSAGES as fallbacks
		langEnv = os.Getenv("LC_ALL")
		if langEnv == "" {
			langEnv = os.Getenv("LC_MESSAGES")
		}
	}
	if langEnv == "" {
		return "", nil
	}

	// Parse LANG format: en_GB.UTF-8 -> en-GB
	baseLang := strings.Split(langEnv, ".")[0]
	baseLang = strings.ReplaceAll(baseLang, "_", "-")

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

// SetLanguage sets the language for translations.
// The language tag should be a valid BCP 47 tag (e.g., "en", "en-GB", "de").
func (s *Service) SetLanguage(lang string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	requestedLang, err := language.Parse(lang)
	if err != nil {
		return fmt.Errorf("invalid language tag %q: %w", lang, err)
	}

	if len(s.availableLangs) == 0 {
		return fmt.Errorf("no languages available")
	}

	matcher := language.NewMatcher(s.availableLangs)
	bestMatch, _, confidence := matcher.Match(requestedLang)

	if confidence == language.No {
		return fmt.Errorf("unsupported language: %s", lang)
	}

	s.localizer = i18n.NewLocalizer(s.bundle, bestMatch.String())
	s.currentLang = bestMatch.String()
	return nil
}

// Language returns the current language code.
func (s *Service) Language() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentLang
}

// AvailableLanguages returns the list of available language codes.
func (s *Service) AvailableLanguages() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	langs := make([]string, len(s.availableLangs))
	for i, tag := range s.availableLangs {
		langs[i] = tag.String()
	}
	return langs
}

// T translates a message by its ID.
// Optional template data can be passed for interpolation.
//
// Examples:
//
//	svc.T("cli.success")
//	svc.T("cli.items_found", map[string]int{"Count": 5})
//	svc.T("cli.greeting", map[string]string{"Name": "Alice"})
func (s *Service) T(messageID string, args ...interface{}) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.localizer == nil {
		return messageID
	}

	config := &i18n.LocalizeConfig{MessageID: messageID}
	if len(args) > 0 {
		config.TemplateData = args[0]
	}

	translation, err := s.localizer.Localize(config)
	if err != nil {
		// Return the message ID if translation not found
		return messageID
	}
	return translation
}

// Translate is an alias for T.
func (s *Service) Translate(messageID string, args ...interface{}) string {
	return s.T(messageID, args...)
}

// MustT translates a message, panicking if not found.
// Use sparingly, mainly for critical messages that must exist.
func (s *Service) MustT(messageID string, args ...interface{}) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.localizer == nil {
		panic(fmt.Sprintf("i18n: localizer not initialized for message %q", messageID))
	}

	config := &i18n.LocalizeConfig{MessageID: messageID}
	if len(args) > 0 {
		config.TemplateData = args[0]
	}

	translation, err := s.localizer.Localize(config)
	if err != nil {
		panic(fmt.Sprintf("i18n: translation not found for %q: %v", messageID, err))
	}
	return translation
}

// Bundle returns the underlying i18n.Bundle.
// Useful for extending with additional translations.
func (s *Service) Bundle() *i18n.Bundle {
	return s.bundle
}

// AddMessages adds additional messages to the bundle.
// This allows runtime extension of translations.
func (s *Service) AddMessages(lang string, messages map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag := language.Make(lang)
	var i18nMessages []*i18n.Message
	for id, text := range messages {
		i18nMessages = append(i18nMessages, &i18n.Message{
			ID:    id,
			Other: text,
		})
	}

	if err := s.bundle.AddMessages(tag, i18nMessages...); err != nil {
		return fmt.Errorf("failed to add messages for %s: %w", lang, err)
	}

	// Check if this is a new language
	found := false
	for _, existing := range s.availableLangs {
		if existing == tag {
			found = true
			break
		}
	}
	if !found {
		s.availableLangs = append(s.availableLangs, tag)
	}

	return nil
}

// LoadFS loads additional locale files from a filesystem.
// Useful for GUI to add its translations on top of CLI translations.
func (s *Service) LoadFS(fsys fs.FS, dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newLangs, err := loadLocalesFromFS(s.bundle, fsys, dir)
	if err != nil {
		return err
	}

	// Merge new languages
	for _, newTag := range newLangs {
		found := false
		for _, existing := range s.availableLangs {
			if existing == newTag {
				found = true
				break
			}
		}
		if !found {
			s.availableLangs = append(s.availableLangs, newTag)
		}
	}

	return nil
}
