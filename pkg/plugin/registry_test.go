package plugin

import (
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_Add_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/home/user/.core/plugins")

	err := reg.Add(&PluginConfig{
		Name:    "my-plugin",
		Version: "1.0.0",
		Source:  "github:org/my-plugin",
		Enabled: true,
	})
	assert.NoError(t, err)

	list := reg.List()
	assert.Len(t, list, 1)
	assert.Equal(t, "my-plugin", list[0].Name)
	assert.Equal(t, "1.0.0", list[0].Version)
}

func TestRegistry_Add_Bad_EmptyName(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/home/user/.core/plugins")

	err := reg.Add(&PluginConfig{
		Version: "1.0.0",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin name is required")
}

func TestRegistry_Remove_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/home/user/.core/plugins")

	_ = reg.Add(&PluginConfig{
		Name:    "my-plugin",
		Version: "1.0.0",
	})

	err := reg.Remove("my-plugin")
	assert.NoError(t, err)
	assert.Empty(t, reg.List())
}

func TestRegistry_Get_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/home/user/.core/plugins")

	_ = reg.Add(&PluginConfig{
		Name:    "test-plugin",
		Version: "2.0.0",
		Source:  "github:org/test-plugin",
	})

	cfg, ok := reg.Get("test-plugin")
	assert.True(t, ok)
	assert.Equal(t, "test-plugin", cfg.Name)
	assert.Equal(t, "2.0.0", cfg.Version)
}

func TestRegistry_Get_Bad_NotFound(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/home/user/.core/plugins")

	cfg, ok := reg.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, cfg)
}

func TestRegistry_Remove_Bad_NotFound(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/home/user/.core/plugins")

	err := reg.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestRegistry_SaveLoad_Good(t *testing.T) {
	m := io.NewMockMedium()
	basePath := "/home/user/.core/plugins"
	reg := NewRegistry(m, basePath)

	_ = reg.Add(&PluginConfig{
		Name:        "plugin-a",
		Version:     "1.0.0",
		Source:      "github:org/plugin-a",
		Enabled:     true,
		InstalledAt: "2025-01-01T00:00:00Z",
	})
	_ = reg.Add(&PluginConfig{
		Name:        "plugin-b",
		Version:     "2.0.0",
		Source:      "github:org/plugin-b",
		Enabled:     false,
		InstalledAt: "2025-01-02T00:00:00Z",
	})

	err := reg.Save()
	assert.NoError(t, err)

	// Load into a fresh registry
	reg2 := NewRegistry(m, basePath)
	err = reg2.Load()
	assert.NoError(t, err)

	list := reg2.List()
	assert.Len(t, list, 2)

	a, ok := reg2.Get("plugin-a")
	assert.True(t, ok)
	assert.Equal(t, "1.0.0", a.Version)
	assert.True(t, a.Enabled)

	b, ok := reg2.Get("plugin-b")
	assert.True(t, ok)
	assert.Equal(t, "2.0.0", b.Version)
	assert.False(t, b.Enabled)
}

func TestRegistry_Load_Good_EmptyWhenNoFile(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/home/user/.core/plugins")

	err := reg.Load()
	assert.NoError(t, err)
	assert.Empty(t, reg.List())
}
