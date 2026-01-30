// Package i18n provides internationalization for the CLI.
package i18n

import (
	"embed"
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

// templateCache stores compiled templates for reuse.
// Key is the template string, value is the compiled template.
var templateCache sync.Map

//go:embed locales/*.json
var localeFS embed.FS

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

// NumberFormat defines locale-specific number formatting rules.
type NumberFormat struct {
	ThousandsSep string // "," for en, "." for de
	DecimalSep   string // "." for en, "," for de
	PercentFmt   string // "%s%%" for en, "%s %%" for de (space before %)
}

// Default number formats by language.
var numberFormats = map[string]NumberFormat{
	"en": {ThousandsSep: ",", DecimalSep: ".", PercentFmt: "%s%%"},
	"de": {ThousandsSep: ".", DecimalSep: ",", PercentFmt: "%s %%"},
	"fr": {ThousandsSep: " ", DecimalSep: ",", PercentFmt: "%s %%"},
	"es": {ThousandsSep: ".", DecimalSep: ",", PercentFmt: "%s%%"},
	"zh": {ThousandsSep: ",", DecimalSep: ".", PercentFmt: "%s%%"},
}

// Mode determines how the i18n service handles missing translation keys.
type Mode int

const (
	// ModeNormal returns the key as-is when a translation is missing (production).
	ModeNormal Mode = iota
	// ModeStrict panics immediately when a translation is missing (dev/CI).
	ModeStrict
	// ModeCollect dispatches MissingKey actions and returns [key] (QA testing).
	ModeCollect
)

// Subject represents a typed subject with metadata for semantic translations.
// Use S() to create a Subject and chain methods for additional context.
//
//	S("file", "config.yaml")
//	S("repo", "core-php").Count(3)
//	S("user", user).Gender("feminine")
//	S("colleague", name).Formal()
type Subject struct {
	Noun      string    // The noun type (e.g., "file", "repo", "user")
	Value     any       // The actual value (e.g., filename, struct, etc.)
	count     int       // Count for pluralization (default 1)
	gender    string    // Grammatical gender for languages that need it
	location  string    // Location context (e.g., "in workspace")
	formality Formality // Formality level override (-1 = use service default)
}

// IntentMeta defines the behaviour and characteristics of an intent.
type IntentMeta struct {
	Type      string   // "action", "question", "info"
	Verb      string   // Reference to verb key (e.g., "delete", "save")
	Dangerous bool     // If true, requires extra confirmation
	Default   string   // Default response: "yes" or "no"
	Supports  []string // Extra options supported by this intent
}

// Composed holds all output forms for an intent after template resolution.
// Each field is ready to display to the user.
type Composed struct {
	Question string     // Question form: "Delete config.yaml?"
	Confirm  string     // Confirmation form: "Really delete config.yaml?"
	Success  string     // Success message: "config.yaml deleted"
	Failure  string     // Failure message: "Failed to delete config.yaml"
	Meta     IntentMeta // Intent metadata for UI decisions
}

// Intent defines a semantic intent with templates for all output forms.
// Templates use Go text/template syntax with Subject data available.
type Intent struct {
	Meta     IntentMeta // Intent behaviour and characteristics
	Question string     // Template for question form
	Confirm  string     // Template for confirmation form
	Success  string     // Template for success message
	Failure  string     // Template for failure message
}

// templateData is passed to intent templates during execution.
type templateData struct {
	Subject   string    // Display value of subject
	Noun      string    // Noun type
	Count     int       // Count for pluralization
	Gender    string    // Grammatical gender
	Location  string    // Location context
	Formality Formality // Formality level
	IsFormal  bool      // Convenience: formality == FormalityFormal
	IsPlural  bool      // Convenience: count != 1
	Value     any       // Raw value (for complex templates)
}

// GrammarData holds language-specific grammar forms loaded from JSON.
type GrammarData struct {
	Verbs    map[string]VerbForms // verb -> forms
	Nouns    map[string]NounForms // noun -> forms
	Articles ArticleForms         // article configuration
	Words    map[string]string    // base word translations
	Punct    PunctuationRules     // language-specific punctuation
}

// PunctuationRules holds language-specific punctuation patterns.
// French uses " :" (space before colon), English uses ":"
type PunctuationRules struct {
	LabelSuffix    string // Suffix for labels (default ":")
	ProgressSuffix string // Suffix for progress (default "...")
}

// NounForms holds plural and gender information for a noun.
type NounForms struct {
	One    string // Singular form
	Other  string // Plural form
	Gender string // Grammatical gender (masculine, feminine, neuter, common)
}

// ArticleForms holds article configuration for a language.
type ArticleForms struct {
	IndefiniteDefault string            // Default indefinite article (e.g., "a")
	IndefiniteVowel   string            // Indefinite article before vowel sounds (e.g., "an")
	Definite          string            // Definite article (e.g., "the")
	ByGender          map[string]string // Gender-specific articles for gendered languages
}

// grammarCache holds loaded grammar data per language.
var (
	grammarCache   = make(map[string]*GrammarData)
	grammarCacheMu sync.RWMutex
)

// VerbForms holds irregular verb conjugations.
type VerbForms struct {
	Past   string // Past tense (e.g., "deleted")
	Gerund string // Present participle (e.g., "deleting")
}

// irregularVerbs maps base verbs to their irregular forms.
var irregularVerbs = map[string]VerbForms{
	"be":        {Past: "was", Gerund: "being"},
	"have":      {Past: "had", Gerund: "having"},
	"do":        {Past: "did", Gerund: "doing"},
	"go":        {Past: "went", Gerund: "going"},
	"make":      {Past: "made", Gerund: "making"},
	"get":       {Past: "got", Gerund: "getting"},
	"run":       {Past: "ran", Gerund: "running"},
	"set":       {Past: "set", Gerund: "setting"},
	"put":       {Past: "put", Gerund: "putting"},
	"cut":       {Past: "cut", Gerund: "cutting"},
	"let":       {Past: "let", Gerund: "letting"},
	"hit":       {Past: "hit", Gerund: "hitting"},
	"shut":      {Past: "shut", Gerund: "shutting"},
	"split":     {Past: "split", Gerund: "splitting"},
	"spread":    {Past: "spread", Gerund: "spreading"},
	"read":      {Past: "read", Gerund: "reading"},
	"write":     {Past: "wrote", Gerund: "writing"},
	"send":      {Past: "sent", Gerund: "sending"},
	"build":     {Past: "built", Gerund: "building"},
	"begin":     {Past: "began", Gerund: "beginning"},
	"find":      {Past: "found", Gerund: "finding"},
	"take":      {Past: "took", Gerund: "taking"},
	"see":       {Past: "saw", Gerund: "seeing"},
	"keep":      {Past: "kept", Gerund: "keeping"},
	"hold":      {Past: "held", Gerund: "holding"},
	"tell":      {Past: "told", Gerund: "telling"},
	"bring":     {Past: "brought", Gerund: "bringing"},
	"think":     {Past: "thought", Gerund: "thinking"},
	"buy":       {Past: "bought", Gerund: "buying"},
	"catch":     {Past: "caught", Gerund: "catching"},
	"teach":     {Past: "taught", Gerund: "teaching"},
	"throw":     {Past: "threw", Gerund: "throwing"},
	"grow":      {Past: "grew", Gerund: "growing"},
	"know":      {Past: "knew", Gerund: "knowing"},
	"show":      {Past: "showed", Gerund: "showing"},
	"draw":      {Past: "drew", Gerund: "drawing"},
	"break":     {Past: "broke", Gerund: "breaking"},
	"speak":     {Past: "spoke", Gerund: "speaking"},
	"choose":    {Past: "chose", Gerund: "choosing"},
	"forget":    {Past: "forgot", Gerund: "forgetting"},
	"lose":      {Past: "lost", Gerund: "losing"},
	"win":       {Past: "won", Gerund: "winning"},
	"swim":      {Past: "swam", Gerund: "swimming"},
	"drive":     {Past: "drove", Gerund: "driving"},
	"rise":      {Past: "rose", Gerund: "rising"},
	"shine":     {Past: "shone", Gerund: "shining"},
	"sing":      {Past: "sang", Gerund: "singing"},
	"ring":      {Past: "rang", Gerund: "ringing"},
	"drink":     {Past: "drank", Gerund: "drinking"},
	"sink":      {Past: "sank", Gerund: "sinking"},
	"sit":       {Past: "sat", Gerund: "sitting"},
	"stand":     {Past: "stood", Gerund: "standing"},
	"hang":      {Past: "hung", Gerund: "hanging"},
	"dig":       {Past: "dug", Gerund: "digging"},
	"stick":     {Past: "stuck", Gerund: "sticking"},
	"bite":      {Past: "bit", Gerund: "biting"},
	"hide":      {Past: "hid", Gerund: "hiding"},
	"feed":      {Past: "fed", Gerund: "feeding"},
	"meet":      {Past: "met", Gerund: "meeting"},
	"lead":      {Past: "led", Gerund: "leading"},
	"sleep":     {Past: "slept", Gerund: "sleeping"},
	"feel":      {Past: "felt", Gerund: "feeling"},
	"leave":     {Past: "left", Gerund: "leaving"},
	"mean":      {Past: "meant", Gerund: "meaning"},
	"lend":      {Past: "lent", Gerund: "lending"},
	"spend":     {Past: "spent", Gerund: "spending"},
	"bend":      {Past: "bent", Gerund: "bending"},
	"deal":      {Past: "dealt", Gerund: "dealing"},
	"lay":       {Past: "laid", Gerund: "laying"},
	"pay":       {Past: "paid", Gerund: "paying"},
	"say":       {Past: "said", Gerund: "saying"},
	"sell":      {Past: "sold", Gerund: "selling"},
	"seek":      {Past: "sought", Gerund: "seeking"},
	"fight":     {Past: "fought", Gerund: "fighting"},
	"fly":       {Past: "flew", Gerund: "flying"},
	"wear":      {Past: "wore", Gerund: "wearing"},
	"tear":      {Past: "tore", Gerund: "tearing"},
	"bear":      {Past: "bore", Gerund: "bearing"},
	"swear":     {Past: "swore", Gerund: "swearing"},
	"wake":      {Past: "woke", Gerund: "waking"},
	"freeze":    {Past: "froze", Gerund: "freezing"},
	"steal":     {Past: "stole", Gerund: "stealing"},
	"overwrite": {Past: "overwritten", Gerund: "overwriting"},
	"reset":     {Past: "reset", Gerund: "resetting"},
	"reboot":    {Past: "rebooted", Gerund: "rebooting"},

	// Multi-syllable verbs with stressed final syllables (double consonant)
	"submit":   {Past: "submitted", Gerund: "submitting"},
	"permit":   {Past: "permitted", Gerund: "permitting"},
	"admit":    {Past: "admitted", Gerund: "admitting"},
	"omit":     {Past: "omitted", Gerund: "omitting"},
	"commit":   {Past: "committed", Gerund: "committing"},
	"transmit": {Past: "transmitted", Gerund: "transmitting"},
	"prefer":   {Past: "preferred", Gerund: "preferring"},
	"refer":    {Past: "referred", Gerund: "referring"},
	"transfer": {Past: "transferred", Gerund: "transferring"},
	"defer":    {Past: "deferred", Gerund: "deferring"},
	"confer":   {Past: "conferred", Gerund: "conferring"},
	"infer":    {Past: "inferred", Gerund: "inferring"},
	"occur":    {Past: "occurred", Gerund: "occurring"},
	"recur":    {Past: "recurred", Gerund: "recurring"},
	"incur":    {Past: "incurred", Gerund: "incurring"},
	"deter":    {Past: "deterred", Gerund: "deterring"},
	"control":  {Past: "controlled", Gerund: "controlling"},
	"patrol":   {Past: "patrolled", Gerund: "patrolling"},
	"compel":   {Past: "compelled", Gerund: "compelling"},
	"expel":    {Past: "expelled", Gerund: "expelling"},
	"propel":   {Past: "propelled", Gerund: "propelling"},
	"repel":    {Past: "repelled", Gerund: "repelling"},
	"rebel":    {Past: "rebelled", Gerund: "rebelling"},
	"excel":    {Past: "excelled", Gerund: "excelling"},
	"cancel":   {Past: "cancelled", Gerund: "cancelling"}, // UK spelling
	"travel":   {Past: "travelled", Gerund: "travelling"}, // UK spelling
	"label":    {Past: "labelled", Gerund: "labelling"},   // UK spelling
	"model":    {Past: "modelled", Gerund: "modelling"},   // UK spelling
	"level":    {Past: "levelled", Gerund: "levelling"},   // UK spelling
}

// noDoubleConsonant contains multi-syllable verbs that don't double the final consonant.
// Note: UK English doubles -l (travelled, cancelled) - those are in irregularVerbs.
var noDoubleConsonant = map[string]bool{
	"open":    true,
	"listen":  true,
	"happen":  true,
	"enter":   true,
	"offer":   true,
	"suffer":  true,
	"differ":  true,
	"cover":   true,
	"deliver": true,
	"develop": true,
	"visit":   true,
	"limit":   true,
	"edit":    true,
	"credit":  true,
	"orbit":   true,
	"total":   true,
	"target":  true,
	"budget":  true,
	"market":  true,
	"benefit": true,
	"focus":   true,
}

// irregularNouns maps singular nouns to their irregular plural forms.
var irregularNouns = map[string]string{
	"child":       "children",
	"person":      "people",
	"man":         "men",
	"woman":       "women",
	"foot":        "feet",
	"tooth":       "teeth",
	"mouse":       "mice",
	"goose":       "geese",
	"ox":          "oxen",
	"index":       "indices",
	"appendix":    "appendices",
	"matrix":      "matrices",
	"vertex":      "vertices",
	"crisis":      "crises",
	"analysis":    "analyses",
	"diagnosis":   "diagnoses",
	"thesis":      "theses",
	"hypothesis":  "hypotheses",
	"parenthesis": "parentheses",
	"datum":       "data",
	"medium":      "media",
	"bacterium":   "bacteria",
	"criterion":   "criteria",
	"phenomenon":  "phenomena",
	"curriculum":  "curricula",
	"alumnus":     "alumni",
	"cactus":      "cacti",
	"focus":       "foci",
	"fungus":      "fungi",
	"nucleus":     "nuclei",
	"radius":      "radii",
	"stimulus":    "stimuli",
	"syllabus":    "syllabi",
	"fish":        "fish",
	"sheep":       "sheep",
	"deer":        "deer",
	"species":     "species",
	"series":      "series",
	"aircraft":    "aircraft",
	"life":        "lives",
	"wife":        "wives",
	"knife":       "knives",
	"leaf":        "leaves",
	"half":        "halves",
	"self":        "selves",
	"shelf":       "shelves",
	"wolf":        "wolves",
	"calf":        "calves",
	"loaf":        "loaves",
	"thief":       "thieves",
}

// vowelSounds contains words that start with consonants but have vowel sounds.
// These take "an" instead of "a".
var vowelSounds = map[string]bool{
	"hour":   true,
	"honest": true,
	"honour": true,
	"honor":  true,
	"heir":   true,
	"herb":   true, // US pronunciation
}

// consonantSounds contains words that start with vowels but have consonant sounds.
// These take "a" instead of "an".
var consonantSounds = map[string]bool{
	"user":       true, // "yoo-zer"
	"union":      true, // "yoon-yon"
	"unique":     true,
	"unit":       true,
	"universe":   true,
	"university": true,
	"uniform":    true,
	"usage":      true,
	"usual":      true,
	"utility":    true,
	"utensil":    true,
	"one":        true, // "wun"
	"once":       true,
	"euro":       true, // "yoo-ro"
	"eulogy":     true,
	"euphemism":  true,
}

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
	CallerFile string         // Source file where T() was called
	CallerLine int            // Line number where T() was called
}

