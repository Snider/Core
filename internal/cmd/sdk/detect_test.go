package sdk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectSpec_Good_ConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "api", "spec.yaml")
	err := os.MkdirAll(filepath.Dir(specPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(specPath, []byte("openapi: 3.0.0"), 0644)
	require.NoError(t, err)

	sdk := New(tmpDir, &Config{Spec: "api/spec.yaml"})
	got, err := sdk.DetectSpec()
	assert.NoError(t, err)
	assert.Equal(t, specPath, got)
}

func TestDetectSpec_Good_CommonPath(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.yaml")
	err := os.WriteFile(specPath, []byte("openapi: 3.0.0"), 0644)
	require.NoError(t, err)

	sdk := New(tmpDir, nil)
	got, err := sdk.DetectSpec()
	assert.NoError(t, err)
	assert.Equal(t, specPath, got)
}

func TestDetectSpec_Bad_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	sdk := New(tmpDir, nil)
	_, err := sdk.DetectSpec()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no OpenAPI spec found")
}

func TestDetectSpec_Bad_ConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	sdk := New(tmpDir, &Config{Spec: "non-existent.yaml"})
	_, err := sdk.DetectSpec()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configured spec not found")
}

func TestContainsScramble(t *testing.T) {
	tests := []struct {
		data     string
		expected bool
	}{
		{`{"require": {"dedoc/scramble": "^0.1"}}`, true},
		{`{"require": {"scramble": "^0.1"}}`, true},
		{`{"require": {"laravel/framework": "^11.0"}}`, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, containsScramble(tt.data))
	}
}

func TestDetectScramble_Bad(t *testing.T) {
	t.Run("no composer.json", func(t *testing.T) {
		sdk := New(t.TempDir(), nil)
		_, err := sdk.detectScramble()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no composer.json")
	})

	t.Run("no scramble in composer.json", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(`{}`), 0644)
		require.NoError(t, err)

		sdk := New(tmpDir, nil)
		_, err = sdk.detectScramble()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scramble not found")
	})
}
