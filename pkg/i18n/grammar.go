// Package i18n provides internationalization for the CLI.
package i18n

import (
	"strings"
	"sync"
	"text/template"
	"unicode"
)

// GrammarData holds language-specific grammar forms loaded from JSON.
type GrammarData struct {
	Verbs    map[string]VerbForms  // verb -> forms
	Nouns    map[string]NounForms  // noun -> forms
	Articles ArticleForms          // article configuration
	Words    map[string]string     // base word translations
	Punct    PunctuationRules      // language-specific punctuation
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

// getGrammarData returns the grammar data for the current language.
// Returns nil if no grammar data is loaded for the language.
func getGrammarData(lang string) *GrammarData {
	grammarCacheMu.RLock()
	defer grammarCacheMu.RUnlock()
	return grammarCache[lang]
}

// SetGrammarData sets the grammar data for a language.
// Called by the Service when loading locale files.
func SetGrammarData(lang string, data *GrammarData) {
	grammarCacheMu.Lock()
	defer grammarCacheMu.Unlock()
	grammarCache[lang] = data
}

// getVerbForm retrieves a verb form from JSON data.
// Returns empty string if not found, allowing fallback to computed form.
func getVerbForm(lang, verb, form string) string {
	data := getGrammarData(lang)
	if data == nil || data.Verbs == nil {
		return ""
	}
	verb = strings.ToLower(verb)
	if forms, ok := data.Verbs[verb]; ok {
		switch form {
		case "past":
			return forms.Past
		case "gerund":
			return forms.Gerund
		}
	}
	return ""
}

// getWord retrieves a base word translation from JSON data.
// Returns empty string if not found, allowing fallback to the key itself.
func getWord(lang, word string) string {
	data := getGrammarData(lang)
	if data == nil || data.Words == nil {
		return ""
	}
	return data.Words[strings.ToLower(word)]
}

// getPunct retrieves a punctuation rule for the language.
// Returns the default if not found.
func getPunct(lang, rule, defaultVal string) string {
	data := getGrammarData(lang)
	if data == nil {
		return defaultVal
	}
	switch rule {
	case "label":
		if data.Punct.LabelSuffix != "" {
			return data.Punct.LabelSuffix
		}
	case "progress":
		if data.Punct.ProgressSuffix != "" {
			return data.Punct.ProgressSuffix
		}
	}
	return defaultVal
}

// getNounForm retrieves a noun form from JSON data.
// Returns empty string if not found, allowing fallback to computed form.
func getNounForm(lang, noun, form string) string {
	data := getGrammarData(lang)
	if data == nil || data.Nouns == nil {
		return ""
	}
	noun = strings.ToLower(noun)
	if forms, ok := data.Nouns[noun]; ok {
		switch form {
		case "one":
			return forms.One
		case "other":
			return forms.Other
		case "gender":
			return forms.Gender
		}
	}
	return ""
}

// currentLangForGrammar returns the current language for grammar lookups.
// Uses the default service's language if available.
func currentLangForGrammar() string {
	if svc := Default(); svc != nil {
		return svc.Language()
	}
	return "en-GB"
}

// VerbForms holds irregular verb conjugations.
type VerbForms struct {
	Past   string // Past tense (e.g., "deleted")
	Gerund string // Present participle (e.g., "deleting")
}

// irregularVerbs maps base verbs to their irregular forms.
var irregularVerbs = map[string]VerbForms{
	"be":     {Past: "was", Gerund: "being"},
	"have":   {Past: "had", Gerund: "having"},
	"do":     {Past: "did", Gerund: "doing"},
	"go":     {Past: "went", Gerund: "going"},
	"make":   {Past: "made", Gerund: "making"},
	"get":    {Past: "got", Gerund: "getting"},
	"run":    {Past: "ran", Gerund: "running"},
	"set":    {Past: "set", Gerund: "setting"},
	"put":    {Past: "put", Gerund: "putting"},
	"cut":    {Past: "cut", Gerund: "cutting"},
	"let":    {Past: "let", Gerund: "letting"},
	"hit":    {Past: "hit", Gerund: "hitting"},
	"shut":   {Past: "shut", Gerund: "shutting"},
	"split":  {Past: "split", Gerund: "splitting"},
	"spread": {Past: "spread", Gerund: "spreading"},
	"read":   {Past: "read", Gerund: "reading"},
	"write":  {Past: "wrote", Gerund: "writing"},
	"send":   {Past: "sent", Gerund: "sending"},
	"build":  {Past: "built", Gerund: "building"},
	"begin":  {Past: "began", Gerund: "beginning"},
	"find":   {Past: "found", Gerund: "finding"},
	"take":   {Past: "took", Gerund: "taking"},
	"see":    {Past: "saw", Gerund: "seeing"},
	"keep":   {Past: "kept", Gerund: "keeping"},
	"hold":   {Past: "held", Gerund: "holding"},
	"tell":   {Past: "told", Gerund: "telling"},
	"bring":  {Past: "brought", Gerund: "bringing"},
	"think":  {Past: "thought", Gerund: "thinking"},
	"buy":    {Past: "bought", Gerund: "buying"},
	"catch":  {Past: "caught", Gerund: "catching"},
	"teach":  {Past: "taught", Gerund: "teaching"},
	"throw":  {Past: "threw", Gerund: "throwing"},
	"grow":   {Past: "grew", Gerund: "growing"},
	"know":   {Past: "knew", Gerund: "knowing"},
	"show":   {Past: "showed", Gerund: "showing"},
	"draw":   {Past: "drew", Gerund: "drawing"},
	"break":  {Past: "broke", Gerund: "breaking"},
	"speak":  {Past: "spoke", Gerund: "speaking"},
	"choose": {Past: "chose", Gerund: "choosing"},
	"forget": {Past: "forgot", Gerund: "forgetting"},
	"lose":   {Past: "lost", Gerund: "losing"},
	"win":    {Past: "won", Gerund: "winning"},
	"swim":   {Past: "swam", Gerund: "swimming"},
	"drive":  {Past: "drove", Gerund: "driving"},
	"rise":   {Past: "rose", Gerund: "rising"},
	"shine":  {Past: "shone", Gerund: "shining"},
	"sing":   {Past: "sang", Gerund: "singing"},
	"ring":   {Past: "rang", Gerund: "ringing"},
	"drink":  {Past: "drank", Gerund: "drinking"},
	"sink":   {Past: "sank", Gerund: "sinking"},
	"sit":    {Past: "sat", Gerund: "sitting"},
	"stand":  {Past: "stood", Gerund: "standing"},
	"hang":   {Past: "hung", Gerund: "hanging"},
	"dig":    {Past: "dug", Gerund: "digging"},
	"stick":  {Past: "stuck", Gerund: "sticking"},
	"bite":   {Past: "bit", Gerund: "biting"},
	"hide":   {Past: "hid", Gerund: "hiding"},
	"feed":   {Past: "fed", Gerund: "feeding"},
	"meet":   {Past: "met", Gerund: "meeting"},
	"lead":   {Past: "led", Gerund: "leading"},
	"sleep":  {Past: "slept", Gerund: "sleeping"},
	"feel":   {Past: "felt", Gerund: "feeling"},
	"leave":  {Past: "left", Gerund: "leaving"},
	"mean":   {Past: "meant", Gerund: "meaning"},
	"lend":   {Past: "lent", Gerund: "lending"},
	"spend":  {Past: "spent", Gerund: "spending"},
	"bend":   {Past: "bent", Gerund: "bending"},
	"deal":   {Past: "dealt", Gerund: "dealing"},
	"lay":    {Past: "laid", Gerund: "laying"},
	"pay":    {Past: "paid", Gerund: "paying"},
	"say":    {Past: "said", Gerund: "saying"},
	"sell":   {Past: "sold", Gerund: "selling"},
	"seek":   {Past: "sought", Gerund: "seeking"},
	"fight":  {Past: "fought", Gerund: "fighting"},
	"fly":    {Past: "flew", Gerund: "flying"},
	"wear":   {Past: "wore", Gerund: "wearing"},
	"tear":   {Past: "tore", Gerund: "tearing"},
	"bear":   {Past: "bore", Gerund: "bearing"},
	"swear":  {Past: "swore", Gerund: "swearing"},
	"wake":   {Past: "woke", Gerund: "waking"},
	"freeze":    {Past: "froze", Gerund: "freezing"},
	"steal":     {Past: "stole", Gerund: "stealing"},
	"overwrite": {Past: "overwritten", Gerund: "overwriting"},
	"reset":     {Past: "reset", Gerund: "resetting"},
	"reboot":    {Past: "rebooted", Gerund: "rebooting"},
}

// PastTense returns the past tense of a verb.
// Checks JSON locale data first, then irregular verbs, then applies regular rules.
//
//	PastTense("delete") // "deleted"
//	PastTense("run")    // "ran"
//	PastTense("copy")   // "copied"
func PastTense(verb string) string {
	verb = strings.ToLower(strings.TrimSpace(verb))
	if verb == "" {
		return ""
	}

	// Check JSON data first (for current language)
	if form := getVerbForm(currentLangForGrammar(), verb, "past"); form != "" {
		return form
	}

	// Check irregular verbs
	if forms, ok := irregularVerbs[verb]; ok {
		return forms.Past
	}

	return applyRegularPastTense(verb)
}

// applyRegularPastTense applies regular past tense rules.
func applyRegularPastTense(verb string) string {
	// Already ends in -ed (but not -eed, -ied which need different handling)
	// Words like "proceed", "succeed", "exceed" end in -eed and are NOT past tense
	if strings.HasSuffix(verb, "ed") && len(verb) > 2 {
		// Check if it's actually a past tense suffix (consonant + ed)
		// vs a word root ending (e.g., "proceed" = proc + eed, "feed" = feed)
		thirdFromEnd := verb[len(verb)-3]
		if !isVowel(rune(thirdFromEnd)) && thirdFromEnd != 'e' {
			// Consonant before -ed means it's likely already past tense
			return verb
		}
		// Words ending in vowel + ed (like "proceed") need -ed added
	}

	// Ends in -e: just add -d
	if strings.HasSuffix(verb, "e") {
		return verb + "d"
	}

	// Ends in consonant + y: change y to ied
	if strings.HasSuffix(verb, "y") && len(verb) > 1 {
		prev := rune(verb[len(verb)-2])
		if !isVowel(prev) {
			return verb[:len(verb)-1] + "ied"
		}
	}

	// Ends in single vowel + single consonant (CVC pattern): double consonant
	if len(verb) >= 2 && shouldDoubleConsonant(verb) {
		return verb + string(verb[len(verb)-1]) + "ed"
	}

	// Default: add -ed
	return verb + "ed"
}

// noDoubleConsonant contains multi-syllable verbs that don't double the final consonant.
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
	"cancel":  true,
	"model":   true,
	"travel":  true,
	"label":   true,
	"level":   true,
	"total":   true,
	"target":  true,
	"budget":  true,
	"market":  true,
	"benefit": true,
	"focus":   true,
}

