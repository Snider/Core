package node

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_Good(t *testing.T) {
	n := New()
	require.NotNil(t, n, "New() must not return nil")
	assert.NotNil(t, n.files, "New() must initialize the files map")
}

// ---------------------------------------------------------------------------
// AddData
// ---------------------------------------------------------------------------

func TestAddData_Good(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))

	file, ok := n.files["foo.txt"]
	require.True(t, ok, "file foo.txt should be present")
	assert.Equal(t, []byte("foo"), file.content)

	info, err := file.Stat()
	require.NoError(t, err)
	assert.Equal(t, "foo.txt", info.Name())
}

func TestAddData_Bad(t *testing.T) {
	n := New()

	// Empty name is silently ignored.
	n.AddData("", []byte("data"))
	assert.Empty(t, n.files, "empty name must not be stored")

	// Directory entry (trailing slash) is silently ignored.
	n.AddData("dir/", nil)
	assert.Empty(t, n.files, "directory entry must not be stored")
}

func TestAddData_Ugly(t *testing.T) {
	t.Run("Overwrite", func(t *testing.T) {
		n := New()
		n.AddData("foo.txt", []byte("foo"))
		n.AddData("foo.txt", []byte("bar"))

		file := n.files["foo.txt"]
		assert.Equal(t, []byte("bar"), file.content, "second AddData should overwrite")
	})

	t.Run("LeadingSlash", func(t *testing.T) {
		n := New()
		n.AddData("/hello.txt", []byte("hi"))
		_, ok := n.files["hello.txt"]
		assert.True(t, ok, "leading slash should be trimmed")
	})
}

// ---------------------------------------------------------------------------
// Open
// ---------------------------------------------------------------------------

func TestOpen_Good(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))

	file, err := n.Open("foo.txt")
	require.NoError(t, err)
	defer file.Close()

	buf := make([]byte, 10)
	nr, err := file.Read(buf)
	require.True(t, nr > 0 || err == io.EOF)
	assert.Equal(t, "foo", string(buf[:nr]))
}

func TestOpen_Bad(t *testing.T) {
	n := New()
	_, err := n.Open("nonexistent.txt")
	require.Error(t, err)
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestOpen_Ugly(t *testing.T) {
	n := New()
	n.AddData("bar/baz.txt", []byte("baz"))

	// Opening a directory should succeed.
	file, err := n.Open("bar")
	require.NoError(t, err)
	defer file.Close()

	// Reading from a directory should fail.
	_, err = file.Read(make([]byte, 1))
	require.Error(t, err)

	var pathErr *fs.PathError
	require.True(t, errors.As(err, &pathErr))
	assert.Equal(t, fs.ErrInvalid, pathErr.Err)
}

// ---------------------------------------------------------------------------
// Stat
// ---------------------------------------------------------------------------

func TestStat_Good(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))
	n.AddData("bar/baz.txt", []byte("baz"))

	// File stat.
	info, err := n.Stat("bar/baz.txt")
	require.NoError(t, err)
	assert.Equal(t, "baz.txt", info.Name())
	assert.Equal(t, int64(3), info.Size())
	assert.False(t, info.IsDir())

	// Directory stat.
	dirInfo, err := n.Stat("bar")
	require.NoError(t, err)
	assert.True(t, dirInfo.IsDir())
	assert.Equal(t, "bar", dirInfo.Name())
}

func TestStat_Bad(t *testing.T) {
	n := New()
	_, err := n.Stat("nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestStat_Ugly(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))

	// Root directory.
	info, err := n.Stat(".")
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, ".", info.Name())
}

// ---------------------------------------------------------------------------
// ReadFile
// ---------------------------------------------------------------------------

func TestReadFile_Good(t *testing.T) {
	n := New()
	n.AddData("hello.txt", []byte("hello world"))

	data, err := n.ReadFile("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello world"), data)
}

func TestReadFile_Bad(t *testing.T) {
	n := New()
	_, err := n.ReadFile("missing.txt")
	require.Error(t, err)
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestReadFile_Ugly(t *testing.T) {
	n := New()
	n.AddData("data.bin", []byte("original"))

	// Returned slice must be a copy — mutating it must not affect internal state.
	data, err := n.ReadFile("data.bin")
	require.NoError(t, err)
	data[0] = 'X'

	data2, err := n.ReadFile("data.bin")
	require.NoError(t, err)
	assert.Equal(t, []byte("original"), data2, "ReadFile must return an independent copy")
}

// ---------------------------------------------------------------------------
// ReadDir
// ---------------------------------------------------------------------------

func TestReadDir_Good(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))
	n.AddData("bar/baz.txt", []byte("baz"))
	n.AddData("bar/qux.txt", []byte("qux"))

	// Root.
	entries, err := n.ReadDir(".")
	require.NoError(t, err)
	assert.Equal(t, []string{"bar", "foo.txt"}, sortedNames(entries))

	// Subdirectory.
	barEntries, err := n.ReadDir("bar")
	require.NoError(t, err)
	assert.Equal(t, []string{"baz.txt", "qux.txt"}, sortedNames(barEntries))
}

