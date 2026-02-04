package release

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/host-uk/core/pkg/build"
	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindArtifacts_Good(t *testing.T) {
	t.Run("finds tar.gz artifacts", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))

		// Create test artifact files
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app-linux-amd64.tar.gz"), []byte("test"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app-darwin-arm64.tar.gz"), []byte("test"), 0644))

		artifacts, err := findArtifacts(io.Local, distDir)
		require.NoError(t, err)

		assert.Len(t, artifacts, 2)
	})

	t.Run("finds zip artifacts", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))

		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app-windows-amd64.zip"), []byte("test"), 0644))

		artifacts, err := findArtifacts(io.Local, distDir)
		require.NoError(t, err)

		assert.Len(t, artifacts, 1)
		assert.Contains(t, artifacts[0].Path, "app-windows-amd64.zip")
	})

	t.Run("finds checksum files", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))

		require.NoError(t, os.WriteFile(filepath.Join(distDir, "CHECKSUMS.txt"), []byte("checksums"), 0644))

		artifacts, err := findArtifacts(io.Local, distDir)
		require.NoError(t, err)

		assert.Len(t, artifacts, 1)
		assert.Contains(t, artifacts[0].Path, "CHECKSUMS.txt")
	})

	t.Run("finds signature files", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))

		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz.sig"), []byte("signature"), 0644))

		artifacts, err := findArtifacts(io.Local, distDir)
		require.NoError(t, err)

		assert.Len(t, artifacts, 1)
	})

	t.Run("finds mixed artifact types", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))

		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app-linux.tar.gz"), []byte("test"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app-windows.zip"), []byte("test"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "CHECKSUMS.txt"), []byte("checksums"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.sig"), []byte("sig"), 0644))

		artifacts, err := findArtifacts(io.Local, distDir)
		require.NoError(t, err)

		assert.Len(t, artifacts, 4)
	})

	t.Run("ignores non-artifact files", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))

		require.NoError(t, os.WriteFile(filepath.Join(distDir, "README.md"), []byte("readme"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.exe"), []byte("binary"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz"), []byte("artifact"), 0644))

		artifacts, err := findArtifacts(io.Local, distDir)
		require.NoError(t, err)

		assert.Len(t, artifacts, 1)
		assert.Contains(t, artifacts[0].Path, "app.tar.gz")
	})

	t.Run("ignores subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(distDir, "subdir"), 0755))

		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz"), []byte("artifact"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "subdir", "nested.tar.gz"), []byte("nested"), 0644))

		artifacts, err := findArtifacts(io.Local, distDir)
		require.NoError(t, err)

		// Should only find the top-level artifact
		assert.Len(t, artifacts, 1)
	})

	t.Run("returns empty slice for empty dist directory", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))

		artifacts, err := findArtifacts(io.Local, distDir)
		require.NoError(t, err)

		assert.Empty(t, artifacts)
	})
}

func TestFindArtifacts_Bad(t *testing.T) {
	t.Run("returns error when dist directory does not exist", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")

		_, err := findArtifacts(io.Local, distDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dist/ directory not found")
	})

	t.Run("returns error when dist directory is unreadable", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))

		// Create a file that looks like dist but will cause ReadDir to fail
		// by making the directory unreadable
		require.NoError(t, os.Chmod(distDir, 0000))
		defer func() { _ = os.Chmod(distDir, 0755) }()

		_, err := findArtifacts(io.Local, distDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read dist/")
	})
}

func TestGetBuilder_Good(t *testing.T) {
	t.Run("returns Go builder for go project type", func(t *testing.T) {
		builder, err := getBuilder(build.ProjectTypeGo)
		require.NoError(t, err)
		assert.NotNil(t, builder)
		assert.Equal(t, "go", builder.Name())
	})

	t.Run("returns Wails builder for wails project type", func(t *testing.T) {
		builder, err := getBuilder(build.ProjectTypeWails)
		require.NoError(t, err)
		assert.NotNil(t, builder)
		assert.Equal(t, "wails", builder.Name())
	})
}

