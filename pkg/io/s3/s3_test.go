package s3

import (
	"bytes"
	"context"
	"fmt"
	goio "io"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockS3 is an in-memory mock implementing the s3API interface.
type mockS3 struct {
	mu      sync.RWMutex
	objects map[string][]byte
	mtimes  map[string]time.Time
}

func newMockS3() *mockS3 {
	return &mockS3{
		objects: make(map[string][]byte),
		mtimes:  make(map[string]time.Time),
	}
}

func (m *mockS3) GetObject(_ context.Context, params *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := aws.ToString(params.Key)
	data, ok := m.objects[key]
	if !ok {
		return nil, fmt.Errorf("NoSuchKey: key %q not found", key)
	}
	mtime := m.mtimes[key]
	return &s3.GetObjectOutput{
		Body:          goio.NopCloser(bytes.NewReader(data)),
		ContentLength: aws.Int64(int64(len(data))),
		LastModified:  &mtime,
	}, nil
}

func (m *mockS3) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := aws.ToString(params.Key)
	data, err := goio.ReadAll(params.Body)
	if err != nil {
		return nil, err
	}
	m.objects[key] = data
	m.mtimes[key] = time.Now()
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3) DeleteObject(_ context.Context, params *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := aws.ToString(params.Key)
	delete(m.objects, key)
	delete(m.mtimes, key)
	return &s3.DeleteObjectOutput{}, nil
}

func (m *mockS3) DeleteObjects(_ context.Context, params *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, obj := range params.Delete.Objects {
		key := aws.ToString(obj.Key)
		delete(m.objects, key)
		delete(m.mtimes, key)
	}
	return &s3.DeleteObjectsOutput{}, nil
}

func (m *mockS3) HeadObject(_ context.Context, params *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := aws.ToString(params.Key)
	data, ok := m.objects[key]
	if !ok {
		return nil, fmt.Errorf("NotFound: key %q not found", key)
	}
	mtime := m.mtimes[key]
	return &s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(len(data))),
		LastModified:  &mtime,
	}, nil
}

func (m *mockS3) ListObjectsV2(_ context.Context, params *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prefix := aws.ToString(params.Prefix)
	delimiter := aws.ToString(params.Delimiter)
	maxKeys := int32(1000)
	if params.MaxKeys != nil {
		maxKeys = *params.MaxKeys
	}

	// Collect all matching keys sorted
	var allKeys []string
	for k := range m.objects {
		if strings.HasPrefix(k, prefix) {
			allKeys = append(allKeys, k)
		}
	}
	sort.Strings(allKeys)

	var contents []types.Object
	commonPrefixes := make(map[string]bool)

	for _, k := range allKeys {
		rest := strings.TrimPrefix(k, prefix)

		if delimiter != "" {
			if idx := strings.Index(rest, delimiter); idx >= 0 {
				// This key has a delimiter after the prefix -> common prefix
				cp := prefix + rest[:idx+len(delimiter)]
				commonPrefixes[cp] = true
				continue
			}
		}

		if int32(len(contents)) >= maxKeys {
			break
		}

		data := m.objects[k]
		mtime := m.mtimes[k]
		contents = append(contents, types.Object{
			Key:          aws.String(k),
			Size:         aws.Int64(int64(len(data))),
			LastModified: &mtime,
		})
	}

	var cpSlice []types.CommonPrefix
	// Sort common prefixes for deterministic output
	var cpKeys []string
	for cp := range commonPrefixes {
		cpKeys = append(cpKeys, cp)
	}
	sort.Strings(cpKeys)
	for _, cp := range cpKeys {
		cpSlice = append(cpSlice, types.CommonPrefix{Prefix: aws.String(cp)})
	}

	return &s3.ListObjectsV2Output{
		Contents:       contents,
		CommonPrefixes: cpSlice,
		IsTruncated:    aws.Bool(false),
	}, nil
}

func (m *mockS3) CopyObject(_ context.Context, params *s3.CopyObjectInput, _ ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// CopySource is "bucket/key"
	source := aws.ToString(params.CopySource)
	parts := strings.SplitN(source, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid CopySource: %s", source)
	}
	srcKey := parts[1]

	data, ok := m.objects[srcKey]
	if !ok {
		return nil, fmt.Errorf("NoSuchKey: source key %q not found", srcKey)
	}

	destKey := aws.ToString(params.Key)
	m.objects[destKey] = append([]byte{}, data...)
	m.mtimes[destKey] = time.Now()

	return &s3.CopyObjectOutput{}, nil
}

// --- Helper ---

func newTestMedium(t *testing.T) (*Medium, *mockS3) {
	t.Helper()
	mock := newMockS3()
	m, err := New("test-bucket", withAPI(mock))
	require.NoError(t, err)
	return m, mock
}

// --- Tests ---

func TestNew_Good(t *testing.T) {
	mock := newMockS3()
	m, err := New("my-bucket", withAPI(mock))
	require.NoError(t, err)
	assert.Equal(t, "my-bucket", m.bucket)
	assert.Equal(t, "", m.prefix)
}

