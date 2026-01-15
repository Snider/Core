package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func newTestBundle() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustParseMessageFileBytes([]byte(`{
		"hello": "Hello",
		"welcome": "Welcome {{.Name}}"
	}`), "en.json")
	bundle.MustParseMessageFileBytes([]byte(`{
		"hello": "Bonjour",
		"welcome": "Bienvenue {{.Name}}"
	}`), "fr.json")
	return bundle
}

func TestNew(t *testing.T) {
	s, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestSetLanguage(t *testing.T) {
	s, err := New()
	require.NoError(t, err)

	s.SetBundle(newTestBundle())

	err = s.SetLanguage("en")
	assert.NoError(t, err)

	err = s.SetLanguage("fr")
	assert.NoError(t, err)

	err = s.SetLanguage("invalid")
	assert.Error(t, err)
}

func TestTranslate(t *testing.T) {
	s, err := New()
	require.NoError(t, err)

	s.SetBundle(newTestBundle())

	err = s.SetLanguage("en")
	require.NoError(t, err)
	assert.Equal(t, "Hello", s.Translate("hello"))

	err = s.SetLanguage("fr")
	require.NoError(t, err)
	assert.Equal(t, "Bonjour", s.Translate("hello"))
}

func TestTranslate_WithArgs(t *testing.T) {
	s, err := New()
	require.NoError(t, err)

	s.SetBundle(newTestBundle())

	err = s.SetLanguage("en")
	require.NoError(t, err)
	assert.Equal(t, "Welcome John", s.Translate("welcome", map[string]string{"Name": "John"}))

	err = s.SetLanguage("fr")
	require.NoError(t, err)
	assert.Equal(t, "Bienvenue John", s.Translate("welcome", map[string]string{"Name": "John"}))
}

func TestTranslate_Good(t *testing.T) {
	s, err := New()
	require.NoError(t, err)

	s.SetBundle(newTestBundle())

	err = s.SetLanguage("en")
	require.NoError(t, err)
	assert.Equal(t, "Hello", s.Translate("hello"))
}

func TestTranslate_Bad(t *testing.T) {
	s, err := New()
	require.NoError(t, err)

	s.SetBundle(newTestBundle())

	err = s.SetLanguage("en")
	require.NoError(t, err)
	assert.Equal(t, "non-existent", s.Translate("non-existent"))
}

func TestTranslate_Ugly(t *testing.T) {
	s, err := New()
	require.NoError(t, err)

	s.SetBundle(newTestBundle())

	err = s.SetLanguage("en")
	require.NoError(t, err)
	assert.Equal(t, "", s.Translate(""))
}

func ExampleNew() {
	i18nService, err := New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(i18nService.Translate("hello"))
	// Output: Hello
}

func ExampleService_SetLanguage() {
	i18nService, err := New()
	if err != nil {
		log.Fatal(err)
	}

	err = i18nService.SetLanguage("es")
	if err != nil {
		log.Printf("Failed to set language: %v", err)
	}

	// This would load a real Spanish locale file in a real application
	// For this example, we'll inject a bundle with Spanish translations
	bundle := i18n.NewBundle(language.Spanish)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustParseMessageFileBytes([]byte(`{
		"hello": "Hola"
	}`), "es.json")
	i18nService.SetBundle(bundle)

	err = i18nService.SetLanguage("es")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(i18nService.Translate("hello"))
	// Output: Hola
}

func ExampleService_Translate() {
	i18nService, err := New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(i18nService.Translate("hello"))
	// Output: Hello
}
