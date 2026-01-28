package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupConfigTestDir creates a temp directory with optional .core/build.yaml content.
func setupConfigTestDir(t *testing.T, configContent string) string {
	t.Helper()
	dir := t.TempDir()

	if configContent != "" {
		coreDir := filepath.Join(dir, ConfigDir)
		err := os.MkdirAll(coreDir, 0755)
		require.NoError(t, err)

		configPath := filepath.Join(coreDir, ConfigFileName)
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)
	}

	return dir
}

func TestLoadConfig_Good(t *testing.T) {
	t.Run("loads valid config", func(t *testing.T) {
		content := `
version: 1
project:
  name: myapp
  description: A test application
  main: ./cmd/myapp
  binary: myapp
build:
  cgo: true
  flags:
    - -trimpath
    - -race
  ldflags:
    - -s
    - -w
  env:
    - FOO=bar
targets:
  - os: linux
    arch: amd64
  - os: darwin
    arch: arm64
`
		dir := setupConfigTestDir(t, content)

		cfg, err := LoadConfig(dir)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 1, cfg.Version)
		assert.Equal(t, "myapp", cfg.Project.Name)
		assert.Equal(t, "A test application", cfg.Project.Description)
		assert.Equal(t, "./cmd/myapp", cfg.Project.Main)
		assert.Equal(t, "myapp", cfg.Project.Binary)
		assert.True(t, cfg.Build.CGO)
		assert.Equal(t, []string{"-trimpath", "-race"}, cfg.Build.Flags)
		assert.Equal(t, []string{"-s", "-w"}, cfg.Build.LDFlags)
		assert.Equal(t, []string{"FOO=bar"}, cfg.Build.Env)
		assert.Len(t, cfg.Targets, 2)
		assert.Equal(t, "linux", cfg.Targets[0].OS)
		assert.Equal(t, "amd64", cfg.Targets[0].Arch)
		assert.Equal(t, "darwin", cfg.Targets[1].OS)
		assert.Equal(t, "arm64", cfg.Targets[1].Arch)
	})

	t.Run("returns defaults when config file missing", func(t *testing.T) {
		dir := t.TempDir()

		cfg, err := LoadConfig(dir)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		defaults := DefaultConfig()
		assert.Equal(t, defaults.Version, cfg.Version)
		assert.Equal(t, defaults.Project.Main, cfg.Project.Main)
		assert.Equal(t, defaults.Build.CGO, cfg.Build.CGO)
		assert.Equal(t, defaults.Build.Flags, cfg.Build.Flags)
		assert.Equal(t, defaults.Build.LDFlags, cfg.Build.LDFlags)
		assert.Equal(t, defaults.Targets, cfg.Targets)
	})

	t.Run("applies defaults for missing fields", func(t *testing.T) {
		content := `
version: 2
project:
  name: partial
`
		dir := setupConfigTestDir(t, content)

		cfg, err := LoadConfig(dir)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Explicit values preserved
		assert.Equal(t, 2, cfg.Version)
		assert.Equal(t, "partial", cfg.Project.Name)

		// Defaults applied
		defaults := DefaultConfig()
		assert.Equal(t, defaults.Project.Main, cfg.Project.Main)
		assert.Equal(t, defaults.Build.Flags, cfg.Build.Flags)
		assert.Equal(t, defaults.Build.LDFlags, cfg.Build.LDFlags)
		assert.Equal(t, defaults.Targets, cfg.Targets)
	})

	t.Run("preserves empty arrays when explicitly set", func(t *testing.T) {
		content := `
version: 1
project:
  name: noflags
build:
  flags: []
  ldflags: []
targets:
  - os: linux
    arch: amd64
`
		dir := setupConfigTestDir(t, content)

		cfg, err := LoadConfig(dir)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Empty arrays are preserved (not replaced with defaults)
		assert.Empty(t, cfg.Build.Flags)
		assert.Empty(t, cfg.Build.LDFlags)
		// Targets explicitly set
		assert.Len(t, cfg.Targets, 1)
	})
}

