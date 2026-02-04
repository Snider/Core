package publishers

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/host-uk/core/pkg/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomebrewPublisher_Name_Good(t *testing.T) {
	t.Run("returns homebrew", func(t *testing.T) {
		p := NewHomebrewPublisher()
		assert.Equal(t, "homebrew", p.Name())
	})
}

func TestHomebrewPublisher_ParseConfig_Good(t *testing.T) {
	p := NewHomebrewPublisher()

	t.Run("uses defaults when no extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{Type: "homebrew"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Tap)
		assert.Empty(t, cfg.Formula)
		assert.Nil(t, cfg.Official)
	})

	t.Run("parses tap and formula from extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "homebrew",
			Extended: map[string]any{
				"tap":     "host-uk/homebrew-tap",
				"formula": "myformula",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Equal(t, "host-uk/homebrew-tap", cfg.Tap)
		assert.Equal(t, "myformula", cfg.Formula)
	})

	t.Run("parses official config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "homebrew",
			Extended: map[string]any{
				"official": map[string]any{
					"enabled": true,
					"output":  "dist/brew",
				},
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		require.NotNil(t, cfg.Official)
		assert.True(t, cfg.Official.Enabled)
		assert.Equal(t, "dist/brew", cfg.Official.Output)
	})

	t.Run("handles missing official fields", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "homebrew",
			Extended: map[string]any{
				"official": map[string]any{},
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		require.NotNil(t, cfg.Official)
		assert.False(t, cfg.Official.Enabled)
		assert.Empty(t, cfg.Official.Output)
	})
}

