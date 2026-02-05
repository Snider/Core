package php

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

// DetectedService represents a service that was detected in a Laravel project.
type DetectedService string

// Detected service constants for Laravel projects.
const (
	// ServiceFrankenPHP indicates FrankenPHP server is detected.
	ServiceFrankenPHP DetectedService = "frankenphp"
	// ServiceVite indicates Vite frontend bundler is detected.
	ServiceVite DetectedService = "vite"
	// ServiceHorizon indicates Laravel Horizon queue dashboard is detected.
	ServiceHorizon DetectedService = "horizon"
	// ServiceReverb indicates Laravel Reverb WebSocket server is detected.
	ServiceReverb DetectedService = "reverb"
	// ServiceRedis indicates Redis cache/queue backend is detected.
	ServiceRedis DetectedService = "redis"
)

// IsLaravelProject checks if the given directory is a Laravel project.
// It looks for the presence of artisan file and laravel in composer.json.
func IsLaravelProject(dir string) bool {
	m := getMedium()

	// Check for artisan file
	artisanPath := filepath.Join(dir, "artisan")
	if !m.Exists(artisanPath) {
		return false
	}

	// Check composer.json for laravel/framework
	composerPath := filepath.Join(dir, "composer.json")
	data, err := m.Read(composerPath)
	if err != nil {
		return false
	}

	var composer struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}

	if err := json.Unmarshal([]byte(data), &composer); err != nil {
		return false
	}

	// Check for laravel/framework in require
	if _, ok := composer.Require["laravel/framework"]; ok {
		return true
	}

	// Also check require-dev (less common but possible)
	if _, ok := composer.RequireDev["laravel/framework"]; ok {
		return true
	}

	return false
}

// IsFrankenPHPProject checks if the project is configured for FrankenPHP.
// It looks for laravel/octane with frankenphp driver.
func IsFrankenPHPProject(dir string) bool {
	m := getMedium()

	// Check composer.json for laravel/octane
	composerPath := filepath.Join(dir, "composer.json")
	data, err := m.Read(composerPath)
	if err != nil {
		return false
	}

	var composer struct {
		Require map[string]string `json:"require"`
	}

	if err := json.Unmarshal([]byte(data), &composer); err != nil {
		return false
	}

	if _, ok := composer.Require["laravel/octane"]; !ok {
		return false
	}

	// Check octane config for frankenphp
	configPath := filepath.Join(dir, "config", "octane.php")
	if !m.Exists(configPath) {
		// If no config exists but octane is installed, assume frankenphp
		return true
	}

	configData, err := m.Read(configPath)
	if err != nil {
		return true // Assume frankenphp if we can't read config
	}

	// Look for frankenphp in the config
	return strings.Contains(configData, "frankenphp")
}

// DetectServices detects which services are needed based on project files.
func DetectServices(dir string) []DetectedService {
	services := []DetectedService{}

	// FrankenPHP/Octane is always needed for a Laravel dev environment
	if IsFrankenPHPProject(dir) || IsLaravelProject(dir) {
		services = append(services, ServiceFrankenPHP)
	}

	// Check for Vite
	if hasVite(dir) {
		services = append(services, ServiceVite)
	}

	// Check for Horizon
	if hasHorizon(dir) {
		services = append(services, ServiceHorizon)
	}

	// Check for Reverb
	if hasReverb(dir) {
		services = append(services, ServiceReverb)
	}

	// Check for Redis
	if needsRedis(dir) {
		services = append(services, ServiceRedis)
	}

	return services
}

// hasVite checks if the project uses Vite.
func hasVite(dir string) bool {
	m := getMedium()
	viteConfigs := []string{
		"vite.config.js",
		"vite.config.ts",
		"vite.config.mjs",
		"vite.config.mts",
	}

	for _, config := range viteConfigs {
		if m.Exists(filepath.Join(dir, config)) {
			return true
		}
	}

	return false
}

// hasHorizon checks if Laravel Horizon is configured.
func hasHorizon(dir string) bool {
	horizonConfig := filepath.Join(dir, "config", "horizon.php")
	return getMedium().Exists(horizonConfig)
}

// hasReverb checks if Laravel Reverb is configured.
func hasReverb(dir string) bool {
	reverbConfig := filepath.Join(dir, "config", "reverb.php")
	return getMedium().Exists(reverbConfig)
}

// needsRedis checks if the project uses Redis based on .env configuration.
func needsRedis(dir string) bool {
	m := getMedium()
	envPath := filepath.Join(dir, ".env")
	content, err := m.Read(envPath)
	if err != nil {
		return false
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Check for Redis-related environment variables
		redisIndicators := []string{
			"REDIS_HOST=",
			"CACHE_DRIVER=redis",
			"QUEUE_CONNECTION=redis",
			"SESSION_DRIVER=redis",
			"BROADCAST_DRIVER=redis",
		}

		for _, indicator := range redisIndicators {
			if strings.HasPrefix(line, indicator) {
				// Check if it's set to localhost or 127.0.0.1
				if strings.Contains(line, "127.0.0.1") || strings.Contains(line, "localhost") ||
					indicator != "REDIS_HOST=" {
					return true
				}
			}
		}
	}

	return false
}

// DetectPackageManager detects which package manager is used in the project.
// Returns "npm", "pnpm", "yarn", or "bun".
func DetectPackageManager(dir string) string {
	m := getMedium()
	// Check for lock files in order of preference
	lockFiles := []struct {
		file    string
		manager string
	}{
		{"bun.lockb", "bun"},
		{"pnpm-lock.yaml", "pnpm"},
		{"yarn.lock", "yarn"},
		{"package-lock.json", "npm"},
	}

	for _, lf := range lockFiles {
		if m.Exists(filepath.Join(dir, lf.file)) {
			return lf.manager
		}
	}

	// Default to npm if no lock file found
	return "npm"
}

// GetLaravelAppName extracts the application name from Laravel's .env file.
func GetLaravelAppName(dir string) string {
	m := getMedium()
	envPath := filepath.Join(dir, ".env")
	content, err := m.Read(envPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "APP_NAME=") {
			value := strings.TrimPrefix(line, "APP_NAME=")
			// Remove quotes if present
			value = strings.Trim(value, `"'`)
			return value
		}
	}

	return ""
}

// GetLaravelAppURL extracts the application URL from Laravel's .env file.
func GetLaravelAppURL(dir string) string {
	m := getMedium()
	envPath := filepath.Join(dir, ".env")
	content, err := m.Read(envPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "APP_URL=") {
			value := strings.TrimPrefix(line, "APP_URL=")
			// Remove quotes if present
			value = strings.Trim(value, `"'`)
			return value
		}
	}

	return ""
}

// ExtractDomainFromURL extracts the domain from a URL string.
func ExtractDomainFromURL(url string) string {
	// Remove protocol
	domain := strings.TrimPrefix(url, "https://")
	domain = strings.TrimPrefix(domain, "http://")

	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove path if present
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}
