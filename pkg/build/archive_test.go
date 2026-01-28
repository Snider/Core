package build

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupArchiveTestFile creates a test binary file in a temp directory with the standard structure.
// Returns the path to the binary and the output directory.
func setupArchiveTestFile(t *testing.T, name, os_, arch string) (binaryPath string, outputDir string) {
	t.Helper()

	outputDir = t.TempDir()

	// Create platform directory: dist/os_arch
	platformDir := filepath.Join(outputDir, os_+"_"+arch)
	err := os.MkdirAll(platformDir, 0755)
	require.NoError(t, err)

	// Create test binary
	binaryPath = filepath.Join(platformDir, name)
	content := []byte("#!/bin/bash\necho 'Hello, World!'\n")
	err = os.WriteFile(binaryPath, content, 0755)
	require.NoError(t, err)

	return binaryPath, outputDir
}

func TestArchive_Good(t *testing.T) {
	t.Run("creates tar.gz for linux", func(t *testing.T) {
		binaryPath, outputDir := setupArchiveTestFile(t, "myapp", "linux", "amd64")

		artifact := Artifact{
			Path: binaryPath,
			OS:   "linux",
			Arch: "amd64",
		}

		result, err := Archive(artifact)
		require.NoError(t, err)

		// Verify archive was created
		expectedPath := filepath.Join(outputDir, "myapp_linux_amd64.tar.gz")
		assert.Equal(t, expectedPath, result.Path)
		assert.FileExists(t, result.Path)

		// Verify OS and Arch are preserved
		assert.Equal(t, "linux", result.OS)
		assert.Equal(t, "amd64", result.Arch)

		// Verify archive content
		verifyTarGzContent(t, result.Path, "myapp")
	})

	t.Run("creates tar.gz for darwin", func(t *testing.T) {
		binaryPath, outputDir := setupArchiveTestFile(t, "myapp", "darwin", "arm64")

		artifact := Artifact{
			Path: binaryPath,
			OS:   "darwin",
			Arch: "arm64",
		}

		result, err := Archive(artifact)
		require.NoError(t, err)

		expectedPath := filepath.Join(outputDir, "myapp_darwin_arm64.tar.gz")
		assert.Equal(t, expectedPath, result.Path)
		assert.FileExists(t, result.Path)

		verifyTarGzContent(t, result.Path, "myapp")
	})

	t.Run("creates zip for windows", func(t *testing.T) {
		binaryPath, outputDir := setupArchiveTestFile(t, "myapp.exe", "windows", "amd64")

		artifact := Artifact{
			Path: binaryPath,
			OS:   "windows",
			Arch: "amd64",
		}

		result, err := Archive(artifact)
		require.NoError(t, err)

		// Windows archives should strip .exe from archive name
		expectedPath := filepath.Join(outputDir, "myapp_windows_amd64.zip")
		assert.Equal(t, expectedPath, result.Path)
		assert.FileExists(t, result.Path)

		verifyZipContent(t, result.Path, "myapp.exe")
	})

	t.Run("preserves checksum field", func(t *testing.T) {
		binaryPath, _ := setupArchiveTestFile(t, "myapp", "linux", "amd64")

		artifact := Artifact{
			Path:     binaryPath,
			OS:       "linux",
			Arch:     "amd64",
			Checksum: "abc123",
		}

		result, err := Archive(artifact)
		require.NoError(t, err)
		assert.Equal(t, "abc123", result.Checksum)
	})
}

func TestArchive_Bad(t *testing.T) {
	t.Run("returns error for empty path", func(t *testing.T) {
		artifact := Artifact{
			Path: "",
			OS:   "linux",
			Arch: "amd64",
		}

		result, err := Archive(artifact)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "artifact path is empty")
		assert.Empty(t, result.Path)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		artifact := Artifact{
			Path: "/nonexistent/path/binary",
			OS:   "linux",
			Arch: "amd64",
		}

		result, err := Archive(artifact)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source file not found")
		assert.Empty(t, result.Path)
	})

	t.Run("returns error for directory path", func(t *testing.T) {
		dir := t.TempDir()

		artifact := Artifact{
			Path: dir,
			OS:   "linux",
			Arch: "amd64",
		}

		result, err := Archive(artifact)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source path is a directory")
		assert.Empty(t, result.Path)
	})
}

