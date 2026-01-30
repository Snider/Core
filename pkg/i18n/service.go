// Package i18n provides internationalization for the CLI.
package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/text/language"
)

// Service provides internationalization and localization.
type Service struct {
	messages       map[string]map[string]Message // lang -> key -> message
	currentLang    string
	fallbackLang   string
	availableLangs []language.Tag
	mode           Mode      // Translation mode (Normal, Strict, Collect)
	debug          bool      // Debug mode shows key prefixes
	formality      Formality // Default formality level for translations
	mu             sync.RWMutex
}

// Default is the global i18n service instance.
var (
	defaultService *Service
	defaultOnce    sync.Once
	defaultErr     error
)

// New creates a new i18n service with embedded locales.
func New() (*Service, error) {
	return NewWithFS(localeFS, "locales")
}

// NewWithFS creates a new i18n service loading locales from the given filesystem.
func NewWithFS(fsys fs.FS, dir string) (*Service, error) {
	s := &Service{
		messages:     make(map[string]map[string]Message),
		fallbackLang: "en-GB",
	}

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read locale %s: %w", entry.Name(), err)
		}

		lang := strings.TrimSuffix(entry.Name(), ".json")
		// Normalise underscore to hyphen (en_GB -> en-GB)
		lang = strings.ReplaceAll(lang, "_", "-")

		if err := s.loadJSON(lang, data); err != nil {
			return nil, fmt.Errorf("failed to parse locale %s: %w", entry.Name(), err)
		}

		tag := language.Make(lang)
		s.availableLangs = append(s.availableLangs, tag)
	}

	if len(s.availableLangs) == 0 {
		return nil, fmt.Errorf("no locale files found in %s", dir)
	}

	// Try to detect system language
	if detected := detectLanguage(s.availableLangs); detected != "" {
		s.currentLang = detected
	} else {
		s.currentLang = s.fallbackLang
	}

	return s, nil
}

// Init initializes the default global service.
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
func SetDefault(s *Service) {
	defaultService = s
}

// loadJSON parses nested JSON and flattens to dot-notation keys.
// Also extracts grammar data (verbs, nouns, articles) for the language.
func (s *Service) loadJSON(lang string, data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	messages := make(map[string]Message)
	grammarData := &GrammarData{
		Verbs: make(map[string]VerbForms),
		Nouns: make(map[string]NounForms),
		Words: make(map[string]string),
	}

	flattenWithGrammar("", raw, messages, grammarData)
	s.messages[lang] = messages

	// Store grammar data if any was found
	if len(grammarData.Verbs) > 0 || len(grammarData.Nouns) > 0 || len(grammarData.Words) > 0 {
		SetGrammarData(lang, grammarData)
	}

	return nil
}

// SetLanguage sets the language for translations.
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

// SetMode sets the translation mode for missing key handling.
func (s *Service) SetMode(m Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = m
}

// Mode returns the current translation mode.
func (s *Service) Mode() Mode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

// SetFormality sets the default formality level for translations.
// This affects languages that distinguish formal/informal address (Sie/du, vous/tu).
//
//	svc.SetFormality(FormalityFormal)  // Use formal address
func (s *Service) SetFormality(f Formality) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.formality = f
}

// Formality returns the current formality level.
func (s *Service) Formality() Formality {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.formality
}

// Direction returns the text direction for the current language.
func (s *Service) Direction() TextDirection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if IsRTLLanguage(s.currentLang) {
		return DirRTL
	}
	return DirLTR
}

// IsRTL returns true if the current language uses right-to-left text direction.
func (s *Service) IsRTL() bool {
	return s.Direction() == DirRTL
}

// PluralCategory returns the plural category for a count in the current language.
func (s *Service) PluralCategory(n int) PluralCategory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return GetPluralCategory(s.currentLang, n)
}

