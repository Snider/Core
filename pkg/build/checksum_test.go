package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupChecksumTestFile creates a test file with known content.
func setupChecksumTestFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	return path
}

func TestChecksum_Good(t *testing.T) {
	t.Run("computes SHA256 checksum", func(t *testing.T) {
		// Known SHA256 of "Hello, World!\n"
		path := setupChecksumTestFile(t, "Hello, World!\n")
		expectedChecksum := "c98c24b677eff44860afea6f493bbaec5bb1c4cbb209c6fc2bbb47f66ff2ad31"

		artifact := Artifact{
			Path: path,
			OS:   "linux",
			Arch: "amd64",
		}

		result, err := Checksum(artifact)
		require.NoError(t, err)
		assert.Equal(t, expectedChecksum, result.Checksum)
	})

	t.Run("preserves artifact fields", func(t *testing.T) {
		path := setupChecksumTestFile(t, "test content")

		artifact := Artifact{
			Path: path,
			OS:   "darwin",
			Arch: "arm64",
		}

		result, err := Checksum(artifact)
		require.NoError(t, err)

		assert.Equal(t, path, result.Path)
		assert.Equal(t, "darwin", result.OS)
		assert.Equal(t, "arm64", result.Arch)
		assert.NotEmpty(t, result.Checksum)
	})

	t.Run("produces 64 character hex string", func(t *testing.T) {
		path := setupChecksumTestFile(t, "any content")

		artifact := Artifact{Path: path, OS: "linux", Arch: "amd64"}

		result, err := Checksum(artifact)
		require.NoError(t, err)

		// SHA256 produces 32 bytes = 64 hex characters
		assert.Len(t, result.Checksum, 64)
	})

	t.Run("different content produces different checksums", func(t *testing.T) {
		path1 := setupChecksumTestFile(t, "content one")
		path2 := setupChecksumTestFile(t, "content two")

		result1, err := Checksum(Artifact{Path: path1, OS: "linux", Arch: "amd64"})
		require.NoError(t, err)

		result2, err := Checksum(Artifact{Path: path2, OS: "linux", Arch: "amd64"})
		require.NoError(t, err)

		assert.NotEqual(t, result1.Checksum, result2.Checksum)
	})

	t.Run("same content produces same checksum", func(t *testing.T) {
		content := "identical content"
		path1 := setupChecksumTestFile(t, content)
		path2 := setupChecksumTestFile(t, content)

		result1, err := Checksum(Artifact{Path: path1, OS: "linux", Arch: "amd64"})
		require.NoError(t, err)

		result2, err := Checksum(Artifact{Path: path2, OS: "linux", Arch: "amd64"})
		require.NoError(t, err)

		assert.Equal(t, result1.Checksum, result2.Checksum)
	})
}

func TestChecksum_Bad(t *testing.T) {
	t.Run("returns error for empty path", func(t *testing.T) {
		artifact := Artifact{
			Path: "",
			OS:   "linux",
			Arch: "amd64",
		}

		result, err := Checksum(artifact)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "artifact path is empty")
		assert.Empty(t, result.Checksum)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		artifact := Artifact{
			Path: "/nonexistent/path/file",
			OS:   "linux",
			Arch: "amd64",
		}

		result, err := Checksum(artifact)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open file")
		assert.Empty(t, result.Checksum)
	})
}

func TestChecksumAll_Good(t *testing.T) {
	t.Run("checksums multiple artifacts", func(t *testing.T) {
		paths := []string{
			setupChecksumTestFile(t, "content one"),
			setupChecksumTestFile(t, "content two"),
			setupChecksumTestFile(t, "content three"),
		}

		artifacts := []Artifact{
			{Path: paths[0], OS: "linux", Arch: "amd64"},
			{Path: paths[1], OS: "darwin", Arch: "arm64"},
			{Path: paths[2], OS: "windows", Arch: "amd64"},
		}

		results, err := ChecksumAll(artifacts)
		require.NoError(t, err)
		require.Len(t, results, 3)

		for i, result := range results {
			assert.Equal(t, artifacts[i].Path, result.Path)
			assert.Equal(t, artifacts[i].OS, result.OS)
			assert.Equal(t, artifacts[i].Arch, result.Arch)
			assert.NotEmpty(t, result.Checksum)
		}
	})

	t.Run("returns nil for empty slice", func(t *testing.T) {
		results, err := ChecksumAll([]Artifact{})
		assert.NoError(t, err)
		assert.Nil(t, results)
	})

	t.Run("returns nil for nil slice", func(t *testing.T) {
		results, err := ChecksumAll(nil)
		assert.NoError(t, err)
		assert.Nil(t, results)
	})
}

