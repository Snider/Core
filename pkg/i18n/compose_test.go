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

	t.Run("NewSubject alias", func(t *testing.T) {
		s := NewSubject("repo", "core-php")
		assert.Equal(t, "repo", s.Noun)
		assert.Equal(t, "core-php", s.Value)
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

	t.Run("with formality", func(t *testing.T) {
		s := S("user", "Hans").Formal()
		data := newTemplateData(s)

		assert.Equal(t, FormalityFormal, data.Formality)
		assert.True(t, data.IsFormal)
	})

	t.Run("with plural", func(t *testing.T) {
		s := S("file", "*.go").Count(5)
		data := newTemplateData(s)

		assert.True(t, data.IsPlural)
		assert.Equal(t, 5, data.Count)
	})
}

func TestSubject_Formality(t *testing.T) {
	t.Run("default is neutral", func(t *testing.T) {
		s := S("user", "name")
		assert.Equal(t, FormalityNeutral, s.GetFormality())
		assert.False(t, s.IsFormal())
		assert.False(t, s.IsInformal())
	})

	t.Run("Formal()", func(t *testing.T) {
		s := S("user", "name").Formal()
		assert.Equal(t, FormalityFormal, s.GetFormality())
		assert.True(t, s.IsFormal())
	})

	t.Run("Informal()", func(t *testing.T) {
		s := S("user", "name").Informal()
		assert.Equal(t, FormalityInformal, s.GetFormality())
		assert.True(t, s.IsInformal())
	})

	t.Run("Formality() explicit", func(t *testing.T) {
		s := S("user", "name").Formality(FormalityFormal)
		assert.Equal(t, FormalityFormal, s.GetFormality())
	})

	t.Run("nil safety", func(t *testing.T) {
		var s *Subject
		assert.Equal(t, FormalityNeutral, s.GetFormality())
		assert.False(t, s.IsFormal())
		assert.False(t, s.IsInformal())
	})
}

func TestIntentBuilder(t *testing.T) {
	// Initialize the default service for tests
	_ = Init()

	t.Run("basic fluent API", func(t *testing.T) {
		builder := I("core.delete").For(S("file", "config.yaml"))
		assert.NotNil(t, builder)
	})

	t.Run("With alias", func(t *testing.T) {
		builder := I("core.delete").With(S("file", "config.yaml"))
		assert.NotNil(t, builder)
	})

	t.Run("Compose returns all forms", func(t *testing.T) {
		result := I("core.delete").For(S("file", "config.yaml")).Compose()
		assert.NotEmpty(t, result.Question)
		assert.NotEmpty(t, result.Success)
		assert.NotEmpty(t, result.Failure)
	})

	t.Run("Question returns string", func(t *testing.T) {
		question := I("core.delete").For(S("file", "config.yaml")).Question()
		assert.Contains(t, question, "config.yaml")
	})

	t.Run("Success returns string", func(t *testing.T) {
		success := I("core.delete").For(S("file", "config.yaml")).Success()
		assert.NotEmpty(t, success)
	})

	t.Run("Failure returns string", func(t *testing.T) {
		failure := I("core.delete").For(S("file", "config.yaml")).Failure()
		assert.Contains(t, failure, "delete")
	})

	t.Run("Meta returns metadata", func(t *testing.T) {
		meta := I("core.delete").Meta()
		assert.True(t, meta.Dangerous)
	})

	t.Run("IsDangerous helper", func(t *testing.T) {
		assert.True(t, I("core.delete").IsDangerous())
		assert.False(t, I("core.save").IsDangerous())
	})
}
