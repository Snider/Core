package local

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	root := t.TempDir()
	m, err := New(root)
	assert.NoError(t, err)
	assert.Equal(t, root, m.root)
}

func TestPath(t *testing.T) {
	m := &Medium{root: "/home/user"}

	// Normal paths
	assert.Equal(t, "/home/user/file.txt", m.path("file.txt"))
	assert.Equal(t, "/home/user/dir/file.txt", m.path("dir/file.txt"))

	// Empty returns root
	assert.Equal(t, "/home/user", m.path(""))

	// Traversal attempts get sanitized
	assert.Equal(t, "/home/user/file.txt", m.path("../file.txt"))
	assert.Equal(t, "/home/user/file.txt", m.path("dir/../file.txt"))

	// Absolute paths are constrained to sandbox (no escape)
	assert.Equal(t, "/home/user/etc/passwd", m.path("/etc/passwd"))
}

func TestPath_RootFilesystem(t *testing.T) {
	m := &Medium{root: "/"}

	// When root is "/", absolute paths pass through
	assert.Equal(t, "/etc/passwd", m.path("/etc/passwd"))
	assert.Equal(t, "/home/user/file.txt", m.path("/home/user/file.txt"))

	// Relative paths are relative to CWD when root is "/"
	cwd, _ := os.Getwd()
	assert.Equal(t, filepath.Join(cwd, "file.txt"), m.path("file.txt"))
}

func TestReadWrite(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	// Write and read back
	err := m.Write("test.txt", "hello")
	assert.NoError(t, err)

	content, err := m.Read("test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello", content)

	// Write creates parent dirs
	err = m.Write("a/b/c.txt", "nested")
	assert.NoError(t, err)

	content, err = m.Read("a/b/c.txt")
	assert.NoError(t, err)
	assert.Equal(t, "nested", content)

	// Read nonexistent
	_, err = m.Read("nope.txt")
	assert.Error(t, err)
}

