package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:laravel
var laravelFiles embed.FS

// extractLaravel copies the embedded Laravel app to a temporary directory.
// FrankenPHP needs real filesystem paths — it cannot serve from embed.FS.
// Returns the path to the extracted Laravel root.
func extractLaravel() (string, error) {
	tmpDir, err := os.MkdirTemp("", "core-app-laravel-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	err = fs.WalkDir(laravelFiles, "laravel", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("laravel", path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(tmpDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		data, err := laravelFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		return os.WriteFile(targetPath, data, 0o644)
	})

	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("extract Laravel: %w", err)
	}

	return tmpDir, nil
}