// shouldDoubleConsonant checks if the final consonant should be doubled.
// Applies to CVC (consonant-vowel-consonant) endings in single-syllable words
// and stressed final syllables in multi-syllable words.
func shouldDoubleConsonant(verb string) bool {
	if len(verb) < 3 {
		return false
	}

	// Check explicit exceptions
	if noDoubleConsonant[verb] {
		return false
	}

	lastChar := rune(verb[len(verb)-1])
	secondLast := rune(verb[len(verb)-2])

	// Last char must be consonant (not w, x, y)
	if isVowel(lastChar) || lastChar == 'w' || lastChar == 'x' || lastChar == 'y' {
		return false
	}

	// Second to last must be a single vowel
	if !isVowel(secondLast) {
		return false
	}

	// For short words (3-4 chars), always double if CVC pattern
	if len(verb) <= 4 {
		thirdLast := rune(verb[len(verb)-3])
		return !isVowel(thirdLast)
	}

	// For longer words, only double if the pattern is strongly CVC
	// (stressed final syllable). This is a simplification - in practice,
	// most common multi-syllable verbs either:
	// 1. End in a doubled consonant already (e.g., "submit" -> "submitted")
	// 2. Don't double (e.g., "open" -> "opened")
	// We err on the side of not doubling for longer words
	return false
}

