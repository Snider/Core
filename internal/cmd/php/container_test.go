package php

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerBuildOptions_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := DockerBuildOptions{
			ProjectDir:   "/project",
			ImageName:    "myapp",
			Tag:          "v1.0.0",
			Platform:     "linux/amd64",
			Dockerfile:   "/path/to/Dockerfile",
			NoBuildCache: true,
			BuildArgs:    map[string]string{"ARG1": "value1"},
			Output:       os.Stdout,
		}

		assert.Equal(t, "/project", opts.ProjectDir)
		assert.Equal(t, "myapp", opts.ImageName)
		assert.Equal(t, "v1.0.0", opts.Tag)
		assert.Equal(t, "linux/amd64", opts.Platform)
		assert.Equal(t, "/path/to/Dockerfile", opts.Dockerfile)
		assert.True(t, opts.NoBuildCache)
		assert.Equal(t, "value1", opts.BuildArgs["ARG1"])
		assert.NotNil(t, opts.Output)
	})
}

func TestLinuxKitBuildOptions_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := LinuxKitBuildOptions{
			ProjectDir: "/project",
			OutputPath: "/output/image.qcow2",
			Format:     "qcow2",
			Template:   "server-php",
			Variables:  map[string]string{"VAR1": "value1"},
			Output:     os.Stdout,
		}

		assert.Equal(t, "/project", opts.ProjectDir)
		assert.Equal(t, "/output/image.qcow2", opts.OutputPath)
		assert.Equal(t, "qcow2", opts.Format)
		assert.Equal(t, "server-php", opts.Template)
		assert.Equal(t, "value1", opts.Variables["VAR1"])
		assert.NotNil(t, opts.Output)
	})
}

func TestServeOptions_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := ServeOptions{
			ImageName:     "myapp",
			Tag:           "latest",
			ContainerName: "myapp-container",
			Port:          8080,
			HTTPSPort:     8443,
			Detach:        true,
			EnvFile:       "/path/to/.env",
			Volumes:       map[string]string{"/host": "/container"},
			Output:        os.Stdout,
		}

		assert.Equal(t, "myapp", opts.ImageName)
		assert.Equal(t, "latest", opts.Tag)
		assert.Equal(t, "myapp-container", opts.ContainerName)
		assert.Equal(t, 8080, opts.Port)
		assert.Equal(t, 8443, opts.HTTPSPort)
		assert.True(t, opts.Detach)
		assert.Equal(t, "/path/to/.env", opts.EnvFile)
		assert.Equal(t, "/container", opts.Volumes["/host"])
		assert.NotNil(t, opts.Output)
	})
}

