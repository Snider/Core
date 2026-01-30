// Package i18n provides internationalization for the CLI.
package i18n

import (
	"os"
	"strings"

	"golang.org/x/text/language"
)

// SetFormality sets the default formality level on the default service.
// Does nothing if the service is not initialized.
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