// Gerund returns the present participle (-ing form) of a verb.
// Checks JSON locale data first, then irregular verbs, then applies regular rules.
//
//	Gerund("delete")  // "deleting"
//	Gerund("run")     // "running"
//	Gerund("die")     // "dying"
func Gerund(verb string) string {
	verb = strings.ToLower(strings.TrimSpace(verb))
	if verb == "" {
		return ""
	}

	// Check JSON data first (for current language)
	if form := getVerbForm(currentLangForGrammar(), verb, "gerund"); form != "" {
		return form
	}

	// Check irregular verbs
	if forms, ok := irregularVerbs[verb]; ok {
		return forms.Gerund
	}

	return applyRegularGerund(verb)
}

// applyRegularGerund applies regular gerund rules.
func applyRegularGerund(verb string) string {
	// Ends in -ie: change to -ying
	if strings.HasSuffix(verb, "ie") {
		return verb[:len(verb)-2] + "ying"
	}

	// Ends in -e (but not -ee, -ye, -oe): drop e, add -ing
	if strings.HasSuffix(verb, "e") && len(verb) > 1 {
		secondLast := rune(verb[len(verb)-2])
		if secondLast != 'e' && secondLast != 'y' && secondLast != 'o' {
			return verb[:len(verb)-1] + "ing"
		}
	}

	// CVC pattern: double final consonant
	if shouldDoubleConsonant(verb) {
		return verb + string(verb[len(verb)-1]) + "ing"
	}

	// Default: add -ing
	return verb + "ing"
}

