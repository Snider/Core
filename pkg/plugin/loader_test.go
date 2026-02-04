package plugin

import (
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestLoader_Discover_Good(t *testing.T) {
	m := io.NewMockMedium()
	baseDir := "/home/user/.core/plugins"

	// Set up mock filesystem with two plugins
	m.Dirs[baseDir] = true
	m.Dirs[baseDir+"/plugin-a"] = true
	m.Dirs[baseDir+"/plugin-b"] = true

	m.Files[baseDir+"/plugin-a/plugin.json"] = `{
		"name": "plugin-a",
		"version": "1.0.0",
		"description": "Plugin A",
		"entrypoint": "main.go"
	}`

	m.Files[baseDir+"/plugin-b/plugin.json"] = `{
		"name": "plugin-b",
		"version": "2.0.0",
		"description": "Plugin B",
		"entrypoint": "run.sh"
	}`

	loader := NewLoader(m, baseDir)
	manifests, err := loader.Discover()
	assert.NoError(t, err)
	assert.Len(t, manifests, 2)

	names := make(map[string]bool)
	for _, manifest := range manifests {
		names[manifest.Name] = true
	}
	assert.True(t, names["plugin-a"])
	assert.True(t, names["plugin-b"])
}

func TestLoader_Discover_Good_SkipsInvalidPlugins(t *testing.T) {
	m := io.NewMockMedium()
	baseDir := "/home/user/.core/plugins"

	m.Dirs[baseDir] = true
	m.Dirs[baseDir+"/good-plugin"] = true
	m.Dirs[baseDir+"/bad-plugin"] = true

	// Valid plugin
	m.Files[baseDir+"/good-plugin/plugin.json"] = `{
		"name": "good-plugin",
		"version": "1.0.0",
		"entrypoint": "main.go"
	}`

	// Invalid plugin (bad JSON)
	m.Files[baseDir+"/bad-plugin/plugin.json"] = `{invalid}`

	loader := NewLoader(m, baseDir)
	manifests, err := loader.Discover()
	assert.NoError(t, err)
	assert.Len(t, manifests, 1)
	assert.Equal(t, "good-plugin", manifests[0].Name)
}

func TestLoader_Discover_Good_SkipsFiles(t *testing.T) {
	m := io.NewMockMedium()
	baseDir := "/home/user/.core/plugins"

	m.Dirs[baseDir] = true
	m.Dirs[baseDir+"/real-plugin"] = true
	m.Files[baseDir+"/registry.json"] = `{}` // A file, not a directory

	m.Files[baseDir+"/real-plugin/plugin.json"] = `{
		"name": "real-plugin",
		"version": "1.0.0",
		"entrypoint": "main.go"
	}`

	loader := NewLoader(m, baseDir)
	manifests, err := loader.Discover()
	assert.NoError(t, err)
	assert.Len(t, manifests, 1)
	assert.Equal(t, "real-plugin", manifests[0].Name)
}

func TestLoader_Discover_Good_EmptyDirectory(t *testing.T) {
	m := io.NewMockMedium()
	baseDir := "/home/user/.core/plugins"
	m.Dirs[baseDir] = true

	loader := NewLoader(m, baseDir)
	manifests, err := loader.Discover()
	assert.NoError(t, err)
	assert.Empty(t, manifests)
}

func TestLoader_LoadPlugin_Good(t *testing.T) {
	m := io.NewMockMedium()
	baseDir := "/home/user/.core/plugins"

	m.Dirs[baseDir+"/my-plugin"] = true
	m.Files[baseDir+"/my-plugin/plugin.json"] = `{
		"name": "my-plugin",
		"version": "1.0.0",
		"description": "My plugin",
		"author": "Test",
		"entrypoint": "main.go"
	}`

	loader := NewLoader(m, baseDir)
	manifest, err := loader.LoadPlugin("my-plugin")
	assert.NoError(t, err)
	assert.Equal(t, "my-plugin", manifest.Name)
	assert.Equal(t, "1.0.0", manifest.Version)
}

func TestLoader_LoadPlugin_Bad_NotFound(t *testing.T) {
	m := io.NewMockMedium()
	loader := NewLoader(m, "/home/user/.core/plugins")

	_, err := loader.LoadPlugin("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load plugin")
}

func TestLoader_LoadPlugin_Bad_InvalidManifest(t *testing.T) {
	m := io.NewMockMedium()
	baseDir := "/home/user/.core/plugins"

	m.Dirs[baseDir+"/bad-plugin"] = true
	m.Files[baseDir+"/bad-plugin/plugin.json"] = `{
		"name": "bad-plugin",
		"version": "1.0.0"
	}` // Missing entrypoint

	loader := NewLoader(m, baseDir)
	_, err := loader.LoadPlugin("bad-plugin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plugin manifest")
}
