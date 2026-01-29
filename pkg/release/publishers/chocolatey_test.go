package publishers

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChocolateyPublisher_Name_Good(t *testing.T) {
	t.Run("returns chocolatey", func(t *testing.T) {
		p := NewChocolateyPublisher()
		assert.Equal(t, "chocolatey", p.Name())
	})
}

func TestChocolateyPublisher_ParseConfig_Good(t *testing.T) {
	p := NewChocolateyPublisher()

	t.Run("uses defaults when no extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{Type: "chocolatey"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Package)
		assert.False(t, cfg.Push)
		assert.Nil(t, cfg.Official)
	})

	t.Run("parses package and push from extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "chocolatey",
			Extended: map[string]any{
				"package": "mypackage",
				"push":    true,
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Equal(t, "mypackage", cfg.Package)
		assert.True(t, cfg.Push)
	})

	t.Run("parses official config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "chocolatey",
			Extended: map[string]any{
				"official": map[string]any{
					"enabled": true,
					"output":  "dist/choco",
				},
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		require.NotNil(t, cfg.Official)
		assert.True(t, cfg.Official.Enabled)
		assert.Equal(t, "dist/choco", cfg.Official.Output)
	})

	t.Run("handles missing official fields", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "chocolatey",
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

	t.Run("handles nil extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type:     "chocolatey",
			Extended: nil,
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Package)
		assert.False(t, cfg.Push)
		assert.Nil(t, cfg.Official)
	})

	t.Run("defaults push to false when not specified", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "chocolatey",
			Extended: map[string]any{
				"package": "mypackage",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.False(t, cfg.Push)
	})
}

func TestChocolateyPublisher_RenderTemplate_Good(t *testing.T) {
	p := NewChocolateyPublisher()

	t.Run("renders nuspec template with data", func(t *testing.T) {
		data := chocolateyTemplateData{
			PackageName: "myapp",
			Title:       "MyApp CLI",
			Description: "My awesome CLI",
			Repository:  "owner/myapp",
			Version:     "1.2.3",
			License:     "MIT",
			BinaryName:  "myapp",
			Authors:     "owner",
			Tags:        "cli myapp",
			Checksums:   ChecksumMap{},
		}

		result, err := p.renderTemplate("templates/chocolatey/package.nuspec.tmpl", data)
		require.NoError(t, err)

		assert.Contains(t, result, `<id>myapp</id>`)
		assert.Contains(t, result, `<version>1.2.3</version>`)
		assert.Contains(t, result, `<title>MyApp CLI</title>`)
		assert.Contains(t, result, `<authors>owner</authors>`)
		assert.Contains(t, result, `<description>My awesome CLI</description>`)
		assert.Contains(t, result, `<tags>cli myapp</tags>`)
		assert.Contains(t, result, "projectUrl>https://github.com/owner/myapp")
		assert.Contains(t, result, "releaseNotes>https://github.com/owner/myapp/releases/tag/v1.2.3")
	})

	t.Run("renders install script template with data", func(t *testing.T) {
		data := chocolateyTemplateData{
			PackageName: "myapp",
			Repository:  "owner/myapp",
			Version:     "1.2.3",
			BinaryName:  "myapp",
			Checksums: ChecksumMap{
				WindowsAmd64: "abc123def456",
			},
		}

		result, err := p.renderTemplate("templates/chocolatey/tools/chocolateyinstall.ps1.tmpl", data)
		require.NoError(t, err)

		assert.Contains(t, result, "$ErrorActionPreference = 'Stop'")
		assert.Contains(t, result, "https://github.com/owner/myapp/releases/download/v1.2.3/myapp-windows-amd64.zip")
		assert.Contains(t, result, "packageName    = 'myapp'")
		assert.Contains(t, result, "checksum64     = 'abc123def456'")
		assert.Contains(t, result, "checksumType64 = 'sha256'")
		assert.Contains(t, result, "Install-ChocolateyZipPackage")
	})
}

