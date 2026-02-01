package php

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDockerfile_Good(t *testing.T) {
	t.Run("basic Laravel project", func(t *testing.T) {
		dir := t.TempDir()

		// Create composer.json
		composerJSON := `{
			"name": "test/laravel-project",
			"require": {
				"php": "^8.2",
				"laravel/framework": "^11.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		// Create composer.lock
		err = os.WriteFile(filepath.Join(dir, "composer.lock"), []byte("{}"), 0644)
		require.NoError(t, err)

		content, err := GenerateDockerfile(dir)
		require.NoError(t, err)

		// Check content
		assert.Contains(t, content, "FROM dunglas/frankenphp")
		assert.Contains(t, content, "php8.2")
		assert.Contains(t, content, "COPY composer.json composer.lock")
		assert.Contains(t, content, "composer install")
		assert.Contains(t, content, "EXPOSE 80 443")
	})

	t.Run("Laravel project with Octane", func(t *testing.T) {
		dir := t.TempDir()

		composerJSON := `{
			"name": "test/laravel-octane",
			"require": {
				"php": "^8.3",
				"laravel/framework": "^11.0",
				"laravel/octane": "^2.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "composer.lock"), []byte("{}"), 0644)
		require.NoError(t, err)

		content, err := GenerateDockerfile(dir)
		require.NoError(t, err)

		assert.Contains(t, content, "php8.3")
		assert.Contains(t, content, "octane:start")
	})

	t.Run("project with frontend assets", func(t *testing.T) {
		dir := t.TempDir()

		composerJSON := `{
			"name": "test/laravel-vite",
			"require": {
				"php": "^8.3",
				"laravel/framework": "^11.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "composer.lock"), []byte("{}"), 0644)
		require.NoError(t, err)

		packageJSON := `{
			"name": "test-app",
			"scripts": {
				"dev": "vite",
				"build": "vite build"
			}
		}`
		err = os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		content, err := GenerateDockerfile(dir)
		require.NoError(t, err)

		// Should have multi-stage build
		assert.Contains(t, content, "FROM node:20-alpine AS frontend")
		assert.Contains(t, content, "npm ci")
		assert.Contains(t, content, "npm run build")
		assert.Contains(t, content, "COPY --from=frontend")
	})

	t.Run("project with pnpm", func(t *testing.T) {
		dir := t.TempDir()

		composerJSON := `{
			"name": "test/laravel-pnpm",
			"require": {
				"php": "^8.3",
				"laravel/framework": "^11.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "composer.lock"), []byte("{}"), 0644)
		require.NoError(t, err)

		packageJSON := `{
			"name": "test-app",
			"scripts": {
				"build": "vite build"
			}
		}`
		err = os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err)

		// Create pnpm-lock.yaml
		err = os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte("lockfileVersion: 6.0"), 0644)
		require.NoError(t, err)

		content, err := GenerateDockerfile(dir)
		require.NoError(t, err)

		assert.Contains(t, content, "pnpm install")
		assert.Contains(t, content, "pnpm run build")
	})

	t.Run("project with Redis dependency", func(t *testing.T) {
		dir := t.TempDir()

		composerJSON := `{
			"name": "test/laravel-redis",
			"require": {
				"php": "^8.3",
				"laravel/framework": "^11.0",
				"predis/predis": "^2.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "composer.lock"), []byte("{}"), 0644)
		require.NoError(t, err)

		content, err := GenerateDockerfile(dir)
		require.NoError(t, err)

		assert.Contains(t, content, "install-php-extensions")
		assert.Contains(t, content, "redis")
	})

	t.Run("project with explicit ext- requirements", func(t *testing.T) {
		dir := t.TempDir()

		composerJSON := `{
			"name": "test/with-extensions",
			"require": {
				"php": "^8.3",
				"ext-gd": "*",
				"ext-imagick": "*",
				"ext-intl": "*"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "composer.lock"), []byte("{}"), 0644)
		require.NoError(t, err)

		content, err := GenerateDockerfile(dir)
		require.NoError(t, err)

		assert.Contains(t, content, "install-php-extensions")
		assert.Contains(t, content, "gd")
		assert.Contains(t, content, "imagick")
		assert.Contains(t, content, "intl")
	})
}

func TestGenerateDockerfile_Bad(t *testing.T) {
	t.Run("missing composer.json", func(t *testing.T) {
		dir := t.TempDir()

		_, err := GenerateDockerfile(dir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "composer.json")
	})

	t.Run("invalid composer.json", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte("not json{"), 0644)
		require.NoError(t, err)

		_, err = GenerateDockerfile(dir)
		assert.Error(t, err)
	})
}

