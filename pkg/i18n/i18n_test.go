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
	result := svc.T("cli.success")
	assert.Equal(t, "Success", result)

	// Missing key returns the key
	result = svc.T("nonexistent.key")
	assert.Equal(t, "nonexistent.key", result)
}

func TestTranslateWithArgs(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Translation with template data
	result := svc.T("error.not_found", map[string]string{"Item": "config.yaml"})
	assert.Equal(t, "Not found: config.yaml", result)

	result = svc.T("cli.time.minutes_ago", map[string]int{"Count": 5})
	assert.Equal(t, "5 minutes ago", result)
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
	result := T("cli.success")
	assert.Equal(t, "Success", result)
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

	// Singular
	result := svc.T("cli.count.items", map[string]any{"Count": 1})
	assert.Equal(t, "1 item", result)

	// Plural
	result = svc.T("cli.count.items", map[string]any{"Count": 5})
	assert.Equal(t, "5 items", result)

	// Zero uses plural
	result = svc.T("cli.count.items", map[string]any{"Count": 0})
	assert.Equal(t, "0 items", result)
}

func TestNestedKeys(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Deeply nested key
	result := svc.T("cmd.dev.work.short")
	assert.Equal(t, "Multi-repo git operations", result)

	// Nested with flag
	result = svc.T("cmd.dev.work.flag.status")
	assert.Equal(t, "Show status only, don't push", result)
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
		result := svc.T("cli.success")
		assert.Equal(t, "Success", result)

		// Enable debug
		svc.SetDebug(true)
		assert.True(t, svc.Debug())

		// With debug - shows key prefix
		result = svc.T("cli.success")
		assert.Equal(t, "[cli.success] Success", result)

		// Disable debug
		svc.SetDebug(false)
		result = svc.T("cli.success")
		assert.Equal(t, "Success", result)
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
		result := T("cli.success")
		assert.Equal(t, "[cli.success] Success", result)

		// Cleanup
		SetDebug(false)
	})
}
