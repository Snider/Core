// Package i18n provides internationalization for the CLI.
//
// Locale files use nested JSON for compatibility with translation tools:
//
//	{
//	    "cli": {
//	        "success": "Operation completed",
//	        "count": {
//	            "items": {
//	                "one": "{{.Count}} item",
//	                "other": "{{.Count}} items"
//	            }
//	        }
//	    }
//	}
//
// Keys are accessed with dot notation: T("cli.success"), T("cli.count.items")
//
// # Getting Started
//
//	svc, err := i18n.New()
//	fmt.Println(svc.T("cli.success"))
//	fmt.Println(svc.T("cli.count.items", map[string]any{"Count": 5}))
package i18n

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

// Message represents a translation - either a simple string or plural forms.
// Supports full CLDR plural categories for languages with complex plural rules.
type Message struct {
	Text  string // Simple string value (non-plural)
	Zero  string // count == 0 (Arabic, Latvian, Welsh)
	One   string // count == 1 (most languages)
	Two   string // count == 2 (Arabic, Welsh)
	Few   string // Small numbers (Slavic: 2-4, Arabic: 3-10)
	Many  string // Larger numbers (Slavic: 5+, Arabic: 11-99)
	Other string // Default/fallback form
}

// IsPlural returns true if this message has any plural forms.
func (m Message) IsPlural() bool {
	return m.Zero != "" || m.One != "" || m.Two != "" ||
		m.Few != "" || m.Many != "" || m.Other != ""
}

// ForCategory returns the appropriate text for a plural category.
// Falls back through the category hierarchy to find a non-empty string.
func (m Message) ForCategory(cat PluralCategory) string {
	switch cat {
	case PluralZero:
		if m.Zero != "" {
			return m.Zero
		}
	case PluralOne:
		if m.One != "" {
			return m.One
		}
	case PluralTwo:
		if m.Two != "" {
			return m.Two
		}
	case PluralFew:
		if m.Few != "" {
			return m.Few
		}
	case PluralMany:
		if m.Many != "" {
			return m.Many
		}
	}
	// Fallback to Other, then One, then Text
	if m.Other != "" {
		return m.Other
	}
	if m.One != "" {
		return m.One
	}
	return m.Text
}

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

// flatten recursively flattens nested maps into dot-notation keys.
func flatten(prefix string, data map[string]any, out map[string]Message) {
	flattenWithGrammar(prefix, data, out, nil)
}

// flattenWithGrammar recursively flattens nested maps and extracts grammar data.
func flattenWithGrammar(prefix string, data map[string]any, out map[string]Message, grammar *GrammarData) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			out[fullKey] = Message{Text: v}

		case map[string]any:
			// Check if this is a verb form object
			// Grammar data lives under "gram.*" (a nod to Gram - grandmother)
			if grammar != nil && isVerbFormObject(v) {
				verbName := key
				if strings.HasPrefix(fullKey, "gram.verb.") {
					verbName = strings.TrimPrefix(fullKey, "gram.verb.")
				}
				forms := VerbForms{}
				if base, ok := v["base"].(string); ok {
					_ = base // base form stored but not used in VerbForms
				}
				if past, ok := v["past"].(string); ok {
					forms.Past = past
				}
				if gerund, ok := v["gerund"].(string); ok {
					forms.Gerund = gerund
				}
				grammar.Verbs[strings.ToLower(verbName)] = forms
				continue
			}

			// Check if this is a noun form object
			if grammar != nil && isNounFormObject(v) {
				nounName := key
				if strings.HasPrefix(fullKey, "gram.noun.") {
					nounName = strings.TrimPrefix(fullKey, "gram.noun.")
				}
				forms := NounForms{}
				if one, ok := v["one"].(string); ok {
					forms.One = one
				}
				if other, ok := v["other"].(string); ok {
					forms.Other = other
				}
				if gender, ok := v["gender"].(string); ok {
					forms.Gender = gender
				}
				grammar.Nouns[strings.ToLower(nounName)] = forms
				continue
			}

			// Check if this is an article object
			if grammar != nil && fullKey == "gram.article" {
				if indef, ok := v["indefinite"].(map[string]any); ok {
					if def, ok := indef["default"].(string); ok {
						grammar.Articles.IndefiniteDefault = def
					}
					if vowel, ok := indef["vowel"].(string); ok {
						grammar.Articles.IndefiniteVowel = vowel
					}
				}
				if def, ok := v["definite"].(string); ok {
					grammar.Articles.Definite = def
				}
				continue
			}

			// Check if this is a punctuation rules object
			if grammar != nil && fullKey == "gram.punct" {
				if label, ok := v["label"].(string); ok {
					grammar.Punct.LabelSuffix = label
				}
				if progress, ok := v["progress"].(string); ok {
					grammar.Punct.ProgressSuffix = progress
				}
				continue
			}

			// Check if this is a base word in gram.word.*
			if grammar != nil && strings.HasPrefix(fullKey, "gram.word.") {
				wordKey := strings.TrimPrefix(fullKey, "gram.word.")
				// v could be a string or a nested object
				if str, ok := value.(string); ok {
					if grammar.Words == nil {
						grammar.Words = make(map[string]string)
					}
					grammar.Words[strings.ToLower(wordKey)] = str
				}
				continue
			}

			// Check if this is a plural object (has CLDR plural category keys)
			if isPluralObject(v) {
				msg := Message{}
				if zero, ok := v["zero"].(string); ok {
					msg.Zero = zero
				}
				if one, ok := v["one"].(string); ok {
					msg.One = one
				}
				if two, ok := v["two"].(string); ok {
					msg.Two = two
				}
				if few, ok := v["few"].(string); ok {
					msg.Few = few
				}
				if many, ok := v["many"].(string); ok {
					msg.Many = many
				}
				if other, ok := v["other"].(string); ok {
					msg.Other = other
				}
				out[fullKey] = msg
			} else {
				// Recurse into nested object
				flattenWithGrammar(fullKey, v, out, grammar)
			}
		}
	}
}

