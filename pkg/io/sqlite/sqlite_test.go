package sqlite

import (
	goio "io"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMedium(t *testing.T) *Medium {
	t.Helper()
	m, err := New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { m.Close() })
	return m
}

// --- Constructor Tests ---

func TestNew_Good(t *testing.T) {
	m, err := New(":memory:")
	require.NoError(t, err)
	defer m.Close()
	assert.Equal(t, "files", m.table)
}

func TestNew_Good_WithTable(t *testing.T) {
	m, err := New(":memory:", WithTable("custom"))
	require.NoError(t, err)
	defer m.Close()
	assert.Equal(t, "custom", m.table)
}

func TestNew_Bad_EmptyPath(t *testing.T) {
	_, err := New("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database path is required")
}

// --- Read/Write Tests ---

func TestReadWrite_Good(t *testing.T) {
	m := newTestMedium(t)

	err := m.Write("hello.txt", "world")
	require.NoError(t, err)

	content, err := m.Read("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "world", content)
}

func TestReadWrite_Good_Overwrite(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("file.txt", "first"))
	require.NoError(t, m.Write("file.txt", "second"))

	content, err := m.Read("file.txt")
	require.NoError(t, err)
	assert.Equal(t, "second", content)
}

func TestReadWrite_Good_NestedPath(t *testing.T) {
	m := newTestMedium(t)

	err := m.Write("a/b/c.txt", "nested")
	require.NoError(t, err)

	content, err := m.Read("a/b/c.txt")
	require.NoError(t, err)
	assert.Equal(t, "nested", content)
}

func TestRead_Bad_NotFound(t *testing.T) {
	m := newTestMedium(t)

	_, err := m.Read("nonexistent.txt")
	assert.Error(t, err)
}

func TestRead_Bad_EmptyPath(t *testing.T) {
	m := newTestMedium(t)

	_, err := m.Read("")
	assert.Error(t, err)
}

func TestWrite_Bad_EmptyPath(t *testing.T) {
	m := newTestMedium(t)

	err := m.Write("", "content")
	assert.Error(t, err)
}

func TestRead_Bad_IsDirectory(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.EnsureDir("mydir"))
	_, err := m.Read("mydir")
	assert.Error(t, err)
}

// --- EnsureDir Tests ---

func TestEnsureDir_Good(t *testing.T) {
	m := newTestMedium(t)

	err := m.EnsureDir("mydir")
	require.NoError(t, err)
	assert.True(t, m.IsDir("mydir"))
}

func TestEnsureDir_Good_EmptyPath(t *testing.T) {
	m := newTestMedium(t)
	// Root always exists, no-op
	err := m.EnsureDir("")
	assert.NoError(t, err)
}

func TestEnsureDir_Good_Idempotent(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.EnsureDir("mydir"))
	require.NoError(t, m.EnsureDir("mydir"))
	assert.True(t, m.IsDir("mydir"))
}

// --- IsFile Tests ---

func TestIsFile_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("file.txt", "content"))
	require.NoError(t, m.EnsureDir("mydir"))

	assert.True(t, m.IsFile("file.txt"))
	assert.False(t, m.IsFile("mydir"))
	assert.False(t, m.IsFile("nonexistent"))
	assert.False(t, m.IsFile(""))
}

// --- FileGet/FileSet Tests ---

func TestFileGetFileSet_Good(t *testing.T) {
	m := newTestMedium(t)

	err := m.FileSet("key.txt", "value")
	require.NoError(t, err)

	val, err := m.FileGet("key.txt")
	require.NoError(t, err)
	assert.Equal(t, "value", val)
}

// --- Delete Tests ---

func TestDelete_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("to-delete.txt", "content"))
	assert.True(t, m.Exists("to-delete.txt"))

	err := m.Delete("to-delete.txt")
	require.NoError(t, err)
	assert.False(t, m.Exists("to-delete.txt"))
}

func TestDelete_Good_EmptyDir(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.EnsureDir("emptydir"))
	assert.True(t, m.IsDir("emptydir"))

	err := m.Delete("emptydir")
	require.NoError(t, err)
	assert.False(t, m.IsDir("emptydir"))
}

func TestDelete_Bad_NotFound(t *testing.T) {
	m := newTestMedium(t)

	err := m.Delete("nonexistent")
	assert.Error(t, err)
}

func TestDelete_Bad_EmptyPath(t *testing.T) {
	m := newTestMedium(t)

	err := m.Delete("")
	assert.Error(t, err)
}

func TestDelete_Bad_NotEmpty(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.EnsureDir("mydir"))
	require.NoError(t, m.Write("mydir/file.txt", "content"))

	err := m.Delete("mydir")
	assert.Error(t, err)
}

// --- DeleteAll Tests ---

func TestDeleteAll_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("dir/file1.txt", "a"))
	require.NoError(t, m.Write("dir/sub/file2.txt", "b"))
	require.NoError(t, m.Write("other.txt", "c"))

	err := m.DeleteAll("dir")
	require.NoError(t, err)

	assert.False(t, m.Exists("dir/file1.txt"))
	assert.False(t, m.Exists("dir/sub/file2.txt"))
	assert.True(t, m.Exists("other.txt"))
}