// PluralRule is a function that determines the plural category for a count.
// Each language has its own plural rule based on CLDR data.
//
//	rule := i18n.GetPluralRule("ru")
//	category := rule(5) // Returns PluralMany for Russian
type PluralRule func(n int) PluralCategory

// PluralCategory represents CLDR plural categories.
// Different languages use different subsets of these categories.
//
// Examples:
//   - English: one, other
//   - Russian: one, few, many, other
//   - Arabic: zero, one, two, few, many, other
//   - Welsh: zero, one, two, few, many, other
type PluralCategory int

const (
	// PluralOther is the default/fallback category
	PluralOther PluralCategory = iota
	// PluralZero is used when count == 0 (Arabic, Latvian, etc.)
	PluralZero
	// PluralOne is used when count == 1 (most languages)
	PluralOne
	// PluralTwo is used when count == 2 (Arabic, Welsh, etc.)
	PluralTwo
	// PluralFew is used for small numbers (Slavic: 2-4, Arabic: 3-10, etc.)
	PluralFew
	// PluralMany is used for larger numbers (Slavic: 5+, Arabic: 11-99, etc.)
	PluralMany
)

// GrammaticalGender represents grammatical gender for nouns.
type GrammaticalGender int

const (
	// GenderNeuter is used for neuter nouns (das in German, it in English)
	GenderNeuter GrammaticalGender = iota
	// GenderMasculine is used for masculine nouns (der in German, le in French)
	GenderMasculine
	// GenderFeminine is used for feminine nouns (die in German, la in French)
	GenderFeminine
	// GenderCommon is used in languages with common gender (Swedish, Dutch)
	GenderCommon
)

