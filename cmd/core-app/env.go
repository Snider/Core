package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// AppEnvironment holds the resolved paths for the running application.
type AppEnvironment struct {
	// DataDir is the persistent data directory (survives app updates).
	DataDir string
	// LaravelRoot is the extracted Laravel app in the temp directory.
	LaravelRoot string
	// DatabasePath is the full path to the SQLite database file.
	DatabasePath string
}

// PrepareEnvironment creates data directories, generates .env, and symlinks
// storage so Laravel can write to persistent locations.
func PrepareEnvironment(laravelRoot string) (*AppEnvironment, error) {
	dataDir, err := resolveDataDir()
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}

	env := &AppEnvironment{
		DataDir:      dataDir,
		LaravelRoot:  laravelRoot,
		DatabasePath: filepath.Join(dataDir, "core-app.sqlite"),
	}

	// Create persistent directories
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "storage", "app"),
		filepath.Join(dataDir, "storage", "framework", "cache", "data"),
		filepath.Join(dataDir, "storage", "framework", "sessions"),
		filepath.Join(dataDir, "storage", "framework", "views"),
		filepath.Join(dataDir, "storage", "logs"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create dir %s: %w", dir, err)
		}
	}

	// Create empty SQLite database if it doesn't exist
	if _, err := os.Stat(env.DatabasePath); os.IsNotExist(err) {
		if err := os.WriteFile(env.DatabasePath, nil, 0o644); err != nil {
			return nil, fmt.Errorf("create database: %w", err)
		}
		log.Printf("Created new database: %s", env.DatabasePath)
	}

	// Replace the extracted storage/ with a symlink to the persistent one
	extractedStorage := filepath.Join(laravelRoot, "storage")
	os.RemoveAll(extractedStorage)
	persistentStorage := filepath.Join(dataDir, "storage")
	if err := os.Symlink(persistentStorage, extractedStorage); err != nil {
		return nil, fmt.Errorf("symlink storage: %w", err)
	}

	// Generate .env file with resolved paths
	if err := writeEnvFile(laravelRoot, env); err != nil {
		return nil, fmt.Errorf("write .env: %w", err)
	}

	return env, nil
}

// resolveDataDir returns the OS-appropriate persistent data directory.
func resolveDataDir() (string, error) {
	var base string
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Application Support", "core-app")
	case "linux":
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			base = filepath.Join(xdg, "core-app")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, ".local", "share", "core-app")
		}
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".core-app")
	}
	return base, nil
}

// writeEnvFile generates the Laravel .env with resolved runtime paths.
func writeEnvFile(laravelRoot string, env *AppEnvironment) error {
	appKey, err := loadOrGenerateAppKey(env.DataDir)
	if err != nil {
		return fmt.Errorf("app key: %w", err)
	}

	content := fmt.Sprintf(`APP_NAME="Core App"
APP_ENV=production
APP_KEY=%s
APP_DEBUG=false
APP_URL=http://localhost

DB_CONNECTION=sqlite
DB_DATABASE="%s"

CACHE_STORE=file
SESSION_DRIVER=file
LOG_CHANNEL=single
LOG_LEVEL=warning
`, appKey, env.DatabasePath)

	return os.WriteFile(filepath.Join(laravelRoot, ".env"), []byte(content), 0o644)
}

// loadOrGenerateAppKey loads an existing APP_KEY from the data dir,
// or generates a new one and persists it.
func loadOrGenerateAppKey(dataDir string) (string, error) {
	keyFile := filepath.Join(dataDir, ".app-key")

	data, err := os.ReadFile(keyFile)
	if err == nil && len(data) > 0 {
		return string(data), nil
	}

	// Generate a new 32-byte key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}
	appKey := "base64:" + base64.StdEncoding.EncodeToString(key)

	if err := os.WriteFile(keyFile, []byte(appKey), 0o600); err != nil {
		return "", fmt.Errorf("save key: %w", err)
	}

	log.Printf("Generated new APP_KEY (saved to %s)", keyFile)
	return appKey, nil
}

// appendEnv appends a key=value pair to the Laravel .env file.
func appendEnv(laravelRoot, key, value string) error {
	envFile := filepath.Join(laravelRoot, ".env")
	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s=\"%s\"\n", key, value)
	return err
}
