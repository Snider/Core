package i18n

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIntent(t *testing.T) {
	t.Run("existing intent", func(t *testing.T) {
		intent := getIntent("core.delete")
		require.NotNil(t, intent)
		assert.Equal(t, "action", intent.Meta.Type)
		assert.Equal(t, "delete", intent.Meta.Verb)
		assert.True(t, intent.Meta.Dangerous)
		assert.Equal(t, "no", intent.Meta.Default)
	})

	t.Run("non-existent intent", func(t *testing.T) {
		intent := getIntent("nonexistent.intent")
		assert.Nil(t, intent)
	})
}

func TestRegisterIntent(t *testing.T) {
	// Register a custom intent
	RegisterIntent("test.custom", Intent{
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "custom",
			Default: "yes",
		},
		Question: "Custom {{.Subject}}?",
		Success:  "{{.Subject | title}} customised",
		Failure:  "Failed to customise {{.Subject}}",
	})

	// Verify it was registered
	intent := getIntent("test.custom")
	require.NotNil(t, intent)
	assert.Equal(t, "custom", intent.Meta.Verb)
	assert.Equal(t, "Custom {{.Subject}}?", intent.Question)

	// Clean up
	UnregisterIntent("test.custom")
}

func TestRegisterIntents_Batch(t *testing.T) {
	// Register multiple intents at once
	RegisterIntents(map[string]Intent{
		"test.batch1": {
			Meta:     IntentMeta{Type: "action", Verb: "batch1", Default: "yes"},
			Question: "Batch 1?",
		},
		"test.batch2": {
			Meta:     IntentMeta{Type: "action", Verb: "batch2", Default: "no"},
			Question: "Batch 2?",
		},
	})

	// Verify both were registered
	assert.True(t, HasIntent("test.batch1"))
	assert.True(t, HasIntent("test.batch2"))

	intent1 := GetIntent("test.batch1")
	require.NotNil(t, intent1)
	assert.Equal(t, "batch1", intent1.Meta.Verb)

	intent2 := GetIntent("test.batch2")
	require.NotNil(t, intent2)
	assert.Equal(t, "batch2", intent2.Meta.Verb)

	// Clean up
	UnregisterIntent("test.batch1")
	UnregisterIntent("test.batch2")

	// Verify cleanup
	assert.False(t, HasIntent("test.batch1"))
	assert.False(t, HasIntent("test.batch2"))
}

func TestCustomIntentOverridesCoreIntent(t *testing.T) {
	// Custom intents should be checked before core intents
	RegisterIntent("core.delete", Intent{
		Meta:     IntentMeta{Type: "action", Verb: "delete", Default: "yes"},
		Question: "Custom delete {{.Subject}}?",
	})

	// Should get custom intent
	intent := getIntent("core.delete")
	require.NotNil(t, intent)
	assert.Equal(t, "Custom delete {{.Subject}}?", intent.Question)
	assert.Equal(t, "yes", intent.Meta.Default) // Changed from core's "no"

	// Clean up
	UnregisterIntent("core.delete")

	// Now should get core intent again
	intent = getIntent("core.delete")
	require.NotNil(t, intent)
	assert.Equal(t, "Delete {{.Subject}}?", intent.Question)
	assert.Equal(t, "no", intent.Meta.Default) // Back to core default
}

func TestHasIntent(t *testing.T) {
	assert.True(t, HasIntent("core.delete"))
	assert.True(t, HasIntent("core.create"))
	assert.False(t, HasIntent("nonexistent.intent"))
}

func TestGetIntent_Public(t *testing.T) {
	intent := GetIntent("core.delete")
	require.NotNil(t, intent)
	assert.Equal(t, "delete", intent.Meta.Verb)

	// Non-existent intent
	intent = GetIntent("nonexistent.intent")
	assert.Nil(t, intent)
}

func TestIntentKeys(t *testing.T) {
	keys := IntentKeys()

	// Should contain core intents
	assert.Contains(t, keys, "core.delete")
	assert.Contains(t, keys, "core.create")
	assert.Contains(t, keys, "core.save")
	assert.Contains(t, keys, "core.commit")
	assert.Contains(t, keys, "core.push")

	// Should have a reasonable number of intents
	assert.GreaterOrEqual(t, len(keys), 20)
}

func TestCoreIntents_Structure(t *testing.T) {
	// Verify all core intents have required fields
	for key, intent := range coreIntents {
		t.Run(key, func(t *testing.T) {
			// Meta should be set
			assert.NotEmpty(t, intent.Meta.Type, "intent %s missing Type", key)
			assert.NotEmpty(t, intent.Meta.Verb, "intent %s missing Verb", key)
			assert.NotEmpty(t, intent.Meta.Default, "intent %s missing Default", key)

			// At least Question and one output should be set
			assert.NotEmpty(t, intent.Question, "intent %s missing Question", key)

			// Default should be valid
			assert.Contains(t, []string{"yes", "no"}, intent.Meta.Default,
				"intent %s has invalid Default: %s", key, intent.Meta.Default)

			// Type should be valid
			assert.Contains(t, []string{"action", "question", "info"}, intent.Meta.Type,
				"intent %s has invalid Type: %s", key, intent.Meta.Type)
		})
	}
}