func TestLoadConfig_Bad(t *testing.T) {
	t.Run("returns error for invalid YAML", func(t *testing.T) {
		content := `
version: 1
project:
  name: [invalid yaml
`
		dir := setupConfigTestDir(t, content)

		cfg, err := LoadConfig(dir)
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})

	t.Run("returns error for unreadable file", func(t *testing.T) {
		dir := t.TempDir()
		coreDir := filepath.Join(dir, ConfigDir)
		err := os.MkdirAll(coreDir, 0755)
		require.NoError(t, err)

		// Create config as a directory instead of file
		configPath := filepath.Join(coreDir, ConfigFileName)
		err = os.Mkdir(configPath, 0755)
		require.NoError(t, err)

		cfg, err := LoadConfig(dir)
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "failed to read config file")
	})
}

func TestDefaultConfig_Good(t *testing.T) {
	t.Run("returns sensible defaults", func(t *testing.T) {
		cfg := DefaultConfig()

		assert.Equal(t, 1, cfg.Version)
		assert.Equal(t, ".", cfg.Project.Main)
		assert.Empty(t, cfg.Project.Name)
		assert.Empty(t, cfg.Project.Binary)
		assert.False(t, cfg.Build.CGO)
		assert.Contains(t, cfg.Build.Flags, "-trimpath")
		assert.Contains(t, cfg.Build.LDFlags, "-s")
		assert.Contains(t, cfg.Build.LDFlags, "-w")
		assert.Empty(t, cfg.Build.Env)

		// Default targets cover common platforms
		assert.Len(t, cfg.Targets, 5)
		hasLinuxAmd64 := false
		hasDarwinArm64 := false
		hasWindowsAmd64 := false
		for _, t := range cfg.Targets {
			if t.OS == "linux" && t.Arch == "amd64" {
				hasLinuxAmd64 = true
			}
			if t.OS == "darwin" && t.Arch == "arm64" {
				hasDarwinArm64 = true
			}
			if t.OS == "windows" && t.Arch == "amd64" {
				hasWindowsAmd64 = true
			}
		}
		assert.True(t, hasLinuxAmd64)
		assert.True(t, hasDarwinArm64)
		assert.True(t, hasWindowsAmd64)
	})
}

func TestConfigPath_Good(t *testing.T) {
	t.Run("returns correct path", func(t *testing.T) {
		path := ConfigPath("/project/root")
		assert.Equal(t, "/project/root/.core/build.yaml", path)
	})
}

func TestConfigExists_Good(t *testing.T) {
	t.Run("returns true when config exists", func(t *testing.T) {
		dir := setupConfigTestDir(t, "version: 1")
		assert.True(t, ConfigExists(dir))
	})

	t.Run("returns false when config missing", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, ConfigExists(dir))
	})

	t.Run("returns false when .core dir missing", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, ConfigExists(dir))
	})
}

func TestBuildConfig_ToTargets_Good(t *testing.T) {
	t.Run("converts TargetConfig to Target", func(t *testing.T) {
		cfg := &BuildConfig{
			Targets: []TargetConfig{
				{OS: "linux", Arch: "amd64"},
				{OS: "darwin", Arch: "arm64"},
				{OS: "windows", Arch: "386"},
			},
		}

		targets := cfg.ToTargets()
		require.Len(t, targets, 3)

		assert.Equal(t, Target{OS: "linux", Arch: "amd64"}, targets[0])
		assert.Equal(t, Target{OS: "darwin", Arch: "arm64"}, targets[1])
		assert.Equal(t, Target{OS: "windows", Arch: "386"}, targets[2])
	})

	t.Run("returns empty slice for no targets", func(t *testing.T) {
		cfg := &BuildConfig{
			Targets: []TargetConfig{},
		}

		targets := cfg.ToTargets()
		assert.Empty(t, targets)
	})
}

// TestLoadConfig_Testdata tests loading from the testdata fixture.
func TestLoadConfig_Testdata(t *testing.T) {
	t.Run("loads config-project fixture", func(t *testing.T) {
		cfg, err := LoadConfig("testdata/config-project")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 1, cfg.Version)
		assert.Equal(t, "example-cli", cfg.Project.Name)
		assert.Equal(t, "An example CLI application", cfg.Project.Description)
		assert.Equal(t, "./cmd/example", cfg.Project.Main)
		assert.Equal(t, "example", cfg.Project.Binary)
		assert.False(t, cfg.Build.CGO)
		assert.Equal(t, []string{"-trimpath"}, cfg.Build.Flags)
		assert.Equal(t, []string{"-s", "-w"}, cfg.Build.LDFlags)
		assert.Len(t, cfg.Targets, 3)
	})
}