// T translates a message by its ID with smart i18n.* namespace handling.
//
// # i18n Namespace Magic
//
// The i18n.* namespace provides auto-composed grammar shortcuts:
//
//	T("i18n.label.status")              // → "Status:"
//	T("i18n.progress.build")            // → "Building..."
//	T("i18n.progress.check", "config")  // → "Checking config..."
//	T("i18n.count.file", 5)             // → "5 files"
//	T("i18n.done.delete", "file")       // → "File deleted"
//	T("i18n.fail.delete", "file")       // → "Failed to delete file"
//
// For semantic intents, pass a Subject:
//
//	T("core.delete", S("file", "config.yaml")) // → "Delete config.yaml?"
//
// Use _() for raw key lookup without i18n.* magic.
func (s *Service) T(messageID string, args ...any) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Handle i18n.* namespace magic
	if strings.HasPrefix(messageID, "i18n.") {
		if result := s.handleI18nNamespace(messageID, args); result != "" {
			if s.debug {
				return debugFormat(messageID, result)
			}
			return result
		}
	}

	// Get template data
	var data any
	if len(args) > 0 {
		data = args[0]
	}

	// Try fallback chain
	text := s.resolveWithFallback(messageID, data)
	if text == "" {
		return s.handleMissingKey(messageID, args)
	}

	// Debug mode: prefix with key
	if s.debug {
		return debugFormat(messageID, text)
	}

	return text
}

// handleI18nNamespace processes i18n.* namespace patterns.
// Returns empty string if pattern not recognized.
// Must be called with s.mu.RLock held.
func (s *Service) handleI18nNamespace(key string, args []any) string {
	// i18n.label.{word} → Label(word)
	if strings.HasPrefix(key, "i18n.label.") {
		word := strings.TrimPrefix(key, "i18n.label.")
		return Label(word)
	}

	// i18n.progress.{verb} → Progress(verb) or ProgressSubject(verb, subj)
	if strings.HasPrefix(key, "i18n.progress.") {
		verb := strings.TrimPrefix(key, "i18n.progress.")
		if len(args) > 0 {
			if subj, ok := args[0].(string); ok {
				return ProgressSubject(verb, subj)
			}
		}
		return Progress(verb)
	}

	// i18n.count.{noun} → "N noun(s)"
	if strings.HasPrefix(key, "i18n.count.") {
		noun := strings.TrimPrefix(key, "i18n.count.")
		if len(args) > 0 {
			count := toInt(args[0])
			return fmt.Sprintf("%d %s", count, Pluralize(noun, count))
		}
		return noun
	}

	// i18n.done.{verb} → ActionResult(verb, subj)
	if strings.HasPrefix(key, "i18n.done.") {
		verb := strings.TrimPrefix(key, "i18n.done.")
		if len(args) > 0 {
			if subj, ok := args[0].(string); ok {
				return ActionResult(verb, subj)
			}
		}
		return Title(PastTense(verb))
	}

	// i18n.fail.{verb} → ActionFailed(verb, subj)
	if strings.HasPrefix(key, "i18n.fail.") {
		verb := strings.TrimPrefix(key, "i18n.fail.")
		if len(args) > 0 {
			if subj, ok := args[0].(string); ok {
				return ActionFailed(verb, subj)
			}
		}
		return ActionFailed(verb, "")
	}

	// i18n.number → FormatNumber(n)
	if key == "i18n.number" && len(args) > 0 {
		return FormatNumber(toInt64(args[0]))
	}

	// i18n.decimal → FormatDecimal(f)
	if key == "i18n.decimal" && len(args) > 0 {
		return FormatDecimal(toFloat64(args[0]))
	}

	// i18n.percent → FormatPercent(f)
	if key == "i18n.percent" && len(args) > 0 {
		return FormatPercent(toFloat64(args[0]))
	}

	// i18n.bytes → FormatBytes(n)
	if key == "i18n.bytes" && len(args) > 0 {
		return FormatBytes(toInt64(args[0]))
	}

	// i18n.ordinal → FormatOrdinal(n)
	if key == "i18n.ordinal" && len(args) > 0 {
		return FormatOrdinal(toInt(args[0]))
	}

	// i18n.ago → FormatAgo(count, unit)
	if key == "i18n.ago" && len(args) >= 2 {
		count := toInt(args[0])
		if unit, ok := args[1].(string); ok {
			return FormatAgo(count, unit)
		}
	}

	return ""
}

