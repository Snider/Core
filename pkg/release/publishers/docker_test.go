package publishers

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerPublisher_Name_Good(t *testing.T) {
	t.Run("returns docker", func(t *testing.T) {
		p := NewDockerPublisher()
		assert.Equal(t, "docker", p.Name())
	})
}

func TestDockerPublisher_ParseConfig_Good(t *testing.T) {
	p := NewDockerPublisher()

	t.Run("uses defaults when no extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{Type: "docker"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg, "/project")

		assert.Equal(t, "ghcr.io", cfg.Registry)
		assert.Equal(t, "owner/repo", cfg.Image)
		assert.Equal(t, "/project/Dockerfile", cfg.Dockerfile)
		assert.Equal(t, []string{"linux/amd64", "linux/arm64"}, cfg.Platforms)
		assert.Equal(t, []string{"latest", "{{.Version}}"}, cfg.Tags)
	})

	t.Run("parses extended config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"registry":   "docker.io",
				"image":      "myorg/myimage",
				"dockerfile": "docker/Dockerfile.prod",
				"platforms":  []any{"linux/amd64"},
				"tags":       []any{"latest", "stable", "{{.Version}}"},
				"build_args": map[string]any{
					"GO_VERSION": "1.21",
				},
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg, "/project")

		assert.Equal(t, "docker.io", cfg.Registry)
		assert.Equal(t, "myorg/myimage", cfg.Image)
		assert.Equal(t, "/project/docker/Dockerfile.prod", cfg.Dockerfile)
		assert.Equal(t, []string{"linux/amd64"}, cfg.Platforms)
		assert.Equal(t, []string{"latest", "stable", "{{.Version}}"}, cfg.Tags)
		assert.Equal(t, "1.21", cfg.BuildArgs["GO_VERSION"])
	})

	t.Run("handles absolute dockerfile path", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"dockerfile": "/absolute/path/Dockerfile",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}
		cfg := p.parseConfig(pubCfg, relCfg, "/project")

		assert.Equal(t, "/absolute/path/Dockerfile", cfg.Dockerfile)
	})
}

func TestDockerPublisher_ResolveTags_Good(t *testing.T) {
	p := NewDockerPublisher()

	t.Run("resolves version template", func(t *testing.T) {
		tags := p.resolveTags([]string{"latest", "{{.Version}}", "stable"}, "v1.2.3")

		assert.Equal(t, []string{"latest", "v1.2.3", "stable"}, tags)
	})

	t.Run("handles simple version syntax", func(t *testing.T) {
		tags := p.resolveTags([]string{"{{Version}}"}, "v1.0.0")

		assert.Equal(t, []string{"v1.0.0"}, tags)
	})

	t.Run("handles no templates", func(t *testing.T) {
		tags := p.resolveTags([]string{"latest", "stable"}, "v1.2.3")

		assert.Equal(t, []string{"latest", "stable"}, tags)
	})
}

func TestDockerPublisher_BuildFullTag_Good(t *testing.T) {
	p := NewDockerPublisher()

	tests := []struct {
		name     string
		registry string
		image    string
		tag      string
		expected string
	}{
		{
			name:     "with registry",
			registry: "ghcr.io",
			image:    "owner/repo",
			tag:      "v1.0.0",
			expected: "ghcr.io/owner/repo:v1.0.0",
		},
		{
			name:     "without registry",
			registry: "",
			image:    "myimage",
			tag:      "latest",
			expected: "myimage:latest",
		},
		{
			name:     "docker hub",
			registry: "docker.io",
			image:    "library/nginx",
			tag:      "alpine",
			expected: "docker.io/library/nginx:alpine",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := p.buildFullTag(tc.registry, tc.image, tc.tag)
			assert.Equal(t, tc.expected, tag)
		})
	}
}

