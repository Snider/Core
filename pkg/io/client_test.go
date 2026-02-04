package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- MockMedium Tests ---

func TestNewMockMedium_Good(t *testing.T) {
	m := NewMockMedium()
	assert.NotNil(t, m)
	assert.NotNil(t, m.Files)
	assert.NotNil(t, m.Dirs)
	assert.Empty(t, m.Files)
	assert.Empty(t, m.Dirs)
}

func TestMockMedium_Read_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["test.txt"] = "hello world"
	content, err := m.Read("test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", content)
}

func TestMockMedium_Read_Bad(t *testing.T) {
	m := NewMockMedium()
	_, err := m.Read("nonexistent.txt")
	assert.Error(t, err)
}

func TestMockMedium_Write_Good(t *testing.T) {
	m := NewMockMedium()
	err := m.Write("test.txt", "content")
	assert.NoError(t, err)
	assert.Equal(t, "content", m.Files["test.txt"])

	// Overwrite existing file
	err = m.Write("test.txt", "new content")
	assert.NoError(t, err)
	assert.Equal(t, "new content", m.Files["test.txt"])
}

func TestMockMedium_EnsureDir_Good(t *testing.T) {
	m := NewMockMedium()
	err := m.EnsureDir("/path/to/dir")
	assert.NoError(t, err)
	assert.True(t, m.Dirs["/path/to/dir"])
}

func TestMockMedium_IsFile_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["exists.txt"] = "content"

	assert.True(t, m.IsFile("exists.txt"))
	assert.False(t, m.IsFile("nonexistent.txt"))
}

func TestMockMedium_FileGet_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["test.txt"] = "content"
	content, err := m.FileGet("test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "content", content)
}

func TestMockMedium_FileSet_Good(t *testing.T) {
	m := NewMockMedium()
	err := m.FileSet("test.txt", "content")
	assert.NoError(t, err)
	assert.Equal(t, "content", m.Files["test.txt"])
}

func TestMockMedium_Delete_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["test.txt"] = "content"

	err := m.Delete("test.txt")
	assert.NoError(t, err)
	assert.False(t, m.IsFile("test.txt"))
}

func TestMockMedium_Delete_Bad_NotFound(t *testing.T) {
	m := NewMockMedium()
	err := m.Delete("nonexistent.txt")
	assert.Error(t, err)
}

func TestMockMedium_Delete_Bad_DirNotEmpty(t *testing.T) {
	m := NewMockMedium()
	m.Dirs["mydir"] = true
	m.Files["mydir/file.txt"] = "content"

	err := m.Delete("mydir")
	assert.Error(t, err)
}

func TestMockMedium_DeleteAll_Good(t *testing.T) {
	m := NewMockMedium()
	m.Dirs["mydir"] = true
	m.Dirs["mydir/subdir"] = true
	m.Files["mydir/file.txt"] = "content"
	m.Files["mydir/subdir/nested.txt"] = "nested"

	err := m.DeleteAll("mydir")
	assert.NoError(t, err)
	assert.Empty(t, m.Dirs)
	assert.Empty(t, m.Files)
}

func TestMockMedium_Rename_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["old.txt"] = "content"

	err := m.Rename("old.txt", "new.txt")
	assert.NoError(t, err)
	assert.False(t, m.IsFile("old.txt"))
	assert.True(t, m.IsFile("new.txt"))
	assert.Equal(t, "content", m.Files["new.txt"])
}

func TestMockMedium_Rename_Good_Dir(t *testing.T) {
	m := NewMockMedium()
	m.Dirs["olddir"] = true
	m.Files["olddir/file.txt"] = "content"

	err := m.Rename("olddir", "newdir")
	assert.NoError(t, err)
	assert.False(t, m.Dirs["olddir"])
	assert.True(t, m.Dirs["newdir"])
	assert.Equal(t, "content", m.Files["newdir/file.txt"])
}

func TestMockMedium_List_Good(t *testing.T) {
	m := NewMockMedium()
	m.Dirs["mydir"] = true
	m.Files["mydir/file1.txt"] = "content1"
	m.Files["mydir/file2.txt"] = "content2"
	m.Dirs["mydir/subdir"] = true

	entries, err := m.List("mydir")
	assert.NoError(t, err)
	assert.Len(t, entries, 3)

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	assert.True(t, names["file1.txt"])
	assert.True(t, names["file2.txt"])
	assert.True(t, names["subdir"])
}

func TestMockMedium_Stat_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["test.txt"] = "hello world"

	info, err := m.Stat("test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "test.txt", info.Name())
	assert.Equal(t, int64(11), info.Size())
	assert.False(t, info.IsDir())
}

func TestMockMedium_Stat_Good_Dir(t *testing.T) {
	m := NewMockMedium()
	m.Dirs["mydir"] = true

	info, err := m.Stat("mydir")
	assert.NoError(t, err)
	assert.Equal(t, "mydir", info.Name())
	assert.True(t, info.IsDir())
}

func TestMockMedium_Exists_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["file.txt"] = "content"
	m.Dirs["mydir"] = true

	assert.True(t, m.Exists("file.txt"))
	assert.True(t, m.Exists("mydir"))
	assert.False(t, m.Exists("nonexistent"))
}

func TestMockMedium_IsDir_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["file.txt"] = "content"
	m.Dirs["mydir"] = true

	assert.False(t, m.IsDir("file.txt"))
	assert.True(t, m.IsDir("mydir"))
	assert.False(t, m.IsDir("nonexistent"))
}

// --- Wrapper Function Tests ---

func TestRead_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["test.txt"] = "hello"
	content, err := Read(m, "test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello", content)
}

func TestWrite_Good(t *testing.T) {
	m := NewMockMedium()
	err := Write(m, "test.txt", "hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello", m.Files["test.txt"])
}

func TestEnsureDir_Good(t *testing.T) {
	m := NewMockMedium()
	err := EnsureDir(m, "/my/dir")
	assert.NoError(t, err)
	assert.True(t, m.Dirs["/my/dir"])
}

func TestIsFile_Good(t *testing.T) {
	m := NewMockMedium()
	m.Files["exists.txt"] = "content"

	assert.True(t, IsFile(m, "exists.txt"))
	assert.False(t, IsFile(m, "nonexistent.txt"))
}

func TestCopy_Good(t *testing.T) {
	source := NewMockMedium()
	dest := NewMockMedium()
	source.Files["test.txt"] = "hello"
	err := Copy(source, "test.txt", dest, "test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello", dest.Files["test.txt"])

	// Copy to different path
	source.Files["original.txt"] = "content"
	err = Copy(source, "original.txt", dest, "copied.txt")
	assert.NoError(t, err)
	assert.Equal(t, "content", dest.Files["copied.txt"])
}

func TestCopy_Bad(t *testing.T) {
	source := NewMockMedium()
	dest := NewMockMedium()
	err := Copy(source, "nonexistent.txt", dest, "dest.txt")
	assert.Error(t, err)
}

// --- Local Global Tests ---

func TestLocalGlobal_Good(t *testing.T) {
	// io.Local should be initialized by init()
	assert.NotNil(t, Local, "io.Local should be initialized")

	// Should be able to use it as a Medium
	var m = Local
	assert.NotNil(t, m)
}