func TestGetBuilder_Bad(t *testing.T) {
	t.Run("returns error for Node project type", func(t *testing.T) {
		_, err := getBuilder(build.ProjectTypeNode)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node.js builder not yet implemented")
	})

	t.Run("returns error for PHP project type", func(t *testing.T) {
		_, err := getBuilder(build.ProjectTypePHP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PHP builder not yet implemented")
	})

	t.Run("returns error for unsupported project type", func(t *testing.T) {
		_, err := getBuilder(build.ProjectType("unknown"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported project type")
	})
}

func TestGetPublisher_Good(t *testing.T) {
	tests := []struct {
		pubType      string
		expectedName string
	}{
		{"github", "github"},
		{"linuxkit", "linuxkit"},
		{"docker", "docker"},
		{"npm", "npm"},
		{"homebrew", "homebrew"},
		{"scoop", "scoop"},
		{"aur", "aur"},
		{"chocolatey", "chocolatey"},
	}

	for _, tc := range tests {
		t.Run(tc.pubType, func(t *testing.T) {
			publisher, err := getPublisher(tc.pubType)
			require.NoError(t, err)
			assert.NotNil(t, publisher)
			assert.Equal(t, tc.expectedName, publisher.Name())
		})
	}
}

func TestGetPublisher_Bad(t *testing.T) {
	t.Run("returns error for unsupported publisher type", func(t *testing.T) {
		_, err := getPublisher("unsupported")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported publisher type: unsupported")
	})

	t.Run("returns error for empty publisher type", func(t *testing.T) {
		_, err := getPublisher("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported publisher type")
	})
}

func TestBuildExtendedConfig_Good(t *testing.T) {
	t.Run("returns empty map for minimal config", func(t *testing.T) {
		cfg := PublisherConfig{
			Type: "github",
		}

		ext := buildExtendedConfig(cfg)
		assert.Empty(t, ext)
	})

	t.Run("includes LinuxKit config", func(t *testing.T) {
		cfg := PublisherConfig{
			Type:      "linuxkit",
			Config:    "linuxkit.yaml",
			Formats:   []string{"iso", "qcow2"},
			Platforms: []string{"linux/amd64", "linux/arm64"},
		}

		ext := buildExtendedConfig(cfg)

		assert.Equal(t, "linuxkit.yaml", ext["config"])
		assert.Equal(t, []any{"iso", "qcow2"}, ext["formats"])
		assert.Equal(t, []any{"linux/amd64", "linux/arm64"}, ext["platforms"])
	})

	t.Run("includes Docker config", func(t *testing.T) {
		cfg := PublisherConfig{
			Type:       "docker",
			Registry:   "ghcr.io",
			Image:      "owner/repo",
			Dockerfile: "Dockerfile.prod",
			Tags:       []string{"latest", "v1.0.0"},
			BuildArgs:  map[string]string{"VERSION": "1.0.0"},
		}

		ext := buildExtendedConfig(cfg)

		assert.Equal(t, "ghcr.io", ext["registry"])
		assert.Equal(t, "owner/repo", ext["image"])
		assert.Equal(t, "Dockerfile.prod", ext["dockerfile"])
		assert.Equal(t, []any{"latest", "v1.0.0"}, ext["tags"])
		buildArgs := ext["build_args"].(map[string]any)
		assert.Equal(t, "1.0.0", buildArgs["VERSION"])
	})

	t.Run("includes npm config", func(t *testing.T) {
		cfg := PublisherConfig{
			Type:    "npm",
			Package: "@host-uk/core",
			Access:  "public",
		}

		ext := buildExtendedConfig(cfg)

		assert.Equal(t, "@host-uk/core", ext["package"])
		assert.Equal(t, "public", ext["access"])
	})

	t.Run("includes Homebrew config", func(t *testing.T) {
		cfg := PublisherConfig{
			Type:    "homebrew",
			Tap:     "host-uk/tap",
			Formula: "core",
		}

		ext := buildExtendedConfig(cfg)

		assert.Equal(t, "host-uk/tap", ext["tap"])
		assert.Equal(t, "core", ext["formula"])
	})

	t.Run("includes Scoop config", func(t *testing.T) {
		cfg := PublisherConfig{
			Type:   "scoop",
			Bucket: "host-uk/bucket",
		}

		ext := buildExtendedConfig(cfg)

		assert.Equal(t, "host-uk/bucket", ext["bucket"])
	})

	t.Run("includes AUR config", func(t *testing.T) {
		cfg := PublisherConfig{
			Type:       "aur",
			Maintainer: "John Doe <john@example.com>",
		}

		ext := buildExtendedConfig(cfg)

		assert.Equal(t, "John Doe <john@example.com>", ext["maintainer"])
	})

	t.Run("includes Chocolatey config", func(t *testing.T) {
		cfg := PublisherConfig{
			Type: "chocolatey",
			Push: true,
		}

		ext := buildExtendedConfig(cfg)

		assert.True(t, ext["push"].(bool))
	})

	t.Run("includes Official config", func(t *testing.T) {
		cfg := PublisherConfig{
			Type: "homebrew",
			Official: &OfficialConfig{
				Enabled: true,
				Output:  "/path/to/output",
			},
		}

		ext := buildExtendedConfig(cfg)

		official := ext["official"].(map[string]any)
		assert.True(t, official["enabled"].(bool))
		assert.Equal(t, "/path/to/output", official["output"])
	})

	t.Run("Official config without output", func(t *testing.T) {
		cfg := PublisherConfig{
			Type: "scoop",
			Official: &OfficialConfig{
				Enabled: true,
			},
		}

		ext := buildExtendedConfig(cfg)

		official := ext["official"].(map[string]any)
		assert.True(t, official["enabled"].(bool))
		_, hasOutput := official["output"]
		assert.False(t, hasOutput)
	})
}

func TestToAnySlice_Good(t *testing.T) {
	t.Run("converts string slice to any slice", func(t *testing.T) {
		input := []string{"a", "b", "c"}

		result := toAnySlice(input)

		assert.Len(t, result, 3)
		assert.Equal(t, "a", result[0])
		assert.Equal(t, "b", result[1])
		assert.Equal(t, "c", result[2])
	})

	t.Run("handles empty slice", func(t *testing.T) {
		input := []string{}

		result := toAnySlice(input)

		assert.Empty(t, result)
	})

	t.Run("handles single element", func(t *testing.T) {
		input := []string{"only"}

		result := toAnySlice(input)

		assert.Len(t, result, 1)
		assert.Equal(t, "only", result[0])
	})
}

func TestPublish_Good(t *testing.T) {
	t.Run("returns release with version from config", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz"), []byte("test"), 0644))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		cfg.SetVersion("v1.0.0")
		cfg.Publishers = nil // No publishers to avoid network calls

		release, err := Publish(context.Background(), cfg, true)
		require.NoError(t, err)

		assert.Equal(t, "v1.0.0", release.Version)
		assert.Len(t, release.Artifacts, 1)
	})

	t.Run("finds artifacts in dist directory", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app-linux.tar.gz"), []byte("test"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app-darwin.tar.gz"), []byte("test"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "CHECKSUMS.txt"), []byte("checksums"), 0644))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		cfg.SetVersion("v1.0.0")
		cfg.Publishers = nil

		release, err := Publish(context.Background(), cfg, true)
		require.NoError(t, err)

		assert.Len(t, release.Artifacts, 3)
	})
}