func TestReadDir_Bad(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))

	// Reading a file as a directory should fail.
	_, err := n.ReadDir("foo.txt")
	require.Error(t, err)
	var pathErr *fs.PathError
	require.True(t, errors.As(err, &pathErr))
	assert.Equal(t, fs.ErrInvalid, pathErr.Err)
}

func TestReadDir_Ugly(t *testing.T) {
	n := New()
	n.AddData("bar/baz.txt", []byte("baz"))
	n.AddData("empty_dir/", nil) // Ignored by AddData.

	entries, err := n.ReadDir(".")
	require.NoError(t, err)
	assert.Equal(t, []string{"bar"}, sortedNames(entries))
}

// ---------------------------------------------------------------------------
// Exists
// ---------------------------------------------------------------------------

func TestExists_Good(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))
	n.AddData("bar/baz.txt", []byte("baz"))

	exists, err := n.Exists("foo.txt")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = n.Exists("bar")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExists_Bad(t *testing.T) {
	n := New()
	exists, err := n.Exists("nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestExists_Ugly(t *testing.T) {
	n := New()
	n.AddData("dummy.txt", []byte("dummy"))

	exists, err := n.Exists(".")
	require.NoError(t, err)
	assert.True(t, exists, "root '.' must exist")

	exists, err = n.Exists("")
	require.NoError(t, err)
	assert.True(t, exists, "empty path (root) must exist")
}

// ---------------------------------------------------------------------------
// Walk
// ---------------------------------------------------------------------------

func TestWalk_Good(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))
	n.AddData("bar/baz.txt", []byte("baz"))
	n.AddData("bar/qux.txt", []byte("qux"))

	var paths []string
	err := n.Walk(".", func(p string, d fs.DirEntry, err error) error {
		paths = append(paths, p)
		return nil
	})
	require.NoError(t, err)

	sort.Strings(paths)
	assert.Equal(t, []string{".", "bar", "bar/baz.txt", "bar/qux.txt", "foo.txt"}, paths)
}

func TestWalk_Bad(t *testing.T) {
	n := New()

	var called bool
	err := n.Walk("nonexistent", func(p string, d fs.DirEntry, err error) error {
		called = true
		assert.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
		return err
	})
	assert.True(t, called, "walk function must be called for nonexistent root")
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestWalk_Ugly(t *testing.T) {
	n := New()
	n.AddData("a/b.txt", []byte("b"))
	n.AddData("a/c.txt", []byte("c"))

	// Stop walk early with a custom error.
	walkErr := errors.New("stop walking")
	var paths []string
	err := n.Walk(".", func(p string, d fs.DirEntry, err error) error {
		if p == "a/b.txt" {
			return walkErr
		}
		paths = append(paths, p)
		return nil
	})

	assert.Equal(t, walkErr, err, "Walk must propagate the callback error")
}

func TestWalk_Options(t *testing.T) {
	n := New()
	n.AddData("root.txt", []byte("root"))
	n.AddData("a/a1.txt", []byte("a1"))
	n.AddData("a/b/b1.txt", []byte("b1"))
	n.AddData("c/c1.txt", []byte("c1"))

	t.Run("MaxDepth", func(t *testing.T) {
		var paths []string
		err := n.Walk(".", func(p string, d fs.DirEntry, err error) error {
			paths = append(paths, p)
			return nil
		}, WalkOptions{MaxDepth: 1})
		require.NoError(t, err)

		sort.Strings(paths)
		assert.Equal(t, []string{".", "a", "c", "root.txt"}, paths)
	})

	t.Run("Filter", func(t *testing.T) {
		var paths []string
		err := n.Walk(".", func(p string, d fs.DirEntry, err error) error {
			paths = append(paths, p)
			return nil
		}, WalkOptions{Filter: func(p string, d fs.DirEntry) bool {
			return !strings.HasPrefix(p, "a")
		}})
		require.NoError(t, err)

		sort.Strings(paths)
		assert.Equal(t, []string{".", "c", "c/c1.txt", "root.txt"}, paths)
	})

	t.Run("SkipErrors", func(t *testing.T) {
		var called bool
		err := n.Walk("nonexistent", func(p string, d fs.DirEntry, err error) error {
			called = true
			return err
		}, WalkOptions{SkipErrors: true})

		assert.NoError(t, err, "SkipErrors should suppress the error")
		assert.False(t, called, "callback should not be called when error is skipped")
	})
}

// ---------------------------------------------------------------------------
// CopyFile
// ---------------------------------------------------------------------------

func TestCopyFile_Good(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))

	tmpfile := filepath.Join(t.TempDir(), "test.txt")
	err := n.CopyFile("foo.txt", tmpfile, 0644)
	require.NoError(t, err)

	content, err := os.ReadFile(tmpfile)
	require.NoError(t, err)
	assert.Equal(t, "foo", string(content))
}

