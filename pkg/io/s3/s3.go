// Package s3 provides an S3-backed implementation of the io.Medium interface.
package s3

import (
	"bytes"
	"context"
	"fmt"
	goio "io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	coreerr "github.com/host-uk/core/pkg/framework/core"
)

// s3API is the subset of the S3 client API used by this package.
// This allows for interface-based mocking in tests.
type s3API interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
}

// Medium is an S3-backed storage backend implementing the io.Medium interface.
type Medium struct {
	client s3API
	bucket string
	prefix string
}

// Option configures a Medium.
type Option func(*Medium)

// WithPrefix sets an optional key prefix for all operations.
func WithPrefix(prefix string) Option {
	return func(m *Medium) {
		// Ensure prefix ends with "/" if non-empty
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		m.prefix = prefix
	}
}

// WithClient sets the S3 client for dependency injection.
func WithClient(client *s3.Client) Option {
	return func(m *Medium) {
		m.client = client
	}
}

// withAPI sets the s3API interface directly (for testing with mocks).
func withAPI(api s3API) Option {
	return func(m *Medium) {
		m.client = api
	}
}

// New creates a new S3 Medium for the given bucket.
func New(bucket string, opts ...Option) (*Medium, error) {
	if bucket == "" {
		return nil, coreerr.E("s3.New", "bucket name is required", nil)
	}
	m := &Medium{bucket: bucket}
	for _, opt := range opts {
		opt(m)
	}
	if m.client == nil {
		return nil, coreerr.E("s3.New", "S3 client is required (use WithClient option)", nil)
	}
	return m, nil
}

// key returns the full S3 object key for a given path.
func (m *Medium) key(p string) string {
	// Clean the path using a leading "/" to sandbox traversal attempts,
	// then strip the "/" prefix. This ensures ".." can't escape.
	clean := path.Clean("/" + p)
	if clean == "/" {
		clean = ""
	}
	clean = strings.TrimPrefix(clean, "/")

	if m.prefix == "" {
		return clean
	}
	if clean == "" {
		return m.prefix
	}
	return m.prefix + clean
}

// Read retrieves the content of a file as a string.
func (m *Medium) Read(p string) (string, error) {
	key := m.key(p)
	if key == "" {
		return "", coreerr.E("s3.Read", "path is required", os.ErrInvalid)
	}

	out, err := m.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", coreerr.E("s3.Read", "failed to get object: "+key, err)
	}
	defer out.Body.Close()

	data, err := goio.ReadAll(out.Body)
	if err != nil {
		return "", coreerr.E("s3.Read", "failed to read body: "+key, err)
	}
	return string(data), nil
}

// Write saves the given content to a file, overwriting it if it exists.
func (m *Medium) Write(p, content string) error {
	key := m.key(p)
	if key == "" {
		return coreerr.E("s3.Write", "path is required", os.ErrInvalid)
	}

	_, err := m.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(content),
	})
	if err != nil {
		return coreerr.E("s3.Write", "failed to put object: "+key, err)
	}
	return nil
}

// EnsureDir is a no-op for S3 (S3 has no real directories).
func (m *Medium) EnsureDir(_ string) error {
	return nil
}

// IsFile checks if a path exists and is a regular file (not a "directory" prefix).
func (m *Medium) IsFile(p string) bool {
	key := m.key(p)
	if key == "" {
		return false
	}
	// A "file" in S3 is an object whose key does not end with "/"
	if strings.HasSuffix(key, "/") {
		return false
	}
	_, err := m.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	return err == nil
}

// FileGet is a convenience function that reads a file from the medium.
func (m *Medium) FileGet(p string) (string, error) {
	return m.Read(p)
}

// FileSet is a convenience function that writes a file to the medium.
func (m *Medium) FileSet(p, content string) error {
	return m.Write(p, content)
}

// Delete removes a single object.
func (m *Medium) Delete(p string) error {
	key := m.key(p)
	if key == "" {
		return coreerr.E("s3.Delete", "path is required", os.ErrInvalid)
	}

	_, err := m.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return coreerr.E("s3.Delete", "failed to delete object: "+key, err)
	}
	return nil
}