// resolveWithFallback implements the fallback chain for message resolution.
// Must be called with s.mu.RLock held.
func (s *Service) resolveWithFallback(messageID string, data any) string {
	// 1. Try exact key in current language
	if text := s.tryResolve(s.currentLang, messageID, data); text != "" {
		return text
	}

	// 2. Try exact key in fallback language
	if text := s.tryResolve(s.fallbackLang, messageID, data); text != "" {
		return text
	}

	// 3. Try fallback patterns for intent-like keys
	if strings.Contains(messageID, ".") {
		parts := strings.Split(messageID, ".")
		verb := parts[len(parts)-1]

		// Try common.action.{verb}
		commonKey := "common.action." + verb
		if text := s.tryResolve(s.currentLang, commonKey, data); text != "" {
			return text
		}
		if text := s.tryResolve(s.fallbackLang, commonKey, data); text != "" {
			return text
		}

		// Try common.{verb}
		commonKey = "common." + verb
		if text := s.tryResolve(s.currentLang, commonKey, data); text != "" {
			return text
		}
		if text := s.tryResolve(s.fallbackLang, commonKey, data); text != "" {
			return text
		}
	}

	return ""
}

// tryResolve attempts to resolve a single key in a single language.
// Returns empty string if not found.
// Must be called with s.mu.RLock held.
func (s *Service) tryResolve(lang, key string, data any) string {
	msg, ok := s.getMessage(lang, key)
	if !ok {
		return ""
	}

	text := msg.Text
	if msg.IsPlural() {
		count := getCount(data)
		category := GetPluralCategory(lang, count)
		text = msg.ForCategory(category)
	}

	if text == "" {
		return ""
	}

	// Apply template if we have data
	if data != nil {
		text = applyTemplate(text, data)
	}

	return text
}

// handleMissingKey handles a missing translation key based on the current mode.
// Must be called with s.mu.RLock held.
func (s *Service) handleMissingKey(key string, args []any) string {
	switch s.mode {
	case ModeStrict:
		panic(fmt.Sprintf("i18n: missing translation key %q", key))
	case ModeCollect:
		// Convert args to map for the action
		var argsMap map[string]any
		if len(args) > 0 {
			if m, ok := args[0].(map[string]any); ok {
				argsMap = m
			}
		}
		dispatchMissingKey(key, argsMap)
		return "[" + key + "]"
	default:
		return key
	}
}

// Raw is the raw translation helper without i18n.* namespace magic.
// Use T() for smart i18n.* handling, Raw() for direct key lookup.
func (s *Service) Raw(messageID string, args ...any) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var data any
	if len(args) > 0 {
		data = args[0]
	}

	text := s.resolveWithFallback(messageID, data)
	if text == "" {
		return s.handleMissingKey(messageID, args)
	}

	if s.debug {
		return debugFormat(messageID, text)
	}
	return text
}

func (s *Service) getMessage(lang, key string) (Message, bool) {
	msgs, ok := s.messages[lang]
	if !ok {
		return Message{}, false
	}
	msg, ok := msgs[key]
	return msg, ok
}

// AddMessages adds messages for a language at runtime.
func (s *Service) AddMessages(lang string, messages map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.messages[lang] == nil {
		s.messages[lang] = make(map[string]Message)
	}
	for key, text := range messages {
		s.messages[lang][key] = Message{Text: text}
	}
}

// LoadFS loads additional locale files from a filesystem.
func (s *Service) LoadFS(fsys fs.FS, dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return fmt.Errorf("failed to read locale %s: %w", entry.Name(), err)
		}

		lang := strings.TrimSuffix(entry.Name(), ".json")
		lang = strings.ReplaceAll(lang, "_", "-")

		if err := s.loadJSON(lang, data); err != nil {
			return fmt.Errorf("failed to parse locale %s: %w", entry.Name(), err)
		}

		// Add to available languages if new
		tag := language.Make(lang)
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
	}

	return nil
}