// irregularNouns maps singular nouns to their irregular plural forms.
var irregularNouns = map[string]string{
	"child":      "children",
	"person":     "people",
	"man":        "men",
	"woman":      "women",
	"foot":       "feet",
	"tooth":      "teeth",
	"mouse":      "mice",
	"goose":      "geese",
	"ox":         "oxen",
	"index":      "indices",
	"appendix":   "appendices",
	"matrix":     "matrices",
	"vertex":     "vertices",
	"crisis":     "crises",
	"analysis":   "analyses",
	"diagnosis":  "diagnoses",
	"thesis":     "theses",
	"hypothesis": "hypotheses",
	"parenthesis":"parentheses",
	"datum":      "data",
	"medium":     "media",
	"bacterium":  "bacteria",
	"criterion":  "criteria",
	"phenomenon": "phenomena",
	"curriculum": "curricula",
	"alumnus":    "alumni",
	"cactus":     "cacti",
	"focus":      "foci",
	"fungus":     "fungi",
	"nucleus":    "nuclei",
	"radius":     "radii",
	"stimulus":   "stimuli",
	"syllabus":   "syllabi",
	"fish":       "fish",
	"sheep":      "sheep",
	"deer":       "deer",
	"species":    "species",
	"series":     "series",
	"aircraft":   "aircraft",
	"life":       "lives",
	"wife":       "wives",
	"knife":      "knives",
	"leaf":       "leaves",
	"half":       "halves",
	"self":       "selves",
	"shelf":      "shelves",
	"wolf":       "wolves",
	"calf":       "calves",
	"loaf":       "loaves",
	"thief":      "thieves",
}

// Pluralize returns the plural form of a noun based on count.
// If count is 1, returns the singular form unchanged.
//
//	Pluralize("file", 1)    // "file"
//	Pluralize("file", 5)    // "files"
//	Pluralize("child", 3)   // "children"
//	Pluralize("box", 2)     // "boxes"
func Pluralize(noun string, count int) string {
	if count == 1 {
		return noun
	}
	return PluralForm(noun)
}

// PluralForm returns the plural form of a noun.
// Checks JSON locale data first, then irregular nouns, then applies regular rules.
//
//	PluralForm("file")   // "files"
//	PluralForm("child")  // "children"
//	PluralForm("box")    // "boxes"
func PluralForm(noun string) string {
	noun = strings.TrimSpace(noun)
	if noun == "" {
		return ""
	}

	lower := strings.ToLower(noun)

	// Check JSON data first (for current language)
	if form := getNounForm(currentLangForGrammar(), lower, "other"); form != "" {
		// Preserve original casing if title case
		if unicode.IsUpper(rune(noun[0])) && len(form) > 0 {
			return strings.ToUpper(string(form[0])) + form[1:]
		}
		return form
	}

	// Check irregular nouns
	if plural, ok := irregularNouns[lower]; ok {
		// Preserve original casing if title case
		if unicode.IsUpper(rune(noun[0])) {
			return strings.ToUpper(string(plural[0])) + plural[1:]
		}
		return plural
	}

	return applyRegularPlural(noun)
}

// applyRegularPlural applies regular plural rules.
func applyRegularPlural(noun string) string {
	lower := strings.ToLower(noun)

	// Words ending in -s, -ss, -sh, -ch, -x, -z: add -es
	if strings.HasSuffix(lower, "s") ||
		strings.HasSuffix(lower, "ss") ||
		strings.HasSuffix(lower, "sh") ||
		strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") {
		return noun + "es"
	}

	// Words ending in consonant + y: change y to ies
	if strings.HasSuffix(lower, "y") && len(noun) > 1 {
		prev := rune(lower[len(lower)-2])
		if !isVowel(prev) {
			return noun[:len(noun)-1] + "ies"
		}
	}

	// Words ending in -f or -fe: change to -ves (some exceptions already in irregulars)
	if strings.HasSuffix(lower, "f") {
		return noun[:len(noun)-1] + "ves"
	}
	if strings.HasSuffix(lower, "fe") {
		return noun[:len(noun)-2] + "ves"
	}

	// Words ending in -o preceded by consonant: add -es
	if strings.HasSuffix(lower, "o") && len(noun) > 1 {
		prev := rune(lower[len(lower)-2])
		if !isVowel(prev) {
			// Many exceptions (photos, pianos) - but common tech terms add -es
			if lower == "hero" || lower == "potato" || lower == "tomato" || lower == "echo" || lower == "veto" {
				return noun + "es"
			}
		}
	}

	// Default: add -s
	return noun + "s"
}