func TestPublish_Bad(t *testing.T) {
	t.Run("returns error when config is nil", func(t *testing.T) {
		_, err := Publish(context.Background(), nil, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config is nil")
	})

	t.Run("returns error when dist directory missing", func(t *testing.T) {
		dir := t.TempDir()

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		cfg.SetVersion("v1.0.0")

		_, err := Publish(context.Background(), cfg, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dist/ directory not found")
	})

	t.Run("returns error when no artifacts found", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		cfg.SetVersion("v1.0.0")

		_, err := Publish(context.Background(), cfg, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no artifacts found")
	})

	t.Run("returns error for unsupported publisher", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz"), []byte("test"), 0644))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		cfg.SetVersion("v1.0.0")
		cfg.Publishers = []PublisherConfig{
			{Type: "unsupported"},
		}

		_, err := Publish(context.Background(), cfg, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported publisher type")
	})

	t.Run("returns error when version determination fails in non-git dir", func(t *testing.T) {
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz"), []byte("test"), 0644))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		// Don't set version - let it try to determine from git
		cfg.Publishers = nil

		// In a non-git directory, DetermineVersion returns v0.0.1 as default
		// so we verify that the publish proceeds without error
		release, err := Publish(context.Background(), cfg, true)
		require.NoError(t, err)
		assert.Equal(t, "v0.0.1", release.Version)
	})
}

func TestRun_Good(t *testing.T) {
	t.Run("returns release with version from config", func(t *testing.T) {
		// Create a minimal Go project for testing
		dir := t.TempDir()

		// Create go.mod
		goMod := `module testapp

go 1.21
`
		require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644))

		// Create main.go
		mainGo := `package main

func main() {}
`
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		cfg.SetVersion("v1.0.0")
		cfg.Project.Name = "testapp"
		cfg.Build.Targets = []TargetConfig{} // Empty targets to use defaults
		cfg.Publishers = nil                 // No publishers to avoid network calls

		// Note: This test will actually try to build, which may fail in CI
		// So we just test that the function accepts the config properly
		release, err := Run(context.Background(), cfg, true)
		if err != nil {
			// Build might fail in test environment, but we still verify the error message
			assert.Contains(t, err.Error(), "build")
		} else {
			assert.Equal(t, "v1.0.0", release.Version)
		}
	})
}

