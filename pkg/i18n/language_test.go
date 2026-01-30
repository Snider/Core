package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormality_String(t *testing.T) {
	tests := []struct {
		f        Formality
		expected string
	}{
		{FormalityNeutral, "neutral"},
		{FormalityInformal, "informal"},
		{FormalityFormal, "formal"},
		{Formality(99), "neutral"}, // Unknown defaults to neutral
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.f.String())
	}
}

func TestTextDirection_String(t *testing.T) {
	assert.Equal(t, "ltr", DirLTR.String())
	assert.Equal(t, "rtl", DirRTL.String())
}

func TestPluralCategory_String(t *testing.T) {
	tests := []struct {
		cat      PluralCategory
		expected string
	}{
		{PluralZero, "zero"},
		{PluralOne, "one"},
		{PluralTwo, "two"},
		{PluralFew, "few"},
		{PluralMany, "many"},
		{PluralOther, "other"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.cat.String())
	}
}

func TestGrammaticalGender_String(t *testing.T) {
	tests := []struct {
		g        GrammaticalGender
		expected string
	}{
		{GenderNeuter, "neuter"},
		{GenderMasculine, "masculine"},
		{GenderFeminine, "feminine"},
		{GenderCommon, "common"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.g.String())
	}
}

func TestIsRTLLanguage(t *testing.T) {
	// RTL languages
	assert.True(t, IsRTLLanguage("ar"))
	assert.True(t, IsRTLLanguage("ar-SA"))
	assert.True(t, IsRTLLanguage("he"))
	assert.True(t, IsRTLLanguage("he-IL"))
	assert.True(t, IsRTLLanguage("fa"))
	assert.True(t, IsRTLLanguage("ur"))

	// LTR languages
	assert.False(t, IsRTLLanguage("en"))
	assert.False(t, IsRTLLanguage("en-GB"))
	assert.False(t, IsRTLLanguage("de"))
	assert.False(t, IsRTLLanguage("fr"))
	assert.False(t, IsRTLLanguage("zh"))
}

func TestPluralRuleEnglish(t *testing.T) {
	tests := []struct {
		n        int
		expected PluralCategory
	}{
		{0, PluralOther},
		{1, PluralOne},
		{2, PluralOther},
		{5, PluralOther},
		{100, PluralOther},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, pluralRuleEnglish(tt.n), "count=%d", tt.n)
	}
}

func TestPluralRuleFrench(t *testing.T) {
	// French uses singular for 0 and 1
	assert.Equal(t, PluralOne, pluralRuleFrench(0))
	assert.Equal(t, PluralOne, pluralRuleFrench(1))
	assert.Equal(t, PluralOther, pluralRuleFrench(2))
}

func TestPluralRuleRussian(t *testing.T) {
	tests := []struct {
		n        int
		expected PluralCategory
	}{
		{1, PluralOne},
		{2, PluralFew},
		{3, PluralFew},
		{4, PluralFew},
		{5, PluralMany},
		{11, PluralMany},
		{12, PluralMany},
		{21, PluralOne},
		{22, PluralFew},
		{25, PluralMany},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, pluralRuleRussian(tt.n), "count=%d", tt.n)
	}
}

func TestPluralRuleArabic(t *testing.T) {
	tests := []struct {
		n        int
		expected PluralCategory
	}{
		{0, PluralZero},
		{1, PluralOne},
		{2, PluralTwo},
		{3, PluralFew},
		{10, PluralFew},
		{11, PluralMany},
		{99, PluralMany},
		{100, PluralOther},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, pluralRuleArabic(tt.n), "count=%d", tt.n)
	}
}

func TestPluralRuleChinese(t *testing.T) {
	// Chinese has no plural distinction
	assert.Equal(t, PluralOther, pluralRuleChinese(0))
	assert.Equal(t, PluralOther, pluralRuleChinese(1))
	assert.Equal(t, PluralOther, pluralRuleChinese(100))
}

func TestGetPluralRule(t *testing.T) {
	// Known languages
	rule := GetPluralRule("en-GB")
	assert.Equal(t, PluralOne, rule(1))

	rule = GetPluralRule("ru")
	assert.Equal(t, PluralFew, rule(2))

	// Unknown language falls back to English
	rule = GetPluralRule("xx-unknown")
	assert.Equal(t, PluralOne, rule(1))
	assert.Equal(t, PluralOther, rule(2))
}

func TestGetPluralCategory(t *testing.T) {
	assert.Equal(t, PluralOne, GetPluralCategory("en", 1))
	assert.Equal(t, PluralOther, GetPluralCategory("en", 5))
	assert.Equal(t, PluralFew, GetPluralCategory("ru", 3))
}