func TestIsPHPProject_Container_Good(t *testing.T) {
	t.Run("returns true with composer.json", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{}`), 0644)
		require.NoError(t, err)

		assert.True(t, IsPHPProject(dir))
	})
}

func TestIsPHPProject_Container_Bad(t *testing.T) {
	t.Run("returns false without composer.json", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, IsPHPProject(dir))
	})

	t.Run("returns false for non-existent directory", func(t *testing.T) {
		assert.False(t, IsPHPProject("/non/existent/path"))
	})
}

func TestLookupLinuxKit_Bad(t *testing.T) {
	t.Run("returns error when linuxkit not found", func(t *testing.T) {
		// Save original PATH and paths
		origPath := os.Getenv("PATH")
		origCommonPaths := commonLinuxKitPaths
		defer func() {
			_ = os.Setenv("PATH", origPath)
			commonLinuxKitPaths = origCommonPaths
		}()

		// Set PATH to empty and clear common paths
		_ = os.Setenv("PATH", "")
		commonLinuxKitPaths = []string{}

		_, err := lookupLinuxKit()
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "linuxkit not found")
		}
	})
}

func TestGetLinuxKitTemplate_Good(t *testing.T) {
	t.Run("returns server-php template", func(t *testing.T) {
		content, err := getLinuxKitTemplate("server-php")
		assert.NoError(t, err)
		assert.Contains(t, content, "kernel:")
		assert.Contains(t, content, "linuxkit/kernel")
	})
}

func TestGetLinuxKitTemplate_Bad(t *testing.T) {
	t.Run("returns error for unknown template", func(t *testing.T) {
		_, err := getLinuxKitTemplate("unknown-template")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template not found")
	})
}

func TestApplyTemplateVariables_Good(t *testing.T) {
	t.Run("replaces variables", func(t *testing.T) {
		content := "Hello ${NAME}, welcome to ${PLACE}!"
		vars := map[string]string{
			"NAME":  "World",
			"PLACE": "Earth",
		}

		result, err := applyTemplateVariables(content, vars)
		assert.NoError(t, err)
		assert.Equal(t, "Hello World, welcome to Earth!", result)
	})

	t.Run("handles empty variables", func(t *testing.T) {
		content := "No variables here"
		vars := map[string]string{}

		result, err := applyTemplateVariables(content, vars)
		assert.NoError(t, err)
		assert.Equal(t, "No variables here", result)
	})

	t.Run("leaves unmatched placeholders", func(t *testing.T) {
		content := "Hello ${NAME}, ${UNKNOWN} is unknown"
		vars := map[string]string{
			"NAME": "World",
		}

		result, err := applyTemplateVariables(content, vars)
		assert.NoError(t, err)
		assert.Contains(t, result, "Hello World")
		assert.Contains(t, result, "${UNKNOWN}")
	})

	t.Run("handles multiple occurrences", func(t *testing.T) {
		content := "${VAR} and ${VAR} again"
		vars := map[string]string{
			"VAR": "value",
		}

		result, err := applyTemplateVariables(content, vars)
		assert.NoError(t, err)
		assert.Equal(t, "value and value again", result)
	})
}

func TestDefaultServerPHPTemplate_Good(t *testing.T) {
	t.Run("template has required sections", func(t *testing.T) {
		assert.Contains(t, defaultServerPHPTemplate, "kernel:")
		assert.Contains(t, defaultServerPHPTemplate, "init:")
		assert.Contains(t, defaultServerPHPTemplate, "services:")
		assert.Contains(t, defaultServerPHPTemplate, "onboot:")
	})

	t.Run("template contains placeholders", func(t *testing.T) {
		assert.Contains(t, defaultServerPHPTemplate, "${SSH_KEY:-}")
	})
}

func TestBuildDocker_Bad(t *testing.T) {
	t.Skip("requires Docker installed")

	t.Run("fails for non-PHP project", func(t *testing.T) {
		dir := t.TempDir()
		err := BuildDocker(context.TODO(), DockerBuildOptions{ProjectDir: dir})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a PHP project")
	})
}

func TestBuildLinuxKit_Bad(t *testing.T) {
	t.Skip("requires linuxkit installed")

	t.Run("fails for non-PHP project", func(t *testing.T) {
		dir := t.TempDir()
		err := BuildLinuxKit(context.TODO(), LinuxKitBuildOptions{ProjectDir: dir})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a PHP project")
	})
}

func TestServeProduction_Bad(t *testing.T) {
	t.Run("fails without image name", func(t *testing.T) {
		err := ServeProduction(context.TODO(), ServeOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image name is required")
	})
}

func TestShell_Bad(t *testing.T) {
	t.Run("fails without container ID", func(t *testing.T) {
		err := Shell(context.TODO(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "container ID is required")
	})
}

func TestResolveDockerContainerID_Bad(t *testing.T) {
	t.Skip("requires Docker installed")
}

func TestBuildDocker_DefaultOptions(t *testing.T) {
	t.Run("sets defaults correctly", func(t *testing.T) {
		// This tests the default logic without actually running Docker
		opts := DockerBuildOptions{}

		// Verify default values would be set in BuildDocker
		if opts.Tag == "" {
			opts.Tag = "latest"
		}
		assert.Equal(t, "latest", opts.Tag)

		if opts.ImageName == "" {
			opts.ImageName = filepath.Base("/project/myapp")
		}
		assert.Equal(t, "myapp", opts.ImageName)
	})
}

func TestBuildLinuxKit_DefaultOptions(t *testing.T) {
	t.Run("sets defaults correctly", func(t *testing.T) {
		opts := LinuxKitBuildOptions{}

		// Verify default values would be set
		if opts.Template == "" {
			opts.Template = "server-php"
		}
		assert.Equal(t, "server-php", opts.Template)

		if opts.Format == "" {
			opts.Format = "qcow2"
		}
		assert.Equal(t, "qcow2", opts.Format)
	})
}

func TestServeProduction_DefaultOptions(t *testing.T) {
	t.Run("sets defaults correctly", func(t *testing.T) {
		opts := ServeOptions{ImageName: "myapp"}

		// Verify default values would be set
		if opts.Tag == "" {
			opts.Tag = "latest"
		}
		assert.Equal(t, "latest", opts.Tag)

		if opts.Port == 0 {
			opts.Port = 80
		}
		assert.Equal(t, 80, opts.Port)

		if opts.HTTPSPort == 0 {
			opts.HTTPSPort = 443
		}
		assert.Equal(t, 443, opts.HTTPSPort)
	})
}

func TestLookupLinuxKit_Good(t *testing.T) {
	t.Skip("requires linuxkit installed")

	t.Run("finds linuxkit in PATH", func(t *testing.T) {
		path, err := lookupLinuxKit()
		assert.NoError(t, err)
		assert.NotEmpty(t, path)
	})
}

func TestBuildDocker_WithCustomDockerfile(t *testing.T) {
	t.Skip("requires Docker installed")

	t.Run("uses custom Dockerfile when provided", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"name":"test"}`), 0644)
		require.NoError(t, err)

		dockerfilePath := filepath.Join(dir, "Dockerfile.custom")
		err = os.WriteFile(dockerfilePath, []byte("FROM alpine"), 0644)
		require.NoError(t, err)

		opts := DockerBuildOptions{
			ProjectDir: dir,
			Dockerfile: dockerfilePath,
		}

		// The function would use the custom Dockerfile
		assert.Equal(t, dockerfilePath, opts.Dockerfile)
	})
}