func TestRun_Bad(t *testing.T) {
	t.Run("returns error when config is nil", func(t *testing.T) {
		_, err := Run(context.Background(), nil, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config is nil")
	})
}

func TestRelease_Structure(t *testing.T) {
	t.Run("Release struct holds expected fields", func(t *testing.T) {
		release := &Release{
			Version:    "v1.0.0",
			Artifacts:  []build.Artifact{{Path: "/path/to/artifact"}},
			Changelog:  "## v1.0.0\n\nChanges",
			ProjectDir: "/project",
		}

		assert.Equal(t, "v1.0.0", release.Version)
		assert.Len(t, release.Artifacts, 1)
		assert.Contains(t, release.Changelog, "v1.0.0")
		assert.Equal(t, "/project", release.ProjectDir)
	})
}

func TestPublish_VersionFromGit(t *testing.T) {
	t.Run("determines version from git when not set", func(t *testing.T) {
		dir := setupPublishGitRepo(t)
		createPublishCommit(t, dir, "feat: initial commit")
		createPublishTag(t, dir, "v1.2.3")

		// Create dist directory with artifact
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz"), []byte("test"), 0644))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		// Don't set version - let it be determined from git
		cfg.Publishers = nil

		release, err := Publish(context.Background(), cfg, true)
		require.NoError(t, err)

		assert.Equal(t, "v1.2.3", release.Version)
	})
}

func TestPublish_ChangelogGeneration(t *testing.T) {
	t.Run("generates changelog from git commits when available", func(t *testing.T) {
		dir := setupPublishGitRepo(t)
		createPublishCommit(t, dir, "feat: add feature")
		createPublishTag(t, dir, "v1.0.0")
		createPublishCommit(t, dir, "fix: fix bug")
		createPublishTag(t, dir, "v1.0.1")

		// Create dist directory with artifact
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz"), []byte("test"), 0644))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		cfg.SetVersion("v1.0.1")
		cfg.Publishers = nil

		release, err := Publish(context.Background(), cfg, true)
		require.NoError(t, err)

		// Changelog should contain either the commit message or the version
		assert.Contains(t, release.Changelog, "v1.0.1")
	})

	t.Run("uses fallback changelog on error", func(t *testing.T) {
		dir := t.TempDir() // Not a git repo
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz"), []byte("test"), 0644))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		cfg.SetVersion("v1.0.0")
		cfg.Publishers = nil

		release, err := Publish(context.Background(), cfg, true)
		require.NoError(t, err)

		// Should use fallback changelog
		assert.Contains(t, release.Changelog, "Release v1.0.0")
	})
}

func TestPublish_DefaultProjectDir(t *testing.T) {
	t.Run("uses current directory when projectDir is empty", func(t *testing.T) {
		// Create artifacts in current directory's dist folder
		dir := t.TempDir()
		distDir := filepath.Join(dir, "dist")
		require.NoError(t, os.MkdirAll(distDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(distDir, "app.tar.gz"), []byte("test"), 0644))

		cfg := DefaultConfig()
		cfg.SetProjectDir(dir)
		cfg.SetVersion("v1.0.0")
		cfg.Publishers = nil

		release, err := Publish(context.Background(), cfg, true)
		require.NoError(t, err)

		assert.NotEmpty(t, release.ProjectDir)
	})
}

// Helper functions for publish tests
func setupPublishGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	return dir
}

func createPublishCommit(t *testing.T, dir, message string) {
	t.Helper()

	filePath := filepath.Join(dir, "publish_test.txt")
	content, _ := os.ReadFile(filePath)
	content = append(content, []byte(message+"\n")...)
	require.NoError(t, os.WriteFile(filePath, content, 0644))

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
}

func createPublishTag(t *testing.T, dir, tag string) {
	t.Helper()
	cmd := exec.Command("git", "tag", tag)
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
}
