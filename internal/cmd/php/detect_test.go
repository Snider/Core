package php

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsLaravelProject_Good(t *testing.T) {
	t.Run("valid Laravel project with artisan and composer.json", func(t *testing.T) {
		dir := t.TempDir()

		// Create artisan file
		artisanPath := filepath.Join(dir, "artisan")
		err := os.WriteFile(artisanPath, []byte("#!/usr/bin/env php\n"), 0755)
		require.NoError(t, err)

		// Create composer.json with laravel/framework
		composerJSON := `{
			"name": "test/laravel-project",
			"require": {
				"php": "^8.2",
				"laravel/framework": "^11.0"
			}
		}`
		composerPath := filepath.Join(dir, "composer.json")
		err = os.WriteFile(composerPath, []byte(composerJSON), 0644)
		require.NoError(t, err)

		assert.True(t, IsLaravelProject(dir))
	})

	t.Run("Laravel in require-dev", func(t *testing.T) {
		dir := t.TempDir()

		// Create artisan file
		artisanPath := filepath.Join(dir, "artisan")
		err := os.WriteFile(artisanPath, []byte("#!/usr/bin/env php\n"), 0755)
		require.NoError(t, err)

		// Create composer.json with laravel/framework in require-dev
		composerJSON := `{
			"name": "test/laravel-project",
			"require-dev": {
				"laravel/framework": "^11.0"
			}
		}`
		composerPath := filepath.Join(dir, "composer.json")
		err = os.WriteFile(composerPath, []byte(composerJSON), 0644)
		require.NoError(t, err)

		assert.True(t, IsLaravelProject(dir))
	})
}

func TestIsLaravelProject_Bad(t *testing.T) {
	t.Run("missing artisan file", func(t *testing.T) {
		dir := t.TempDir()

		// Create composer.json but no artisan
		composerJSON := `{
			"name": "test/laravel-project",
			"require": {
				"laravel/framework": "^11.0"
			}
		}`
		composerPath := filepath.Join(dir, "composer.json")
		err := os.WriteFile(composerPath, []byte(composerJSON), 0644)
		require.NoError(t, err)

		assert.False(t, IsLaravelProject(dir))
	})

	t.Run("missing composer.json", func(t *testing.T) {
		dir := t.TempDir()

		// Create artisan but no composer.json
		artisanPath := filepath.Join(dir, "artisan")
		err := os.WriteFile(artisanPath, []byte("#!/usr/bin/env php\n"), 0755)
		require.NoError(t, err)

		assert.False(t, IsLaravelProject(dir))
	})

	t.Run("composer.json without Laravel", func(t *testing.T) {
		dir := t.TempDir()

		// Create artisan file
		artisanPath := filepath.Join(dir, "artisan")
		err := os.WriteFile(artisanPath, []byte("#!/usr/bin/env php\n"), 0755)
		require.NoError(t, err)

		// Create composer.json without laravel/framework
		composerJSON := `{
			"name": "test/symfony-project",
			"require": {
				"symfony/framework-bundle": "^7.0"
			}
		}`
		composerPath := filepath.Join(dir, "composer.json")
		err = os.WriteFile(composerPath, []byte(composerJSON), 0644)
		require.NoError(t, err)

		assert.False(t, IsLaravelProject(dir))
	})

	t.Run("invalid composer.json", func(t *testing.T) {
		dir := t.TempDir()

		// Create artisan file
		artisanPath := filepath.Join(dir, "artisan")
		err := os.WriteFile(artisanPath, []byte("#!/usr/bin/env php\n"), 0755)
		require.NoError(t, err)

		// Create invalid composer.json
		composerPath := filepath.Join(dir, "composer.json")
		err = os.WriteFile(composerPath, []byte("not valid json{"), 0644)
		require.NoError(t, err)

		assert.False(t, IsLaravelProject(dir))
	})

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, IsLaravelProject(dir))
	})

	t.Run("non-existent directory", func(t *testing.T) {
		assert.False(t, IsLaravelProject("/non/existent/path"))
	})
}