func TestDeleteAll_Good_SingleFile(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("file.txt", "content"))

	err := m.DeleteAll("file.txt")
	require.NoError(t, err)
	assert.False(t, m.Exists("file.txt"))
}

func TestDeleteAll_Bad_NotFound(t *testing.T) {
	m := newTestMedium(t)

	err := m.DeleteAll("nonexistent")
	assert.Error(t, err)
}

func TestDeleteAll_Bad_EmptyPath(t *testing.T) {
	m := newTestMedium(t)

	err := m.DeleteAll("")
	assert.Error(t, err)
}

// --- Rename Tests ---

func TestRename_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("old.txt", "content"))

	err := m.Rename("old.txt", "new.txt")
	require.NoError(t, err)

	assert.False(t, m.Exists("old.txt"))
	assert.True(t, m.IsFile("new.txt"))

	content, err := m.Read("new.txt")
	require.NoError(t, err)
	assert.Equal(t, "content", content)
}

func TestRename_Good_Directory(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.EnsureDir("olddir"))
	require.NoError(t, m.Write("olddir/file.txt", "content"))

	err := m.Rename("olddir", "newdir")
	require.NoError(t, err)

	assert.False(t, m.Exists("olddir"))
	assert.False(t, m.Exists("olddir/file.txt"))
	assert.True(t, m.IsDir("newdir"))
	assert.True(t, m.IsFile("newdir/file.txt"))

	content, err := m.Read("newdir/file.txt")
	require.NoError(t, err)
	assert.Equal(t, "content", content)
}

func TestRename_Bad_SourceNotFound(t *testing.T) {
	m := newTestMedium(t)

	err := m.Rename("nonexistent", "new")
	assert.Error(t, err)
}

func TestRename_Bad_EmptyPath(t *testing.T) {
	m := newTestMedium(t)

	err := m.Rename("", "new")
	assert.Error(t, err)

	err = m.Rename("old", "")
	assert.Error(t, err)
}

// --- List Tests ---

func TestList_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("dir/file1.txt", "a"))
	require.NoError(t, m.Write("dir/file2.txt", "b"))
	require.NoError(t, m.Write("dir/sub/file3.txt", "c"))

	entries, err := m.List("dir")
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}

	assert.True(t, names["file1.txt"])
	assert.True(t, names["file2.txt"])
	assert.True(t, names["sub"])
	assert.Len(t, entries, 3)
}

func TestList_Good_Root(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("root.txt", "content"))
	require.NoError(t, m.Write("dir/nested.txt", "nested"))

	entries, err := m.List("")
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}

	assert.True(t, names["root.txt"])
	assert.True(t, names["dir"])
}

func TestList_Good_DirectoryEntry(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("dir/sub/file.txt", "content"))

	entries, err := m.List("dir")
	require.NoError(t, err)

	require.Len(t, entries, 1)
	assert.Equal(t, "sub", entries[0].Name())
	assert.True(t, entries[0].IsDir())

	info, err := entries[0].Info()
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// --- Stat Tests ---

func TestStat_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("file.txt", "hello world"))

	info, err := m.Stat("file.txt")
	require.NoError(t, err)
	assert.Equal(t, "file.txt", info.Name())
	assert.Equal(t, int64(11), info.Size())
	assert.False(t, info.IsDir())
}

func TestStat_Good_Directory(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.EnsureDir("mydir"))

	info, err := m.Stat("mydir")
	require.NoError(t, err)
	assert.Equal(t, "mydir", info.Name())
	assert.True(t, info.IsDir())
}

func TestStat_Bad_NotFound(t *testing.T) {
	m := newTestMedium(t)

	_, err := m.Stat("nonexistent")
	assert.Error(t, err)
}

func TestStat_Bad_EmptyPath(t *testing.T) {
	m := newTestMedium(t)

	_, err := m.Stat("")
	assert.Error(t, err)
}

// --- Open Tests ---

func TestOpen_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("file.txt", "open me"))

	f, err := m.Open("file.txt")
	require.NoError(t, err)
	defer f.Close()

	data, err := goio.ReadAll(f.(goio.Reader))
	require.NoError(t, err)
	assert.Equal(t, "open me", string(data))

	stat, err := f.Stat()
	require.NoError(t, err)
	assert.Equal(t, "file.txt", stat.Name())
}

func TestOpen_Bad_NotFound(t *testing.T) {
	m := newTestMedium(t)

	_, err := m.Open("nonexistent.txt")
	assert.Error(t, err)
}

func TestOpen_Bad_IsDirectory(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.EnsureDir("mydir"))
	_, err := m.Open("mydir")
	assert.Error(t, err)
}

// --- Create Tests ---

func TestCreate_Good(t *testing.T) {
	m := newTestMedium(t)

	w, err := m.Create("new.txt")
	require.NoError(t, err)

	n, err := w.Write([]byte("created"))
	require.NoError(t, err)
	assert.Equal(t, 7, n)

	err = w.Close()
	require.NoError(t, err)

	content, err := m.Read("new.txt")
	require.NoError(t, err)
	assert.Equal(t, "created", content)
}

