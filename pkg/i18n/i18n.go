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
	"os"
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

// --- Global convenience functions ---

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

// --- JSON parsing helpers ---

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

// --- Template helpers ---

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
