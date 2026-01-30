// Package i18n provides internationalization for the CLI.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/text/language"
)

// Service provides internationalization and localization.
type Service struct {
	loader         Loader                        // Source for loading translations
	messages       map[string]map[string]Message // lang -> key -> message
	currentLang    string
	fallbackLang   string
	availableLangs []language.Tag
	mode           Mode         // Translation mode (Normal, Strict, Collect)
	debug          bool         // Debug mode shows key prefixes
	formality      Formality    // Default formality level for translations
	handlers       []KeyHandler // Handler chain for dynamic key patterns
	mu             sync.RWMutex
}

// Option configures a Service during construction.
type Option func(*Service)

// WithFallback sets the fallback language for missing translations.
func WithFallback(lang string) Option {
	return func(s *Service) {
		s.fallbackLang = lang
	}
}

// WithFormality sets the default formality level.
func WithFormality(f Formality) Option {
	return func(s *Service) {
		s.formality = f
	}
}

// WithHandlers sets custom handlers (replaces default handlers).
func WithHandlers(handlers ...KeyHandler) Option {
	return func(s *Service) {
		s.handlers = handlers
	}
}

// WithDefaultHandlers adds the default i18n.* namespace handlers.
// Use this after WithHandlers to add defaults back, or to ensure defaults are present.
func WithDefaultHandlers() Option {
	return func(s *Service) {
		s.handlers = append(s.handlers, DefaultHandlers()...)
	}
}

// WithMode sets the translation mode.
func WithMode(m Mode) Option {
	return func(s *Service) {
		s.mode = m
	}
}

// WithDebug enables or disables debug mode.
func WithDebug(enabled bool) Option {
	return func(s *Service) {
		s.debug = enabled
	}
}

// Default is the global i18n service instance.
var (
	defaultService atomic.Pointer[Service]
	defaultOnce    sync.Once
	defaultErr     error
)

//go:embed locales/*.json
var localeFS embed.FS

// Ensure Service implements Translator at compile time.
var _ Translator = (*Service)(nil)

// New creates a new i18n service with embedded locales and default options.
func New(opts ...Option) (*Service, error) {
	return NewWithLoader(NewFSLoader(localeFS, "locales"), opts...)
}

// NewWithFS creates a new i18n service loading locales from the given filesystem.
func NewWithFS(fsys fs.FS, dir string, opts ...Option) (*Service, error) {
	return NewWithLoader(NewFSLoader(fsys, dir), opts...)
}

// NewWithLoader creates a new i18n service with a custom loader.
// Use this for custom storage backends (database, remote API, etc.).
//
//	loader := NewFSLoader(customFS, "translations")
//	svc, err := NewWithLoader(loader, WithFallback("de-DE"))
func NewWithLoader(loader Loader, opts ...Option) (*Service, error) {
	s := &Service{
		loader:       loader,
		messages:     make(map[string]map[string]Message),
		fallbackLang: "en-GB",
		handlers:     DefaultHandlers(),
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Load all available languages
	langs := loader.Languages()
	if len(langs) == 0 {
		return nil, fmt.Errorf("no languages available from loader")
	}

	for _, lang := range langs {
		messages, grammar, err := loader.Load(lang)
		if err != nil {
			return nil, fmt.Errorf("failed to load locale %q: %w", lang, err)
		}

		s.messages[lang] = messages
		if grammar != nil && (len(grammar.Verbs) > 0 || len(grammar.Nouns) > 0 || len(grammar.Words) > 0) {
			SetGrammarData(lang, grammar)
		}

		tag := language.Make(lang)
		s.availableLangs = append(s.availableLangs, tag)
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
		svc, err := New()
		if err == nil {
			defaultService.Store(svc)
		}
		defaultErr = err
	})
	return defaultErr
}

// Default returns the global i18n service, initializing if needed.
// Thread-safe: can be called concurrently.
func Default() *Service {
	_ = Init() // sync.Once handles idempotency
	return defaultService.Load()
}