func TestCreate_Good_Overwrite(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("file.txt", "old content"))

	w, err := m.Create("file.txt")
	require.NoError(t, err)
	_, err = w.Write([]byte("new"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	content, err := m.Read("file.txt")
	require.NoError(t, err)
	assert.Equal(t, "new", content)
}

func TestCreate_Bad_EmptyPath(t *testing.T) {
	m := newTestMedium(t)

	_, err := m.Create("")
	assert.Error(t, err)
}

// --- Append Tests ---

func TestAppend_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("append.txt", "hello"))

	w, err := m.Append("append.txt")
	require.NoError(t, err)

	_, err = w.Write([]byte(" world"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	content, err := m.Read("append.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello world", content)
}

func TestAppend_Good_NewFile(t *testing.T) {
	m := newTestMedium(t)

	w, err := m.Append("new.txt")
	require.NoError(t, err)

	_, err = w.Write([]byte("fresh"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	content, err := m.Read("new.txt")
	require.NoError(t, err)
	assert.Equal(t, "fresh", content)
}

func TestAppend_Bad_EmptyPath(t *testing.T) {
	m := newTestMedium(t)

	_, err := m.Append("")
	assert.Error(t, err)
}

// --- ReadStream Tests ---

func TestReadStream_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("stream.txt", "streaming content"))

	reader, err := m.ReadStream("stream.txt")
	require.NoError(t, err)
	defer reader.Close()

	data, err := goio.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "streaming content", string(data))
}

func TestReadStream_Bad_NotFound(t *testing.T) {
	m := newTestMedium(t)

	_, err := m.ReadStream("nonexistent.txt")
	assert.Error(t, err)
}

func TestReadStream_Bad_IsDirectory(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.EnsureDir("mydir"))
	_, err := m.ReadStream("mydir")
	assert.Error(t, err)
}

// --- WriteStream Tests ---

func TestWriteStream_Good(t *testing.T) {
	m := newTestMedium(t)

	writer, err := m.WriteStream("output.txt")
	require.NoError(t, err)

	_, err = goio.Copy(writer, strings.NewReader("piped data"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	content, err := m.Read("output.txt")
	require.NoError(t, err)
	assert.Equal(t, "piped data", content)
}

// --- Exists Tests ---

func TestExists_Good(t *testing.T) {
	m := newTestMedium(t)

	assert.False(t, m.Exists("nonexistent"))

	require.NoError(t, m.Write("file.txt", "content"))
	assert.True(t, m.Exists("file.txt"))

	require.NoError(t, m.EnsureDir("mydir"))
	assert.True(t, m.Exists("mydir"))
}

func TestExists_Good_EmptyPath(t *testing.T) {
	m := newTestMedium(t)
	// Root always exists
	assert.True(t, m.Exists(""))
}

// --- IsDir Tests ---

func TestIsDir_Good(t *testing.T) {
	m := newTestMedium(t)

	require.NoError(t, m.Write("file.txt", "content"))
	require.NoError(t, m.EnsureDir("mydir"))

	assert.True(t, m.IsDir("mydir"))
	assert.False(t, m.IsDir("file.txt"))
	assert.False(t, m.IsDir("nonexistent"))
	assert.False(t, m.IsDir(""))
}

// --- cleanPath Tests ---

func TestCleanPath_Good(t *testing.T) {
	assert.Equal(t, "file.txt", cleanPath("file.txt"))
	assert.Equal(t, "dir/file.txt", cleanPath("dir/file.txt"))
	assert.Equal(t, "file.txt", cleanPath("/file.txt"))
	assert.Equal(t, "file.txt", cleanPath("../file.txt"))
	assert.Equal(t, "file.txt", cleanPath("dir/../file.txt"))
	assert.Equal(t, "", cleanPath(""))
	assert.Equal(t, "", cleanPath("."))
	assert.Equal(t, "", cleanPath("/"))
}

// --- Interface Compliance ---

func TestInterfaceCompliance_Ugly(t *testing.T) {
	m := newTestMedium(t)

	// Verify all methods exist by asserting the interface shape.
	var _ interface {
		Read(string) (string, error)
		Write(string, string) error
		EnsureDir(string) error
		IsFile(string) bool
		FileGet(string) (string, error)
		FileSet(string, string) error
		Delete(string) error
		DeleteAll(string) error
		Rename(string, string) error
		List(string) ([]fs.DirEntry, error)
		Stat(string) (fs.FileInfo, error)
		Open(string) (fs.File, error)
		Create(string) (goio.WriteCloser, error)
		Append(string) (goio.WriteCloser, error)
		ReadStream(string) (goio.ReadCloser, error)
		WriteStream(string) (goio.WriteCloser, error)
		Exists(string) bool
		IsDir(string) bool
	} = m
}

// --- Custom Table ---

func TestCustomTable_Good(t *testing.T) {
	m, err := New(":memory:", WithTable("my_files"))
	require.NoError(t, err)
	defer m.Close()

	require.NoError(t, m.Write("file.txt", "content"))

	content, err := m.Read("file.txt")
	require.NoError(t, err)
	assert.Equal(t, "content", content)
}