func TestDetectDockerfileConfig_Good(t *testing.T) {
	t.Run("full Laravel project", func(t *testing.T) {
		dir := t.TempDir()

		composerJSON := `{
			"name": "test/full-laravel",
			"require": {
				"php": "^8.3",
				"laravel/framework": "^11.0",
				"laravel/octane": "^2.0",
				"predis/predis": "^2.0",
				"intervention/image": "^3.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		packageJSON := `{"scripts": {"build": "vite build"}}`
		err = os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte(""), 0644)
		require.NoError(t, err)

		config, err := DetectDockerfileConfig(dir)
		require.NoError(t, err)

		assert.Equal(t, "8.3", config.PHPVersion)
		assert.True(t, config.IsLaravel)
		assert.True(t, config.HasOctane)
		assert.True(t, config.HasAssets)
		assert.Equal(t, "yarn", config.PackageManager)
		assert.Contains(t, config.PHPExtensions, "redis")
		assert.Contains(t, config.PHPExtensions, "gd")
	})
}

func TestDetectDockerfileConfig_Bad(t *testing.T) {
	t.Run("non-existent directory", func(t *testing.T) {
		_, err := DetectDockerfileConfig("/non/existent/path")
		assert.Error(t, err)
	})
}

func TestExtractPHPVersion_Good(t *testing.T) {
	tests := []struct {
		constraint string
		expected   string
	}{
		{"^8.2", "8.2"},
		{"^8.3", "8.3"},
		{">=8.2", "8.2"},
		{"~8.2", "8.2"},
		{"8.2.*", "8.2"},
		{"8.2.0", "8.2"},
		{"8", "8.0"},
	}

	for _, tt := range tests {
		t.Run(tt.constraint, func(t *testing.T) {
			result := extractPHPVersion(tt.constraint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectPHPExtensions_Good(t *testing.T) {
	t.Run("detects Redis from predis", func(t *testing.T) {
		composer := ComposerJSON{
			Require: map[string]string{
				"predis/predis": "^2.0",
			},
		}

		extensions := detectPHPExtensions(composer)
		assert.Contains(t, extensions, "redis")
	})

	t.Run("detects GD from intervention/image", func(t *testing.T) {
		composer := ComposerJSON{
			Require: map[string]string{
				"intervention/image": "^3.0",
			},
		}

		extensions := detectPHPExtensions(composer)
		assert.Contains(t, extensions, "gd")
	})

	t.Run("detects multiple extensions from Laravel", func(t *testing.T) {
		composer := ComposerJSON{
			Require: map[string]string{
				"laravel/framework": "^11.0",
			},
		}

		extensions := detectPHPExtensions(composer)
		assert.Contains(t, extensions, "pdo_mysql")
		assert.Contains(t, extensions, "bcmath")
	})

	t.Run("detects explicit ext- requirements", func(t *testing.T) {
		composer := ComposerJSON{
			Require: map[string]string{
				"ext-gd":      "*",
				"ext-imagick": "*",
			},
		}

		extensions := detectPHPExtensions(composer)
		assert.Contains(t, extensions, "gd")
		assert.Contains(t, extensions, "imagick")
	})

	t.Run("skips built-in extensions", func(t *testing.T) {
		composer := ComposerJSON{
			Require: map[string]string{
				"ext-json":    "*",
				"ext-session": "*",
				"ext-pdo":     "*",
			},
		}

		extensions := detectPHPExtensions(composer)
		assert.NotContains(t, extensions, "json")
		assert.NotContains(t, extensions, "session")
		assert.NotContains(t, extensions, "pdo")
	})

	t.Run("sorts extensions alphabetically", func(t *testing.T) {
		composer := ComposerJSON{
			Require: map[string]string{
				"ext-zip":  "*",
				"ext-gd":   "*",
				"ext-intl": "*",
			},
		}

		extensions := detectPHPExtensions(composer)

		// Check they are sorted
		for i := 1; i < len(extensions); i++ {
			assert.True(t, extensions[i-1] < extensions[i],
				"extensions should be sorted: %v", extensions)
		}
	})
}

func TestHasNodeAssets_Good(t *testing.T) {
	t.Run("with build script", func(t *testing.T) {
		dir := t.TempDir()

		packageJSON := `{
			"name": "test",
			"scripts": {
				"dev": "vite",
				"build": "vite build"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err)

		assert.True(t, hasNodeAssets(dir))
	})
}

func TestHasNodeAssets_Bad(t *testing.T) {
	t.Run("no package.json", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, hasNodeAssets(dir))
	})

	t.Run("no build script", func(t *testing.T) {
		dir := t.TempDir()

		packageJSON := `{
			"name": "test",
			"scripts": {
				"dev": "vite"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err)

		assert.False(t, hasNodeAssets(dir))
	})

	t.Run("invalid package.json", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("invalid{"), 0644)
		require.NoError(t, err)

		assert.False(t, hasNodeAssets(dir))
	})
}

func TestGenerateDockerignore_Good(t *testing.T) {
	t.Run("generates complete dockerignore", func(t *testing.T) {
		dir := t.TempDir()
		content := GenerateDockerignore(dir)

		// Check key entries
		assert.Contains(t, content, ".git")
		assert.Contains(t, content, "node_modules")
		assert.Contains(t, content, ".env")
		assert.Contains(t, content, "vendor")
		assert.Contains(t, content, "storage/logs/*")
		assert.Contains(t, content, ".idea")
		assert.Contains(t, content, ".vscode")
	})
}

func TestGenerateDockerfileFromConfig_Good(t *testing.T) {
	t.Run("minimal config", func(t *testing.T) {
		config := &DockerfileConfig{
			PHPVersion: "8.3",
			BaseImage:  "dunglas/frankenphp",
			UseAlpine:  true,
		}

		content := GenerateDockerfileFromConfig(config)

		assert.Contains(t, content, "FROM dunglas/frankenphp:latest-php8.3-alpine")
		assert.Contains(t, content, "WORKDIR /app")
		assert.Contains(t, content, "COPY composer.json composer.lock")
		assert.Contains(t, content, "EXPOSE 80 443")
	})

	t.Run("with extensions", func(t *testing.T) {
		config := &DockerfileConfig{
			PHPVersion:    "8.3",
			BaseImage:     "dunglas/frankenphp",
			UseAlpine:     true,
			PHPExtensions: []string{"redis", "gd", "intl"},
		}

		content := GenerateDockerfileFromConfig(config)

		assert.Contains(t, content, "install-php-extensions redis gd intl")
	})

	t.Run("Laravel with Octane", func(t *testing.T) {
		config := &DockerfileConfig{
			PHPVersion: "8.3",
			BaseImage:  "dunglas/frankenphp",
			UseAlpine:  true,
			IsLaravel:  true,
			HasOctane:  true,
		}

		content := GenerateDockerfileFromConfig(config)

		assert.Contains(t, content, "php artisan config:cache")
		assert.Contains(t, content, "php artisan route:cache")
		assert.Contains(t, content, "php artisan view:cache")
		assert.Contains(t, content, "chown -R www-data:www-data storage")
		assert.Contains(t, content, "octane:start")
	})

	t.Run("with frontend assets", func(t *testing.T) {
		config := &DockerfileConfig{
			PHPVersion:     "8.3",
			BaseImage:      "dunglas/frankenphp",
			UseAlpine:      true,
			HasAssets:      true,
			PackageManager: "npm",
		}

		content := GenerateDockerfileFromConfig(config)

		// Multi-stage build
		assert.Contains(t, content, "FROM node:20-alpine AS frontend")
		assert.Contains(t, content, "COPY package.json package-lock.json")
		assert.Contains(t, content, "RUN npm ci")
		assert.Contains(t, content, "RUN npm run build")
		assert.Contains(t, content, "COPY --from=frontend /app/public/build public/build")
	})

	t.Run("with yarn", func(t *testing.T) {
		config := &DockerfileConfig{
			PHPVersion:     "8.3",
			BaseImage:      "dunglas/frankenphp",
			UseAlpine:      true,
			HasAssets:      true,
			PackageManager: "yarn",
		}

		content := GenerateDockerfileFromConfig(config)

		assert.Contains(t, content, "COPY package.json yarn.lock")
		assert.Contains(t, content, "yarn install --frozen-lockfile")
		assert.Contains(t, content, "yarn build")
	})

	t.Run("with bun", func(t *testing.T) {
		config := &DockerfileConfig{
			PHPVersion:     "8.3",
			BaseImage:      "dunglas/frankenphp",
			UseAlpine:      true,
			HasAssets:      true,
			PackageManager: "bun",
		}

		content := GenerateDockerfileFromConfig(config)

		assert.Contains(t, content, "npm install -g bun")
		assert.Contains(t, content, "COPY package.json bun.lockb")
		assert.Contains(t, content, "bun install --frozen-lockfile")
		assert.Contains(t, content, "bun run build")
	})

	t.Run("non-alpine image", func(t *testing.T) {
		config := &DockerfileConfig{
			PHPVersion: "8.3",
			BaseImage:  "dunglas/frankenphp",
			UseAlpine:  false,
		}

		content := GenerateDockerfileFromConfig(config)

		assert.Contains(t, content, "FROM dunglas/frankenphp:latest-php8.3 AS app")
		assert.NotContains(t, content, "alpine")
	})
}

func TestIsPHPProject_Good(t *testing.T) {
	t.Run("project with composer.json", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		assert.True(t, IsPHPProject(dir))
	})
}

func TestIsPHPProject_Bad(t *testing.T) {
	t.Run("project without composer.json", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, IsPHPProject(dir))
	})

	t.Run("non-existent directory", func(t *testing.T) {
		assert.False(t, IsPHPProject("/non/existent/path"))
	})
}

func TestExtractPHPVersion_Edge(t *testing.T) {
	t.Run("handles single major version", func(t *testing.T) {
		result := extractPHPVersion("8")
		assert.Equal(t, "8.0", result)
	})
}

func TestDetectPHPExtensions_RequireDev(t *testing.T) {
	t.Run("detects extensions from require-dev", func(t *testing.T) {
		composer := ComposerJSON{
			RequireDev: map[string]string{
				"predis/predis": "^2.0",
			},
		}

		extensions := detectPHPExtensions(composer)
		assert.Contains(t, extensions, "redis")
	})
}

func TestDockerfileStructure_Good(t *testing.T) {
	t.Run("Dockerfile has proper structure", func(t *testing.T) {
		dir := t.TempDir()

		composerJSON := `{
			"name": "test/app",
			"require": {
				"php": "^8.3",
				"laravel/framework": "^11.0",
				"laravel/octane": "^2.0",
				"predis/predis": "^2.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "composer.lock"), []byte("{}"), 0644)
		require.NoError(t, err)

		packageJSON := `{"scripts": {"build": "vite build"}}`
		err = os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		content, err := GenerateDockerfile(dir)
		require.NoError(t, err)

		lines := strings.Split(content, "\n")
		var fromCount, workdirCount, copyCount, runCount, exposeCount, cmdCount int

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(trimmed, "FROM "):
				fromCount++
			case strings.HasPrefix(trimmed, "WORKDIR "):
				workdirCount++
			case strings.HasPrefix(trimmed, "COPY "):
				copyCount++
			case strings.HasPrefix(trimmed, "RUN "):
				runCount++
			case strings.HasPrefix(trimmed, "EXPOSE "):
				exposeCount++
			case strings.HasPrefix(trimmed, "CMD ["):
				// Only count actual CMD instructions, not HEALTHCHECK CMD
				cmdCount++
			}
		}

		// Multi-stage build should have 2 FROM statements
		assert.Equal(t, 2, fromCount, "should have 2 FROM statements for multi-stage build")

		// Should have proper structure
		assert.GreaterOrEqual(t, workdirCount, 1, "should have WORKDIR")
		assert.GreaterOrEqual(t, copyCount, 3, "should have multiple COPY statements")
		assert.GreaterOrEqual(t, runCount, 2, "should have multiple RUN statements")
		assert.Equal(t, 1, exposeCount, "should have exactly one EXPOSE")
		assert.Equal(t, 1, cmdCount, "should have exactly one CMD")
	})
}
