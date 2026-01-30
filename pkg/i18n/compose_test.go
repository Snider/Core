package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// stringerValue is a test helper that implements fmt.Stringer
type stringerValue struct {
	value string
}

func (s stringerValue) String() string {
	return s.value
}

func TestSubject_Good(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		s := S("file", "config.yaml")
		assert.Equal(t, "file", s.Noun)
		assert.Equal(t, "config.yaml", s.Value)
		assert.Equal(t, 1, s.count)
		assert.Equal(t, "", s.gender)
		assert.Equal(t, "", s.location)
	})

	t.Run("with count", func(t *testing.T) {
		s := S("file", "*.go").Count(5)
		assert.Equal(t, 5, s.GetCount())
		assert.True(t, s.IsPlural())
	})

	t.Run("with gender", func(t *testing.T) {
		s := S("user", "alice").Gender("female")
		assert.Equal(t, "female", s.GetGender())
	})

	t.Run("with location", func(t *testing.T) {
		s := S("file", "config.yaml").In("workspace")
		assert.Equal(t, "workspace", s.GetLocation())
	})

	t.Run("chained methods", func(t *testing.T) {
		s := S("repo", "core-php").Count(3).Gender("neuter").In("organisation")
		assert.Equal(t, "repo", s.GetNoun())
		assert.Equal(t, 3, s.GetCount())
		assert.Equal(t, "neuter", s.GetGender())
		assert.Equal(t, "organisation", s.GetLocation())
	})
}

func TestSubject_String(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		s := S("file", "config.yaml")
		assert.Equal(t, "config.yaml", s.String())
	})

	t.Run("stringer interface", func(t *testing.T) {
		// Using a struct that implements Stringer via embedded method
		s := S("item", stringerValue{"test"})
		assert.Equal(t, "test", s.String())
	})

	t.Run("nil subject", func(t *testing.T) {
		var s *Subject
		assert.Equal(t, "", s.String())
	})

	t.Run("int value", func(t *testing.T) {
		s := S("count", 42)
		assert.Equal(t, "42", s.String())
	})
}

func TestSubject_IsPlural(t *testing.T) {
	t.Run("singular (count 1)", func(t *testing.T) {
		s := S("file", "test.go")
		assert.False(t, s.IsPlural())
	})

	t.Run("plural (count 0)", func(t *testing.T) {
		s := S("file", "*.go").Count(0)
		assert.True(t, s.IsPlural())
	})

	t.Run("plural (count > 1)", func(t *testing.T) {
		s := S("file", "*.go").Count(5)
		assert.True(t, s.IsPlural())
	})

	t.Run("nil subject", func(t *testing.T) {
		var s *Subject
		assert.False(t, s.IsPlural())
	})
}

func TestSubject_Getters(t *testing.T) {
	t.Run("nil safety", func(t *testing.T) {
		var s *Subject
		assert.Equal(t, "", s.GetNoun())
		assert.Equal(t, 1, s.GetCount())
		assert.Equal(t, "", s.GetGender())
		assert.Equal(t, "", s.GetLocation())
	})
}

func TestIntentMeta(t *testing.T) {
	meta := IntentMeta{
		Type:      "action",
		Verb:      "delete",
		Dangerous: true,
		Default:   "no",
		Supports:  []string{"force", "recursive"},
	}

	assert.Equal(t, "action", meta.Type)
	assert.Equal(t, "delete", meta.Verb)
	assert.True(t, meta.Dangerous)
	assert.Equal(t, "no", meta.Default)
	assert.Contains(t, meta.Supports, "force")
	assert.Contains(t, meta.Supports, "recursive")
}

func TestComposed(t *testing.T) {
	composed := Composed{
		Question: "Delete config.yaml?",
		Confirm:  "Really delete config.yaml?",
		Success:  "Config.yaml deleted",
		Failure:  "Failed to delete config.yaml",
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "delete",
			Dangerous: true,
			Default:   "no",
		},
	}

	assert.Equal(t, "Delete config.yaml?", composed.Question)
	assert.Equal(t, "Really delete config.yaml?", composed.Confirm)
	assert.Equal(t, "Config.yaml deleted", composed.Success)
	assert.Equal(t, "Failed to delete config.yaml", composed.Failure)
	assert.True(t, composed.Meta.Dangerous)
}

func TestNewTemplateData(t *testing.T) {
	t.Run("from subject", func(t *testing.T) {
		s := S("file", "config.yaml").Count(3).Gender("neuter").In("workspace")
		data := newTemplateData(s)

		assert.Equal(t, "config.yaml", data.Subject)
		assert.Equal(t, "file", data.Noun)
		assert.Equal(t, 3, data.Count)
		assert.Equal(t, "neuter", data.Gender)
		assert.Equal(t, "workspace", data.Location)
		assert.Equal(t, "config.yaml", data.Value)
	})

	t.Run("from nil subject", func(t *testing.T) {
		data := newTemplateData(nil)

		assert.Equal(t, "", data.Subject)
		assert.Equal(t, "", data.Noun)
		assert.Equal(t, 1, data.Count)
		assert.Equal(t, "", data.Gender)
		assert.Equal(t, "", data.Location)
		assert.Nil(t, data.Value)
	})
}