func TestNew_Bad_NoBucket(t *testing.T) {
	_, err := New("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket name is required")
}

func TestNew_Bad_NoClient(t *testing.T) {
	_, err := New("bucket")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "S3 client is required")
}

func TestWithPrefix_Good(t *testing.T) {
	mock := newMockS3()
	m, err := New("bucket", withAPI(mock), WithPrefix("data/"))
	require.NoError(t, err)
	assert.Equal(t, "data/", m.prefix)

	// Prefix without trailing slash gets one added
	m2, err := New("bucket", withAPI(mock), WithPrefix("data"))
	require.NoError(t, err)
	assert.Equal(t, "data/", m2.prefix)
}

func TestReadWrite_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	err := m.Write("hello.txt", "world")
	require.NoError(t, err)

	content, err := m.Read("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "world", content)
}

func TestReadWrite_Bad_NotFound(t *testing.T) {
	m, _ := newTestMedium(t)

	_, err := m.Read("nonexistent.txt")
	assert.Error(t, err)
}

func TestReadWrite_Bad_EmptyPath(t *testing.T) {
	m, _ := newTestMedium(t)

	_, err := m.Read("")
	assert.Error(t, err)

	err = m.Write("", "content")
	assert.Error(t, err)
}

func TestReadWrite_Good_WithPrefix(t *testing.T) {
	mock := newMockS3()
	m, err := New("bucket", withAPI(mock), WithPrefix("pfx"))
	require.NoError(t, err)

	err = m.Write("file.txt", "data")
	require.NoError(t, err)

	// Verify the key has the prefix
	_, ok := mock.objects["pfx/file.txt"]
	assert.True(t, ok, "object should be stored with prefix")

	content, err := m.Read("file.txt")
	require.NoError(t, err)
	assert.Equal(t, "data", content)
}

func TestEnsureDir_Good(t *testing.T) {
	m, _ := newTestMedium(t)
	// EnsureDir is a no-op for S3
	err := m.EnsureDir("any/path")
	assert.NoError(t, err)
}

func TestIsFile_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	err := m.Write("file.txt", "content")
	require.NoError(t, err)

	assert.True(t, m.IsFile("file.txt"))
	assert.False(t, m.IsFile("nonexistent.txt"))
	assert.False(t, m.IsFile(""))
}

func TestFileGetFileSet_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	err := m.FileSet("key.txt", "value")
	require.NoError(t, err)

	val, err := m.FileGet("key.txt")
	require.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestDelete_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	err := m.Write("to-delete.txt", "content")
	require.NoError(t, err)
	assert.True(t, m.Exists("to-delete.txt"))

	err = m.Delete("to-delete.txt")
	require.NoError(t, err)
	assert.False(t, m.IsFile("to-delete.txt"))
}

func TestDelete_Bad_EmptyPath(t *testing.T) {
	m, _ := newTestMedium(t)
	err := m.Delete("")
	assert.Error(t, err)
}

func TestDeleteAll_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	// Create nested structure
	require.NoError(t, m.Write("dir/file1.txt", "a"))
	require.NoError(t, m.Write("dir/sub/file2.txt", "b"))
	require.NoError(t, m.Write("other.txt", "c"))

	err := m.DeleteAll("dir")
	require.NoError(t, err)

	assert.False(t, m.IsFile("dir/file1.txt"))
	assert.False(t, m.IsFile("dir/sub/file2.txt"))
	assert.True(t, m.IsFile("other.txt"))
}

func TestDeleteAll_Bad_EmptyPath(t *testing.T) {
	m, _ := newTestMedium(t)
	err := m.DeleteAll("")
	assert.Error(t, err)
}

func TestRename_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	require.NoError(t, m.Write("old.txt", "content"))
	assert.True(t, m.IsFile("old.txt"))

	err := m.Rename("old.txt", "new.txt")
	require.NoError(t, err)

	assert.False(t, m.IsFile("old.txt"))
	assert.True(t, m.IsFile("new.txt"))

	content, err := m.Read("new.txt")
	require.NoError(t, err)
	assert.Equal(t, "content", content)
}

func TestRename_Bad_EmptyPath(t *testing.T) {
	m, _ := newTestMedium(t)
	err := m.Rename("", "new.txt")
	assert.Error(t, err)

	err = m.Rename("old.txt", "")
	assert.Error(t, err)
}

func TestRename_Bad_SourceNotFound(t *testing.T) {
	m, _ := newTestMedium(t)
	err := m.Rename("nonexistent.txt", "new.txt")
	assert.Error(t, err)
}

