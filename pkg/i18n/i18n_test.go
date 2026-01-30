package i18n

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	require.NotNil(t, svc)

	// Should have English available
	langs := svc.AvailableLanguages()
	assert.Contains(t, langs, "en-GB")
}

func TestTranslate(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Basic translation
	result := svc.T("cmd.dev.short")
	assert.Equal(t, "Multi-repo development workflow", result)

	// Missing key returns the key
	result = svc.T("nonexistent.key")
	assert.Equal(t, "nonexistent.key", result)
}

func TestTranslateWithArgs(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Translation with template data
	result := svc.T("error.repo_not_found", map[string]string{"Name": "config.yaml"})
	assert.Equal(t, "Repository 'config.yaml' not found", result)

	result = svc.T("cmd.ai.task_pr.branch_error", map[string]string{"Branch": "main"})
	assert.Equal(t, "cannot create PR from main branch; create a feature branch first", result)
}

func TestSetLanguage(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Default is en-GB
	assert.Equal(t, "en-GB", svc.Language())

	// Setting invalid language should error
	err = svc.SetLanguage("xx-invalid")
	assert.Error(t, err)

	// Language should still be en-GB
	assert.Equal(t, "en-GB", svc.Language())
}

func TestDefaultService(t *testing.T) {
	// Reset default for test
	defaultService = nil
	defaultOnce = sync.Once{}
	defaultErr = nil

	err := Init()
	require.NoError(t, err)

	svc := Default()
	require.NotNil(t, svc)

	// Global T function should work
	result := T("cmd.dev.short")
	assert.Equal(t, "Multi-repo development workflow", result)
}

func TestAddMessages(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Add custom messages
	svc.AddMessages("en-GB", map[string]string{
		"custom.greeting": "Hello, {{.Name}}!",
	})

	result := svc.T("custom.greeting", map[string]string{"Name": "World"})
	assert.Equal(t, "Hello, World!", result)
}