func TestDockerPublisher_BuildBuildxArgs_Good(t *testing.T) {
	p := NewDockerPublisher()

	t.Run("builds basic args", func(t *testing.T) {
		cfg := DockerConfig{
			Registry:   "ghcr.io",
			Image:      "owner/repo",
			Dockerfile: "/project/Dockerfile",
			Platforms:  []string{"linux/amd64", "linux/arm64"},
			BuildArgs:  make(map[string]string),
		}
		tags := []string{"latest", "v1.0.0"}

		args := p.buildBuildxArgs(cfg, tags, "v1.0.0")

		assert.Contains(t, args, "buildx")
		assert.Contains(t, args, "build")
		assert.Contains(t, args, "--platform")
		assert.Contains(t, args, "linux/amd64,linux/arm64")
		assert.Contains(t, args, "-t")
		assert.Contains(t, args, "ghcr.io/owner/repo:latest")
		assert.Contains(t, args, "ghcr.io/owner/repo:v1.0.0")
		assert.Contains(t, args, "-f")
		assert.Contains(t, args, "/project/Dockerfile")
		assert.Contains(t, args, "--push")
		assert.Contains(t, args, ".")
	})

	t.Run("includes build args", func(t *testing.T) {
		cfg := DockerConfig{
			Registry:   "ghcr.io",
			Image:      "owner/repo",
			Dockerfile: "/project/Dockerfile",
			Platforms:  []string{"linux/amd64"},
			BuildArgs: map[string]string{
				"GO_VERSION": "1.21",
				"APP_NAME":   "myapp",
			},
		}
		tags := []string{"latest"}

		args := p.buildBuildxArgs(cfg, tags, "v1.0.0")

		assert.Contains(t, args, "--build-arg")
		// Check that build args are present (order may vary)
		foundGoVersion := false
		foundAppName := false
		foundVersion := false
		for i, arg := range args {
			if arg == "--build-arg" && i+1 < len(args) {
				if args[i+1] == "GO_VERSION=1.21" {
					foundGoVersion = true
				}
				if args[i+1] == "APP_NAME=myapp" {
					foundAppName = true
				}
				if args[i+1] == "VERSION=v1.0.0" {
					foundVersion = true
				}
			}
		}
		assert.True(t, foundGoVersion, "GO_VERSION build arg not found")
		assert.True(t, foundAppName, "APP_NAME build arg not found")
		assert.True(t, foundVersion, "VERSION build arg not found")
	})

	t.Run("expands version in build args", func(t *testing.T) {
		cfg := DockerConfig{
			Registry:   "ghcr.io",
			Image:      "owner/repo",
			Dockerfile: "/project/Dockerfile",
			Platforms:  []string{"linux/amd64"},
			BuildArgs: map[string]string{
				"APP_VERSION": "{{.Version}}",
			},
		}
		tags := []string{"latest"}

		args := p.buildBuildxArgs(cfg, tags, "v2.0.0")

		foundExpandedVersion := false
		for i, arg := range args {
			if arg == "--build-arg" && i+1 < len(args) {
				if args[i+1] == "APP_VERSION=v2.0.0" {
					foundExpandedVersion = true
				}
			}
		}
		assert.True(t, foundExpandedVersion, "APP_VERSION should be expanded to v2.0.0")
	})
}

func TestDockerPublisher_Publish_Bad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	p := NewDockerPublisher()

	t.Run("fails when dockerfile not found", func(t *testing.T) {
		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/nonexistent",
		}
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"dockerfile": "/nonexistent/Dockerfile",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err := p.Publish(context.TODO(), release, pubCfg, relCfg, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Dockerfile not found")
	})
}