func TestIsFrankenPHPProject_Good(t *testing.T) {
	t.Run("project with octane and frankenphp config", func(t *testing.T) {
		dir := t.TempDir()

		// Create composer.json with laravel/octane
		composerJSON := `{
			"require": {
				"laravel/octane": "^2.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		// Create config directory and octane.php
		configDir := filepath.Join(dir, "config")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		octaneConfig := `<?php
return [
    'server' => 'frankenphp',
];`
		err = os.WriteFile(filepath.Join(configDir, "octane.php"), []byte(octaneConfig), 0644)
		require.NoError(t, err)

		assert.True(t, IsFrankenPHPProject(dir))
	})

	t.Run("project with octane but no config file", func(t *testing.T) {
		dir := t.TempDir()

		// Create composer.json with laravel/octane
		composerJSON := `{
			"require": {
				"laravel/octane": "^2.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		// No config file - should still return true (assume frankenphp)
		assert.True(t, IsFrankenPHPProject(dir))
	})

	t.Run("project with octane but unreadable config file", func(t *testing.T) {
		dir := t.TempDir()

		// Create composer.json with laravel/octane
		composerJSON := `{
			"require": {
				"laravel/octane": "^2.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		// Create config directory and octane.php with no read permissions
		configDir := filepath.Join(dir, "config")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		octanePath := filepath.Join(configDir, "octane.php")
		err = os.WriteFile(octanePath, []byte("<?php return [];"), 0000)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(octanePath, 0644) }() // Clean up

		// Should return true (assume frankenphp if unreadable)
		assert.True(t, IsFrankenPHPProject(dir))
	})
}

func TestIsFrankenPHPProject_Bad(t *testing.T) {
	t.Run("project without octane", func(t *testing.T) {
		dir := t.TempDir()

		composerJSON := `{
			"require": {
				"laravel/framework": "^11.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		assert.False(t, IsFrankenPHPProject(dir))
	})

	t.Run("missing composer.json", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, IsFrankenPHPProject(dir))
	})
}

func TestDetectServices_Good(t *testing.T) {
	t.Run("full Laravel project with all services", func(t *testing.T) {
		dir := t.TempDir()

		// Setup Laravel project
		err := os.WriteFile(filepath.Join(dir, "artisan"), []byte("#!/usr/bin/env php\n"), 0755)
		require.NoError(t, err)

		composerJSON := `{
			"require": {
				"laravel/framework": "^11.0",
				"laravel/octane": "^2.0"
			}
		}`
		err = os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		// Add vite.config.js
		err = os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte("export default {}"), 0644)
		require.NoError(t, err)

		// Add config directory
		configDir := filepath.Join(dir, "config")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Add horizon.php
		err = os.WriteFile(filepath.Join(configDir, "horizon.php"), []byte("<?php return [];"), 0644)
		require.NoError(t, err)

		// Add reverb.php
		err = os.WriteFile(filepath.Join(configDir, "reverb.php"), []byte("<?php return [];"), 0644)
		require.NoError(t, err)

		// Add .env with Redis
		envContent := `APP_NAME=TestApp
CACHE_DRIVER=redis
REDIS_HOST=127.0.0.1`
		err = os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		services := DetectServices(dir)

		assert.Contains(t, services, ServiceFrankenPHP)
		assert.Contains(t, services, ServiceVite)
		assert.Contains(t, services, ServiceHorizon)
		assert.Contains(t, services, ServiceReverb)
		assert.Contains(t, services, ServiceRedis)
	})

	t.Run("minimal Laravel project", func(t *testing.T) {
		dir := t.TempDir()

		// Setup minimal Laravel project
		err := os.WriteFile(filepath.Join(dir, "artisan"), []byte("#!/usr/bin/env php\n"), 0755)
		require.NoError(t, err)

		composerJSON := `{
			"require": {
				"laravel/framework": "^11.0"
			}
		}`
		err = os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		services := DetectServices(dir)

		assert.Contains(t, services, ServiceFrankenPHP)
		assert.NotContains(t, services, ServiceVite)
		assert.NotContains(t, services, ServiceHorizon)
		assert.NotContains(t, services, ServiceReverb)
		assert.NotContains(t, services, ServiceRedis)
	})
}

func TestHasHorizon_Good(t *testing.T) {
	t.Run("horizon config exists", func(t *testing.T) {
		dir := t.TempDir()
		configDir := filepath.Join(dir, "config")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(configDir, "horizon.php"), []byte("<?php return [];"), 0644)
		require.NoError(t, err)

		assert.True(t, hasHorizon(dir))
	})
}