func TestHomebrewPublisher_ToFormulaClass_Good(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "core",
			expected: "Core",
		},
		{
			name:     "kebab case",
			input:    "my-cli-tool",
			expected: "MyCliTool",
		},
		{
			name:     "already capitalised",
			input:    "CLI",
			expected: "CLI",
		},
		{
			name:     "single letter",
			input:    "x",
			expected: "X",
		},
		{
			name:     "multiple dashes",
			input:    "my-super-cool-app",
			expected: "MySuperCoolApp",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := toFormulaClass(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHomebrewPublisher_BuildChecksumMap_Good(t *testing.T) {
	t.Run("maps artifacts to checksums by platform", func(t *testing.T) {
		artifacts := []build.Artifact{
			{Path: "/dist/myapp-darwin-amd64.tar.gz", OS: "darwin", Arch: "amd64", Checksum: "abc123"},
			{Path: "/dist/myapp-darwin-arm64.tar.gz", OS: "darwin", Arch: "arm64", Checksum: "def456"},
			{Path: "/dist/myapp-linux-amd64.tar.gz", OS: "linux", Arch: "amd64", Checksum: "ghi789"},
			{Path: "/dist/myapp-linux-arm64.tar.gz", OS: "linux", Arch: "arm64", Checksum: "jkl012"},
			{Path: "/dist/myapp-windows-amd64.zip", OS: "windows", Arch: "amd64", Checksum: "mno345"},
			{Path: "/dist/myapp-windows-arm64.zip", OS: "windows", Arch: "arm64", Checksum: "pqr678"},
		}

		checksums := buildChecksumMap(artifacts)

		assert.Equal(t, "abc123", checksums.DarwinAmd64)
		assert.Equal(t, "def456", checksums.DarwinArm64)
		assert.Equal(t, "ghi789", checksums.LinuxAmd64)
		assert.Equal(t, "jkl012", checksums.LinuxArm64)
		assert.Equal(t, "mno345", checksums.WindowsAmd64)
		assert.Equal(t, "pqr678", checksums.WindowsArm64)
	})

	t.Run("handles empty artifacts", func(t *testing.T) {
		checksums := buildChecksumMap([]build.Artifact{})

		assert.Empty(t, checksums.DarwinAmd64)
		assert.Empty(t, checksums.DarwinArm64)
		assert.Empty(t, checksums.LinuxAmd64)
		assert.Empty(t, checksums.LinuxArm64)
	})

	t.Run("handles partial platform coverage", func(t *testing.T) {
		artifacts := []build.Artifact{
			{Path: "/dist/myapp-darwin-arm64.tar.gz", Checksum: "def456"},
			{Path: "/dist/myapp-linux-amd64.tar.gz", Checksum: "ghi789"},
		}

		checksums := buildChecksumMap(artifacts)

		assert.Empty(t, checksums.DarwinAmd64)
		assert.Equal(t, "def456", checksums.DarwinArm64)
		assert.Equal(t, "ghi789", checksums.LinuxAmd64)
		assert.Empty(t, checksums.LinuxArm64)
	})
}

func TestHomebrewPublisher_RenderTemplate_Good(t *testing.T) {
	p := NewHomebrewPublisher()

	t.Run("renders formula template with data", func(t *testing.T) {
		data := homebrewTemplateData{
			FormulaClass: "MyApp",
			Description:  "My awesome CLI",
			Repository:   "owner/myapp",
			Version:      "1.2.3",
			License:      "MIT",
			BinaryName:   "myapp",
			Checksums: ChecksumMap{
				DarwinAmd64: "abc123",
				DarwinArm64: "def456",
				LinuxAmd64:  "ghi789",
				LinuxArm64:  "jkl012",
			},
		}

		result, err := p.renderTemplate("templates/homebrew/formula.rb.tmpl", data)
		require.NoError(t, err)

		assert.Contains(t, result, "class MyApp < Formula")
		assert.Contains(t, result, `desc "My awesome CLI"`)
		assert.Contains(t, result, `version "1.2.3"`)
		assert.Contains(t, result, `license "MIT"`)
		assert.Contains(t, result, "owner/myapp")
		assert.Contains(t, result, "abc123")
		assert.Contains(t, result, "def456")
		assert.Contains(t, result, "ghi789")
		assert.Contains(t, result, "jkl012")
		assert.Contains(t, result, `bin.install "myapp"`)
	})
}

func TestHomebrewPublisher_RenderTemplate_Bad(t *testing.T) {
	p := NewHomebrewPublisher()

	t.Run("returns error for non-existent template", func(t *testing.T) {
		data := homebrewTemplateData{}
		_, err := p.renderTemplate("templates/homebrew/nonexistent.tmpl", data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read template")
	})
}

func TestHomebrewPublisher_DryRunPublish_Good(t *testing.T) {
	p := NewHomebrewPublisher()

	t.Run("outputs expected dry run information", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		data := homebrewTemplateData{
			FormulaClass: "MyApp",
			Description:  "My CLI",
			Repository:   "owner/repo",
			Version:      "1.0.0",
			License:      "MIT",
			BinaryName:   "myapp",
			Checksums:    ChecksumMap{},
		}
		cfg := HomebrewConfig{
			Tap: "owner/homebrew-tap",
		}

		err := p.dryRunPublish(data, cfg)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "DRY RUN: Homebrew Publish")
		assert.Contains(t, output, "Formula:    MyApp")
		assert.Contains(t, output, "Version:    1.0.0")
		assert.Contains(t, output, "Tap:        owner/homebrew-tap")
		assert.Contains(t, output, "Repository: owner/repo")
		assert.Contains(t, output, "Would commit to tap: owner/homebrew-tap")
		assert.Contains(t, output, "END DRY RUN")
	})

	t.Run("shows official output path when enabled", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		data := homebrewTemplateData{
			FormulaClass: "MyApp",
			Version:      "1.0.0",
			BinaryName:   "myapp",
			Checksums:    ChecksumMap{},
		}
		cfg := HomebrewConfig{
			Official: &OfficialConfig{
				Enabled: true,
				Output:  "custom/path",
			},
		}

		err := p.dryRunPublish(data, cfg)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Would write files for official PR to: custom/path")
	})

	t.Run("uses default official output path when not specified", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		data := homebrewTemplateData{
			FormulaClass: "MyApp",
			Version:      "1.0.0",
			BinaryName:   "myapp",
			Checksums:    ChecksumMap{},
		}
		cfg := HomebrewConfig{
			Official: &OfficialConfig{
				Enabled: true,
			},
		}

		err := p.dryRunPublish(data, cfg)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Would write files for official PR to: dist/homebrew")
	})
}

func TestHomebrewPublisher_Publish_Bad(t *testing.T) {
	p := NewHomebrewPublisher()

	t.Run("fails when tap not configured and not official mode", func(t *testing.T) {
		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/project",
		}
		pubCfg := PublisherConfig{Type: "homebrew"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err := p.Publish(context.TODO(), release, pubCfg, relCfg, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tap is required")
	})
}

func TestHomebrewConfig_Defaults_Good(t *testing.T) {
	t.Run("has sensible defaults", func(t *testing.T) {
		p := NewHomebrewPublisher()
		pubCfg := PublisherConfig{Type: "homebrew"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Tap)
		assert.Empty(t, cfg.Formula)
		assert.Nil(t, cfg.Official)
	})
}