// SetDefault sets the global i18n service.
// Thread-safe: can be called concurrently with Default().
// Panics if s is nil.
func SetDefault(s *Service) {
	if s == nil {
		panic("i18n: SetDefault called with nil service")
	}
	defaultService.Store(s)
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
		return fmt.Errorf("unsupported language: %q", lang)
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

// AddHandler appends a handler to the end of the handler chain.
// Later handlers have lower priority (run if earlier handlers don't match).
func (s *Service) AddHandler(h KeyHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, h)
}

// PrependHandler inserts a handler at the start of the handler chain.
// Prepended handlers have highest priority (run first).
func (s *Service) PrependHandler(h KeyHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append([]KeyHandler{h}, s.handlers...)
}

// ClearHandlers removes all handlers from the chain.
// Useful for testing or disabling all i18n.* magic.
func (s *Service) ClearHandlers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = nil
}

// Handlers returns a copy of the current handler chain.
func (s *Service) Handlers() []KeyHandler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]KeyHandler, len(s.handlers))
	copy(result, s.handlers)
	return result
}

// T translates a message by its ID with handler chain support.
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
// Use Raw() for direct key lookup without handler chain processing.
func (s *Service) T(messageID string, args ...any) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Run handler chain - handlers can intercept and process keys
	result := RunHandlerChain(s.handlers, messageID, args, func() string {
		// Fallback: standard message lookup
		var data any
		if len(args) > 0 {
			data = args[0]
		}
		text := s.resolveWithFallback(messageID, data)
		if text == "" {
			return s.handleMissingKey(messageID, args)
		}
		return text
	})

	// Debug mode: prefix with key
	if s.debug {
		return debugFormat(messageID, result)
	}

	return result
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
	// Determine effective formality
	formality := s.getEffectiveFormality(data)

	// Try formality-specific key first (key._formal or key._informal)
	if formality != FormalityNeutral {
		formalityKey := key + "._" + formality.String()
		if text := s.resolveMessage(lang, formalityKey, data); text != "" {
			return text
		}
	}

	// Fall back to base key
	return s.resolveMessage(lang, key, data)
}

// resolveMessage resolves a single message key without formality fallback.
// Must be called with s.mu.RLock held.
func (s *Service) resolveMessage(lang, key string, data any) string {
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

// getEffectiveFormality returns the formality to use for translation.
// Priority: TranslationContext > Subject > map["Formality"] > Service.formality
// Must be called with s.mu.RLock held.
func (s *Service) getEffectiveFormality(data any) Formality {
	// Check if data is a TranslationContext with explicit formality
	if ctx, ok := data.(*TranslationContext); ok && ctx != nil {
		if ctx.Formality != FormalityNeutral {
			return ctx.Formality
		}
	}

	// Check if data is a Subject with explicit formality
	if subj, ok := data.(*Subject); ok && subj != nil {
		if subj.formality != FormalityNeutral {
			return subj.formality
		}
	}

	// Check if data is a map with Formality field
	if m, ok := data.(map[string]any); ok {
		if f, ok := m["Formality"].(Formality); ok && f != FormalityNeutral {
			return f
		}
	}

	// Fall back to service default
	return s.formality
}

// handleMissingKey handles a missing translation key based on the current mode.
// Must be called with s.mu.RLock held.
//
// In ModeStrict, this panics - use only in development/CI to catch missing keys.
// In ModeCollect, this dispatches to OnMissingKey handler for logging/collection.
// In ModeNormal (default), this returns the key as-is.
func (s *Service) handleMissingKey(key string, args []any) string {
	switch s.mode {
	case ModeStrict:
		// WARNING: Panics! Use ModeStrict only in development/CI environments.
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

// getMessage retrieves a message by language and key.
// Returns the message and true if found, or empty Message and false if not.
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
			return fmt.Errorf("failed to read locale %q: %w", entry.Name(), err)
		}

		lang := strings.TrimSuffix(entry.Name(), ".json")
		lang = strings.ReplaceAll(lang, "_", "-")

		if err := s.loadJSON(lang, data); err != nil {
			return fmt.Errorf("failed to parse locale %q: %w", entry.Name(), err)
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
