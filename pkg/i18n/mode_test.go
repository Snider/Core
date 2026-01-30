package i18n

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMode_String(t *testing.T) {
	tests := []struct {
		mode     Mode
		expected string
	}{
		{ModeNormal, "normal"},
		{ModeStrict, "strict"},
		{ModeCollect, "collect"},
		{Mode(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.String())
		})
	}
}

func TestMissingKeyAction(t *testing.T) {
	action := MissingKeyAction{
		Key:        "test.missing.key",
		Args:       map[string]any{"Name": "test"},
		CallerFile: "/path/to/file.go",
		CallerLine: 42,
	}

	assert.Equal(t, "test.missing.key", action.Key)
	assert.Equal(t, "test", action.Args["Name"])
	assert.Equal(t, "/path/to/file.go", action.CallerFile)
	assert.Equal(t, 42, action.CallerLine)
}

func TestSetActionHandler(t *testing.T) {
	// Reset handler after test
	defer SetActionHandler(nil)

	t.Run("sets handler", func(t *testing.T) {
		var received MissingKeyAction
		SetActionHandler(func(action MissingKeyAction) {
			received = action
		})

		dispatchMissingKey("test.key", map[string]any{"foo": "bar"})

		assert.Equal(t, "test.key", received.Key)
		assert.Equal(t, "bar", received.Args["foo"])
	})

	t.Run("nil handler", func(t *testing.T) {
		SetActionHandler(nil)
		// Should not panic
		dispatchMissingKey("test.key", nil)
	})

	t.Run("replaces previous handler", func(t *testing.T) {
		called1 := false
		called2 := false

		SetActionHandler(func(action MissingKeyAction) {
			called1 = true
		})
		SetActionHandler(func(action MissingKeyAction) {
			called2 = true
		})

		dispatchMissingKey("test.key", nil)

		assert.False(t, called1)
		assert.True(t, called2)
	})
}

func TestServiceMode(t *testing.T) {
	// Reset default service after tests
	originalService := defaultService
	defer func() {
		defaultService = originalService
	}()

	t.Run("default mode is normal", func(t *testing.T) {
		defaultService = nil
		defaultOnce = sync.Once{}
		defaultErr = nil

		svc, err := New()
		require.NoError(t, err)

		assert.Equal(t, ModeNormal, svc.Mode())
	})

	t.Run("set mode", func(t *testing.T) {
		svc, err := New()
		require.NoError(t, err)

		svc.SetMode(ModeStrict)
		assert.Equal(t, ModeStrict, svc.Mode())

		svc.SetMode(ModeCollect)
		assert.Equal(t, ModeCollect, svc.Mode())

		svc.SetMode(ModeNormal)
		assert.Equal(t, ModeNormal, svc.Mode())
	})
}

func TestModeNormal_MissingKey(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	svc.SetMode(ModeNormal)

	// Missing key should return the key itself
	result := svc.T("nonexistent.key")
	assert.Equal(t, "nonexistent.key", result)
}

func TestModeStrict_MissingKey(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	svc.SetMode(ModeStrict)

	// Missing key should panic
	assert.Panics(t, func() {
		svc.T("nonexistent.key")
	})
}

func TestModeCollect_MissingKey(t *testing.T) {
	// Reset handler after test
	defer SetActionHandler(nil)

	svc, err := New()
	require.NoError(t, err)

	svc.SetMode(ModeCollect)

	var received MissingKeyAction
	SetActionHandler(func(action MissingKeyAction) {
		received = action
	})

	// Missing key should dispatch action and return [key]
	result := svc.T("nonexistent.key", map[string]any{"arg": "value"})

	assert.Equal(t, "[nonexistent.key]", result)
	assert.Equal(t, "nonexistent.key", received.Key)
	assert.Equal(t, "value", received.Args["arg"])
	assert.NotEmpty(t, received.CallerFile)
	assert.Greater(t, received.CallerLine, 0)
}