func TestCopyFile_Bad(t *testing.T) {
	n := New()
	tmpfile := filepath.Join(t.TempDir(), "test.txt")

	// Source does not exist.
	err := n.CopyFile("nonexistent.txt", tmpfile, 0644)
	assert.Error(t, err)

	// Destination not writable.
	n.AddData("foo.txt", []byte("foo"))
	err = n.CopyFile("foo.txt", "/nonexistent_dir/test.txt", 0644)
	assert.Error(t, err)
}

func TestCopyFile_Ugly(t *testing.T) {
	n := New()
	n.AddData("bar/baz.txt", []byte("baz"))
	tmpfile := filepath.Join(t.TempDir(), "test.txt")

	// Attempting to copy a directory should fail.
	err := n.CopyFile("bar", tmpfile, 0644)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ToTar / FromTar
// ---------------------------------------------------------------------------

func TestToTar_Good(t *testing.T) {
	n := New()
	n.AddData("foo.txt", []byte("foo"))
	n.AddData("bar/baz.txt", []byte("baz"))

	tarball, err := n.ToTar()
	require.NoError(t, err)
	require.NotEmpty(t, tarball)

	// Verify tar content.
	tr := tar.NewReader(bytes.NewReader(tarball))
	files := make(map[string]string)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		content, err := io.ReadAll(tr)
		require.NoError(t, err)
		files[header.Name] = string(content)
	}

	assert.Equal(t, "foo", files["foo.txt"])
	assert.Equal(t, "baz", files["bar/baz.txt"])
}

func TestFromTar_Good(t *testing.T) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, f := range []struct{ Name, Body string }{
		{"foo.txt", "foo"},
		{"bar/baz.txt", "baz"},
	} {
		hdr := &tar.Header{
			Name:     f.Name,
			Mode:     0600,
			Size:     int64(len(f.Body)),
			Typeflag: tar.TypeReg,
		}
		require.NoError(t, tw.WriteHeader(hdr))
		_, err := tw.Write([]byte(f.Body))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())

	n, err := FromTar(buf.Bytes())
	require.NoError(t, err)

	exists, _ := n.Exists("foo.txt")
	assert.True(t, exists, "foo.txt should exist")

	exists, _ = n.Exists("bar/baz.txt")
	assert.True(t, exists, "bar/baz.txt should exist")
}

func TestFromTar_Bad(t *testing.T) {
	// Truncated data that cannot be a valid tar.
	truncated := make([]byte, 100)
	_, err := FromTar(truncated)
	assert.Error(t, err, "truncated data should produce an error")
}

func TestTarRoundTrip_Good(t *testing.T) {
	n1 := New()
	n1.AddData("a.txt", []byte("alpha"))
	n1.AddData("b/c.txt", []byte("charlie"))

	tarball, err := n1.ToTar()
	require.NoError(t, err)

	n2, err := FromTar(tarball)
	require.NoError(t, err)

	// Verify n2 matches n1.
	data, err := n2.ReadFile("a.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("alpha"), data)

	data, err = n2.ReadFile("b/c.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("charlie"), data)
}

// ---------------------------------------------------------------------------
// fs.FS interface compliance
// ---------------------------------------------------------------------------

func TestFSInterface_Good(t *testing.T) {
	n := New()
	n.AddData("hello.txt", []byte("world"))

	// fs.FS
	var fsys fs.FS = n
	file, err := fsys.Open("hello.txt")
	require.NoError(t, err)
	defer file.Close()

	// fs.StatFS
	var statFS fs.StatFS = n
	info, err := statFS.Stat("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello.txt", info.Name())
	assert.Equal(t, int64(5), info.Size())

	// fs.ReadFileFS
	var readFS fs.ReadFileFS = n
	data, err := readFS.ReadFile("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("world"), data)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func sortedNames(entries []fs.DirEntry) []string {
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}