func TestDockerConfig_Defaults_Good(t *testing.T) {
	t.Run("has sensible defaults", func(t *testing.T) {
		p := NewDockerPublisher()
		pubCfg := PublisherConfig{Type: "docker"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		cfg := p.parseConfig(pubCfg, relCfg, "/project")

		// Verify defaults
		assert.Equal(t, "ghcr.io", cfg.Registry)
		assert.Equal(t, "owner/repo", cfg.Image)
		assert.Len(t, cfg.Platforms, 2)
		assert.Contains(t, cfg.Platforms, "linux/amd64")
		assert.Contains(t, cfg.Platforms, "linux/arm64")
		assert.Contains(t, cfg.Tags, "latest")
	})
}

func TestDockerPublisher_DryRunPublish_Good(t *testing.T) {
	p := NewDockerPublisher()

	t.Run("outputs expected dry run information", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/project",
		}
		cfg := DockerConfig{
			Registry:   "ghcr.io",
			Image:      "owner/repo",
			Dockerfile: "/project/Dockerfile",
			Platforms:  []string{"linux/amd64", "linux/arm64"},
			Tags:       []string{"latest", "{{.Version}}"},
			BuildArgs:  make(map[string]string),
		}

		err := p.dryRunPublish(release, cfg)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "DRY RUN: Docker Build & Push")
		assert.Contains(t, output, "Version:       v1.0.0")
		assert.Contains(t, output, "Registry:      ghcr.io")
		assert.Contains(t, output, "Image:         owner/repo")
		assert.Contains(t, output, "Dockerfile:    /project/Dockerfile")
		assert.Contains(t, output, "Platforms:     linux/amd64, linux/arm64")
		assert.Contains(t, output, "Tags to be applied:")
		assert.Contains(t, output, "ghcr.io/owner/repo:latest")
		assert.Contains(t, output, "ghcr.io/owner/repo:v1.0.0")
		assert.Contains(t, output, "Would execute command:")
		assert.Contains(t, output, "docker buildx build")
		assert.Contains(t, output, "END DRY RUN")
	})

	t.Run("shows build args when present", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/project",
		}
		cfg := DockerConfig{
			Registry:   "docker.io",
			Image:      "myorg/myapp",
			Dockerfile: "/project/Dockerfile",
			Platforms:  []string{"linux/amd64"},
			Tags:       []string{"latest"},
			BuildArgs: map[string]string{
				"GO_VERSION": "1.21",
				"APP_NAME":   "myapp",
			},
		}

		err := p.dryRunPublish(release, cfg)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "Build arguments:")
		assert.Contains(t, output, "GO_VERSION=1.21")
		assert.Contains(t, output, "APP_NAME=myapp")
	})

	t.Run("handles single platform", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v2.0.0",
			ProjectDir: "/project",
		}
		cfg := DockerConfig{
			Registry:   "ghcr.io",
			Image:      "owner/repo",
			Dockerfile: "/project/Dockerfile.prod",
			Platforms:  []string{"linux/amd64"},
			Tags:       []string{"stable"},
			BuildArgs:  make(map[string]string),
		}

		err := p.dryRunPublish(release, cfg)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "Platforms:     linux/amd64")
		assert.Contains(t, output, "ghcr.io/owner/repo:stable")
	})
}

func TestDockerPublisher_ParseConfig_EdgeCases_Good(t *testing.T) {
	p := NewDockerPublisher()

	t.Run("handles nil release config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"image": "custom/image",
			},
		}

		cfg := p.parseConfig(pubCfg, nil, "/project")

		assert.Equal(t, "custom/image", cfg.Image)
		assert.Equal(t, "ghcr.io", cfg.Registry)
	})

	t.Run("handles empty repository in release config", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"image": "fallback/image",
			},
		}
		relCfg := &mockReleaseConfig{repository: ""}

		cfg := p.parseConfig(pubCfg, relCfg, "/project")

		assert.Equal(t, "fallback/image", cfg.Image)
	})

	t.Run("extended config overrides repository image", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"image": "override/image",
			},
		}
		relCfg := &mockReleaseConfig{repository: "original/repo"}

		cfg := p.parseConfig(pubCfg, relCfg, "/project")

		assert.Equal(t, "override/image", cfg.Image)
	})

	t.Run("handles mixed build args types", func(t *testing.T) {
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"build_args": map[string]any{
					"STRING_ARG": "value",
					"INT_ARG":    123, // Non-string value should be skipped
				},
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		cfg := p.parseConfig(pubCfg, relCfg, "/project")

		assert.Equal(t, "value", cfg.BuildArgs["STRING_ARG"])
		_, exists := cfg.BuildArgs["INT_ARG"]
		assert.False(t, exists, "non-string build arg should not be included")
	})
}