func TestHasHorizon_Bad(t *testing.T) {
	t.Run("horizon config missing", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, hasHorizon(dir))
	})
}

func TestHasReverb_Good(t *testing.T) {
	t.Run("reverb config exists", func(t *testing.T) {
		dir := t.TempDir()
		configDir := filepath.Join(dir, "config")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(configDir, "reverb.php"), []byte("<?php return [];"), 0644)
		require.NoError(t, err)

		assert.True(t, hasReverb(dir))
	})
}

func TestHasReverb_Bad(t *testing.T) {
	t.Run("reverb config missing", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, hasReverb(dir))
	})
}

func TestDetectServices_Bad(t *testing.T) {
	t.Run("non-Laravel project", func(t *testing.T) {
		dir := t.TempDir()

		services := DetectServices(dir)
		assert.Empty(t, services)
	})
}

func TestDetectPackageManager_Good(t *testing.T) {
	tests := []struct {
		name     string
		lockFile string
		expected string
	}{
		{"bun detected", "bun.lockb", "bun"},
		{"pnpm detected", "pnpm-lock.yaml", "pnpm"},
		{"yarn detected", "yarn.lock", "yarn"},
		{"npm detected", "package-lock.json", "npm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			err := os.WriteFile(filepath.Join(dir, tt.lockFile), []byte(""), 0644)
			require.NoError(t, err)

			result := DetectPackageManager(dir)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("no lock file defaults to npm", func(t *testing.T) {
		dir := t.TempDir()

		result := DetectPackageManager(dir)
		assert.Equal(t, "npm", result)
	})

	t.Run("bun takes priority over npm", func(t *testing.T) {
		dir := t.TempDir()

		// Create both lock files
		err := os.WriteFile(filepath.Join(dir, "bun.lockb"), []byte(""), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(""), 0644)
		require.NoError(t, err)

		result := DetectPackageManager(dir)
		assert.Equal(t, "bun", result)
	})
}

func TestGetLaravelAppName_Good(t *testing.T) {
	t.Run("simple app name", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME=MyApp
APP_ENV=local`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.Equal(t, "MyApp", GetLaravelAppName(dir))
	})

	t.Run("quoted app name", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME="My Awesome App"
APP_ENV=local`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.Equal(t, "My Awesome App", GetLaravelAppName(dir))
	})

	t.Run("single quoted app name", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME='My App'
APP_ENV=local`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.Equal(t, "My App", GetLaravelAppName(dir))
	})
}

func TestGetLaravelAppName_Bad(t *testing.T) {
	t.Run("no .env file", func(t *testing.T) {
		dir := t.TempDir()
		assert.Equal(t, "", GetLaravelAppName(dir))
	})

	t.Run("no APP_NAME in .env", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_ENV=local
APP_DEBUG=true`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.Equal(t, "", GetLaravelAppName(dir))
	})
}

