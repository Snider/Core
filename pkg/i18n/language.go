// Package i18n provides internationalization for the CLI.
package i18n

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

// String returns the string representation of a TextDirection.
func (d TextDirection) String() string {
	if d == DirRTL {
		return "rtl"
	}
	return "ltr"
}

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
