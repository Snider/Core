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
	var m Medium = Local
	assert.NotNil(t, m)
}
