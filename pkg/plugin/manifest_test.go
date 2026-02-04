package plugin

import (
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestLoadManifest_Good(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["plugins/test/plugin.json"] = `{
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "A test plugin",
		"author": "Test Author",
		"entrypoint": "main.go",
		"dependencies": ["dep-a", "dep-b"],
		"min_version": "0.5.0"
	}`

	manifest, err := LoadManifest(m, "plugins/test/plugin.json")
	assert.NoError(t, err)
	assert.Equal(t, "test-plugin", manifest.Name)
	assert.Equal(t, "1.0.0", manifest.Version)
	assert.Equal(t, "A test plugin", manifest.Description)
	assert.Equal(t, "Test Author", manifest.Author)
	assert.Equal(t, "main.go", manifest.Entrypoint)
	assert.Equal(t, []string{"dep-a", "dep-b"}, manifest.Dependencies)
	assert.Equal(t, "0.5.0", manifest.MinVersion)
}

func TestLoadManifest_Good_MinimalFields(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["plugin.json"] = `{
		"name": "minimal",
		"version": "0.1.0",
		"entrypoint": "run.sh"
	}`

	manifest, err := LoadManifest(m, "plugin.json")
	assert.NoError(t, err)
	assert.Equal(t, "minimal", manifest.Name)
	assert.Equal(t, "0.1.0", manifest.Version)
	assert.Equal(t, "run.sh", manifest.Entrypoint)
	assert.Empty(t, manifest.Dependencies)
	assert.Empty(t, manifest.MinVersion)
}

func TestLoadManifest_Bad_FileNotFound(t *testing.T) {
	m := io.NewMockMedium()

	_, err := LoadManifest(m, "nonexistent/plugin.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read manifest")
}

func TestLoadManifest_Bad_InvalidJSON(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["plugin.json"] = `{invalid json}`

	_, err := LoadManifest(m, "plugin.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse manifest JSON")
}

func TestManifest_Validate_Good(t *testing.T) {
	manifest := &Manifest{
		Name:       "test-plugin",
		Version:    "1.0.0",
		Entrypoint: "main.go",
	}

	err := manifest.Validate()
	assert.NoError(t, err)
}

func TestManifest_Validate_Bad_MissingName(t *testing.T) {
	manifest := &Manifest{
		Version:    "1.0.0",
		Entrypoint: "main.go",
	}

	err := manifest.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestManifest_Validate_Bad_MissingVersion(t *testing.T) {
	manifest := &Manifest{
		Name:       "test-plugin",
		Entrypoint: "main.go",
	}

	err := manifest.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version is required")
}

func TestManifest_Validate_Bad_MissingEntrypoint(t *testing.T) {
	manifest := &Manifest{
		Name:    "test-plugin",
		Version: "1.0.0",
	}

	err := manifest.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entrypoint is required")
}