func TestEnsureDir(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	err := m.EnsureDir("one/two/three")
	assert.NoError(t, err)

	info, err := os.Stat(filepath.Join(root, "one/two/three"))
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestIsDir(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	_ = os.Mkdir(filepath.Join(root, "mydir"), 0755)
	_ = os.WriteFile(filepath.Join(root, "myfile"), []byte("x"), 0644)

	assert.True(t, m.IsDir("mydir"))
	assert.False(t, m.IsDir("myfile"))
	assert.False(t, m.IsDir("nope"))
	assert.False(t, m.IsDir(""))
}

func TestIsFile(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	_ = os.Mkdir(filepath.Join(root, "mydir"), 0755)
	_ = os.WriteFile(filepath.Join(root, "myfile"), []byte("x"), 0644)

	assert.True(t, m.IsFile("myfile"))
	assert.False(t, m.IsFile("mydir"))
	assert.False(t, m.IsFile("nope"))
	assert.False(t, m.IsFile(""))
}

func TestExists(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	_ = os.WriteFile(filepath.Join(root, "exists"), []byte("x"), 0644)

	assert.True(t, m.Exists("exists"))
	assert.False(t, m.Exists("nope"))
}

func TestList(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	_ = os.WriteFile(filepath.Join(root, "a.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(root, "b.txt"), []byte("b"), 0644)
	_ = os.Mkdir(filepath.Join(root, "subdir"), 0755)

	entries, err := m.List("")
	assert.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestStat(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	_ = os.WriteFile(filepath.Join(root, "file"), []byte("content"), 0644)

	info, err := m.Stat("file")
	assert.NoError(t, err)
	assert.Equal(t, int64(7), info.Size())
}

func TestDelete(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	_ = os.WriteFile(filepath.Join(root, "todelete"), []byte("x"), 0644)
	assert.True(t, m.Exists("todelete"))

	err := m.Delete("todelete")
	assert.NoError(t, err)
	assert.False(t, m.Exists("todelete"))
}

func TestDeleteAll(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	_ = os.MkdirAll(filepath.Join(root, "dir/sub"), 0755)
	_ = os.WriteFile(filepath.Join(root, "dir/sub/file"), []byte("x"), 0644)

	err := m.DeleteAll("dir")
	assert.NoError(t, err)
	assert.False(t, m.Exists("dir"))
}

func TestRename(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	_ = os.WriteFile(filepath.Join(root, "old"), []byte("x"), 0644)

	err := m.Rename("old", "new")
	assert.NoError(t, err)
	assert.False(t, m.Exists("old"))
	assert.True(t, m.Exists("new"))
}

func TestFileGetFileSet(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	err := m.FileSet("data", "value")
	assert.NoError(t, err)

	val, err := m.FileGet("data")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestDelete_Good(t *testing.T) {
	testRoot, err := os.MkdirTemp("", "local_delete_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testRoot) }()

	medium, err := New(testRoot)
	assert.NoError(t, err)

	// Create and delete a file
	err = medium.Write("file.txt", "content")
	assert.NoError(t, err)
	assert.True(t, medium.IsFile("file.txt"))

	err = medium.Delete("file.txt")
	assert.NoError(t, err)
	assert.False(t, medium.IsFile("file.txt"))

	// Create and delete an empty directory
	err = medium.EnsureDir("emptydir")
	assert.NoError(t, err)
	err = medium.Delete("emptydir")
	assert.NoError(t, err)
	assert.False(t, medium.IsDir("emptydir"))
}

func TestDelete_Bad_NotEmpty(t *testing.T) {
	testRoot, err := os.MkdirTemp("", "local_delete_notempty_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testRoot) }()

	medium, err := New(testRoot)
	assert.NoError(t, err)

	// Create a directory with a file
	err = medium.Write("mydir/file.txt", "content")
	assert.NoError(t, err)

	// Try to delete non-empty directory
	err = medium.Delete("mydir")
	assert.Error(t, err)
}

func TestDeleteAll_Good(t *testing.T) {
	testRoot, err := os.MkdirTemp("", "local_deleteall_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testRoot) }()

	medium, err := New(testRoot)
	assert.NoError(t, err)

	// Create nested structure
	err = medium.Write("mydir/file1.txt", "content1")
	assert.NoError(t, err)
	err = medium.Write("mydir/subdir/file2.txt", "content2")
	assert.NoError(t, err)

	// Delete all
	err = medium.DeleteAll("mydir")
	assert.NoError(t, err)
	assert.False(t, medium.Exists("mydir"))
	assert.False(t, medium.Exists("mydir/file1.txt"))
	assert.False(t, medium.Exists("mydir/subdir/file2.txt"))
}

func TestRename_Good(t *testing.T) {
	testRoot, err := os.MkdirTemp("", "local_rename_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testRoot) }()

	medium, err := New(testRoot)
	assert.NoError(t, err)

	// Rename a file
	err = medium.Write("old.txt", "content")
	assert.NoError(t, err)
	err = medium.Rename("old.txt", "new.txt")
	assert.NoError(t, err)
	assert.False(t, medium.IsFile("old.txt"))
	assert.True(t, medium.IsFile("new.txt"))

	content, err := medium.Read("new.txt")
	assert.NoError(t, err)
	assert.Equal(t, "content", content)
}

func TestRename_Traversal_Sanitized(t *testing.T) {
	testRoot, err := os.MkdirTemp("", "local_rename_traversal_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testRoot) }()

	medium, err := New(testRoot)
	assert.NoError(t, err)

	err = medium.Write("file.txt", "content")
	assert.NoError(t, err)

	// Traversal attempts are sanitized (.. becomes .), so this renames to "./escaped.txt"
	// which is just "escaped.txt" in the root
	err = medium.Rename("file.txt", "../escaped.txt")
	assert.NoError(t, err)
	assert.False(t, medium.Exists("file.txt"))
	assert.True(t, medium.Exists("escaped.txt"))
}

func TestList_Good(t *testing.T) {
	testRoot, err := os.MkdirTemp("", "local_list_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testRoot) }()

	medium, err := New(testRoot)
	assert.NoError(t, err)

	// Create some files and directories
	err = medium.Write("file1.txt", "content1")
	assert.NoError(t, err)
	err = medium.Write("file2.txt", "content2")
	assert.NoError(t, err)
	err = medium.EnsureDir("subdir")
	assert.NoError(t, err)

	// List root
	entries, err := medium.List(".")
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

func TestStat_Good(t *testing.T) {
	testRoot, err := os.MkdirTemp("", "local_stat_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testRoot) }()

	medium, err := New(testRoot)
	assert.NoError(t, err)

	// Stat a file
	err = medium.Write("file.txt", "hello world")
	assert.NoError(t, err)
	info, err := medium.Stat("file.txt")
	assert.NoError(t, err)
	assert.Equal(t, "file.txt", info.Name())
	assert.Equal(t, int64(11), info.Size())
	assert.False(t, info.IsDir())

	// Stat a directory
	err = medium.EnsureDir("mydir")
	assert.NoError(t, err)
	info, err = medium.Stat("mydir")
	assert.NoError(t, err)
	assert.Equal(t, "mydir", info.Name())
	assert.True(t, info.IsDir())
}

func TestExists_Good(t *testing.T) {
	testRoot, err := os.MkdirTemp("", "local_exists_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testRoot) }()

	medium, err := New(testRoot)
	assert.NoError(t, err)

	assert.False(t, medium.Exists("nonexistent"))

	err = medium.Write("file.txt", "content")
	assert.NoError(t, err)
	assert.True(t, medium.Exists("file.txt"))

	err = medium.EnsureDir("mydir")
	assert.NoError(t, err)
	assert.True(t, medium.Exists("mydir"))
}

func TestIsDir_Good(t *testing.T) {
	testRoot, err := os.MkdirTemp("", "local_isdir_test")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testRoot) }()

	medium, err := New(testRoot)
	assert.NoError(t, err)

	err = medium.Write("file.txt", "content")
	assert.NoError(t, err)
	assert.False(t, medium.IsDir("file.txt"))

	err = medium.EnsureDir("mydir")
	assert.NoError(t, err)
	assert.True(t, medium.IsDir("mydir"))

	assert.False(t, medium.IsDir("nonexistent"))
}

func TestReadStream(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	content := "streaming content"
	err := m.Write("stream.txt", content)
	assert.NoError(t, err)

	reader, err := m.ReadStream("stream.txt")
	assert.NoError(t, err)
	defer reader.Close()

	// Read only first 9 bytes
	limitReader := io.LimitReader(reader, 9)
	data, err := io.ReadAll(limitReader)
	assert.NoError(t, err)
	assert.Equal(t, "streaming", string(data))
}

func TestWriteStream(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	writer, err := m.WriteStream("output.txt")
	assert.NoError(t, err)

	_, err = io.Copy(writer, strings.NewReader("piped data"))
	assert.NoError(t, err)
	err = writer.Close()
	assert.NoError(t, err)

	content, err := m.Read("output.txt")
	assert.NoError(t, err)
	assert.Equal(t, "piped data", content)
}
