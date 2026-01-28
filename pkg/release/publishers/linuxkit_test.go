package publishers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinuxKitPublisher_Name_Good(t *testing.T) {
	t.Run("returns linuxkit", func(t *testing.T) {
		p := NewLinuxKitPublisher()
		assert.Equal(t, "linuxkit", p.Name())
	})
}

func TestLinuxKitPublisher_ParseConfig_Good(t *testing.T) {
	p := NewLinuxKitPublisher()

	t.Run("uses defaults when no extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{Type: "linuxkit"}
		cfg := p.parseConfig(pubCfg, "/project")

		assert.Equal(t, "/project/.core/linuxkit/server.yml", cfg.Config)
		assert.Equal(t, []string{"iso"}, cfg.Formats)
		assert.Equal(t, []string{"linux/amd64"}, cfg.Platforms)
	})

	t.Run("parses extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "linuxkit",
			Extended: map[string]any{
				"config":    ".core/linuxkit/custom.yml",
				"formats":   []any{"iso", "qcow2", "vmdk"},
				"platforms": []any{"linux/amd64", "linux/arm64"},
			},
		}
		cfg := p.parseConfig(pubCfg, "/project")

		assert.Equal(t, "/project/.core/linuxkit/custom.yml", cfg.Config)
		assert.Equal(t, []string{"iso", "qcow2", "vmdk"}, cfg.Formats)
		assert.Equal(t, []string{"linux/amd64", "linux/arm64"}, cfg.Platforms)
	})

	t.Run("handles absolute config path", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "linuxkit",
			Extended: map[string]any{
				"config": "/absolute/path/to/config.yml",
			},
		}
		cfg := p.parseConfig(pubCfg, "/project")

		assert.Equal(t, "/absolute/path/to/config.yml", cfg.Config)
	})
}

func TestLinuxKitPublisher_BuildLinuxKitArgs_Good(t *testing.T) {
	p := NewLinuxKitPublisher()

	t.Run("builds basic args for amd64", func(t *testing.T) {
		args := p.buildLinuxKitArgs("/config/server.yml", "iso", "linuxkit-1.0.0-amd64", "/output", "amd64")

		assert.Contains(t, args, "build")
		assert.Contains(t, args, "-format")
		assert.Contains(t, args, "iso")
		assert.Contains(t, args, "-name")
		assert.Contains(t, args, "linuxkit-1.0.0-amd64")
		assert.Contains(t, args, "-dir")
		assert.Contains(t, args, "/output")
		assert.Contains(t, args, "/config/server.yml")
		// Should not contain -arch for amd64 (default)
		assert.NotContains(t, args, "-arch")
	})

	t.Run("builds args with arch for arm64", func(t *testing.T) {
		args := p.buildLinuxKitArgs("/config/server.yml", "qcow2", "linuxkit-1.0.0-arm64", "/output", "arm64")

		assert.Contains(t, args, "-arch")
		assert.Contains(t, args, "arm64")
		assert.Contains(t, args, "qcow2")
	})
}

func TestLinuxKitPublisher_BuildBaseName_Good(t *testing.T) {
	p := NewLinuxKitPublisher()

	t.Run("strips v prefix", func(t *testing.T) {
		name := p.buildBaseName("v1.2.3")
		assert.Equal(t, "linuxkit-1.2.3", name)
	})

	t.Run("handles version without v prefix", func(t *testing.T) {
		name := p.buildBaseName("1.2.3")
		assert.Equal(t, "linuxkit-1.2.3", name)
	})
}

func TestLinuxKitPublisher_GetArtifactPath_Good(t *testing.T) {
	p := NewLinuxKitPublisher()

	tests := []struct {
		name       string
		outputDir  string
		outputName string
		format     string
		expected   string
	}{
		{
			name:       "ISO format",
			outputDir:  "/dist/linuxkit",
			outputName: "linuxkit-1.0.0-amd64",
			format:     "iso",
			expected:   "/dist/linuxkit/linuxkit-1.0.0-amd64.iso",
		},
		{
			name:       "raw format",
			outputDir:  "/dist/linuxkit",
			outputName: "linuxkit-1.0.0-amd64",
			format:     "raw",
			expected:   "/dist/linuxkit/linuxkit-1.0.0-amd64.raw",
		},
		{
			name:       "qcow2 format",
			outputDir:  "/dist/linuxkit",
			outputName: "linuxkit-1.0.0-arm64",
			format:     "qcow2",
			expected:   "/dist/linuxkit/linuxkit-1.0.0-arm64.qcow2",
		},
		{
			name:       "vmdk format",
			outputDir:  "/dist/linuxkit",
			outputName: "linuxkit-1.0.0-amd64",
			format:     "vmdk",
			expected:   "/dist/linuxkit/linuxkit-1.0.0-amd64.vmdk",
		},
		{
			name:       "gcp format",
			outputDir:  "/dist/linuxkit",
			outputName: "linuxkit-1.0.0-amd64",
			format:     "gcp",
			expected:   "/dist/linuxkit/linuxkit-1.0.0-amd64.img.tar.gz",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := p.getArtifactPath(tc.outputDir, tc.outputName, tc.format)
			assert.Equal(t, tc.expected, path)
		})
	}
}

func TestLinuxKitPublisher_GetFormatExtension_Good(t *testing.T) {
	p := NewLinuxKitPublisher()

	tests := []struct {
		format   string
		expected string
	}{
		{"iso", ".iso"},
		{"raw", ".raw"},
		{"qcow2", ".qcow2"},
		{"vmdk", ".vmdk"},
		{"vhd", ".vhd"},
		{"gcp", ".img.tar.gz"},
		{"aws", ".raw"},
		{"unknown", ".unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			ext := p.getFormatExtension(tc.format)
			assert.Equal(t, tc.expected, ext)
		})
	}
}

func TestLinuxKitPublisher_Publish_Bad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	p := NewLinuxKitPublisher()

	t.Run("fails when config file not found", func(t *testing.T) {
		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/nonexistent",
		}
		pubCfg := PublisherConfig{
			Type: "linuxkit",
			Extended: map[string]any{
				"config": "/nonexistent/config.yml",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		// Note: This test requires linuxkit to NOT be installed
		// or uses a non-existent config path
		err := p.Publish(nil, release, pubCfg, relCfg, false)
		assert.Error(t, err)
	})
}

// mockReleaseConfig implements ReleaseConfig for testing.
type mockReleaseConfig struct {
	repository  string
	projectName string
}

func (m *mockReleaseConfig) GetRepository() string {
	return m.repository
}

func (m *mockReleaseConfig) GetProjectName() string {
	return m.projectName
}
