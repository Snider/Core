package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestDetectLanguage(t *testing.T) {
	supported := []language.Tag{
		language.English,
		language.French,
		language.Spanish,
	}

	tests := []struct {
		name          string
		langEnv       string
		supported     []language.Tag
		expectedLang  string
		expectedError string
	}{
		{
			name:         "Exact match",
			langEnv:      "fr",
			supported:    supported,
			expectedLang: "fr",
		},
		{
			name:         "Match with region",
			langEnv:      "fr_CA.UTF-8",
			supported:    supported,
			expectedLang: "fr",
		},
		{
			name:         "Unsupported language",
			langEnv:      "de",
			supported:    supported,
			expectedLang: "",
		},
		{
			name:         "Empty LANG",
			langEnv:      "",
			supported:    supported,
			expectedLang: "",
		},
		{
			name:          "Invalid LANG",
			langEnv:       "invalid-lang-tag",
			supported:     supported,
			expectedLang:  "",
			expectedError: "failed to parse language tag 'invalid-lang-tag': language: tag is not well-formed",
		},
		{
			name:         "Empty supported languages",
			langEnv:      "en",
			supported:    []language.Tag{},
			expectedLang: "",
		},
		{
			name:         "Match with low confidence",
			langEnv:      "it", // Italian is not supported, confidence should be No or Low?
			supported:    supported,
			expectedLang: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LANG", tt.langEnv)

			lang, err := detectLanguage(tt.supported)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLang, lang)
			}
		})
	}
}
