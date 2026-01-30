// Package i18n provides internationalization for the CLI.
package i18n

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

// String returns the string representation of a Formality level.
func (f Formality) String() string {
	switch f {
	case FormalityInformal:
		return "informal"
	case FormalityFormal:
		return "formal"
	default:
		return "neutral"
	}
}

// TextDirection represents text directionality.
type TextDirection int

const (
	// DirLTR is left-to-right text direction (English, German, etc.)
	DirLTR TextDirection = iota
	// DirRTL is right-to-left text direction (Arabic, Hebrew, etc.)
	DirRTL
)

// String returns the string representation of a TextDirection.
func (d TextDirection) String() string {
	if d == DirRTL {
		return "rtl"
	}
	return "ltr"
}

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

// String returns the string representation of a PluralCategory.
func (p PluralCategory) String() string {
	switch p {
	case PluralZero:
		return "zero"
	case PluralOne:
		return "one"
	case PluralTwo:
		return "two"
	case PluralFew:
		return "few"
	case PluralMany:
		return "many"
	default:
		return "other"
	}
}

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

// String returns the string representation of a GrammaticalGender.
func (g GrammaticalGender) String() string {
	switch g {
	case GenderMasculine:
		return "masculine"
	case GenderFeminine:
		return "feminine"
	case GenderCommon:
		return "common"
	default:
		return "neuter"
	}
}

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

// IsRTLLanguage returns true if the language code uses right-to-left text.
func IsRTLLanguage(lang string) bool {
	// Check exact match first
	if rtlLanguages[lang] {
		return true
	}
	// Check base language (e.g., "ar" for "ar-SA")
	if len(lang) > 2 {
		base := lang[:2]
		return rtlLanguages[base]
	}
	return false
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

// English: one (n=1), other
func pluralRuleEnglish(n int) PluralCategory {
	if n == 1 {
		return PluralOne
	}
	return PluralOther
}

// German: same as English
func pluralRuleGerman(n int) PluralCategory {
	return pluralRuleEnglish(n)
}

// French: one (n=0,1), other
func pluralRuleFrench(n int) PluralCategory {
	if n == 0 || n == 1 {
		return PluralOne
	}
	return PluralOther
}

// Spanish: one (n=1), many (n=0 or n>=1000000), other
func pluralRuleSpanish(n int) PluralCategory {
	if n == 1 {
		return PluralOne
	}
	return PluralOther
}

// Russian: one (n%10=1, n%100!=11), few (n%10=2-4, n%100!=12-14), many (others)
func pluralRuleRussian(n int) PluralCategory {
	mod10 := n % 10
	mod100 := n % 100

	if mod10 == 1 && mod100 != 11 {
		return PluralOne
	}
	if mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14) {
		return PluralFew
	}
	return PluralMany
}

// Polish: one (n=1), few (n%10=2-4, n%100!=12-14), many (others)
func pluralRulePolish(n int) PluralCategory {
	if n == 1 {
		return PluralOne
	}
	mod10 := n % 10
	mod100 := n % 100
	if mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14) {
		return PluralFew
	}
	return PluralMany
}

// Arabic: zero (n=0), one (n=1), two (n=2), few (n%100=3-10), many (n%100=11-99), other
func pluralRuleArabic(n int) PluralCategory {
	if n == 0 {
		return PluralZero
	}
	if n == 1 {
		return PluralOne
	}
	if n == 2 {
		return PluralTwo
	}
	mod100 := n % 100
	if mod100 >= 3 && mod100 <= 10 {
		return PluralFew
	}
	if mod100 >= 11 && mod100 <= 99 {
		return PluralMany
	}
	return PluralOther
}

// Chinese/Japanese/Korean: other (no plural distinction)
func pluralRuleChinese(n int) PluralCategory {
	return PluralOther
}

func pluralRuleJapanese(n int) PluralCategory {
	return PluralOther
}

func pluralRuleKorean(n int) PluralCategory {
	return PluralOther
}

// GetPluralRule returns the plural rule for a language code.
// Falls back to English rules if the language is not found.
func GetPluralRule(lang string) PluralRule {
	if rule, ok := pluralRules[lang]; ok {
		return rule
	}
	// Try base language
	if len(lang) > 2 {
		base := lang[:2]
		if rule, ok := pluralRules[base]; ok {
			return rule
		}
	}
	// Default to English
	return pluralRuleEnglish
}

// GetPluralCategory returns the plural category for a count in the given language.
func GetPluralCategory(lang string, n int) PluralCategory {
	return GetPluralRule(lang)(n)
}