func TestCoreIntents_Categories(t *testing.T) {
	// Destructive actions should be dangerous
	destructive := []string{"core.delete", "core.remove", "core.discard", "core.reset", "core.overwrite"}
	for _, key := range destructive {
		intent := getIntent(key)
		require.NotNil(t, intent, "missing intent: %s", key)
		assert.True(t, intent.Meta.Dangerous, "%s should be marked as dangerous", key)
		assert.Equal(t, "no", intent.Meta.Default, "%s should default to no", key)
	}

	// Creation actions should not be dangerous
	creation := []string{"core.create", "core.add", "core.clone", "core.copy"}
	for _, key := range creation {
		intent := getIntent(key)
		require.NotNil(t, intent, "missing intent: %s", key)
		assert.False(t, intent.Meta.Dangerous, "%s should not be marked as dangerous", key)
		assert.Equal(t, "yes", intent.Meta.Default, "%s should default to yes", key)
	}
}

func TestCoreIntents_Templates(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	tests := []struct {
		intent          string
		subject         *Subject
		expectedQ       string
		expectedSuccess string
	}{
		{
			intent:          "core.delete",
			subject:         S("file", "config.yaml"),
			expectedQ:       "Delete config.yaml?",
			expectedSuccess: "Config.Yaml deleted", // strings.Title capitalizes after dots
		},
		{
			intent:          "core.create",
			subject:         S("directory", "src"),
			expectedQ:       "Create src?",
			expectedSuccess: "Src created",
		},
		{
			intent:          "core.commit",
			subject:         S("changes", "3 files"),
			expectedQ:       "Commit 3 files?",
			expectedSuccess: "3 Files committed",
		},
		{
			intent:          "core.push",
			subject:         S("commits", "5 commits"),
			expectedQ:       "Push 5 commits?",
			expectedSuccess: "5 Commits pushed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.intent, func(t *testing.T) {
			result := svc.C(tt.intent, tt.subject)

			assert.Equal(t, tt.expectedQ, result.Question)
			assert.Equal(t, tt.expectedSuccess, result.Success)
		})
	}
}

func TestCoreIntents_TemplateErrors(t *testing.T) {
	// Templates with invalid syntax should return the original template
	RegisterIntent("test.invalid", Intent{
		Meta:     IntentMeta{Type: "action", Verb: "test", Default: "yes"},
		Question: "{{.Invalid",  // Invalid template syntax
		Success:  "Success",
		Failure:  "Failure",
	})
	defer delete(coreIntents, "test.invalid")

	svc, err := New()
	require.NoError(t, err)

	result := svc.C("test.invalid", S("item", "test"))
	// Should return the original invalid template
	assert.Equal(t, "{{.Invalid", result.Question)
}

func TestCoreIntents_TemplateFunctions(t *testing.T) {
	// Register an intent that uses template functions
	RegisterIntent("test.funcs", Intent{
		Meta: IntentMeta{Type: "action", Verb: "test", Default: "yes"},
		Question: "Process {{.Subject | quote}}?",
		Success:  "{{.Subject | upper}} processed",
		Failure:  "{{.Subject | lower}} failed",
	})
	defer delete(coreIntents, "test.funcs")

	svc, err := New()
	require.NoError(t, err)

	result := svc.C("test.funcs", S("item", "Test"))

	assert.Equal(t, `Process "Test"?`, result.Question)
	assert.Equal(t, "TEST processed", result.Success)
	assert.Equal(t, "test failed", result.Failure)
}

func TestIntentT_Integration(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Using T with core.* prefix and Subject should return Question form
	result := svc.T("core.delete", S("file", "config.yaml"))
	assert.Equal(t, "Delete config.yaml?", result)

	// Using T with regular key should work normally
	result = svc.T("cli.success")
	assert.Equal(t, "Success", result)
}

func TestIntent_EmptyTemplates(t *testing.T) {
	RegisterIntent("test.empty", Intent{
		Meta:     IntentMeta{Type: "info", Verb: "info", Default: "yes"},
		Question: "Question",
		Confirm:  "", // Empty confirm
		Success:  "", // Empty success
		Failure:  "", // Empty failure
	})
	defer delete(coreIntents, "test.empty")

	svc, err := New()
	require.NoError(t, err)

	result := svc.C("test.empty", S("item", "test"))

	assert.Equal(t, "Question", result.Question)
	assert.Equal(t, "", result.Confirm)
	assert.Equal(t, "", result.Success)
	assert.Equal(t, "", result.Failure)
}

func TestCoreIntents_AllKeysPrefixed(t *testing.T) {
	for key := range coreIntents {
		assert.True(t, strings.HasPrefix(key, "core."),
			"intent key %q should be prefixed with 'core.'", key)
	}
}
