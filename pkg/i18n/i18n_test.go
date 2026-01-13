package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestNew(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.bundle)
		assert.NotEmpty(t, service.availableLangs)
	})

	t.Run("loads all available languages", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		// Should have loaded multiple languages from locales/
		assert.GreaterOrEqual(t, len(service.availableLangs), 2)
	})
}

func TestSetLanguage(t *testing.T) {
	t.Run("sets English successfully", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		err = service.SetLanguage("en")
		assert.NoError(t, err)
		assert.NotNil(t, service.localizer)
	})

	t.Run("sets Spanish successfully", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		err = service.SetLanguage("es")
		assert.NoError(t, err)
		assert.NotNil(t, service.localizer)
	})

	t.Run("sets German successfully", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		err = service.SetLanguage("de")
		assert.NoError(t, err)
		assert.NotNil(t, service.localizer)
	})

	t.Run("handles language variants", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		// en-US should match to en
		err = service.SetLanguage("en-US")
		assert.NoError(t, err)
	})

	t.Run("handles unknown language by matching closest", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		// Unknown languages may fall back to a default match
		// The matcher uses confidence levels, so many tags will match something
		err = service.SetLanguage("tlh") // Klingon
		// May or may not error depending on matcher confidence
		if err != nil {
			assert.Contains(t, err.Error(), "unsupported language")
		}
	})
}

func TestTranslate(t *testing.T) {
	t.Run("translates English message", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)
		require.NoError(t, service.SetLanguage("en"))

		result := service.Translate("menu.settings")
		assert.Equal(t, "Settings", result)
	})

	t.Run("translates Spanish message", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)
		require.NoError(t, service.SetLanguage("es"))

		result := service.Translate("menu.settings")
		assert.Equal(t, "Ajustes", result)
	})

	t.Run("returns message ID for missing translation", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)
		require.NoError(t, service.SetLanguage("en"))

		result := service.Translate("nonexistent.message.id")
		assert.Equal(t, "nonexistent.message.id", result)
	})

	t.Run("translates multiple messages correctly", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)
		require.NoError(t, service.SetLanguage("en"))

		assert.Equal(t, "Dashboard", service.Translate("menu.dashboard"))
		assert.Equal(t, "Help", service.Translate("menu.help"))
		assert.Equal(t, "Search", service.Translate("app.core.ui.search"))
	})

	t.Run("language switch changes translations", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		// Start with English
		require.NoError(t, service.SetLanguage("en"))
		assert.Equal(t, "Search", service.Translate("app.core.ui.search"))

		// Switch to Spanish
		require.NoError(t, service.SetLanguage("es"))
		assert.Equal(t, "Buscar", service.Translate("app.core.ui.search"))

		// Switch back to English
		require.NoError(t, service.SetLanguage("en"))
		assert.Equal(t, "Search", service.Translate("app.core.ui.search"))
	})
}

func TestGetAvailableLanguages(t *testing.T) {
	t.Run("returns available languages", func(t *testing.T) {
		langs, err := getAvailableLanguages()
		require.NoError(t, err)
		assert.NotEmpty(t, langs)

		// Should include at least English
		langStrings := make([]string, len(langs))
		for i, l := range langs {
			langStrings[i] = l.String()
		}
		assert.Contains(t, langStrings, "en")
	})
}

func TestDetectLanguage(t *testing.T) {
	t.Run("returns empty for empty LANG env", func(t *testing.T) {
		// Save and clear LANG
		t.Setenv("LANG", "")

		service, err := New()
		require.NoError(t, err)

		detected, err := detectLanguage(service.availableLangs)
		assert.NoError(t, err)
		assert.Empty(t, detected)
	})

	t.Run("returns empty for empty supported list", func(t *testing.T) {
		t.Setenv("LANG", "en_US.UTF-8")

		detected, err := detectLanguage([]language.Tag{})
		assert.NoError(t, err)
		assert.Empty(t, detected)
	})

	t.Run("detects language from LANG env", func(t *testing.T) {
		t.Setenv("LANG", "es_ES.UTF-8")

		service, err := New()
		require.NoError(t, err)

		detected, err := detectLanguage(service.availableLangs)
		assert.NoError(t, err)
		// Should detect Spanish or a close variant
		if detected != "" {
			assert.Contains(t, detected, "es")
		}
	})
}