func TestDockerPublisher_ResolveTags_EdgeCases_Good(t *testing.T) {
	p := NewDockerPublisher()

	t.Run("handles empty tags", func(t *testing.T) {
		tags := p.resolveTags([]string{}, "v1.0.0")
		assert.Empty(t, tags)
	})

	t.Run("handles multiple version placeholders", func(t *testing.T) {
		tags := p.resolveTags([]string{"{{.Version}}", "prefix-{{.Version}}", "{{.Version}}-suffix"}, "v1.2.3")
		assert.Equal(t, []string{"v1.2.3", "prefix-v1.2.3", "v1.2.3-suffix"}, tags)
	})

	t.Run("handles mixed template formats", func(t *testing.T) {
		tags := p.resolveTags([]string{"{{.Version}}", "{{Version}}", "latest"}, "v3.0.0")
		assert.Equal(t, []string{"v3.0.0", "v3.0.0", "latest"}, tags)
	})
}

func TestDockerPublisher_BuildBuildxArgs_EdgeCases_Good(t *testing.T) {
	p := NewDockerPublisher()

	t.Run("handles empty platforms", func(t *testing.T) {
		cfg := DockerConfig{
			Registry:   "ghcr.io",
			Image:      "owner/repo",
			Dockerfile: "/project/Dockerfile",
			Platforms:  []string{},
			BuildArgs:  make(map[string]string),
		}

		args := p.buildBuildxArgs(cfg, []string{"latest"}, "v1.0.0")

		assert.Contains(t, args, "buildx")
		assert.Contains(t, args, "build")
		// Should not have --platform if empty
		foundPlatform := false
		for i, arg := range args {
			if arg == "--platform" {
				foundPlatform = true
				// Check the next arg exists (it shouldn't be empty)
				if i+1 < len(args) && args[i+1] == "" {
					t.Error("platform argument should not be empty string")
				}
			}
		}
		assert.False(t, foundPlatform, "should not include --platform when platforms is empty")
	})

	t.Run("handles version expansion in build args", func(t *testing.T) {
		cfg := DockerConfig{
			Registry:   "ghcr.io",
			Image:      "owner/repo",
			Dockerfile: "/Dockerfile",
			Platforms:  []string{"linux/amd64"},
			BuildArgs: map[string]string{
				"VERSION":      "{{.Version}}",
				"SIMPLE_VER":   "{{Version}}",
				"STATIC_VALUE": "static",
			},
		}

		args := p.buildBuildxArgs(cfg, []string{"latest"}, "v2.5.0")

		foundVersionArg := false
		foundSimpleArg := false
		foundStaticArg := false
		foundAutoVersion := false

		for i, arg := range args {
			if arg == "--build-arg" && i+1 < len(args) {
				switch args[i+1] {
				case "VERSION=v2.5.0":
					foundVersionArg = true
				case "SIMPLE_VER=v2.5.0":
					foundSimpleArg = true
				case "STATIC_VALUE=static":
					foundStaticArg = true
				}
				// Auto-added VERSION build arg
				if args[i+1] == "VERSION=v2.5.0" {
					foundAutoVersion = true
				}
			}
		}

		// Note: VERSION is both in BuildArgs and auto-added, so we just check it exists
		assert.True(t, foundVersionArg || foundAutoVersion, "VERSION build arg not found")
		assert.True(t, foundSimpleArg, "SIMPLE_VER build arg not expanded")
		assert.True(t, foundStaticArg, "STATIC_VALUE build arg not found")
	})

	t.Run("handles empty registry", func(t *testing.T) {
		cfg := DockerConfig{
			Registry:   "",
			Image:      "localimage",
			Dockerfile: "/Dockerfile",
			Platforms:  []string{"linux/amd64"},
			BuildArgs:  make(map[string]string),
		}

		args := p.buildBuildxArgs(cfg, []string{"latest"}, "v1.0.0")

		assert.Contains(t, args, "-t")
		assert.Contains(t, args, "localimage:latest")
	})
}

