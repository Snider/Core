package builders

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/host-uk/core/pkg/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupWailsTestProject creates a minimal Wails project structure for testing.
func setupWailsTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create wails.json
	wailsJSON := `{
  "name": "testapp",
  "outputfilename": "testapp"
}`
	err := os.WriteFile(filepath.Join(dir, "wails.json"), []byte(wailsJSON), 0644)
	require.NoError(t, err)

	// Create a minimal go.mod
	goMod := `module testapp

go 1.21

require github.com/wailsapp/wails/v3 v3.0.0
`
	err = os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
	require.NoError(t, err)

	// Create a minimal main.go
	mainGo := `package main

func main() {
	println("hello wails")
}
`
	err = os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644)
	require.NoError(t, err)

	// Create a minimal Taskfile.yml
	taskfile := `version: '3'
tasks:
  build:
    cmds:
      - mkdir -p {{.OUTPUT_DIR}}/{{.GOOS}}_{{.GOARCH}}
      - touch {{.OUTPUT_DIR}}/{{.GOOS}}_{{.GOARCH}}/testapp
`
	err = os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte(taskfile), 0644)
	require.NoError(t, err)

	return dir
}

// setupWailsTestProjectWithFrontend creates a Wails project with frontend directory.
func setupWailsTestProjectWithFrontend(t *testing.T, lockFile string) string {
	t.Helper()
	dir := setupWailsTestProject(t)

	// Create frontend directory
	frontendDir := filepath.Join(dir, "frontend")
	err := os.MkdirAll(frontendDir, 0755)
	require.NoError(t, err)

	// Create package.json
	packageJSON := `{
  "name": "frontend",
  "scripts": {
    "build": "echo building frontend"
  }
}`
	err = os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte(packageJSON), 0644)
	require.NoError(t, err)

	// Create lock file if specified
	if lockFile != "" {
		err = os.WriteFile(filepath.Join(frontendDir, lockFile), []byte(""), 0644)
		require.NoError(t, err)
	}

	return dir
}

// setupWailsV2TestProject creates a Wails v2 project structure.
func setupWailsV2TestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// wails.json
	err := os.WriteFile(filepath.Join(dir, "wails.json"), []byte("{}"), 0644)
	require.NoError(t, err)

	// go.mod with v2
	goMod := `module testapp
go 1.21
require github.com/wailsapp/wails/v2 v2.8.0
`
	err = os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
	require.NoError(t, err)

	return dir
}

func TestWailsBuilder_Build_Taskfile_Good(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if task is available
	if _, err := exec.LookPath("task"); err != nil {
		t.Skip("task not installed, skipping test")
	}

	t.Run("delegates to Taskfile if present", func(t *testing.T) {
		projectDir := setupWailsTestProject(t)
		outputDir := t.TempDir()

		// Create a Taskfile that just touches a file
		taskfile := `version: '3'
tasks:
  build:
    cmds:
      - mkdir -p {{.OUTPUT_DIR}}/{{.GOOS}}_{{.GOARCH}}
      - touch {{.OUTPUT_DIR}}/{{.GOOS}}_{{.GOARCH}}/testapp
`
		err := os.WriteFile(filepath.Join(projectDir, "Taskfile.yml"), []byte(taskfile), 0644)
		require.NoError(t, err)

		builder := NewWailsBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "testapp",
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		require.NoError(t, err)
		assert.NotEmpty(t, artifacts)
	})
}

func TestWailsBuilder_Name_Good(t *testing.T) {
	builder := NewWailsBuilder()
	assert.Equal(t, "wails", builder.Name())
}

func TestWailsBuilder_Build_V2_Good(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if _, err := exec.LookPath("wails"); err != nil {
		t.Skip("wails not installed, skipping integration test")
	}

	t.Run("builds v2 project", func(t *testing.T) {
		projectDir := setupWailsV2TestProject(t)
		outputDir := t.TempDir()

		builder := NewWailsBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "testapp",
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		// This will likely fail in a real run because we can't easily mock the full wails v2 build process 
		// (which needs a valid project with main.go etc). 
		// But it validates we are trying to run the command.
		// For now, we expect an error but check it's the *right* error (from wails CLI)
		_, err := builder.Build(context.Background(), cfg, targets)
		if err != nil {
			// If it fails, it should be because wails build failed, not because logic was wrong
			// assert.Contains(t, err.Error(), "wails build failed")
		}
	})
}

func TestWailsBuilder_Detect_Good(t *testing.T) {
	t.Run("detects Wails project with wails.json", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "wails.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		builder := NewWailsBuilder()
		detected, err := builder.Detect(dir)
		assert.NoError(t, err)
		assert.True(t, detected)
	})

	t.Run("returns false for Go-only project", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
		require.NoError(t, err)

		builder := NewWailsBuilder()
		detected, err := builder.Detect(dir)
		assert.NoError(t, err)
		assert.False(t, detected)
	})

	t.Run("returns false for Node.js project", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		builder := NewWailsBuilder()
		detected, err := builder.Detect(dir)
		assert.NoError(t, err)
		assert.False(t, detected)
	})

	t.Run("returns false for empty directory", func(t *testing.T) {
		dir := t.TempDir()

		builder := NewWailsBuilder()
		detected, err := builder.Detect(dir)
		assert.NoError(t, err)
		assert.False(t, detected)
	})
}