func TestAvailableLanguages(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	langs := svc.AvailableLanguages()
	assert.NotEmpty(t, langs)
	assert.Contains(t, langs, "en-GB")
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		langEnv  string
		expected string
	}{
		{
			name:     "English exact",
			langEnv:  "en-GB",
			expected: "en-GB",
		},
		{
			name:     "English with encoding",
			langEnv:  "en_GB.UTF-8",
			expected: "en-GB",
		},
		{
			name:     "Empty LANG",
			langEnv:  "",
			expected: "",
		},
	}

	svc, err := New()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LANG", tt.langEnv)
			t.Setenv("LC_ALL", "")
			t.Setenv("LC_MESSAGES", "")

			result := detectLanguage(svc.availableLangs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPluralization(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	// Singular - uses i18n.count.* magic
	result := svc.T("i18n.count.item", 1)
	assert.Equal(t, "1 item", result)

	// Plural
	result = svc.T("i18n.count.item", 5)
	assert.Equal(t, "5 items", result)

	// Zero uses plural
	result = svc.T("i18n.count.item", 0)
	assert.Equal(t, "0 items", result)
}

func TestNestedKeys(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Nested key
	result := svc.T("cmd.dev.short")
	assert.Equal(t, "Multi-repo development workflow", result)

	// Deeper nested key (flat key with dots)
	result = svc.T("cmd.dev.push.short")
	assert.Equal(t, "Push commits across all repos", result)
}

func TestMessage_ForCategory(t *testing.T) {
	t.Run("basic categories", func(t *testing.T) {
		msg := Message{
			Zero:  "no items",
			One:   "1 item",
			Two:   "2 items",
			Few:   "a few items",
			Many:  "many items",
			Other: "some items",
		}

		assert.Equal(t, "no items", msg.ForCategory(PluralZero))
		assert.Equal(t, "1 item", msg.ForCategory(PluralOne))
		assert.Equal(t, "2 items", msg.ForCategory(PluralTwo))
		assert.Equal(t, "a few items", msg.ForCategory(PluralFew))
		assert.Equal(t, "many items", msg.ForCategory(PluralMany))
		assert.Equal(t, "some items", msg.ForCategory(PluralOther))
	})

	t.Run("fallback to other", func(t *testing.T) {
		msg := Message{
			One:   "1 item",
			Other: "items",
		}

		// Categories without explicit values fall back to Other
		assert.Equal(t, "items", msg.ForCategory(PluralZero))
		assert.Equal(t, "1 item", msg.ForCategory(PluralOne))
		assert.Equal(t, "items", msg.ForCategory(PluralFew))
	})

	t.Run("fallback to one then text", func(t *testing.T) {
		msg := Message{
			One: "single item",
		}

		// Falls back to One when Other is empty
		assert.Equal(t, "single item", msg.ForCategory(PluralOther))
		assert.Equal(t, "single item", msg.ForCategory(PluralMany))
	})
}

func TestServiceFormality(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	t.Run("default is neutral", func(t *testing.T) {
		assert.Equal(t, FormalityNeutral, svc.Formality())
	})

	t.Run("set formality", func(t *testing.T) {
		svc.SetFormality(FormalityFormal)
		assert.Equal(t, FormalityFormal, svc.Formality())

		svc.SetFormality(FormalityInformal)
		assert.Equal(t, FormalityInformal, svc.Formality())
	})
}

func TestServiceDirection(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	t.Run("English is LTR", func(t *testing.T) {
		err := svc.SetLanguage("en-GB")
		require.NoError(t, err)

		assert.Equal(t, DirLTR, svc.Direction())
		assert.False(t, svc.IsRTL())
	})
}

func TestServicePluralCategory(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	t.Run("English plural rules", func(t *testing.T) {
		assert.Equal(t, PluralOne, svc.PluralCategory(1))
		assert.Equal(t, PluralOther, svc.PluralCategory(0))
		assert.Equal(t, PluralOther, svc.PluralCategory(5))
	})
}

func TestDebugMode(t *testing.T) {
	t.Run("default is disabled", func(t *testing.T) {
		svc, err := New()
		require.NoError(t, err)
		assert.False(t, svc.Debug())
	})

	t.Run("T with debug mode", func(t *testing.T) {
		svc, err := New()
		require.NoError(t, err)

		// Without debug
		result := svc.T("cmd.dev.short")
		assert.Equal(t, "Multi-repo development workflow", result)

		// Enable debug
		svc.SetDebug(true)
		assert.True(t, svc.Debug())

		// With debug - shows key prefix
		result = svc.T("cmd.dev.short")
		assert.Equal(t, "[cmd.dev.short] Multi-repo development workflow", result)

		// Disable debug
		svc.SetDebug(false)
		result = svc.T("cmd.dev.short")
		assert.Equal(t, "Multi-repo development workflow", result)
	})

	t.Run("C with debug mode", func(t *testing.T) {
		svc, err := New()
		require.NoError(t, err)

		subject := S("file", "config.yaml")

		// Without debug
		result := svc.C("core.delete", subject)
		assert.NotContains(t, result.Question, "[core.delete]")

		// Enable debug
		svc.SetDebug(true)

		// With debug - shows key prefix on all forms
		result = svc.C("core.delete", subject)
		assert.Contains(t, result.Question, "[core.delete]")
		assert.Contains(t, result.Success, "[core.delete]")
		assert.Contains(t, result.Failure, "[core.delete]")
	})

	t.Run("package-level SetDebug", func(t *testing.T) {
		// Reset default
		defaultService = nil
		defaultOnce = sync.Once{}
		defaultErr = nil

		err := Init()
		require.NoError(t, err)

		// Enable debug via package function
		SetDebug(true)
		assert.True(t, Default().Debug())

		// Translate
		result := T("cmd.dev.short")
		assert.Equal(t, "[cmd.dev.short] Multi-repo development workflow", result)

		// Cleanup
		SetDebug(false)
	})
}

func TestI18nNamespaceMagic(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	tests := []struct {
		name     string
		key      string
		args     []any
		expected string
	}{
		{"label", "i18n.label.status", nil, "Status:"},
		{"label version", "i18n.label.version", nil, "Version:"},
		{"progress", "i18n.progress.build", nil, "Building..."},
		{"progress check", "i18n.progress.check", nil, "Checking..."},
		{"progress with subject", "i18n.progress.check", []any{"config"}, "Checking config..."},
		{"count singular", "i18n.count.file", []any{1}, "1 file"},
		{"count plural", "i18n.count.file", []any{5}, "5 files"},
		{"done", "i18n.done.delete", []any{"file"}, "File deleted"},
		{"done build", "i18n.done.build", []any{"project"}, "Project built"},
		{"fail", "i18n.fail.delete", []any{"file"}, "Failed to delete file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.T(tt.key, tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRawBypassesI18nNamespace(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Raw() should return key as-is since i18n.label.status isn't in JSON
	result := svc.Raw("i18n.label.status")
	assert.Equal(t, "i18n.label.status", result)

	// T() should compose it
	result = svc.T("i18n.label.status")
	assert.Equal(t, "Status:", result)
}
