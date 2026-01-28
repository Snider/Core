package builders

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/host-uk/core/pkg/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGoTestProject creates a minimal Go project for testing.
func setupGoTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create a minimal go.mod
	goMod := `module testproject

go 1.21
`
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
	require.NoError(t, err)

	// Create a minimal main.go
	mainGo := `package main

func main() {
	println("hello")
}
`
	err = os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644)
	require.NoError(t, err)

	return dir
}

func TestGoBuilder_Name_Good(t *testing.T) {
	builder := NewGoBuilder()
	assert.Equal(t, "go", builder.Name())
}

func TestGoBuilder_Detect_Good(t *testing.T) {
	t.Run("detects Go project with go.mod", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
		require.NoError(t, err)

		builder := NewGoBuilder()
		detected, err := builder.Detect(dir)
		assert.NoError(t, err)
		assert.True(t, detected)
	})

	t.Run("detects Wails project", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "wails.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		builder := NewGoBuilder()
		detected, err := builder.Detect(dir)
		assert.NoError(t, err)
		assert.True(t, detected)
	})

	t.Run("returns false for non-Go project", func(t *testing.T) {
		dir := t.TempDir()
		// Create a Node.js project instead
		err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		builder := NewGoBuilder()
		detected, err := builder.Detect(dir)
		assert.NoError(t, err)
		assert.False(t, detected)
	})

	t.Run("returns false for empty directory", func(t *testing.T) {
		dir := t.TempDir()

		builder := NewGoBuilder()
		detected, err := builder.Detect(dir)
		assert.NoError(t, err)
		assert.False(t, detected)
	})
}

func TestGoBuilder_Build_Good(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("builds for current platform", func(t *testing.T) {
		projectDir := setupGoTestProject(t)
		outputDir := t.TempDir()

		builder := NewGoBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "testbinary",
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

		// Verify binary was created
		assert.FileExists(t, artifact.Path)

		// Verify the path is in the expected location
		expectedName := "testbinary"
		if runtime.GOOS == "windows" {
			expectedName += ".exe"
		}
		assert.Contains(t, artifact.Path, expectedName)
	})

	t.Run("builds multiple targets", func(t *testing.T) {
		projectDir := setupGoTestProject(t)
		outputDir := t.TempDir()

		builder := NewGoBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "multitest",
		}
		targets := []build.Target{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		require.NoError(t, err)
		require.Len(t, artifacts, 2)

		// Verify both artifacts were created
		for i, artifact := range artifacts {
			assert.Equal(t, targets[i].OS, artifact.OS)
			assert.Equal(t, targets[i].Arch, artifact.Arch)
			assert.FileExists(t, artifact.Path)
		}
	})

	t.Run("adds .exe extension for Windows", func(t *testing.T) {
		projectDir := setupGoTestProject(t)
		outputDir := t.TempDir()

		builder := NewGoBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "wintest",
		}
		targets := []build.Target{
			{OS: "windows", Arch: "amd64"},
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		require.NoError(t, err)
		require.Len(t, artifacts, 1)

		// Verify .exe extension
		assert.True(t, filepath.Ext(artifacts[0].Path) == ".exe")
		assert.FileExists(t, artifacts[0].Path)
	})

	t.Run("uses directory name when Name not specified", func(t *testing.T) {
		projectDir := setupGoTestProject(t)
		outputDir := t.TempDir()

		builder := NewGoBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "", // Empty name
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		require.NoError(t, err)
		require.Len(t, artifacts, 1)

		// Binary should use the project directory base name
		baseName := filepath.Base(projectDir)
		if runtime.GOOS == "windows" {
			baseName += ".exe"
		}
		assert.Contains(t, artifacts[0].Path, baseName)
	})

	t.Run("applies ldflags", func(t *testing.T) {
		projectDir := setupGoTestProject(t)
		outputDir := t.TempDir()

		builder := NewGoBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "ldflagstest",
			LDFlags:    []string{"-s", "-w"}, // Strip debug info
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		require.NoError(t, err)
		require.Len(t, artifacts, 1)
		assert.FileExists(t, artifacts[0].Path)
	})

	t.Run("creates output directory if missing", func(t *testing.T) {
		projectDir := setupGoTestProject(t)
		outputDir := filepath.Join(t.TempDir(), "nested", "output")

		builder := NewGoBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "nestedtest",
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		require.NoError(t, err)
		require.Len(t, artifacts, 1)
		assert.FileExists(t, artifacts[0].Path)
		assert.DirExists(t, outputDir)
	})
}

func TestGoBuilder_Build_Bad(t *testing.T) {
	t.Run("returns error for nil config", func(t *testing.T) {
		builder := NewGoBuilder()

		artifacts, err := builder.Build(context.Background(), nil, []build.Target{{OS: "linux", Arch: "amd64"}})
		assert.Error(t, err)
		assert.Nil(t, artifacts)
		assert.Contains(t, err.Error(), "config is nil")
	})

	t.Run("returns error for empty targets", func(t *testing.T) {
		projectDir := setupGoTestProject(t)

		builder := NewGoBuilder()
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

	t.Run("returns error for invalid project directory", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		builder := NewGoBuilder()
		cfg := &build.Config{
			ProjectDir: "/nonexistent/path",
			OutputDir:  t.TempDir(),
			Name:       "test",
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		assert.Error(t, err)
		assert.Empty(t, artifacts)
	})

	t.Run("returns error for invalid Go code", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		dir := t.TempDir()

		// Create go.mod
		err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21"), 0644)
		require.NoError(t, err)

		// Create invalid Go code
		err = os.WriteFile(filepath.Join(dir, "main.go"), []byte("this is not valid go code"), 0644)
		require.NoError(t, err)

		builder := NewGoBuilder()
		cfg := &build.Config{
			ProjectDir: dir,
			OutputDir:  t.TempDir(),
			Name:       "test",
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "go build failed")
		assert.Empty(t, artifacts)
	})

	t.Run("returns partial artifacts on partial failure", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		// Create a project that will fail on one target
		// Using an invalid arch for linux
		projectDir := setupGoTestProject(t)
		outputDir := t.TempDir()

		builder := NewGoBuilder()
		cfg := &build.Config{
			ProjectDir: projectDir,
			OutputDir:  outputDir,
			Name:       "partialtest",
		}
		targets := []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH}, // This should succeed
			{OS: "linux", Arch: "invalid_arch"},      // This should fail
		}

		artifacts, err := builder.Build(context.Background(), cfg, targets)
		// Should return error for the failed build
		assert.Error(t, err)
		// Should have the successful artifact
		assert.Len(t, artifacts, 1)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		projectDir := setupGoTestProject(t)

		builder := NewGoBuilder()
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

func TestGoBuilder_Interface_Good(t *testing.T) {
	// Verify GoBuilder implements Builder interface
	var _ build.Builder = (*GoBuilder)(nil)
	var _ build.Builder = NewGoBuilder()
}
