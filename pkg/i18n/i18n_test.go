package i18n

import (
	"encoding/json"
	"testing"

	"github.com/Snider/Core/pkg/core"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func newTestBundle() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustParseMessageFileBytes([]byte(`{
		"hello": "Hello"
	}`), "en.json")
	bundle.MustParseMessageFileBytes([]byte(`{
		"hello": "Bonjour"
	}`), "fr.json")
	return bundle
}

func TestNew(t *testing.T) {
	s, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestRegister(t *testing.T) {
	c, err := core.New()
	require.NoError(t, err)
	s, err := Register(c)
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
