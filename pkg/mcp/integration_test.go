package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration_FileTools(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := New(WithWorkspaceRoot(tmpDir))
	assert.NoError(t, err)

	ctx := context.Background()

	// 1. Test file_write
	writeInput := WriteFileInput{
		Path:    "test.txt",
		Content: "hello world",
	}
	_, writeOutput, err := s.writeFile(ctx, nil, writeInput)
	assert.NoError(t, err)
	assert.True(t, writeOutput.Success)
	assert.Equal(t, "test.txt", writeOutput.Path)

	// Verify on disk
	content, _ := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
	assert.Equal(t, "hello world", string(content))

	// 2. Test file_read
	readInput := ReadFileInput{
		Path: "test.txt",
	}
	_, readOutput, err := s.readFile(ctx, nil, readInput)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", readOutput.Content)
	assert.Equal(t, "plaintext", readOutput.Language)

	// 3. Test file_edit (replace_all=false)
	editInput := EditDiffInput{
		Path:      "test.txt",
		OldString: "world",
		NewString: "mcp",
	}
	_, editOutput, err := s.editDiff(ctx, nil, editInput)
	assert.NoError(t, err)
	assert.True(t, editOutput.Success)
	assert.Equal(t, 1, editOutput.Replacements)

	// Verify change
	_, readOutput, _ = s.readFile(ctx, nil, readInput)
	assert.Equal(t, "hello mcp", readOutput.Content)

	// 4. Test file_edit (replace_all=true)
	_ = s.medium.Write("multi.txt", "abc abc abc")
	editInputMulti := EditDiffInput{
		Path:       "multi.txt",
		OldString:  "abc",
		NewString:  "xyz",
		ReplaceAll: true,
	}
	_, editOutput, err = s.editDiff(ctx, nil, editInputMulti)
	assert.NoError(t, err)
	assert.Equal(t, 3, editOutput.Replacements)

	content, _ = os.ReadFile(filepath.Join(tmpDir, "multi.txt"))
	assert.Equal(t, "xyz xyz xyz", string(content))

	// 5. Test dir_list
	_ = s.medium.EnsureDir("subdir")
	_ = s.medium.Write("subdir/file1.txt", "content1")

	listInput := ListDirectoryInput{
		Path: "subdir",
	}
	_, listOutput, err := s.listDirectory(ctx, nil, listInput)
	assert.NoError(t, err)
	assert.Len(t, listOutput.Entries, 1)
	assert.Equal(t, "file1.txt", listOutput.Entries[0].Name)
	assert.False(t, listOutput.Entries[0].IsDir)
}

func TestIntegration_ErrorPaths(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := New(WithWorkspaceRoot(tmpDir))
	assert.NoError(t, err)

	ctx := context.Background()

	// Read nonexistent file
	_, _, err = s.readFile(ctx, nil, ReadFileInput{Path: "nonexistent.txt"})
	assert.Error(t, err)

	// Edit nonexistent file
	_, _, err = s.editDiff(ctx, nil, EditDiffInput{
		Path:      "nonexistent.txt",
		OldString: "foo",
		NewString: "bar",
	})
	assert.Error(t, err)

	// Edit with empty old_string
	_, _, err = s.editDiff(ctx, nil, EditDiffInput{
		Path:      "test.txt",
		OldString: "",
		NewString: "bar",
	})
	assert.Error(t, err)

	// Edit with old_string not found
	_ = s.medium.Write("test.txt", "hello")
	_, _, err = s.editDiff(ctx, nil, EditDiffInput{
		Path:      "test.txt",
		OldString: "missing",
		NewString: "bar",
	})
	assert.Error(t, err)
}