func TestList_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	require.NoError(t, m.Write("dir/file1.txt", "a"))
	require.NoError(t, m.Write("dir/file2.txt", "b"))
	require.NoError(t, m.Write("dir/sub/file3.txt", "c"))

	entries, err := m.List("dir")
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}

	assert.True(t, names["file1.txt"], "should list file1.txt")
	assert.True(t, names["file2.txt"], "should list file2.txt")
	assert.True(t, names["sub"], "should list sub directory")
	assert.Len(t, entries, 3)

	// Check that sub is a directory
	for _, e := range entries {
		if e.Name() == "sub" {
			assert.True(t, e.IsDir())
			info, err := e.Info()
			require.NoError(t, err)
			assert.True(t, info.IsDir())
		}
	}
}

func TestList_Good_Root(t *testing.T) {
	m, _ := newTestMedium(t)

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

func TestStat_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	require.NoError(t, m.Write("file.txt", "hello world"))

	info, err := m.Stat("file.txt")
	require.NoError(t, err)
	assert.Equal(t, "file.txt", info.Name())
	assert.Equal(t, int64(11), info.Size())
	assert.False(t, info.IsDir())
}

func TestStat_Bad_NotFound(t *testing.T) {
	m, _ := newTestMedium(t)

	_, err := m.Stat("nonexistent.txt")
	assert.Error(t, err)
}

func TestStat_Bad_EmptyPath(t *testing.T) {
	m, _ := newTestMedium(t)
	_, err := m.Stat("")
	assert.Error(t, err)
}

func TestOpen_Good(t *testing.T) {
	m, _ := newTestMedium(t)

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
	m, _ := newTestMedium(t)

	_, err := m.Open("nonexistent.txt")
	assert.Error(t, err)
}

func TestCreate_Good(t *testing.T) {
	m, _ := newTestMedium(t)

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

func TestAppend_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	require.NoError(t, m.Write("append.txt", "hello"))

	w, err := m.Append("append.txt")
	require.NoError(t, err)

	_, err = w.Write([]byte(" world"))
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)

	content, err := m.Read("append.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello world", content)
}

func TestAppend_Good_NewFile(t *testing.T) {
	m, _ := newTestMedium(t)

	w, err := m.Append("new.txt")
	require.NoError(t, err)

	_, err = w.Write([]byte("fresh"))
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)

	content, err := m.Read("new.txt")
	require.NoError(t, err)
	assert.Equal(t, "fresh", content)
}

func TestReadStream_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	require.NoError(t, m.Write("stream.txt", "streaming content"))

	reader, err := m.ReadStream("stream.txt")
	require.NoError(t, err)
	defer reader.Close()

	data, err := goio.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "streaming content", string(data))
}

func TestReadStream_Bad_NotFound(t *testing.T) {
	m, _ := newTestMedium(t)
	_, err := m.ReadStream("nonexistent.txt")
	assert.Error(t, err)
}

func TestWriteStream_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	writer, err := m.WriteStream("output.txt")
	require.NoError(t, err)

	_, err = goio.Copy(writer, strings.NewReader("piped data"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	content, err := m.Read("output.txt")
	require.NoError(t, err)
	assert.Equal(t, "piped data", content)
}

func TestExists_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	assert.False(t, m.Exists("nonexistent.txt"))

	require.NoError(t, m.Write("file.txt", "content"))
	assert.True(t, m.Exists("file.txt"))
}

func TestExists_Good_DirectoryPrefix(t *testing.T) {
	m, _ := newTestMedium(t)

	require.NoError(t, m.Write("dir/file.txt", "content"))
	// "dir" should exist as a directory prefix
	assert.True(t, m.Exists("dir"))
}

func TestIsDir_Good(t *testing.T) {
	m, _ := newTestMedium(t)

	require.NoError(t, m.Write("dir/file.txt", "content"))

	assert.True(t, m.IsDir("dir"))
	assert.False(t, m.IsDir("dir/file.txt"))
	assert.False(t, m.IsDir("nonexistent"))
	assert.False(t, m.IsDir(""))
}

func TestKey_Good(t *testing.T) {
	mock := newMockS3()

	// No prefix
	m, _ := New("bucket", withAPI(mock))
	assert.Equal(t, "file.txt", m.key("file.txt"))
	assert.Equal(t, "dir/file.txt", m.key("dir/file.txt"))
	assert.Equal(t, "", m.key(""))
	assert.Equal(t, "file.txt", m.key("/file.txt"))
	assert.Equal(t, "file.txt", m.key("../file.txt"))

	// With prefix
	m2, _ := New("bucket", withAPI(mock), WithPrefix("pfx"))
	assert.Equal(t, "pfx/file.txt", m2.key("file.txt"))
	assert.Equal(t, "pfx/dir/file.txt", m2.key("dir/file.txt"))
	assert.Equal(t, "pfx/", m2.key(""))
}

// Ugly: verify the Medium interface is satisfied at compile time.
func TestInterfaceCompliance_Ugly(t *testing.T) {
	mock := newMockS3()
	m, err := New("bucket", withAPI(mock))
	require.NoError(t, err)

	// Verify all methods exist by calling them in a way that
	// proves compile-time satisfaction of the interface.
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