// rtlLanguages contains language codes that use right-to-left text direction.
var rtlLanguages = map[string]bool{
	"ar":    true, // Arabic
	"ar-SA": true,
	"ar-EG": true,
	"he":    true, // Hebrew
	"he-IL": true,
	"fa":    true, // Persian/Farsi
	"fa-IR": true,
	"ur":    true, // Urdu
	"ur-PK": true,
	"yi":    true, // Yiddish
	"ps":    true, // Pashto
	"sd":    true, // Sindhi
	"ug":    true, // Uyghur
}

// pluralRules contains CLDR plural rules for supported languages.
var pluralRules = map[string]PluralRule{
	"en":    pluralRuleEnglish,
	"en-GB": pluralRuleEnglish,
	"en-US": pluralRuleEnglish,
	"de":    pluralRuleGerman,
	"de-DE": pluralRuleGerman,
	"de-AT": pluralRuleGerman,
	"de-CH": pluralRuleGerman,
	"fr":    pluralRuleFrench,
	"fr-FR": pluralRuleFrench,
	"fr-CA": pluralRuleFrench,
	"es":    pluralRuleSpanish,
	"es-ES": pluralRuleSpanish,
	"es-MX": pluralRuleSpanish,
	"ru":    pluralRuleRussian,
	"ru-RU": pluralRuleRussian,
	"pl":    pluralRulePolish,
	"pl-PL": pluralRulePolish,
	"ar":    pluralRuleArabic,
	"ar-SA": pluralRuleArabic,
	"zh":    pluralRuleChinese,
	"zh-CN": pluralRuleChinese,
	"zh-TW": pluralRuleChinese,
	"ja":    pluralRuleJapanese,
	"ja-JP": pluralRuleJapanese,
	"ko":    pluralRuleKorean,
	"ko-KR": pluralRuleKorean,
}

// Formality represents the level of formality in translations.
// Used for languages that distinguish formal/informal address (Sie/du, vous/tu).
type Formality int

const (
	// FormalityNeutral uses context-appropriate formality (default)
	FormalityNeutral Formality = iota
	// FormalityInformal uses informal address (du, tu, you)
	FormalityInformal
	// FormalityFormal uses formal address (Sie, vous, usted)
	FormalityFormal
)

// TextDirection represents text directionality.
type TextDirection int

const (
	// DirLTR is left-to-right text direction (English, German, etc.)
	DirLTR TextDirection = iota
	// DirRTL is right-to-left text direction (Arabic, Hebrew, etc.)
	DirRTL
)

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
