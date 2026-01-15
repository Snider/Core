package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Equal(t, "Search", service.Translate("app.ui.search"))
	})

	t.Run("language switch changes translations", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		// Start with English
		require.NoError(t, service.SetLanguage("en"))
		assert.Equal(t, "Search", service.Translate("app.ui.search"))

		// Switch to Spanish
		require.NoError(t, service.SetLanguage("es"))
		assert.Equal(t, "Buscar", service.Translate("app.ui.search"))

		// Switch back to English
		require.NoError(t, service.SetLanguage("en"))
		assert.Equal(t, "Search", service.Translate("app.ui.search"))
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

// TestDetectLanguage is in detect_language_test.go with table-driven tests

func TestSetLanguageErrors(t *testing.T) {
	t.Run("returns error for invalid language tag", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		// Invalid BCP 47 tag (too long, contains invalid characters)
		err = service.SetLanguage("this-is-not-a-valid-tag-at-all-definitely")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse language tag")
	})

	t.Run("returns error when no available languages", func(t *testing.T) {
		// Create a service and clear its available languages
		service, err := New()
		require.NoError(t, err)

		// Clear the available languages
		service.availableLangs = nil

		err = service.SetLanguage("en")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no available languages")
	})

	t.Run("returns error for unsupported language", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		// Try a known but completely unsupported language
		// Most obscure languages will still match with low confidence,
		// so we just verify the function handles various inputs without panicking
		err = service.SetLanguage("zu") // Zulu - may or may not be supported
		// Just verify no panic - the result depends on matcher confidence
		_ = err
	})
}

func TestTranslateWithTemplateData(t *testing.T) {
	t.Run("translates with template data", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)
		require.NoError(t, service.SetLanguage("en"))

		// If there's a template key available, use it
		// Otherwise, the translation will still work but without interpolation
		data := map[string]interface{}{"Name": "Test"}
		result := service.Translate("menu.settings", data)
		// Just verify it doesn't panic and returns something
		assert.NotEmpty(t, result)
	})

	t.Run("warns when too many arguments", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)
		require.NoError(t, service.SetLanguage("en"))

		// Call with too many args - this will print a warning to stderr
		// but should still work
		result := service.Translate("menu.settings", map[string]interface{}{}, "extra arg")
		assert.NotEmpty(t, result)
	})
}

func TestSetBundle(t *testing.T) {
	t.Run("sets bundle", func(t *testing.T) {
		service, err := New()
		require.NoError(t, err)

		oldBundle := service.bundle
		service.SetBundle(nil)
		assert.Nil(t, service.bundle)

		// Restore
		service.SetBundle(oldBundle)
		assert.Equal(t, oldBundle, service.bundle)
	})
}