func TestChocolateyPublisher_RenderTemplate_Bad(t *testing.T) {
	p := NewChocolateyPublisher()

	t.Run("returns error for non-existent template", func(t *testing.T) {
		data := chocolateyTemplateData{}
		_, err := p.renderTemplate("templates/chocolatey/nonexistent.tmpl", data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read template")
	})
}

func TestChocolateyPublisher_DryRunPublish_Good(t *testing.T) {
	p := NewChocolateyPublisher()

	t.Run("outputs expected dry run information", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		data := chocolateyTemplateData{
			PackageName: "myapp",
			Version:     "1.0.0",
			Repository:  "owner/repo",
			BinaryName:  "myapp",
			Authors:     "owner",
			Tags:        "cli myapp",
			Checksums:   ChecksumMap{},
		}
		cfg := ChocolateyConfig{
			Push: false,
		}

		err := p.dryRunPublish(data, cfg)

		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "DRY RUN: Chocolatey Publish")
		assert.Contains(t, output, "Package:    myapp")
		assert.Contains(t, output, "Version:    1.0.0")
		assert.Contains(t, output, "Push:       false")
		assert.Contains(t, output, "Repository: owner/repo")
		assert.Contains(t, output, "Generated package.nuspec:")
		assert.Contains(t, output, "Generated chocolateyinstall.ps1:")
		assert.Contains(t, output, "Would generate package files only (push=false)")
		assert.Contains(t, output, "END DRY RUN")
	})

	t.Run("shows push message when push is enabled", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		data := chocolateyTemplateData{
			PackageName: "myapp",
			Version:     "1.0.0",
			BinaryName:  "myapp",
			Authors:     "owner",
			Tags:        "cli",
			Checksums:   ChecksumMap{},
		}
		cfg := ChocolateyConfig{
			Push: true,
		}

		err := p.dryRunPublish(data, cfg)

		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Push:       true")
		assert.Contains(t, output, "Would push to Chocolatey community repo")
	})
}

func TestChocolateyPublisher_ExecutePublish_Bad(t *testing.T) {
	p := NewChocolateyPublisher()

	t.Run("fails when CHOCOLATEY_API_KEY not set for push", func(t *testing.T) {
		// Ensure CHOCOLATEY_API_KEY is not set
		oldKey := os.Getenv("CHOCOLATEY_API_KEY")
		os.Unsetenv("CHOCOLATEY_API_KEY")
		defer func() {
			if oldKey != "" {
				os.Setenv("CHOCOLATEY_API_KEY", oldKey)
			}
		}()

		// Create a temp directory for the test
		tmpDir, err := os.MkdirTemp("", "choco-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		data := chocolateyTemplateData{
			PackageName: "testpkg",
			Version:     "1.0.0",
			BinaryName:  "testpkg",
			Repository:  "owner/repo",
			Authors:     "owner",
			Tags:        "cli",
			Checksums:   ChecksumMap{},
		}

		err = p.pushToChocolatey(nil, tmpDir, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CHOCOLATEY_API_KEY environment variable is required")
	})
}

func TestChocolateyConfig_Defaults_Good(t *testing.T) {
	t.Run("has sensible defaults", func(t *testing.T) {
		p := NewChocolateyPublisher()
		pubCfg := PublisherConfig{Type: "chocolatey"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Package)
		assert.False(t, cfg.Push)
		assert.Nil(t, cfg.Official)
	})
}

func TestChocolateyTemplateData_Good(t *testing.T) {
	t.Run("struct has all expected fields", func(t *testing.T) {
		data := chocolateyTemplateData{
			PackageName: "myapp",
			Title:       "MyApp CLI",
			Description: "description",
			Repository:  "org/repo",
			Version:     "1.0.0",
			License:     "MIT",
			BinaryName:  "myapp",
			Authors:     "org",
			Tags:        "cli tool",
			Checksums: ChecksumMap{
				WindowsAmd64: "hash1",
			},
		}

		assert.Equal(t, "myapp", data.PackageName)
		assert.Equal(t, "MyApp CLI", data.Title)
		assert.Equal(t, "description", data.Description)
		assert.Equal(t, "org/repo", data.Repository)
		assert.Equal(t, "1.0.0", data.Version)
		assert.Equal(t, "MIT", data.License)
		assert.Equal(t, "myapp", data.BinaryName)
		assert.Equal(t, "org", data.Authors)
		assert.Equal(t, "cli tool", data.Tags)
		assert.Equal(t, "hash1", data.Checksums.WindowsAmd64)
	})
}