// DeleteAll removes all objects under the given prefix.
func (m *Medium) DeleteAll(p string) error {
	key := m.key(p)
	if key == "" {
		return coreerr.E("s3.DeleteAll", "path is required", os.ErrInvalid)
	}

	// First, try deleting the exact key
	_, _ = m.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})

	// Then delete all objects under the prefix
	prefix := key
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	paginator := true
	var continuationToken *string

	for paginator {
		listOut, err := m.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket:            aws.String(m.bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return coreerr.E("s3.DeleteAll", "failed to list objects: "+prefix, err)
		}

		if len(listOut.Contents) == 0 {
			break
		}

		objects := make([]types.ObjectIdentifier, len(listOut.Contents))
		for i, obj := range listOut.Contents {
			objects[i] = types.ObjectIdentifier{Key: obj.Key}
		}

		_, err = m.client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
			Bucket: aws.String(m.bucket),
			Delete: &types.Delete{Objects: objects, Quiet: aws.Bool(true)},
		})
		if err != nil {
			return coreerr.E("s3.DeleteAll", "failed to delete objects", err)
		}

		if listOut.IsTruncated != nil && *listOut.IsTruncated {
			continuationToken = listOut.NextContinuationToken
		} else {
			paginator = false
		}
	}

	return nil
}

// Rename moves an object by copying then deleting the original.
func (m *Medium) Rename(oldPath, newPath string) error {
	oldKey := m.key(oldPath)
	newKey := m.key(newPath)
	if oldKey == "" || newKey == "" {
		return coreerr.E("s3.Rename", "both old and new paths are required", os.ErrInvalid)
	}

	copySource := m.bucket + "/" + oldKey

	_, err := m.client.CopyObject(context.Background(), &s3.CopyObjectInput{
		Bucket:     aws.String(m.bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(newKey),
	})
	if err != nil {
		return coreerr.E("s3.Rename", "failed to copy object: "+oldKey+" -> "+newKey, err)
	}

	_, err = m.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(oldKey),
	})
	if err != nil {
		return coreerr.E("s3.Rename", "failed to delete source object: "+oldKey, err)
	}

	return nil
}

// List returns directory entries for the given path using ListObjectsV2 with delimiter.
func (m *Medium) List(p string) ([]fs.DirEntry, error) {
	prefix := m.key(p)
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var entries []fs.DirEntry

	listOut, err := m.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket:    aws.String(m.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, coreerr.E("s3.List", "failed to list objects: "+prefix, err)
	}

	// Common prefixes are "directories"
	for _, cp := range listOut.CommonPrefixes {
		if cp.Prefix == nil {
			continue
		}
		name := strings.TrimPrefix(*cp.Prefix, prefix)
		name = strings.TrimSuffix(name, "/")
		if name == "" {
			continue
		}
		entries = append(entries, &dirEntry{
			name:  name,
			isDir: true,
			mode:  fs.ModeDir | 0755,
			info: &fileInfo{
				name:  name,
				isDir: true,
				mode:  fs.ModeDir | 0755,
			},
		})
	}

	// Contents are "files" (excluding the prefix itself)
	for _, obj := range listOut.Contents {
		if obj.Key == nil {
			continue
		}
		name := strings.TrimPrefix(*obj.Key, prefix)
		if name == "" || strings.Contains(name, "/") {
			continue
		}
		var size int64
		if obj.Size != nil {
			size = *obj.Size
		}
		var modTime time.Time
		if obj.LastModified != nil {
			modTime = *obj.LastModified
		}
		entries = append(entries, &dirEntry{
			name:  name,
			isDir: false,
			mode:  0644,
			info: &fileInfo{
				name:    name,
				size:    size,
				mode:    0644,
				modTime: modTime,
			},
		})
	}

	return entries, nil
}

// Stat returns file information for the given path using HeadObject.
func (m *Medium) Stat(p string) (fs.FileInfo, error) {
	key := m.key(p)
	if key == "" {
		return nil, coreerr.E("s3.Stat", "path is required", os.ErrInvalid)
	}

	out, err := m.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, coreerr.E("s3.Stat", "failed to head object: "+key, err)
	}

	var size int64
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	var modTime time.Time
	if out.LastModified != nil {
		modTime = *out.LastModified
	}

	name := path.Base(key)
	return &fileInfo{
		name:    name,
		size:    size,
		mode:    0644,
		modTime: modTime,
	}, nil
}

// Open opens the named file for reading.
func (m *Medium) Open(p string) (fs.File, error) {
	key := m.key(p)
	if key == "" {
		return nil, coreerr.E("s3.Open", "path is required", os.ErrInvalid)
	}

	out, err := m.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, coreerr.E("s3.Open", "failed to get object: "+key, err)
	}

	data, err := goio.ReadAll(out.Body)
	out.Body.Close()
	if err != nil {
		return nil, coreerr.E("s3.Open", "failed to read body: "+key, err)
	}

	var size int64
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	var modTime time.Time
	if out.LastModified != nil {
		modTime = *out.LastModified
	}

	return &s3File{
		name:    path.Base(key),
		content: data,
		size:    size,
		modTime: modTime,
	}, nil
}