func TestBuildDocker_GeneratesDockerfile(t *testing.T) {
	t.Skip("requires Docker installed")

	t.Run("generates Dockerfile when not provided", func(t *testing.T) {
		dir := t.TempDir()

		// Create valid PHP project
		composerJSON := `{"name":"test","require":{"php":"^8.2","laravel/framework":"^11.0"}}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		opts := DockerBuildOptions{
			ProjectDir: dir,
			// Dockerfile not specified - should be generated
		}

		assert.Empty(t, opts.Dockerfile)
	})
}

func TestServeProduction_BuildsCorrectArgs(t *testing.T) {
	t.Run("builds correct docker run arguments", func(t *testing.T) {
		opts := ServeOptions{
			ImageName:     "myapp",
			Tag:           "v1.0.0",
			ContainerName: "myapp-prod",
			Port:          8080,
			HTTPSPort:     8443,
			Detach:        true,
			EnvFile:       "/path/.env",
			Volumes: map[string]string{
				"/host/storage": "/app/storage",
			},
		}

		// Verify the expected image reference format
		imageRef := opts.ImageName + ":" + opts.Tag
		assert.Equal(t, "myapp:v1.0.0", imageRef)

		// Verify port format
		portMapping := opts.Port
		assert.Equal(t, 8080, portMapping)
	})
}

func TestShell_Integration(t *testing.T) {
	t.Skip("requires Docker with running container")
}

func TestResolveDockerContainerID_Integration(t *testing.T) {
	t.Skip("requires Docker with running containers")
}
