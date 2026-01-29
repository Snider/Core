package publishers

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAURPublisher_Name_Good(t *testing.T) {
	t.Run("returns aur", func(t *testing.T) {
		p := NewAURPublisher()
		assert.Equal(t, "aur", p.Name())
	})
}

func TestAURPublisher_ParseConfig_Good(t *testing.T) {
	p := NewAURPublisher()

	t.Run("uses defaults when no extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{Type: "aur"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Package)
		assert.Empty(t, cfg.Maintainer)
		assert.Nil(t, cfg.Official)
	})

	t.Run("parses package and maintainer from extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "aur",
			Extended: map[string]any{
				"package":    "mypackage",
				"maintainer": "John Doe <john@example.com>",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Equal(t, "mypackage", cfg.Package)
		assert.Equal(t, "John Doe <john@example.com>", cfg.Maintainer)
	})

	t.Run("parses official config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "aur",
			Extended: map[string]any{
				"official": map[string]any{
					"enabled": true,
					"output":  "dist/aur-files",
				},
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		require.NotNil(t, cfg.Official)
		assert.True(t, cfg.Official.Enabled)
		assert.Equal(t, "dist/aur-files", cfg.Official.Output)
	})

	t.Run("handles missing official fields", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "aur",
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

func TestAURPublisher_RenderTemplate_Good(t *testing.T) {
	p := NewAURPublisher()

	t.Run("renders PKGBUILD template with data", func(t *testing.T) {
		data := aurTemplateData{
			PackageName: "myapp",
			Description: "My awesome CLI",
			Repository:  "owner/myapp",
			Version:     "1.2.3",
			License:     "MIT",
			BinaryName:  "myapp",
			Maintainer:  "John Doe <john@example.com>",
			Checksums: ChecksumMap{
				LinuxAmd64: "abc123",
				LinuxArm64: "def456",
			},
		}

		result, err := p.renderTemplate("templates/aur/PKGBUILD.tmpl", data)
		require.NoError(t, err)

		assert.Contains(t, result, "# Maintainer: John Doe <john@example.com>")
		assert.Contains(t, result, "pkgname=myapp-bin")
		assert.Contains(t, result, "pkgver=1.2.3")
		assert.Contains(t, result, `pkgdesc="My awesome CLI"`)
		assert.Contains(t, result, "url=\"https://github.com/owner/myapp\"")
		assert.Contains(t, result, "license=('MIT')")
		assert.Contains(t, result, "sha256sums_x86_64=('abc123')")
		assert.Contains(t, result, "sha256sums_aarch64=('def456')")
	})

	t.Run("renders .SRCINFO template with data", func(t *testing.T) {
		data := aurTemplateData{
			PackageName: "myapp",
			Description: "My CLI",
			Repository:  "owner/myapp",
			Version:     "1.0.0",
			License:     "MIT",
			BinaryName:  "myapp",
			Maintainer:  "Test <test@test.com>",
			Checksums: ChecksumMap{
				LinuxAmd64: "checksum1",
				LinuxArm64: "checksum2",
			},
		}

		result, err := p.renderTemplate("templates/aur/.SRCINFO.tmpl", data)
		require.NoError(t, err)

		assert.Contains(t, result, "pkgbase = myapp-bin")
		assert.Contains(t, result, "pkgdesc = My CLI")
		assert.Contains(t, result, "pkgver = 1.0.0")
		assert.Contains(t, result, "arch = x86_64")
		assert.Contains(t, result, "arch = aarch64")
		assert.Contains(t, result, "sha256sums_x86_64 = checksum1")
		assert.Contains(t, result, "sha256sums_aarch64 = checksum2")
		assert.Contains(t, result, "pkgname = myapp-bin")
	})
}

func TestAURPublisher_RenderTemplate_Bad(t *testing.T) {
	p := NewAURPublisher()

	t.Run("returns error for non-existent template", func(t *testing.T) {
		data := aurTemplateData{}
		_, err := p.renderTemplate("templates/aur/nonexistent.tmpl", data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read template")
	})
}

func TestAURPublisher_DryRunPublish_Good(t *testing.T) {
	p := NewAURPublisher()

	t.Run("outputs expected dry run information", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		data := aurTemplateData{
			PackageName: "myapp",
			Version:     "1.0.0",
			Maintainer:  "John Doe <john@example.com>",
			Repository:  "owner/repo",
			BinaryName:  "myapp",
			Checksums:   ChecksumMap{},
		}
		cfg := AURConfig{
			Maintainer: "John Doe <john@example.com>",
		}

		err := p.dryRunPublish(data, cfg)

		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "DRY RUN: AUR Publish")
		assert.Contains(t, output, "Package:    myapp-bin")
		assert.Contains(t, output, "Version:    1.0.0")
		assert.Contains(t, output, "Maintainer: John Doe <john@example.com>")
		assert.Contains(t, output, "Repository: owner/repo")
		assert.Contains(t, output, "Generated PKGBUILD:")
		assert.Contains(t, output, "Generated .SRCINFO:")
		assert.Contains(t, output, "Would push to AUR: ssh://aur@aur.archlinux.org/myapp-bin.git")
		assert.Contains(t, output, "END DRY RUN")
	})
}

func TestAURPublisher_Publish_Bad(t *testing.T) {
	p := NewAURPublisher()

	t.Run("fails when maintainer not configured", func(t *testing.T) {
		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/project",
		}
		pubCfg := PublisherConfig{Type: "aur"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err := p.Publish(nil, release, pubCfg, relCfg, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maintainer is required")
	})
}

func TestAURConfig_Defaults_Good(t *testing.T) {
	t.Run("has sensible defaults", func(t *testing.T) {
		p := NewAURPublisher()
		pubCfg := PublisherConfig{Type: "aur"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Package)
		assert.Empty(t, cfg.Maintainer)
		assert.Nil(t, cfg.Official)
	})
}