// Create creates or truncates the named file. Returns a writer that
// uploads the content on Close.
func (m *Medium) Create(p string) (goio.WriteCloser, error) {
	key := m.key(p)
	if key == "" {
		return nil, coreerr.E("s3.Create", "path is required", os.ErrInvalid)
	}
	return &s3WriteCloser{
		medium: m,
		key:    key,
	}, nil
}

// Append opens the named file for appending. It downloads the existing
// content (if any) and re-uploads the combined content on Close.
func (m *Medium) Append(p string) (goio.WriteCloser, error) {
	key := m.key(p)
	if key == "" {
		return nil, coreerr.E("s3.Append", "path is required", os.ErrInvalid)
	}

	var existing []byte
	out, err := m.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		existing, _ = goio.ReadAll(out.Body)
		out.Body.Close()
	}

	return &s3WriteCloser{
		medium: m,
		key:    key,
		data:   existing,
	}, nil
}

// ReadStream returns a reader for the file content.
func (m *Medium) ReadStream(p string) (goio.ReadCloser, error) {
	key := m.key(p)
	if key == "" {
		return nil, coreerr.E("s3.ReadStream", "path is required", os.ErrInvalid)
	}

	out, err := m.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, coreerr.E("s3.ReadStream", "failed to get object: "+key, err)
	}
	return out.Body, nil
}

// WriteStream returns a writer for the file content. Content is uploaded on Close.
func (m *Medium) WriteStream(p string) (goio.WriteCloser, error) {
	return m.Create(p)
}

// Exists checks if a path exists (file or directory prefix).
func (m *Medium) Exists(p string) bool {
	key := m.key(p)
	if key == "" {
		return false
	}

	// Check as an exact object
	_, err := m.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true
	}

	// Check as a "directory" prefix
	prefix := key
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	listOut, err := m.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket:  aws.String(m.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false
	}
	return len(listOut.Contents) > 0 || len(listOut.CommonPrefixes) > 0
}

// IsDir checks if a path exists and is a directory (has objects under it as a prefix).
func (m *Medium) IsDir(p string) bool {
	key := m.key(p)
	if key == "" {
		return false
	}

	prefix := key
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	listOut, err := m.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket:  aws.String(m.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false
	}
	return len(listOut.Contents) > 0 || len(listOut.CommonPrefixes) > 0
}

// --- Internal types ---

// fileInfo implements fs.FileInfo for S3 objects.
type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.modTime }
func (fi *fileInfo) IsDir() bool        { return fi.isDir }
func (fi *fileInfo) Sys() any           { return nil }

// dirEntry implements fs.DirEntry for S3 listings.
type dirEntry struct {
	name  string
	isDir bool
	mode  fs.FileMode
	info  fs.FileInfo
}

func (de *dirEntry) Name() string               { return de.name }
func (de *dirEntry) IsDir() bool                { return de.isDir }
func (de *dirEntry) Type() fs.FileMode          { return de.mode.Type() }
func (de *dirEntry) Info() (fs.FileInfo, error)  { return de.info, nil }

// s3File implements fs.File for S3 objects.
type s3File struct {
	name    string
	content []byte
	offset  int64
	size    int64
	modTime time.Time
}

func (f *s3File) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		name:    f.name,
		size:    int64(len(f.content)),
		mode:    0644,
		modTime: f.modTime,
	}, nil
}

func (f *s3File) Read(b []byte) (int, error) {
	if f.offset >= int64(len(f.content)) {
		return 0, goio.EOF
	}
	n := copy(b, f.content[f.offset:])
	f.offset += int64(n)
	return n, nil
}

func (f *s3File) Close() error {
	return nil
}

// s3WriteCloser buffers writes and uploads to S3 on Close.
type s3WriteCloser struct {
	medium *Medium
	key    string
	data   []byte
}

func (w *s3WriteCloser) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *s3WriteCloser) Close() error {
	_, err := w.medium.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(w.medium.bucket),
		Key:    aws.String(w.key),
		Body:   bytes.NewReader(w.data),
	})
	if err != nil {
		return fmt.Errorf("s3: failed to upload on close: %w", err)
	}
	return nil
}