// vowelSounds contains words that start with consonants but have vowel sounds.
// These take "an" instead of "a".
var vowelSounds = map[string]bool{
	"hour":    true,
	"honest":  true,
	"honour":  true,
	"honor":   true,
	"heir":    true,
	"herb":    true, // US pronunciation
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

// Article returns the appropriate indefinite article ("a" or "an") for a word.
//
//	Article("file")     // "a"
//	Article("error")    // "an"
//	Article("user")     // "a" (sounds like "yoo-zer")
//	Article("hour")     // "an" (silent h)
func Article(word string) string {
	if word == "" {
		return "a"
	}

	lower := strings.ToLower(strings.TrimSpace(word))

	// Check for consonant sounds (words starting with vowels but sounding like consonants)
	for key := range consonantSounds {
		if strings.HasPrefix(lower, key) {
			return "a"
		}
	}

	// Check for vowel sounds (words starting with consonants but sounding like vowels)
	for key := range vowelSounds {
		if strings.HasPrefix(lower, key) {
			return "an"
		}
	}

	// Check first letter
	if len(lower) > 0 && isVowel(rune(lower[0])) {
		return "an"
	}

	return "a"
}

// isVowel returns true if the rune is a vowel (a, e, i, o, u).
func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// Title capitalizes the first letter of each word.
func Title(s string) string {
	return strings.Title(s) //nolint:staticcheck // strings.Title is fine for our use case
}

// Quote wraps a string in double quotes.
func Quote(s string) string {
	return `"` + s + `"`
}

// TemplateFuncs returns the template.FuncMap with all grammar functions.
// Use this to add grammar helpers to your templates.
//
//	tmpl := template.New("").Funcs(i18n.TemplateFuncs())
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"title":      Title,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"past":       PastTense,
		"gerund":     Gerund,
		"plural":     Pluralize,
		"pluralForm": PluralForm,
		"article":    Article,
		"quote":      Quote,
	}
}

// Progress returns a progress message for a verb.
// Generates "Verbing..." form using language-specific punctuation.
//
//	Progress("build")  // "Building..."
//	Progress("check")  // "Checking..."
//	Progress("fetch")  // "Fetching..."
func Progress(verb string) string {
	lang := currentLangForGrammar()

	// Try translated word first
	word := getWord(lang, verb)
	if word == "" {
		word = verb
	}

	g := Gerund(word)
	if g == "" {
		return ""
	}

	suffix := getPunct(lang, "progress", "...")
	return Title(g) + suffix
}

// ProgressSubject returns a progress message with a subject.
// Generates "Verbing subject..." form using language-specific punctuation.
//
//	ProgressSubject("build", "project")    // "Building project..."
//	ProgressSubject("check", "config.yaml") // "Checking config.yaml..."
func ProgressSubject(verb, subject string) string {
	lang := currentLangForGrammar()

	// Try translated word first
	word := getWord(lang, verb)
	if word == "" {
		word = verb
	}

	g := Gerund(word)
	if g == "" {
		return ""
	}

	suffix := getPunct(lang, "progress", "...")
	return Title(g) + " " + subject + suffix
}

// ActionResult returns a result message for a completed action.
// Generates "Subject verbed" form.
//
//	ActionResult("delete", "file")  // "File deleted"
//	ActionResult("commit", "changes") // "Changes committed"
func ActionResult(verb, subject string) string {
	p := PastTense(verb)
	if p == "" || subject == "" {
		return ""
	}
	return Title(subject) + " " + p
}

// ActionFailed returns a failure message for an action.
// Generates "Failed to verb subject" form.
//
//	ActionFailed("delete", "file")  // "Failed to delete file"
//	ActionFailed("push", "commits") // "Failed to push commits"
func ActionFailed(verb, subject string) string {
	if verb == "" {
		return ""
	}
	if subject == "" {
		return "Failed to " + verb
	}
	return "Failed to " + verb + " " + subject
}

// Label returns a label with a colon suffix.
// Generates "Word:" form using language-specific punctuation.
// French uses " :" (space before colon), English uses ":".
//
//	Label("status")   // EN: "Status:"  FR: "Statut :"
//	Label("version")  // EN: "Version:" FR: "Version :"
func Label(word string) string {
	if word == "" {
		return ""
	}

	lang := currentLangForGrammar()

	// Try translated word first
	translated := getWord(lang, word)
	if translated == "" {
		translated = word
	}

	suffix := getPunct(lang, "label", ":")
	return Title(translated) + suffix
}
