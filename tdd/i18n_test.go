package tdd

import (
	"embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Snider/Core/pkg/i18n"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

//go:embed en.json es.json
var testLocaleFS embed.FS

func setupI18nService(t *testing.T) *i18n.Service {
	s, err := i18n.New()
	assert.NoError(t, err)

	bundle := goi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	availableLangs := []language.Tag{language.English, language.Spanish}

	for _, lang := range availableLangs {
		filePath := fmt.Sprintf("%s.json", lang.String())
		_, err := bundle.LoadMessageFileFS(testLocaleFS, filePath)
		assert.NoError(t, err)
	}

	s.SetBundle(bundle, availableLangs)
	return s
}

// TestI18n_Translate_Good tests the happy path for translation.
func TestI18n_Translate_Good(t *testing.T) {
	s := setupI18nService(t)

	// English is the default
	err := s.SetLanguage("en")
	assert.NoError(t, err)
	assert.Equal(t, "Hello", s.Translate("hello"))

	// Switch to Spanish
	err = s.SetLanguage("es")
	assert.NoError(t, err)
	assert.Equal(t, "Hola", s.Translate("hello"))
}

// TestI18n_SetLanguage_Bad tests setting an unsupported language.
func TestI18n_SetLanguage_Bad(t *testing.T) {
	s := setupI18nService(t)

	err := s.SetLanguage("xx") // An unsupported language
	assert.Error(t, err)
}

// TestI18n_Translate_Ugly tests translating a non-existent key.
func TestI18n_Translate_Ugly(t *testing.T) {
	s := setupI18nService(t)

	err := s.SetLanguage("en")
	assert.NoError(t, err)

	// The service should return the key itself if the translation is not found.
	assert.Equal(t, "nonexistent", s.Translate("nonexistent"))
}