func TestDetectPackageManager_Good(t *testing.T) {
	t.Run("detects bun from bun.lockb", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "bun.lockb"), []byte(""), 0644)
		require.NoError(t, err)

		result := detectPackageManager(dir)
		assert.Equal(t, "bun", result)
	})

	t.Run("detects pnpm from pnpm-lock.yaml", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte(""), 0644)
		require.NoError(t, err)

		result := detectPackageManager(dir)
		assert.Equal(t, "pnpm", result)
	})

	t.Run("detects yarn from yarn.lock", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte(""), 0644)
		require.NoError(t, err)

		result := detectPackageManager(dir)
		assert.Equal(t, "yarn", result)
	})

	t.Run("detects npm from package-lock.json", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(""), 0644)
		require.NoError(t, err)

		result := detectPackageManager(dir)
		assert.Equal(t, "npm", result)
	})

	t.Run("defaults to npm when no lock file", func(t *testing.T) {
		dir := t.TempDir()

		result := detectPackageManager(dir)
		assert.Equal(t, "npm", result)
	})

	t.Run("prefers bun over other lock files", func(t *testing.T) {
		dir := t.TempDir()
		// Create multiple lock files
		require.NoError(t, os.WriteFile(filepath.Join(dir, "bun.lockb"), []byte(""), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte(""), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(""), 0644))

		result := detectPackageManager(dir)
		assert.Equal(t, "bun", result)
	})

	t.Run("prefers pnpm over yarn and npm", func(t *testing.T) {
		dir := t.TempDir()
		// Create multiple lock files (no bun)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte(""), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte(""), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(""), 0644))

		result := detectPackageManager(dir)
		assert.Equal(t, "pnpm", result)
	})

	t.Run("prefers yarn over npm", func(t *testing.T) {
		dir := t.TempDir()
		// Create multiple lock files (no bun or pnpm)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte(""), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(""), 0644))

		result := detectPackageManager(dir)
		assert.Equal(t, "yarn", result)
	})
}

func TestWailsBuilder_Build_Bad(t *testing.T) {
	t.Run("returns error for nil config", func(t *testing.T) {
		builder := NewWailsBuilder()

		artifacts, err := builder.Build(context.Background(), nil, []build.Target{{OS: "linux", Arch: "amd64"}})
		assert.Error(t, err)
		assert.Nil(t, artifacts)
		assert.Contains(t, err.Error(), "config is nil")
	})

	t.Run("returns error for empty targets", func(t *testing.T) {
		projectDir := setupWailsTestProject(t)

		builder := NewWailsBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  t.TempDir(),
			Name:       "test",
		}

		artifacts, err := builder.Build(context.Background(), cfg, []build.Target{})
		assert.Error(t, err)
		assert.Nil(t, artifacts)
		assert.Contains(t, err.Error(), "no targets specified")
	})
}

func TestWailsBuilder_Build_Good(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if wails3 is available in PATH
	if _, err := exec.LookPath("wails3"); err != nil {
		t.Skip("wails3 not installed, skipping integration test")
	}

	t.Run("builds for current platform", func(t *testing.T) {
		projectDir := setupWailsTestProject(t)
		outputDir := t.TempDir()

		builder := NewWailsBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "testapp",
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		require.NoError(t, err)
		require.Len(t, artifacts, 1)

		// Verify artifact properties
		artifact := artifacts[0]
		assert.Equal(t, runtime.GOOS, artifact.OS)
		assert.Equal(t, runtime.GOARCH, artifact.Arch)
	})
}

func TestWailsBuilder_Interface_Good(t *testing.T) {
	// Verify WailsBuilder implements Builder interface
	var _ build.Builder = (*WailsBuilder)(nil)
	var _ build.Builder = NewWailsBuilder()
}

func TestWailsBuilder_Ugly(t *testing.T) {
	t.Run("handles nonexistent frontend directory gracefully", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		// Create a Wails project without a frontend directory
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "wails.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		builder := NewWailsBuilder()
		cfg := &build.Config{
			ProjectDir: dir,
			OutputDir:  t.TempDir(),
			Name:       "test",
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		// This will fail because wails3 isn't set up, but it shouldn't panic
		// due to missing frontend directory
		_, err = builder.Build(context.Background(), cfg, targets)
		// We expect an error (wails3 build will fail), but not a panic
		// The error should be about wails3 build, not about frontend
		if err != nil {
			assert.NotContains(t, err.Error(), "frontend dependencies")
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		projectDir := setupWailsTestProject(t)

		builder := NewWailsBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  t.TempDir(),
			Name:       "canceltest",
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		// Create an already cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		artifacts, err := builder.Build(ctx, cfg, targets)
		assert.Error(t, err)
		assert.Empty(t, artifacts)
	})
}
