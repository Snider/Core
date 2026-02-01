package sdk

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSDK_Good_SetVersion(t *testing.T) {
	s := New("/tmp", nil)
	s.SetVersion("v1.2.3")

	assert.Equal(t, "v1.2.3", s.version)
}

func TestSDK_Good_VersionPassedToGenerator(t *testing.T) {
	config := &Config{
		Languages: []string{"typescript"},
		Output:    "sdk",
		Package: PackageConfig{
			Name: "test-sdk",
		},
	}
	s := New("/tmp", config)
	s.SetVersion("v2.0.0")

	assert.Equal(t, "v2.0.0", s.config.Package.Version)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Contains(t, cfg.Languages, "typescript")
	assert.Equal(t, "sdk", cfg.Output)
	assert.True(t, cfg.Diff.Enabled)
}

func TestSDK_New(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		s := New("/tmp", nil)
		assert.NotNil(t, s.config)
		assert.Equal(t, "sdk", s.config.Output)
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := &Config{Output: "custom"}
		s := New("/tmp", cfg)
		assert.Equal(t, "custom", s.config.Output)
	})
}

func TestSDK_GenerateLanguage_Bad(t *testing.T) {

	t.Run("unknown language", func(t *testing.T) {

		tmpDir := t.TempDir()

		specPath := filepath.Join(tmpDir, "openapi.yaml")

		err := os.WriteFile(specPath, []byte("openapi: 3.0.0"), 0644)

		require.NoError(t, err)

		s := New(tmpDir, nil)

		err = s.GenerateLanguage(context.Background(), "invalid-lang")

		assert.Error(t, err)

		assert.Contains(t, err.Error(), "unknown language")

	})

}
