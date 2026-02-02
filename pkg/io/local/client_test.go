package local

import (
	"os"
	"path/filepath"
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

	// Traversal attempts get sanitized (.. becomes ., then cleaned by Join)
	assert.Equal(t, "/home/user/file.txt", m.path("../file.txt"))
	assert.Equal(t, "/home/user/dir/file.txt", m.path("dir/../file.txt"))

	// Absolute paths pass through
	assert.Equal(t, "/etc/passwd", m.path("/etc/passwd"))
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

	os.Mkdir(filepath.Join(root, "mydir"), 0755)
	os.WriteFile(filepath.Join(root, "myfile"), []byte("x"), 0644)

	assert.True(t, m.IsDir("mydir"))
	assert.False(t, m.IsDir("myfile"))
	assert.False(t, m.IsDir("nope"))
	assert.False(t, m.IsDir(""))
}

func TestIsFile(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	os.Mkdir(filepath.Join(root, "mydir"), 0755)
	os.WriteFile(filepath.Join(root, "myfile"), []byte("x"), 0644)

	assert.True(t, m.IsFile("myfile"))
	assert.False(t, m.IsFile("mydir"))
	assert.False(t, m.IsFile("nope"))
	assert.False(t, m.IsFile(""))
}

func TestExists(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	os.WriteFile(filepath.Join(root, "exists"), []byte("x"), 0644)

	assert.True(t, m.Exists("exists"))
	assert.False(t, m.Exists("nope"))
}

func TestList(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	os.WriteFile(filepath.Join(root, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(root, "b.txt"), []byte("b"), 0644)
	os.Mkdir(filepath.Join(root, "subdir"), 0755)

	entries, err := m.List("")
	assert.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestStat(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	os.WriteFile(filepath.Join(root, "file"), []byte("content"), 0644)

	info, err := m.Stat("file")
	assert.NoError(t, err)
	assert.Equal(t, int64(7), info.Size())
}

func TestDelete(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	os.WriteFile(filepath.Join(root, "todelete"), []byte("x"), 0644)
	assert.True(t, m.Exists("todelete"))

	err := m.Delete("todelete")
	assert.NoError(t, err)
	assert.False(t, m.Exists("todelete"))
}

func TestDeleteAll(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	os.MkdirAll(filepath.Join(root, "dir/sub"), 0755)
	os.WriteFile(filepath.Join(root, "dir/sub/file"), []byte("x"), 0644)

	err := m.DeleteAll("dir")
	assert.NoError(t, err)
	assert.False(t, m.Exists("dir"))
}

func TestRename(t *testing.T) {
	root := t.TempDir()
	m, _ := New(root)

	os.WriteFile(filepath.Join(root, "old"), []byte("x"), 0644)

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