// isVerbFormObject checks if a map represents verb conjugation forms.
func isVerbFormObject(m map[string]any) bool {
	_, hasBase := m["base"]
	_, hasPast := m["past"]
	_, hasGerund := m["gerund"]
	return (hasBase || hasPast || hasGerund) && !isPluralObject(m)
}

// isNounFormObject checks if a map represents noun forms (with gender).
// Noun form objects have "gender" field, distinguishing them from CLDR plural objects.
func isNounFormObject(m map[string]any) bool {
	_, hasGender := m["gender"]
	// Only consider it a noun form if it has a gender field
	// This distinguishes noun forms from CLDR plural objects which use one/other
	return hasGender
}

// hasPluralCategories checks if a map has CLDR plural categories beyond one/other.
func hasPluralCategories(m map[string]any) bool {
	_, hasZero := m["zero"]
	_, hasTwo := m["two"]
	_, hasFew := m["few"]
	_, hasMany := m["many"]
	return hasZero || hasTwo || hasFew || hasMany
}

// isPluralObject checks if a map represents plural forms.
// Recognizes all CLDR plural categories: zero, one, two, few, many, other.
func isPluralObject(m map[string]any) bool {
	_, hasZero := m["zero"]
	_, hasOne := m["one"]
	_, hasTwo := m["two"]
	_, hasFew := m["few"]
	_, hasMany := m["many"]
	_, hasOther := m["other"]

	// It's a plural object if it has any plural category key
	if !hasZero && !hasOne && !hasTwo && !hasFew && !hasMany && !hasOther {
		return false
	}
	// But not if it contains nested objects (those are namespace containers)
	for _, v := range m {
		if _, isMap := v.(map[string]any); isMap {
			return false
		}
	}
	return true
}

func detectLanguage(supported []language.Tag) string {
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		langEnv = os.Getenv("LC_ALL")
		if langEnv == "" {
			langEnv = os.Getenv("LC_MESSAGES")
		}
	}
	if langEnv == "" {
		return ""
	}

	// Parse LANG format: en_GB.UTF-8 -> en-GB
	baseLang := strings.Split(langEnv, ".")[0]
	baseLang = strings.ReplaceAll(baseLang, "_", "-")

	parsedLang, err := language.Parse(baseLang)
	if err != nil {
		return ""
	}

	if len(supported) == 0 {
		return ""
	}

	matcher := language.NewMatcher(supported)
	bestMatch, _, confidence := matcher.Match(parsedLang)

	if confidence >= language.Low {
		return bestMatch.String()
	}
	return ""
}

// --- Global convenience functions ---

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

// SetFormality sets the default formality level on the default service.
//
//	SetFormality(FormalityFormal)  // Use formal address (Sie, vous)
func SetFormality(f Formality) {
	if svc := Default(); svc != nil {
		svc.SetFormality(f)
	}
}

// Direction returns the text direction for the current language.
func Direction() TextDirection {
	if svc := Default(); svc != nil {
		return svc.Direction()
	}
	return DirLTR
}

// IsRTL returns true if the current language uses right-to-left text.
func IsRTL() bool {
	return Direction() == DirRTL
}

// T translates a message using the default service.
// For semantic intents (core.* namespace), pass a Subject as the first argument.
//
//	T("cli.success")                           // Simple translation
//	T("core.delete", S("file", "config.yaml")) // Semantic intent
func T(messageID string, args ...any) string {
	if svc := Default(); svc != nil {
		return svc.T(messageID, args...)
	}
	return messageID
}

// C composes a semantic intent using the default service.
// Returns all output forms (Question, Confirm, Success, Failure) for the intent.
//
//	result := C("core.delete", S("file", "config.yaml"))
//	fmt.Println(result.Question) // "Delete config.yaml?"
func C(intent string, subject *Subject) *Composed {
	if svc := Default(); svc != nil {
		return svc.C(intent, subject)
	}
	return &Composed{
		Question: intent,
		Confirm:  intent,
		Success:  intent,
		Failure:  intent,
	}
}

// --- Grammar convenience functions (package-level) ---
// These provide direct access to grammar functions without needing a service instance.