func TestChecksumAll_Bad(t *testing.T) {
	t.Run("returns partial results on error", func(t *testing.T) {
		path := setupChecksumTestFile(t, "valid content")

		artifacts := []Artifact{
			{Path: path, OS: "linux", Arch: "amd64"},
			{Path: "/nonexistent/file", OS: "linux", Arch: "arm64"}, // This will fail
		}

		results, err := ChecksumAll(artifacts)
		assert.Error(t, err)
		// Should have the first successful result
		assert.Len(t, results, 1)
		assert.NotEmpty(t, results[0].Checksum)
	})
}

func TestWriteChecksumFile_Good(t *testing.T) {
	t.Run("writes checksum file with correct format", func(t *testing.T) {
		dir := t.TempDir()
		checksumPath := filepath.Join(dir, "CHECKSUMS.txt")

		artifacts := []Artifact{
			{Path: "/output/app_linux_amd64.tar.gz", Checksum: "abc123def456", OS: "linux", Arch: "amd64"},
			{Path: "/output/app_darwin_arm64.tar.gz", Checksum: "789xyz000111", OS: "darwin", Arch: "arm64"},
		}

		err := WriteChecksumFile(artifacts, checksumPath)
		require.NoError(t, err)

		// Read and verify content
		content, err := os.ReadFile(checksumPath)
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		require.Len(t, lines, 2)

		// Lines should be sorted alphabetically
		assert.Equal(t, "789xyz000111  app_darwin_arm64.tar.gz", lines[0])
		assert.Equal(t, "abc123def456  app_linux_amd64.tar.gz", lines[1])
	})

	t.Run("creates parent directories", func(t *testing.T) {
		dir := t.TempDir()
		checksumPath := filepath.Join(dir, "nested", "deep", "CHECKSUMS.txt")

		artifacts := []Artifact{
			{Path: "/output/app.tar.gz", Checksum: "abc123", OS: "linux", Arch: "amd64"},
		}

		err := WriteChecksumFile(artifacts, checksumPath)
		require.NoError(t, err)
		assert.FileExists(t, checksumPath)
	})

	t.Run("does nothing for empty artifacts", func(t *testing.T) {
		dir := t.TempDir()
		checksumPath := filepath.Join(dir, "CHECKSUMS.txt")

		err := WriteChecksumFile([]Artifact{}, checksumPath)
		require.NoError(t, err)

		// File should not exist
		_, err = os.Stat(checksumPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("does nothing for nil artifacts", func(t *testing.T) {
		dir := t.TempDir()
		checksumPath := filepath.Join(dir, "CHECKSUMS.txt")

		err := WriteChecksumFile(nil, checksumPath)
		require.NoError(t, err)
	})

	t.Run("uses only basename for filenames", func(t *testing.T) {
		dir := t.TempDir()
		checksumPath := filepath.Join(dir, "CHECKSUMS.txt")

		artifacts := []Artifact{
			{Path: "/some/deep/nested/path/myapp_linux_amd64.tar.gz", Checksum: "checksum123", OS: "linux", Arch: "amd64"},
		}

		err := WriteChecksumFile(artifacts, checksumPath)
		require.NoError(t, err)

		content, err := os.ReadFile(checksumPath)
		require.NoError(t, err)

		// Should only contain the basename
		assert.Contains(t, string(content), "myapp_linux_amd64.tar.gz")
		assert.NotContains(t, string(content), "/some/deep/nested/path/")
	})
}

func TestWriteChecksumFile_Bad(t *testing.T) {
	t.Run("returns error for artifact without checksum", func(t *testing.T) {
		dir := t.TempDir()
		checksumPath := filepath.Join(dir, "CHECKSUMS.txt")

		artifacts := []Artifact{
			{Path: "/output/app.tar.gz", Checksum: "", OS: "linux", Arch: "amd64"}, // No checksum
		}

		err := WriteChecksumFile(artifacts, checksumPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has no checksum")
	})
}