func TestGetLaravelAppURL_Good(t *testing.T) {
	t.Run("standard URL", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME=MyApp
APP_URL=https://myapp.test`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.Equal(t, "https://myapp.test", GetLaravelAppURL(dir))
	})

	t.Run("quoted URL", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_URL="http://localhost:8000"`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.Equal(t, "http://localhost:8000", GetLaravelAppURL(dir))
	})
}

func TestExtractDomainFromURL_Good(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com", "example.com"},
		{"http://example.com", "example.com"},
		{"https://example.com:8080", "example.com"},
		{"https://example.com/path/to/page", "example.com"},
		{"https://example.com:443/path", "example.com"},
		{"localhost", "localhost"},
		{"localhost:8000", "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := ExtractDomainFromURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNeedsRedis_Good(t *testing.T) {
	t.Run("CACHE_DRIVER=redis", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME=Test
CACHE_DRIVER=redis`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.True(t, needsRedis(dir))
	})

	t.Run("QUEUE_CONNECTION=redis", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME=Test
QUEUE_CONNECTION=redis`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.True(t, needsRedis(dir))
	})

	t.Run("REDIS_HOST localhost", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME=Test
REDIS_HOST=localhost`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.True(t, needsRedis(dir))
	})

	t.Run("REDIS_HOST 127.0.0.1", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME=Test
REDIS_HOST=127.0.0.1`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.True(t, needsRedis(dir))
	})

	t.Run("SESSION_DRIVER=redis", func(t *testing.T) {
		dir := t.TempDir()
		envContent := "SESSION_DRIVER=redis"
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)
		assert.True(t, needsRedis(dir))
	})

	t.Run("BROADCAST_DRIVER=redis", func(t *testing.T) {
		dir := t.TempDir()
		envContent := "BROADCAST_DRIVER=redis"
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)
		assert.True(t, needsRedis(dir))
	})

	t.Run("REDIS_HOST remote (should be false for local dev env)", func(t *testing.T) {
		dir := t.TempDir()
		envContent := "REDIS_HOST=redis.example.com"
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)
		assert.False(t, needsRedis(dir))
	})
}

func TestNeedsRedis_Bad(t *testing.T) {
	t.Run("no .env file", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, needsRedis(dir))
	})

	t.Run("no redis configuration", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME=Test
CACHE_DRIVER=file
QUEUE_CONNECTION=sync`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.False(t, needsRedis(dir))
	})

	t.Run("commented redis config", func(t *testing.T) {
		dir := t.TempDir()

		envContent := `APP_NAME=Test
# CACHE_DRIVER=redis`
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		assert.False(t, needsRedis(dir))
	})
}

func TestHasVite_Good(t *testing.T) {
	viteFiles := []string{
		"vite.config.js",
		"vite.config.ts",
		"vite.config.mjs",
		"vite.config.mts",
	}

	for _, file := range viteFiles {
		t.Run(file, func(t *testing.T) {
			dir := t.TempDir()

			err := os.WriteFile(filepath.Join(dir, file), []byte("export default {}"), 0644)
			require.NoError(t, err)

			assert.True(t, hasVite(dir))
		})
	}
}

func TestHasVite_Bad(t *testing.T) {
	t.Run("no vite config", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, hasVite(dir))
	})

	t.Run("wrong file name", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "vite.config.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		assert.False(t, hasVite(dir))
	})
}

func TestIsFrankenPHPProject_ConfigWithoutFrankenPHP(t *testing.T) {
	t.Run("octane config without frankenphp", func(t *testing.T) {
		dir := t.TempDir()

		// Create composer.json with laravel/octane
		composerJSON := `{
			"require": {
				"laravel/octane": "^2.0"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		// Create config directory and octane.php without frankenphp
		configDir := filepath.Join(dir, "config")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		octaneConfig := `<?php
return [
    'server' => 'swoole',
];`
		err = os.WriteFile(filepath.Join(configDir, "octane.php"), []byte(octaneConfig), 0644)
		require.NoError(t, err)

		assert.False(t, IsFrankenPHPProject(dir))
	})
}