// P returns a progress message for a verb: "Building...", "Checking..."
// Use this instead of T("cli.progress.building") for dynamic progress messages.
//
//	P("build")  // "Building..."
//	P("fetch")  // "Fetching..."
func P(verb string) string {
	return Progress(verb)
}

// PS returns a progress message with a subject: "Building project...", "Checking config..."
//
//	PS("build", "project")     // "Building project..."
//	PS("check", "config.yaml") // "Checking config.yaml..."
func PS(verb, subject string) string {
	return ProgressSubject(verb, subject)
}

// L returns a label with colon: "Status:", "Version:"
// Use this instead of T("common.label.status") for simple labels.
//
//	L("status")  // "Status:"
//	L("version") // "Version:"
func L(word string) string {
	return Label(word)
}

// _ is the raw gettext-style translation helper.
// Unlike T(), this does NOT handle core.* namespace magic.
// Use this for direct key lookups without auto-composition.
//
//	i18n._("cli.success")           // Raw lookup
//	i18n.T("i18n.label.status")     // Smart: returns "Status:"
func _(messageID string, args ...any) string {
	if svc := Default(); svc != nil {
		return svc.Raw(messageID, args...)
	}
	return messageID
}

// --- Service methods ---

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

// T translates a message by its ID with smart core.* namespace handling.
//
// # Core Namespace Magic
//
// The core.* namespace provides auto-composed grammar shortcuts:
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

// C composes a semantic intent with a subject.
// Returns all output forms (Question, Confirm, Success, Failure) for the intent.
//
//	result := svc.C("core.delete", S("file", "config.yaml"))
//	fmt.Println(result.Question) // "Delete config.yaml?"
//	fmt.Println(result.Success)  // "Config.yaml deleted"
func (s *Service) C(intent string, subject *Subject) *Composed {
	// Look up the intent definition
	intentDef := getIntent(intent)
	if intentDef == nil {
		// Intent not found, handle as missing key
		s.mu.RLock()
		mode := s.mode
		s.mu.RUnlock()

		switch mode {
		case ModeStrict:
			panic(fmt.Sprintf("i18n: missing intent %q", intent))
		case ModeCollect:
			dispatchMissingKey(intent, nil)
			return &Composed{
				Question: "[" + intent + "]",
				Confirm:  "[" + intent + "]",
				Success:  "[" + intent + "]",
				Failure:  "[" + intent + "]",
			}
		default:
			return &Composed{
				Question: intent,
				Confirm:  intent,
				Success:  intent,
				Failure:  intent,
			}
		}
	}

	// Create template data from subject
	data := newTemplateData(subject)

	result := &Composed{
		Question: executeIntentTemplate(intentDef.Question, data),
		Confirm:  executeIntentTemplate(intentDef.Confirm, data),
		Success:  executeIntentTemplate(intentDef.Success, data),
		Failure:  executeIntentTemplate(intentDef.Failure, data),
		Meta:     intentDef.Meta,
	}

	// Debug mode: prefix each form with the intent key
	s.mu.RLock()
	debug := s.debug
	s.mu.RUnlock()
	if debug {
		result.Question = debugFormat(intent, result.Question)
		result.Confirm = debugFormat(intent, result.Confirm)
		result.Success = debugFormat(intent, result.Success)
		result.Failure = debugFormat(intent, result.Failure)
	}

	return result
}

// templateCache stores compiled templates for reuse.
// Key is the template string, value is the compiled template.
var templateCache sync.Map

// executeIntentTemplate executes an intent template with the given data.
// Templates are cached for performance - repeated calls with the same template
// string will reuse the compiled template.
func executeIntentTemplate(tmplStr string, data templateData) string {
	if tmplStr == "" {
		return ""
	}

	// Check cache first
	if cached, ok := templateCache.Load(tmplStr); ok {
		var buf bytes.Buffer
		if err := cached.(*template.Template).Execute(&buf, data); err != nil {
			return tmplStr
		}
		return buf.String()
	}

	// Parse and cache
	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(tmplStr)
	if err != nil {
		return tmplStr
	}

	// Store in cache (safe even if another goroutine stored it first)
	templateCache.Store(tmplStr, tmpl)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return tmplStr
	}
	return buf.String()
}

// Raw is the raw translation helper without core.* namespace magic.
// Use T() for smart core.* handling, Raw() for direct key lookup.
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

func getCount(data any) int {
	if data == nil {
		return 0
	}
	switch d := data.(type) {
	case map[string]any:
		if c, ok := d["Count"]; ok {
			return toInt(c)
		}
	case map[string]int:
		if c, ok := d["Count"]; ok {
			return c
		}
	}
	return 0
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case int32:
		return int64(n)
	case float64:
		return int64(n)
	case float32:
		return int64(n)
	}
	return 0
}

func toFloat64(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	}
	return 0
}

func applyTemplate(text string, data any) string {
	// Quick check for template syntax
	if !strings.Contains(text, "{{") {
		return text
	}

	tmpl, err := template.New("").Parse(text)
	if err != nil {
		return text
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return text
	}
	return buf.String()
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