func TestDockerPublisher_Publish_DryRun_Good(t *testing.T) {
	// Skip if docker CLI is not available - dry run still validates docker is installed
	if err := validateDockerCli(); err != nil {
		t.Skip("skipping test: docker CLI not available")
	}

	p := NewDockerPublisher()

	t.Run("dry run succeeds with valid Dockerfile", func(t *testing.T) {
		// Create temp directory with Dockerfile
		tmpDir, err := os.MkdirTemp("", "docker-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		err = os.WriteFile(dockerfilePath, []byte("FROM alpine:latest\n"), 0644)
		require.NoError(t, err)

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: tmpDir,
		}
		pubCfg := PublisherConfig{Type: "docker"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err = p.Publish(context.TODO(), release, pubCfg, relCfg, true)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "DRY RUN: Docker Build & Push")
	})

	t.Run("dry run uses custom dockerfile path", func(t *testing.T) {
		// Create temp directory with custom Dockerfile
		tmpDir, err := os.MkdirTemp("", "docker-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		customDir := filepath.Join(tmpDir, "docker")
		err = os.MkdirAll(customDir, 0755)
		require.NoError(t, err)

		dockerfilePath := filepath.Join(customDir, "Dockerfile.prod")
		err = os.WriteFile(dockerfilePath, []byte("FROM alpine:latest\n"), 0644)
		require.NoError(t, err)

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: tmpDir,
		}
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"dockerfile": "docker/Dockerfile.prod",
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err = p.Publish(context.TODO(), release, pubCfg, relCfg, true)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Dockerfile.prod")
	})
}

func TestDockerPublisher_Publish_Validation_Bad(t *testing.T) {
	p := NewDockerPublisher()

	t.Run("fails when Dockerfile not found with docker installed", func(t *testing.T) {
		if err := validateDockerCli(); err != nil {
			t.Skip("skipping test: docker CLI not available")
		}

		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/nonexistent/path",
		}
		pubCfg := PublisherConfig{Type: "docker"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err := p.Publish(context.TODO(), release, pubCfg, relCfg, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Dockerfile not found")
	})

	t.Run("fails when docker CLI not available", func(t *testing.T) {
		if err := validateDockerCli(); err == nil {
			t.Skip("skipping test: docker CLI is available")
		}

		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: "/tmp",
		}
		pubCfg := PublisherConfig{Type: "docker"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err := p.Publish(context.TODO(), release, pubCfg, relCfg, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "docker CLI not found")
	})
}

func TestValidateDockerCli_Good(t *testing.T) {
	t.Run("returns nil when docker is installed", func(t *testing.T) {
		err := validateDockerCli()
		if err != nil {
			// Docker is not installed, which is fine for this test
			assert.Contains(t, err.Error(), "docker CLI not found")
		}
		// If err is nil, docker is installed - that's OK
	})
}

func TestDockerPublisher_Publish_WithCLI_Good(t *testing.T) {
	// These tests run only when docker CLI is available
	if err := validateDockerCli(); err != nil {
		t.Skip("skipping test: docker CLI not available")
	}

	p := NewDockerPublisher()

	t.Run("dry run succeeds with all config options", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "docker-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		err = os.WriteFile(dockerfilePath, []byte("FROM alpine:latest\n"), 0644)
		require.NoError(t, err)

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: tmpDir,
		}
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"registry":   "docker.io",
				"image":      "myorg/myapp",
				"platforms":  []any{"linux/amd64", "linux/arm64"},
				"tags":       []any{"latest", "{{.Version}}", "stable"},
				"build_args": map[string]any{"GO_VERSION": "1.21"},
			},
		}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err = p.Publish(context.TODO(), release, pubCfg, relCfg, true)

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "DRY RUN: Docker Build & Push")
		assert.Contains(t, output, "docker.io")
		assert.Contains(t, output, "myorg/myapp")
	})

	t.Run("dry run with nil relCfg uses extended image", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "docker-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		err = os.WriteFile(dockerfilePath, []byte("FROM alpine:latest\n"), 0644)
		require.NoError(t, err)

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: tmpDir,
		}
		pubCfg := PublisherConfig{
			Type: "docker",
			Extended: map[string]any{
				"image": "standalone/image",
			},
		}

		err = p.Publish(context.TODO(), release, pubCfg, nil, true) // nil relCfg

		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		os.Stdout = oldStdout

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "standalone/image")
	})

	t.Run("fails with non-existent Dockerfile in non-dry-run", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "docker-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Don't create a Dockerfile
		release := &Release{
			Version:    "v1.0.0",
			ProjectDir: tmpDir,
		}
		pubCfg := PublisherConfig{Type: "docker"}
		relCfg := &mockReleaseConfig{repository: "owner/repo"}

		err = p.Publish(context.TODO(), release, pubCfg, relCfg, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Dockerfile not found")
	})
}