func TestArchiveAll_Good(t *testing.T) {
	t.Run("archives multiple artifacts", func(t *testing.T) {
		outputDir := t.TempDir()

		// Create multiple binaries
		var artifacts []Artifact
		targets := []struct {
			os_  string
			arch string
		}{
			{"linux", "amd64"},
			{"linux", "arm64"},
			{"darwin", "arm64"},
			{"windows", "amd64"},
		}

		for _, target := range targets {
			platformDir := filepath.Join(outputDir, target.os_+"_"+target.arch)
			err := os.MkdirAll(platformDir, 0755)
			require.NoError(t, err)

			name := "myapp"
			if target.os_ == "windows" {
				name = "myapp.exe"
			}

			binaryPath := filepath.Join(platformDir, name)
			err = os.WriteFile(binaryPath, []byte("binary content"), 0755)
			require.NoError(t, err)

			artifacts = append(artifacts, Artifact{
				Path: binaryPath,
				OS:   target.os_,
				Arch: target.arch,
			})
		}

		results, err := ArchiveAll(artifacts)
		require.NoError(t, err)
		require.Len(t, results, 4)

		// Verify all archives were created
		for i, result := range results {
			assert.FileExists(t, result.Path)
			assert.Equal(t, artifacts[i].OS, result.OS)
			assert.Equal(t, artifacts[i].Arch, result.Arch)
		}
	})

	t.Run("returns nil for empty slice", func(t *testing.T) {
		results, err := ArchiveAll([]Artifact{})
		assert.NoError(t, err)
		assert.Nil(t, results)
	})

	t.Run("returns nil for nil slice", func(t *testing.T) {
		results, err := ArchiveAll(nil)
		assert.NoError(t, err)
		assert.Nil(t, results)
	})
}

func TestArchiveAll_Bad(t *testing.T) {
	t.Run("returns partial results on error", func(t *testing.T) {
		binaryPath, _ := setupArchiveTestFile(t, "myapp", "linux", "amd64")

		artifacts := []Artifact{
			{Path: binaryPath, OS: "linux", Arch: "amd64"},
			{Path: "/nonexistent/binary", OS: "linux", Arch: "arm64"}, // This will fail
		}

		results, err := ArchiveAll(artifacts)
		assert.Error(t, err)
		// Should have the first successful result
		assert.Len(t, results, 1)
		assert.FileExists(t, results[0].Path)
	})
}

func TestArchiveFilename_Good(t *testing.T) {
	t.Run("generates correct tar.gz filename", func(t *testing.T) {
		artifact := Artifact{
			Path: "/output/linux_amd64/myapp",
			OS:   "linux",
			Arch: "amd64",
		}

		filename := archiveFilename(artifact, ".tar.gz")
		assert.Equal(t, "/output/myapp_linux_amd64.tar.gz", filename)
	})

	t.Run("generates correct zip filename", func(t *testing.T) {
		artifact := Artifact{
			Path: "/output/windows_amd64/myapp.exe",
			OS:   "windows",
			Arch: "amd64",
		}

		filename := archiveFilename(artifact, ".zip")
		assert.Equal(t, "/output/myapp_windows_amd64.zip", filename)
	})

	t.Run("handles nested output directories", func(t *testing.T) {
		artifact := Artifact{
			Path: "/project/dist/linux_arm64/cli",
			OS:   "linux",
			Arch: "arm64",
		}

		filename := archiveFilename(artifact, ".tar.gz")
		assert.Equal(t, "/project/dist/cli_linux_arm64.tar.gz", filename)
	})
}

// verifyTarGzContent opens a tar.gz file and verifies it contains the expected file.
func verifyTarGzContent(t *testing.T, archivePath, expectedName string) {
	t.Helper()

	file, err := os.Open(archivePath)
	require.NoError(t, err)
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	require.NoError(t, err)
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	header, err := tarReader.Next()
	require.NoError(t, err)
	assert.Equal(t, expectedName, header.Name)

	// Verify there's only one file
	_, err = tarReader.Next()
	assert.Equal(t, io.EOF, err)
}

// verifyZipContent opens a zip file and verifies it contains the expected file.
func verifyZipContent(t *testing.T, archivePath, expectedName string) {
	t.Helper()

	reader, err := zip.OpenReader(archivePath)
	require.NoError(t, err)
	defer reader.Close()

	require.Len(t, reader.File, 1)
	assert.Equal(t, expectedName, reader.File[0].Name)
}
