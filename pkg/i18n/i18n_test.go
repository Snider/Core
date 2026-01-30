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
	assert.Contains(t, langs, "en")
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

	// Default is English
	assert.Equal(t, "en", svc.Language())

	// Setting invalid language should error
	err = svc.SetLanguage("xx-invalid")
	assert.Error(t, err)

	// Language should still be English
	assert.Equal(t, "en", svc.Language())
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
	err = svc.AddMessages("en", map[string]string{
		"custom.greeting": "Hello, {{.Name}}!",
	})
	require.NoError(t, err)

	result := svc.T("custom.greeting", map[string]string{"Name": "World"})
	assert.Equal(t, "Hello, World!", result)
}

func TestAvailableLanguages(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	langs := svc.AvailableLanguages()
	assert.NotEmpty(t, langs)
	assert.Contains(t, langs, "en")
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		langEnv  string
		expected string
	}{
		{
			name:     "English exact",
			langEnv:  "en",
			expected: "en",
		},
		{
			name:     "English with region and encoding",
			langEnv:  "en_GB.UTF-8",
			expected: "en",
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

			result, _ := detectLanguage(svc.availableLangs)
			assert.Equal(t, tt.expected, result)
		})
	}
}
