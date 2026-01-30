// Package i18n provides internationalization for the CLI.
package i18n

// Translator defines the interface for translation services.
// Implement this interface to provide custom translation backends
// or mock implementations for testing.
//
// Example usage in tests:
//
//	type mockTranslator struct {
//	    translations map[string]string
//	}
//
//	func (m *mockTranslator) T(key string, args ...any) string {
//	    if v, ok := m.translations[key]; ok {
//	        return v
//	    }
//	    return key
//	}
//
//	func TestSomething(t *testing.T) {
//	    mock := &mockTranslator{translations: map[string]string{
//	        "cli.success": "Test Success",
//	    }}
//	    // Use mock in your tests
//	}
type Translator interface {
	// T translates a message by its ID.
	// Optional template data can be passed for interpolation.
	//
	//	svc.T("cli.success")
	//	svc.T("cli.count.items", map[string]any{"Count": 5})
	T(messageID string, args ...any) string

	// C composes a semantic intent with a subject.
	// Returns all output forms (Question, Confirm, Success, Failure).
	//
	//	result := svc.C("core.delete", S("file", "config.yaml"))
	C(intent string, subject *Subject) *Composed

	// SetLanguage sets the language for translations.
	// Returns an error if the language is not supported.
	SetLanguage(lang string) error

	// Language returns the current language code.
	Language() string

	// SetMode sets the translation mode for missing key handling.
	SetMode(m Mode)

	// Mode returns the current translation mode.
	Mode() Mode

	// SetDebug enables or disables debug mode.
	SetDebug(enabled bool)

	// Debug returns whether debug mode is enabled.
	Debug() bool

	// SetFormality sets the default formality level for translations.
	SetFormality(f Formality)

	// Formality returns the current formality level.
	Formality() Formality

	// Direction returns the text direction for the current language.
	Direction() TextDirection

	// IsRTL returns true if the current language uses RTL text.
	IsRTL() bool

	// PluralCategory returns the plural category for a count.
	PluralCategory(n int) PluralCategory

	// AvailableLanguages returns the list of available language codes.
	AvailableLanguages() []string
}

// Ensure Service implements Translator at compile time.
var _ Translator = (*Service)(nil)

// --- Function type interfaces ---

// MissingKeyHandler receives missing key events for analysis.
// Used in ModeCollect to capture translation keys that need to be added.
//
//	i18n.OnMissingKey(func(m i18n.MissingKey) {
//	    log.Printf("MISSING: %s at %s:%d", m.Key, m.CallerFile, m.CallerLine)
//	})
type MissingKeyHandler func(missing MissingKey)

// MissingKey is dispatched when a translation key is not found in ModeCollect.
// Used by QA tools to collect and report missing translations.
type MissingKey struct {
	Key        string         // The missing translation key
	Args       map[string]any // Arguments passed to the translation
	CallerFile string         // Source file where T()/C() was called
	CallerLine int            // Line number where T()/C() was called
}

// MissingKeyAction is an alias for backwards compatibility.
// Deprecated: Use MissingKey instead.
type MissingKeyAction = MissingKey

// PluralRule is a function that determines the plural category for a count.
// Each language has its own plural rule based on CLDR data.
//
//	rule := i18n.GetPluralRule("ru")
//	category := rule(5) // Returns PluralMany for Russian
type PluralRule func(n int) PluralCategory

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
