package config

import (
	"os"
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Get_Good(t *testing.T) {
	m := io.NewMockMedium()

	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	err = cfg.Set("app.name", "core")
	assert.NoError(t, err)

	var name string
	err = cfg.Get("app.name", &name)
	assert.NoError(t, err)
	assert.Equal(t, "core", name)
}

func TestConfig_Get_Bad(t *testing.T) {
	m := io.NewMockMedium()

	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	var value string
	err = cfg.Get("nonexistent.key", &value)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestConfig_Set_Good(t *testing.T) {
	m := io.NewMockMedium()

	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	err = cfg.Set("dev.editor", "vim")
	assert.NoError(t, err)

	// Verify the value was saved to the medium
	content, readErr := m.Read("/tmp/test/config.yaml")
	assert.NoError(t, readErr)
	assert.Contains(t, content, "editor: vim")

	// Verify we can read it back
	var editor string
	err = cfg.Get("dev.editor", &editor)
	assert.NoError(t, err)
	assert.Equal(t, "vim", editor)
}

func TestConfig_Set_Nested_Good(t *testing.T) {
	m := io.NewMockMedium()

	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	err = cfg.Set("a.b.c", "deep")
	assert.NoError(t, err)

	var val string
	err = cfg.Get("a.b.c", &val)
	assert.NoError(t, err)
	assert.Equal(t, "deep", val)
}

func TestConfig_All_Good(t *testing.T) {
	m := io.NewMockMedium()

	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	_ = cfg.Set("key1", "val1")
	_ = cfg.Set("key2", "val2")

	all := cfg.All()
	assert.Equal(t, "val1", all["key1"])
	assert.Equal(t, "val2", all["key2"])
}

func TestConfig_Path_Good(t *testing.T) {
	m := io.NewMockMedium()

	cfg, err := New(WithMedium(m), WithPath("/custom/path/config.yaml"))
	assert.NoError(t, err)

	assert.Equal(t, "/custom/path/config.yaml", cfg.Path())
}

func TestConfig_Load_Existing_Good(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["/tmp/test/config.yaml"] = "app:\n  name: existing\n"

	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	var name string
	err = cfg.Get("app.name", &name)
	assert.NoError(t, err)
	assert.Equal(t, "existing", name)
}

func TestConfig_Env_Good(t *testing.T) {
	// Set environment variable
	t.Setenv("CORE_CONFIG_DEV_EDITOR", "nano")

	m := io.NewMockMedium()
	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	var editor string
	err = cfg.Get("dev.editor", &editor)
	assert.NoError(t, err)
	assert.Equal(t, "nano", editor)
}

func TestConfig_Env_Overrides_File_Good(t *testing.T) {
	// Set file config
	m := io.NewMockMedium()
	m.Files["/tmp/test/config.yaml"] = "dev:\n  editor: vim\n"

	// Set environment override
	t.Setenv("CORE_CONFIG_DEV_EDITOR", "nano")

	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	var editor string
	err = cfg.Get("dev.editor", &editor)
	assert.NoError(t, err)
	assert.Equal(t, "nano", editor)
}

func TestConfig_Assign_Types_Good(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["/tmp/test/config.yaml"] = "count: 42\nenabled: true\nratio: 3.14\n"

	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	var count int
	err = cfg.Get("count", &count)
	assert.NoError(t, err)
	assert.Equal(t, 42, count)

	var enabled bool
	err = cfg.Get("enabled", &enabled)
	assert.NoError(t, err)
	assert.True(t, enabled)

	var ratio float64
	err = cfg.Get("ratio", &ratio)
	assert.NoError(t, err)
	assert.InDelta(t, 3.14, ratio, 0.001)
}

func TestConfig_Assign_Any_Good(t *testing.T) {
	m := io.NewMockMedium()

	cfg, err := New(WithMedium(m), WithPath("/tmp/test/config.yaml"))
	assert.NoError(t, err)

	_ = cfg.Set("key", "value")

	var val any
	err = cfg.Get("key", &val)
	assert.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestConfig_DefaultPath_Good(t *testing.T) {
	m := io.NewMockMedium()

	cfg, err := New(WithMedium(m))
	assert.NoError(t, err)

	home, _ := os.UserHomeDir()
	assert.Equal(t, home+"/.core/config.yaml", cfg.Path())
}

func TestLoadEnv_Good(t *testing.T) {
	t.Setenv("CORE_CONFIG_FOO_BAR", "baz")
	t.Setenv("CORE_CONFIG_SIMPLE", "value")

	result := LoadEnv("CORE_CONFIG_")
	assert.Equal(t, "baz", result["foo.bar"])
	assert.Equal(t, "value", result["simple"])
}

func TestLoad_Bad(t *testing.T) {
	m := io.NewMockMedium()

	_, err := Load(m, "/nonexistent/file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoad_InvalidYAML_Bad(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["/tmp/test/config.yaml"] = "invalid: yaml: content: [[[["

	_, err := Load(m, "/tmp/test/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestSave_Good(t *testing.T) {
	m := io.NewMockMedium()

	data := map[string]any{
		"key": "value",
	}

	err := Save(m, "/tmp/test/config.yaml", data)
	assert.NoError(t, err)

	content, readErr := m.Read("/tmp/test/config.yaml")
	assert.NoError(t, readErr)
	assert.Contains(t, content, "key: value")
}

func TestConfig_LoadFile_Env(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["/.env"] = "FOO=bar\nBAZ=qux"

	cfg, err := New(WithMedium(m), WithPath("/config.yaml"))
	assert.NoError(t, err)

	err = cfg.LoadFile(m, "/.env")
	assert.NoError(t, err)

	var foo string
	err = cfg.Get("foo", &foo)
	assert.NoError(t, err)
	assert.Equal(t, "bar", foo)
}

func TestConfig_WithEnvPrefix(t *testing.T) {
	t.Setenv("MYAPP_SETTING", "secret")

	m := io.NewMockMedium()
	cfg, err := New(WithMedium(m), WithEnvPrefix("MYAPP"))
	assert.NoError(t, err)

	var setting string
	err = cfg.Get("setting", &setting)
	assert.NoError(t, err)
	assert.Equal(t, "secret", setting)
}

func TestConfig_Get_EmptyKey(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["/config.yaml"] = "app:\n  name: test\nversion: 1"

	cfg, err := New(WithMedium(m), WithPath("/config.yaml"))
	assert.NoError(t, err)

	type AppConfig struct {
		App struct {
			Name string `mapstructure:"name"`
		} `mapstructure:"app"`
		Version int `mapstructure:"version"`
	}

	var full AppConfig
	err = cfg.Get("", &full)
	assert.NoError(t, err)
	assert.Equal(t, "test", full.App.Name)
	assert.Equal(t, 1, full.Version)
}
