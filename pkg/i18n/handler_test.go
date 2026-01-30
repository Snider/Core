package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelHandler(t *testing.T) {
	h := LabelHandler{}

	t.Run("matches i18n.label prefix", func(t *testing.T) {
		assert.True(t, h.Match("i18n.label.status"))
		assert.True(t, h.Match("i18n.label.version"))
		assert.False(t, h.Match("i18n.progress.build"))
		assert.False(t, h.Match("cli.label.status"))
	})

	t.Run("handles label", func(t *testing.T) {
		result := h.Handle("i18n.label.status", nil, func() string { return "fallback" })
		assert.Equal(t, "Status:", result)
	})
}

func TestProgressHandler(t *testing.T) {
	h := ProgressHandler{}

	t.Run("matches i18n.progress prefix", func(t *testing.T) {
		assert.True(t, h.Match("i18n.progress.build"))
		assert.True(t, h.Match("i18n.progress.check"))
		assert.False(t, h.Match("i18n.label.status"))
	})

	t.Run("handles progress without subject", func(t *testing.T) {
		result := h.Handle("i18n.progress.build", nil, func() string { return "fallback" })
		assert.Equal(t, "Building...", result)
	})

	t.Run("handles progress with subject", func(t *testing.T) {
		result := h.Handle("i18n.progress.check", []any{"config"}, func() string { return "fallback" })
		assert.Equal(t, "Checking config...", result)
	})
}

func TestCountHandler(t *testing.T) {
	h := CountHandler{}

	t.Run("matches i18n.count prefix", func(t *testing.T) {
		assert.True(t, h.Match("i18n.count.file"))
		assert.True(t, h.Match("i18n.count.repo"))
		assert.False(t, h.Match("i18n.label.count"))
	})

	t.Run("handles count with number", func(t *testing.T) {
		result := h.Handle("i18n.count.file", []any{5}, func() string { return "fallback" })
		assert.Equal(t, "5 files", result)
	})

	t.Run("handles singular count", func(t *testing.T) {
		result := h.Handle("i18n.count.file", []any{1}, func() string { return "fallback" })
		assert.Equal(t, "1 file", result)
	})

	t.Run("handles no args", func(t *testing.T) {
		result := h.Handle("i18n.count.file", nil, func() string { return "fallback" })
		assert.Equal(t, "file", result)
	})
}

func TestDoneHandler(t *testing.T) {
	h := DoneHandler{}

	t.Run("matches i18n.done prefix", func(t *testing.T) {
		assert.True(t, h.Match("i18n.done.delete"))
		assert.True(t, h.Match("i18n.done.save"))
		assert.False(t, h.Match("i18n.fail.delete"))
	})

	t.Run("handles done with subject", func(t *testing.T) {
		result := h.Handle("i18n.done.delete", []any{"config.yaml"}, func() string { return "fallback" })
		// ActionResult title-cases the subject
		assert.Equal(t, "Config.Yaml deleted", result)
	})

	t.Run("handles done without subject", func(t *testing.T) {
		result := h.Handle("i18n.done.delete", nil, func() string { return "fallback" })
		assert.Equal(t, "Deleted", result)
	})
}

func TestFailHandler(t *testing.T) {
	h := FailHandler{}

	t.Run("matches i18n.fail prefix", func(t *testing.T) {
		assert.True(t, h.Match("i18n.fail.delete"))
		assert.True(t, h.Match("i18n.fail.save"))
		assert.False(t, h.Match("i18n.done.delete"))
	})

	t.Run("handles fail with subject", func(t *testing.T) {
		result := h.Handle("i18n.fail.delete", []any{"config.yaml"}, func() string { return "fallback" })
		assert.Equal(t, "Failed to delete config.yaml", result)
	})

	t.Run("handles fail without subject", func(t *testing.T) {
		result := h.Handle("i18n.fail.delete", nil, func() string { return "fallback" })
		assert.Contains(t, result, "Failed to delete")
	})
}

func TestNumericHandler(t *testing.T) {
	h := NumericHandler{}

	t.Run("matches i18n.numeric prefix", func(t *testing.T) {
		assert.True(t, h.Match("i18n.numeric.number"))
		assert.True(t, h.Match("i18n.numeric.bytes"))
		assert.False(t, h.Match("i18n.count.file"))
	})

	t.Run("handles number format", func(t *testing.T) {
		result := h.Handle("i18n.numeric.number", []any{1234567}, func() string { return "fallback" })
		assert.Equal(t, "1,234,567", result)
	})

	t.Run("handles bytes format", func(t *testing.T) {
		result := h.Handle("i18n.numeric.bytes", []any{1024}, func() string { return "fallback" })
		assert.Equal(t, "1 KB", result)
	})

	t.Run("handles ordinal format", func(t *testing.T) {
		result := h.Handle("i18n.numeric.ordinal", []any{3}, func() string { return "fallback" })
		assert.Equal(t, "3rd", result)
	})

	t.Run("falls through on no args", func(t *testing.T) {
		result := h.Handle("i18n.numeric.number", nil, func() string { return "fallback" })
		assert.Equal(t, "fallback", result)
	})

	t.Run("falls through on unknown format", func(t *testing.T) {
		result := h.Handle("i18n.numeric.unknown", []any{123}, func() string { return "fallback" })
		assert.Equal(t, "fallback", result)
	})
}

func TestDefaultHandlers(t *testing.T) {
	handlers := DefaultHandlers()
	assert.Len(t, handlers, 6)
}

func TestRunHandlerChain(t *testing.T) {
	handlers := DefaultHandlers()

	t.Run("label handler matches", func(t *testing.T) {
		result := RunHandlerChain(handlers, "i18n.label.status", nil, func() string { return "fallback" })
		assert.Equal(t, "Status:", result)
	})

	t.Run("progress handler matches", func(t *testing.T) {
		result := RunHandlerChain(handlers, "i18n.progress.build", nil, func() string { return "fallback" })
		assert.Equal(t, "Building...", result)
	})

	t.Run("falls back for unknown key", func(t *testing.T) {
		result := RunHandlerChain(handlers, "cli.unknown", nil, func() string { return "fallback" })
		assert.Equal(t, "fallback", result)
	})

	t.Run("empty handler chain uses fallback", func(t *testing.T) {
		result := RunHandlerChain(nil, "any.key", nil, func() string { return "fallback" })
		assert.Equal(t, "fallback", result)
	})
}
