package publishers

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

		err := p.Publish(nil, release, pubCfg, relCfg, false)
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
