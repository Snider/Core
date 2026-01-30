package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslationContext_C(t *testing.T) {
	t.Run("creates context", func(t *testing.T) {
		ctx := C("navigation")
		assert.NotNil(t, ctx)
		assert.Equal(t, "navigation", ctx.Context)
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := C("")
		assert.NotNil(t, ctx)
		assert.Empty(t, ctx.Context)
	})
}

func TestTranslationContext_WithGender(t *testing.T) {
	t.Run("sets gender", func(t *testing.T) {
		ctx := C("context").WithGender("masculine")
		assert.Equal(t, "masculine", ctx.Gender)
	})

	t.Run("nil safety", func(t *testing.T) {
		var ctx *TranslationContext
		result := ctx.WithGender("masculine")
		assert.Nil(t, result)
	})
}

func TestTranslationContext_Formality(t *testing.T) {
	t.Run("Formal", func(t *testing.T) {
		ctx := C("context").Formal()
		assert.Equal(t, FormalityFormal, ctx.Formality)
	})

	t.Run("Informal", func(t *testing.T) {
		ctx := C("context").Informal()
		assert.Equal(t, FormalityInformal, ctx.Formality)
	})

	t.Run("WithFormality", func(t *testing.T) {
		ctx := C("context").WithFormality(FormalityFormal)
		assert.Equal(t, FormalityFormal, ctx.Formality)
	})

	t.Run("nil safety", func(t *testing.T) {
		var ctx *TranslationContext
		assert.Nil(t, ctx.Formal())
		assert.Nil(t, ctx.Informal())
		assert.Nil(t, ctx.WithFormality(FormalityFormal))
	})
}

func TestTranslationContext_Extra(t *testing.T) {
	t.Run("Set and Get", func(t *testing.T) {
		ctx := C("context").Set("key", "value")
		assert.Equal(t, "value", ctx.Get("key"))
	})

	t.Run("Get missing key", func(t *testing.T) {
		ctx := C("context")
		assert.Nil(t, ctx.Get("missing"))
	})

	t.Run("nil safety Set", func(t *testing.T) {
		var ctx *TranslationContext
		result := ctx.Set("key", "value")
		assert.Nil(t, result)
	})

	t.Run("nil safety Get", func(t *testing.T) {
		var ctx *TranslationContext
		assert.Nil(t, ctx.Get("key"))
	})
}

func TestTranslationContext_Getters(t *testing.T) {
	t.Run("ContextString", func(t *testing.T) {
		ctx := C("navigation")
		assert.Equal(t, "navigation", ctx.ContextString())
	})

	t.Run("ContextString nil", func(t *testing.T) {
		var ctx *TranslationContext
		assert.Empty(t, ctx.ContextString())
	})

	t.Run("GenderString", func(t *testing.T) {
		ctx := C("context").WithGender("feminine")
		assert.Equal(t, "feminine", ctx.GenderString())
	})

	t.Run("GenderString nil", func(t *testing.T) {
		var ctx *TranslationContext
		assert.Empty(t, ctx.GenderString())
	})

	t.Run("FormalityValue", func(t *testing.T) {
		ctx := C("context").Formal()
		assert.Equal(t, FormalityFormal, ctx.FormalityValue())
	})

	t.Run("FormalityValue nil", func(t *testing.T) {
		var ctx *TranslationContext
		assert.Equal(t, FormalityNeutral, ctx.FormalityValue())
	})
}

func TestTranslationContext_Chaining(t *testing.T) {
	ctx := C("navigation").
		WithGender("masculine").
		Formal().
		Set("locale", "de-DE")

	assert.Equal(t, "navigation", ctx.Context)
	assert.Equal(t, "masculine", ctx.Gender)
	assert.Equal(t, FormalityFormal, ctx.Formality)
	assert.Equal(t, "de-DE", ctx.Get("locale"))
}
