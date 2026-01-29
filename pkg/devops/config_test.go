package devops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "auto", cfg.Images.Source)
	assert.Equal(t, "host-uk/core-images", cfg.Images.GitHub.Repo)
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	assert.NoError(t, err)
	assert.Contains(t, path, ".core/config.yaml")
}

func TestLoadConfig_Good(t *testing.T) {
	t.Run("returns default if not exists", func(t *testing.T) {
		// Mock HOME to a temp dir
		tempHome := t.TempDir()
		origHome := os.Getenv("HOME")
		t.Setenv("HOME", tempHome)
		defer os.Setenv("HOME", origHome)

		cfg, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, DefaultConfig(), cfg)
	})

	t.Run("loads existing config", func(t *testing.T) {
		tempHome := t.TempDir()
		t.Setenv("HOME", tempHome)

		coreDir := filepath.Join(tempHome, ".core")
		err := os.MkdirAll(coreDir, 0755)
		require.NoError(t, err)

		configData := `
version: 2
images:
  source: cdn
  cdn:
    url: https://cdn.example.com
`
		err = os.WriteFile(filepath.Join(coreDir, "config.yaml"), []byte(configData), 0644)
		require.NoError(t, err)

		cfg, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, 2, cfg.Version)
		assert.Equal(t, "cdn", cfg.Images.Source)
		assert.Equal(t, "https://cdn.example.com", cfg.Images.CDN.URL)
	})
}

func TestLoadConfig_Bad(t *testing.T) {
	t.Run("invalid yaml", func(t *testing.T) {
		tempHome := t.TempDir()
		t.Setenv("HOME", tempHome)

		coreDir := filepath.Join(tempHome, ".core")
		err := os.MkdirAll(coreDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(coreDir, "config.yaml"), []byte("invalid: yaml: :"), 0644)
		require.NoError(t, err)

		_, err = LoadConfig()
		assert.Error(t, err)
	})
}

func TestConfig_Struct(t *testing.T) {
	cfg := &Config{
		Version: 2,
		Images: ImagesConfig{
			Source: "github",
			GitHub: GitHubConfig{
				Repo: "owner/repo",
			},
			Registry: RegistryConfig{
				Image: "ghcr.io/owner/image",
			},
			CDN: CDNConfig{
				URL: "https://cdn.example.com",
			},
		},
	}
	assert.Equal(t, 2, cfg.Version)
	assert.Equal(t, "github", cfg.Images.Source)
	assert.Equal(t, "owner/repo", cfg.Images.GitHub.Repo)
	assert.Equal(t, "ghcr.io/owner/image", cfg.Images.Registry.Image)
	assert.Equal(t, "https://cdn.example.com", cfg.Images.CDN.URL)
}

func TestDefaultConfig_Complete(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "auto", cfg.Images.Source)
	assert.Equal(t, "host-uk/core-images", cfg.Images.GitHub.Repo)
	assert.Equal(t, "ghcr.io/host-uk/core-devops", cfg.Images.Registry.Image)
	assert.Empty(t, cfg.Images.CDN.URL)
}

func TestLoadConfig_Good_PartialConfig(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	coreDir := filepath.Join(tempHome, ".core")
	err := os.MkdirAll(coreDir, 0755)
	require.NoError(t, err)

	// Config only specifies source, should merge with defaults
	configData := `
version: 1
images:
  source: github
`
	err = os.WriteFile(filepath.Join(coreDir, "config.yaml"), []byte(configData), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "github", cfg.Images.Source)
	// Default values should be preserved
	assert.Equal(t, "host-uk/core-images", cfg.Images.GitHub.Repo)
}

func TestLoadConfig_Good_AllSourceTypes(t *testing.T) {
	tests := []struct {
		name   string
		config string
		check  func(*testing.T, *Config)
	}{
		{
			name: "github source",
			config: `
version: 1
images:
  source: github
  github:
    repo: custom/repo
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "github", cfg.Images.Source)
				assert.Equal(t, "custom/repo", cfg.Images.GitHub.Repo)
			},
		},
		{
			name: "cdn source",
			config: `
version: 1
images:
  source: cdn
  cdn:
    url: https://custom-cdn.com
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "cdn", cfg.Images.Source)
				assert.Equal(t, "https://custom-cdn.com", cfg.Images.CDN.URL)
			},
		},
		{
			name: "registry source",
			config: `
version: 1
images:
  source: registry
  registry:
    image: docker.io/custom/image
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "registry", cfg.Images.Source)
				assert.Equal(t, "docker.io/custom/image", cfg.Images.Registry.Image)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempHome := t.TempDir()
			t.Setenv("HOME", tempHome)

			coreDir := filepath.Join(tempHome, ".core")
			err := os.MkdirAll(coreDir, 0755)
			require.NoError(t, err)

			err = os.WriteFile(filepath.Join(coreDir, "config.yaml"), []byte(tt.config), 0644)
			require.NoError(t, err)

			cfg, err := LoadConfig()
			assert.NoError(t, err)
			tt.check(t, cfg)
		})
	}
}

func TestImagesConfig_Struct(t *testing.T) {
	ic := ImagesConfig{
		Source: "auto",
		GitHub: GitHubConfig{Repo: "test/repo"},
	}
	assert.Equal(t, "auto", ic.Source)
	assert.Equal(t, "test/repo", ic.GitHub.Repo)
}

func TestGitHubConfig_Struct(t *testing.T) {
	gc := GitHubConfig{Repo: "owner/repo"}
	assert.Equal(t, "owner/repo", gc.Repo)
}

func TestRegistryConfig_Struct(t *testing.T) {
	rc := RegistryConfig{Image: "ghcr.io/owner/image:latest"}
	assert.Equal(t, "ghcr.io/owner/image:latest", rc.Image)
}

func TestCDNConfig_Struct(t *testing.T) {
	cc := CDNConfig{URL: "https://cdn.example.com/images"}
	assert.Equal(t, "https://cdn.example.com/images", cc.URL)
}

func TestLoadConfig_Bad_UnreadableFile(t *testing.T) {
	// This test is platform-specific and may not work on all systems
	// Skip if we can't test file permissions properly
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	coreDir := filepath.Join(tempHome, ".core")
	err := os.MkdirAll(coreDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(coreDir, "config.yaml")
	err = os.WriteFile(configPath, []byte("version: 1"), 0000)
	require.NoError(t, err)

	_, err = LoadConfig()
	assert.Error(t, err)

	// Restore permissions so cleanup works
	os.Chmod(configPath, 0644)
}