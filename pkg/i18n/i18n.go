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
type Message struct {
	Text  string // Simple string value
	One   string // Singular form (count == 1)
	Other string // Plural form (count != 1)
}

// IsPlural returns true if this message has plural forms.
func (m Message) IsPlural() bool {
	return m.One != "" || m.Other != ""
}

// Service provides internationalization and localization.
type Service struct {
	messages       map[string]map[string]Message // lang -> key -> message
	currentLang    string
	fallbackLang   string
	availableLangs []language.Tag
	mode           Mode // Translation mode (Normal, Strict, Collect)
	debug          bool // Debug mode shows key prefixes
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
func (s *Service) loadJSON(lang string, data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	messages := make(map[string]Message)
	flatten("", raw, messages)
	s.messages[lang] = messages
	return nil
}

// flatten recursively flattens nested maps into dot-notation keys.
func flatten(prefix string, data map[string]any, out map[string]Message) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			out[fullKey] = Message{Text: v}

		case map[string]any:
			// Check if this is a plural object (has "one" or "other" keys)
			if isPluralObject(v) {
				msg := Message{}
				if one, ok := v["one"].(string); ok {
					msg.One = one
				}
				if other, ok := v["other"].(string); ok {
					msg.Other = other
				}
				out[fullKey] = msg
			} else {
				// Recurse into nested object
				flatten(fullKey, v, out)
			}
		}
	}
}

// isPluralObject checks if a map represents plural forms.
func isPluralObject(m map[string]any) bool {
	_, hasOne := m["one"]
	_, hasOther := m["other"]
	// It's a plural object if it has one/other and no nested objects
	if !hasOne && !hasOther {
		return false
	}
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

// SetDebug enables or disables debug mode on the default service.
// In debug mode, translations show their keys: [key] translation
//
//	SetDebug(true)
//	T("cli.success") // "[cli.success] Success"
func SetDebug(enabled bool) {
	if svc := Default(); svc != nil {
		svc.SetDebug(enabled)
	}
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

// _ is the standard gettext-style translation helper.
// Alias for T() - use whichever you prefer.
//
//	i18n._("cli.success")
//	i18n._("cli.greeting", map[string]any{"Name": "World"})
func _(messageID string, args ...any) string {
	return T(messageID, args...)
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

// SetDebug enables or disables debug mode.
// In debug mode, translations are prefixed with their key:
//
//	[cli.success] Success
//	[core.delete] Delete config.yaml?
func (s *Service) SetDebug(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debug = enabled
}

// Debug returns whether debug mode is enabled.
func (s *Service) Debug() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.debug
}

// T translates a message by its ID.
// Optional template data can be passed for interpolation.
//
// For plural messages, pass a map with "Count" to select the form:
//
//	svc.T("cli.count.items", map[string]any{"Count": 5})
//
// For semantic intents (core.* namespace), pass a Subject to get the Question form:
//
//	svc.T("core.delete", S("file", "config.yaml")) // "Delete config.yaml?"
func (s *Service) T(messageID string, args ...any) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check for semantic intent with Subject
	if strings.HasPrefix(messageID, "core.") && len(args) > 0 {
		if subject, ok := args[0].(*Subject); ok {
			// Use C() to resolve the intent, return Question form
			s.mu.RUnlock()
			result := s.C(messageID, subject)
			s.mu.RLock()
			return result.Question
		}
	}

	// Try current language, then fallback
	msg, ok := s.getMessage(s.currentLang, messageID)
	if !ok {
		msg, ok = s.getMessage(s.fallbackLang, messageID)
		if !ok {
			return s.handleMissingKey(messageID, args)
		}
	}

	// Get template data
	var data any
	if len(args) > 0 {
		data = args[0]
	}

	// Get the appropriate text
	text := msg.Text
	if msg.IsPlural() {
		count := getCount(data)
		if count == 1 {
			text = msg.One
		} else {
			text = msg.Other
		}
		if text == "" {
			text = msg.Other // Fallback to other
		}
	}

	if text == "" {
		return s.handleMissingKey(messageID, args)
	}

	// Apply template if we have data
	if data != nil {
		text = applyTemplate(text, data)
	}

	// Debug mode: prefix with key
	if s.debug {
		return "[" + messageID + "] " + text
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
		prefix := "[" + intent + "] "
		result.Question = prefix + result.Question
		result.Confirm = prefix + result.Confirm
		result.Success = prefix + result.Success
		result.Failure = prefix + result.Failure
	}

	return result
}

// executeIntentTemplate executes an intent template with the given data.
func executeIntentTemplate(tmplStr string, data templateData) string {
	if tmplStr == "" {
		return ""
	}

	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(tmplStr)
	if err != nil {
		return tmplStr
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return tmplStr
	}
	return buf.String()
}

// _ is the standard gettext-style translation helper. Alias for T().
func (s *Service) _(messageID string, args ...any) string {
	return s.T(messageID, args...)
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
