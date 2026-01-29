package devops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectServeCommand_Good_Laravel(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "artisan"), []byte("#!/usr/bin/env php"), 0644)
	assert.NoError(t, err)

	cmd := DetectServeCommand(tmpDir)
	assert.Equal(t, "php artisan octane:start --host=0.0.0.0 --port=8000", cmd)
}

func TestDetectServeCommand_Good_NodeDev(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := `{"scripts":{"dev":"vite","start":"node index.js"}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644)
	assert.NoError(t, err)

	cmd := DetectServeCommand(tmpDir)
	assert.Equal(t, "npm run dev -- --host 0.0.0.0", cmd)
}

func TestDetectServeCommand_Good_NodeStart(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := `{"scripts":{"start":"node server.js"}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644)
	assert.NoError(t, err)

	cmd := DetectServeCommand(tmpDir)
	assert.Equal(t, "npm start", cmd)
}

func TestDetectServeCommand_Good_PHP(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(`{"require":{}}`), 0644)
	assert.NoError(t, err)

	cmd := DetectServeCommand(tmpDir)
	assert.Equal(t, "frankenphp php-server -l :8000", cmd)
}

func TestDetectServeCommand_Good_GoMain(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	assert.NoError(t, err)

	cmd := DetectServeCommand(tmpDir)
	assert.Equal(t, "go run .", cmd)
}

func TestDetectServeCommand_Good_GoWithoutMain(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example"), 0644)
	assert.NoError(t, err)

	// No main.go, so falls through to fallback
	cmd := DetectServeCommand(tmpDir)
	assert.Equal(t, "python3 -m http.server 8000", cmd)
}

func TestDetectServeCommand_Good_Django(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "manage.py"), []byte("#!/usr/bin/env python"), 0644)
	assert.NoError(t, err)

	cmd := DetectServeCommand(tmpDir)
	assert.Equal(t, "python manage.py runserver 0.0.0.0:8000", cmd)
}

func TestDetectServeCommand_Good_Fallback(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := DetectServeCommand(tmpDir)
	assert.Equal(t, "python3 -m http.server 8000", cmd)
}

func TestDetectServeCommand_Good_Priority(t *testing.T) {
	// Laravel (artisan) should take priority over PHP (composer.json)
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "artisan"), []byte("#!/usr/bin/env php"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(`{"require":{}}`), 0644)
	assert.NoError(t, err)

	cmd := DetectServeCommand(tmpDir)
	assert.Equal(t, "php artisan octane:start --host=0.0.0.0 --port=8000", cmd)
}

func TestServeOptions_Default(t *testing.T) {
	opts := ServeOptions{}
	assert.Equal(t, 0, opts.Port)
	assert.Equal(t, "", opts.Path)
}

func TestServeOptions_Custom(t *testing.T) {
	opts := ServeOptions{
		Port: 3000,
		Path: "public",
	}
	assert.Equal(t, 3000, opts.Port)
	assert.Equal(t, "public", opts.Path)
}

func TestHasFile_Good(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	assert.NoError(t, err)

	assert.True(t, hasFile(tmpDir, "test.txt"))
}

func TestHasFile_Bad(t *testing.T) {
	tmpDir := t.TempDir()

	assert.False(t, hasFile(tmpDir, "nonexistent.txt"))
}

func TestHasFile_Bad_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	assert.NoError(t, err)

	// hasFile returns true for directories too (it's just checking existence)
	assert.True(t, hasFile(tmpDir, "subdir"))
}
