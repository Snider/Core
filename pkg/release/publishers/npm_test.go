package publishers

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNpmPublisher_Name_Good(t *testing.T) {
	t.Run("returns npm", func(t *testing.T) {
		p := NewNpmPublisher()
		assert.Equal(t, "npm", p.Name())
	})
}

func TestNpmPublisher_ParseConfig_Good(t *testing.T) {
	p := NewNpmPublisher()

	t.Run("uses defaults when no extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{Type: "npm"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Package)
		assert.Equal(t, "public", cfg.Access)
	})

	t.Run("parses package and access from extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "npm",
			Extended: map[string]any{
				"package": "@myorg/mypackage",
				"access":  "restricted",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Equal(t, "@myorg/mypackage", cfg.Package)
		assert.Equal(t, "restricted", cfg.Access)
	})

	t.Run("keeps default access when not specified", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "npm",
			Extended: map[string]any{
				"package": "@myorg/mypackage",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Equal(t, "@myorg/mypackage", cfg.Package)
		assert.Equal(t, "public", cfg.Access)
	})

	t.Run("handles nil extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type:     "npm",
			Extended: nil,
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Package)
		assert.Equal(t, "public", cfg.Access)
	})

	t.Run("handles empty strings in config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "npm",
			Extended: map[string]any{
				"package": "",
				"access":  "",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Package)
		assert.Equal(t, "public", cfg.Access)
	})
}

func TestNpmPublisher_RenderTemplate_Good(t *testing.T) {
	p := NewNpmPublisher()

	t.Run("renders package.json template with data", func(t *testing.T) {
		data := npmTemplateData{
			Package:     "@myorg/mycli",
			Version:     "1.2.3",
			Description: "My awesome CLI",
			License:     "MIT",
			Repository:  "owner/myapp",
			BinaryName:  "myapp",
			ProjectName: "myapp",
			Access:      "public",
		}

		result, err := p.renderTemplate("templates/npm/package.json.tmpl", data)
		require.NoError(t, err)

		assert.Contains(t, result, `"name": "@myorg/mycli"`)
		assert.Contains(t, result, `"version": "1.2.3"`)
		assert.Contains(t, result, `"description": "My awesome CLI"`)
		assert.Contains(t, result, `"license": "MIT"`)
		assert.Contains(t, result, "owner/myapp")
		assert.Contains(t, result, `"myapp": "./bin/run.js"`)
		assert.Contains(t, result, `"access": "public"`)
	})

	t.Run("renders restricted access correctly", func(t *testing.T) {
		data := npmTemplateData{
			Package:     "@private/cli",
			Version:     "1.0.0",
			Description: "Private CLI",
			License:     "MIT",
			Repository:  "org/repo",
			BinaryName:  "cli",
			ProjectName: "cli",
			Access:      "restricted",
		}

		result, err := p.renderTemplate("templates/npm/package.json.tmpl", data)
		require.NoError(t, err)

		assert.Contains(t, result, `"access": "restricted"`)
	})
}

func TestNpmPublisher_RenderTemplate_Bad(t *testing.T) {
	p := NewNpmPublisher()

	t.Run("returns error for non-existent template", func(t *testing.T) {
		data := npmTemplateData{}
		_, err := p.renderTemplate("templates/npm/nonexistent.tmpl", data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read template")
	})
}

func TestNpmPublisher_DryRunPublish_Good(t *testing.T) {
	p := NewNpmPublisher()

	t.Run("outputs expected dry run information", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		data := npmTemplateData{
			Package:     "@myorg/mycli",
			Version:     "1.0.0",
			Access:      "public",
			Repository:  "owner/repo",
			BinaryName:  "mycli",
			Description: "My CLI",
		}
		cfg := &NpmConfig{
			Package: "@myorg/mycli",
			Access:  "public",
		}

		err := p.dryRunPublish(data, cfg)

		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "DRY RUN: npm Publish")
		assert.Contains(t, output, "Package:    @myorg/mycli")
		assert.Contains(t, output, "Version:    1.0.0")
		assert.Contains(t, output, "Access:     public")
		assert.Contains(t, output, "Repository: owner/repo")
		assert.Contains(t, output, "Binary:     mycli")
		assert.Contains(t, output, "Generated package.json:")
		assert.Contains(t, output, "Would run: npm publish --access public")
		assert.Contains(t, output, "END DRY RUN")
	})

	t.Run("shows restricted access correctly", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		data := npmTemplateData{
			Package:    "@private/cli",
			Version:    "2.0.0",
			Access:     "restricted",
			Repository: "org/repo",
			BinaryName: "cli",
		}
		cfg := &NpmConfig{
			Package: "@private/cli",
			Access:  "restricted",
		}

		err := p.dryRunPublish(data, cfg)

		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "Access:     restricted")
		assert.Contains(t, output, "Would run: npm publish --access restricted")
	})
}

func TestNpmPublisher_Publish_Bad(t *testing.T) {
	p := NewNpmPublisher()

	t.Run("fails when package name not configured", func(t *testing.T) {
		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/project",
		}
		pubCfg := PublisherConfig{Type: "npm"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err := p.Publish(nil, release, pubCfg, relCfg, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "package name is required")
	})

	t.Run("fails when NPM_TOKEN not set in non-dry-run", func(t *testing.T) {
		// Ensure NPM_TOKEN is not set
		oldToken := os.Getenv("NPM_TOKEN")
		os.Unsetenv("NPM_TOKEN")
		defer func() {
			if oldToken != "" {
				os.Setenv("NPM_TOKEN", oldToken)
			}
		}()

		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/project",
		}
		pubCfg := PublisherConfig{
			Type: "npm",
			Extended: map[string]any{
				"package": "@test/package",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err := p.Publish(nil, release, pubCfg, relCfg, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "NPM_TOKEN environment variable is required")
	})
}

func TestNpmConfig_Defaults_Good(t *testing.T) {
	t.Run("has sensible defaults", func(t *testing.T) {
		p := NewNpmPublisher()
		pubCfg := PublisherConfig{Type: "npm"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		cfg := p.parseConfig(pubCfg, relCfg)

		assert.Empty(t, cfg.Package)
		assert.Equal(t, "public", cfg.Access)
	})
}

func TestNpmTemplateData_Good(t *testing.T) {
	t.Run("struct has all expected fields", func(t *testing.T) {
		data := npmTemplateData{
			Package:     "@myorg/package",
			Version:     "1.0.0",
			Description: "description",
			License:     "MIT",
			Repository:  "org/repo",
			BinaryName:  "cli",
			ProjectName: "cli",
			Access:      "public",
		}

		assert.Equal(t, "@myorg/package", data.Package)
		assert.Equal(t, "1.0.0", data.Version)
		assert.Equal(t, "description", data.Description)
		assert.Equal(t, "MIT", data.License)
		assert.Equal(t, "org/repo", data.Repository)
		assert.Equal(t, "cli", data.BinaryName)
		assert.Equal(t, "cli", data.ProjectName)
		assert.Equal(t, "public", data.Access)
	})
}
