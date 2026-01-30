package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceImplementsTranslator(t *testing.T) {
	// This test verifies at compile time that Service implements Translator
	var _ Translator = (*Service)(nil)

	// Create a service and use it through the interface
	var translator Translator
	svc, err := New()
	require.NoError(t, err)

	translator = svc

	// Test interface methods
	assert.Equal(t, "Multi-repo development workflow", translator.T("cmd.dev.short"))
	assert.NotEmpty(t, translator.Language())
	assert.NotNil(t, translator.Direction())
	assert.NotNil(t, translator.Formality())
}

// MockTranslator demonstrates how to create a mock for testing
type MockTranslator struct {
	translations map[string]string
	language     string
}

func (m *MockTranslator) T(key string, args ...any) string {
	if v, ok := m.translations[key]; ok {
		return v
	}
	return key
}

func (m *MockTranslator) SetLanguage(lang string) error {
	m.language = lang
	return nil
}

func (m *MockTranslator) Language() string {
	return m.language
}

func (m *MockTranslator) SetMode(mode Mode)        {}
func (m *MockTranslator) Mode() Mode               { return ModeNormal }
func (m *MockTranslator) SetDebug(enabled bool)    {}
func (m *MockTranslator) Debug() bool              { return false }
func (m *MockTranslator) SetFormality(f Formality) {}
func (m *MockTranslator) Formality() Formality     { return FormalityNeutral }
func (m *MockTranslator) Direction() TextDirection { return DirLTR }
func (m *MockTranslator) IsRTL() bool              { return false }
func (m *MockTranslator) PluralCategory(n int) PluralCategory {
	return PluralOther
}
func (m *MockTranslator) AvailableLanguages() []string { return []string{"en-GB"} }

func TestMockTranslator(t *testing.T) {
	var translator Translator = &MockTranslator{
		translations: map[string]string{
			"test.hello": "Hello from mock",
		},
		language: "en-GB",
	}

	assert.Equal(t, "Hello from mock", translator.T("test.hello"))
	assert.Equal(t, "test.missing", translator.T("test.missing"))
	assert.Equal(t, "en-GB", translator.Language())
}
