package node

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// Node is an in-memory filesystem that implements fs.FS, fs.StatFS,
// and fs.ReadFileFS. It stores files as byte slices keyed by their
// path, with directories being implicit based on path prefixes.
//
// Ported from github.com/Snider/Borg/pkg/datanode.
type Node struct {
	files map[string]*nodeFile
}

// Compile-time interface checks.
var (
	_ fs.FS         = (*Node)(nil)
	_ fs.StatFS     = (*Node)(nil)
	_ fs.ReadFileFS = (*Node)(nil)
)

// New creates a new, empty Node.
func New() *Node {
	return &Node{files: make(map[string]*nodeFile)}
}

// FromTar creates a new Node from a tarball.
func FromTar(tarball []byte) (*Node, error) {
	n := New()
	tarReader := tar.NewReader(bytes.NewReader(tarball))

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag == tar.TypeReg {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}
			n.AddData(header.Name, data)
		}
	}

	return n, nil
}

// ToTar serializes the Node to a tarball.
func (n *Node) ToTar() ([]byte, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, file := range n.files {
		hdr := &tar.Header{
			Name:    file.name,
			Mode:    0600,
			Size:    int64(len(file.content)),
			ModTime: file.modTime,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write(file.content); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// AddData adds a file to the Node. If a file with the same name
// already exists it is overwritten. Directory entries (names ending
// in "/") and empty names are silently ignored.
func (n *Node) AddData(name string, content []byte) {
	name = strings.TrimPrefix(name, "/")
	if name == "" {
		return
	}
	// Directories are implicit, so we don't store them.
	// A name ending in "/" is treated as a directory.
	if strings.HasSuffix(name, "/") {
		return
	}
	n.files[name] = &nodeFile{
		name:    name,
		content: content,
		modTime: time.Now(),
	}
}

// Open opens a file from the Node, satisfying the fs.FS interface.
func (n *Node) Open(name string) (fs.File, error) {
	name = strings.TrimPrefix(name, "/")
	if file, ok := n.files[name]; ok {
		return &nodeFileReader{file: file}, nil
	}
	// Check if it's a directory.
	prefix := name + "/"
	if name == "." || name == "" {
		prefix = ""
	}
	for p := range n.files {
		if strings.HasPrefix(p, prefix) {
			return &dirFile{path: name, modTime: time.Now()}, nil
		}
	}
	return nil, fs.ErrNotExist
}

// Stat returns the FileInfo for the named file, satisfying the
// fs.StatFS interface.
func (n *Node) Stat(name string) (fs.FileInfo, error) {
	name = strings.TrimPrefix(name, "/")
	if file, ok := n.files[name]; ok {
		return file.Stat()
	}
	// Check if it's a directory.
	prefix := name + "/"
	if name == "." || name == "" {
		prefix = ""
	}
	for p := range n.files {
		if strings.HasPrefix(p, prefix) {
			return &dirInfo{name: path.Base(name), modTime: time.Now()}, nil
		}
	}
	return nil, fs.ErrNotExist
}

// ReadFile reads the named file and returns its contents, satisfying
// the fs.ReadFileFS interface.
func (n *Node) ReadFile(name string) ([]byte, error) {
	name = strings.TrimPrefix(name, "/")
	if file, ok := n.files[name]; ok {
		// Return a copy so callers cannot mutate the internal state.
		out := make([]byte, len(file.content))
		copy(out, file.content)
		return out, nil
	}
	return nil, fs.ErrNotExist
}

// ReadDir reads and returns all directory entries for the named directory.
func (n *Node) ReadDir(name string) ([]fs.DirEntry, error) {
	name = strings.TrimPrefix(name, "/")
	if name == "." {
		name = ""
	}

	// Disallow reading a file as a directory.
	if info, err := n.Stat(name); err == nil && !info.IsDir() {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}

	entries := []fs.DirEntry{}
	seen := make(map[string]bool)

	prefix := ""
	if name != "" {
		prefix = name + "/"
	}

	for p := range n.files {
		if !strings.HasPrefix(p, prefix) {
			continue
		}

		relPath := strings.TrimPrefix(p, prefix)
		firstComponent := strings.Split(relPath, "/")[0]

		if seen[firstComponent] {
			continue
		}
		seen[firstComponent] = true

		if strings.Contains(relPath, "/") {
			// It's a directory.
			dir := &dirInfo{name: firstComponent, modTime: time.Now()}
			entries = append(entries, fs.FileInfoToDirEntry(dir))
		} else {
			// It's a file.
			file := n.files[p]
			info, _ := file.Stat()
			entries = append(entries, fs.FileInfoToDirEntry(info))
		}
	}

	// Sort for stable order.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	return entries, nil
}

// Exists returns true if the file or directory exists in the Node.
func (n *Node) Exists(name string, opts ...ExistsOptions) (bool, error) {
	info, err := n.Stat(name)
	if err != nil {
		if err == fs.ErrNotExist || os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if len(opts) > 0 {
		if opts[0].WantType == fs.ModeDir && !info.IsDir() {
			return false, nil
		}
		if opts[0].WantType != fs.ModeDir && info.IsDir() {
			return false, nil
		}
	}
	return true, nil
}

// ExistsOptions allows customizing the Exists check.
type ExistsOptions struct {
	WantType fs.FileMode
}

// WalkOptions allows customizing the Walk behavior.
type WalkOptions struct {
	MaxDepth   int
	Filter     func(path string, d fs.DirEntry) bool
	SkipErrors bool
}

// Walk recursively descends the file tree rooted at root, calling fn
// for each file or directory in the tree.
func (n *Node) Walk(root string, fn fs.WalkDirFunc, opts ...WalkOptions) error {
	var maxDepth int
	var filter func(string, fs.DirEntry) bool
	var skipErrors bool
	if len(opts) > 0 {
		maxDepth = opts[0].MaxDepth
		filter = opts[0].Filter
		skipErrors = opts[0].SkipErrors
	}

	return fs.WalkDir(n, root, func(p string, de fs.DirEntry, err error) error {
		if err != nil {
			if skipErrors {
				return nil
			}
			return fn(p, de, err)
		}
		if filter != nil && !filter(p, de) {
			if de.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Process the entry first.
		if err := fn(p, de, nil); err != nil {
			return err
		}

		if maxDepth > 0 {
			// Calculate depth relative to root.
			cleanedPath := strings.TrimPrefix(p, root)
			cleanedPath = strings.TrimPrefix(cleanedPath, "/")

			currentDepth := 0
			if p != root {
				if cleanedPath == "" {
					currentDepth = 0
				} else {
					currentDepth = strings.Count(cleanedPath, "/") + 1
				}
			}

			if de.IsDir() && currentDepth >= maxDepth {
				return fs.SkipDir
			}
		}
		return nil
	})
}

// CopyFile copies a file from the Node to the local filesystem.
func (n *Node) CopyFile(sourcePath string, target string, perm os.FileMode) error {
	sourceFile, err := n.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	return err
}

// ---------------------------------------------------------------------------
// Internal types
// ---------------------------------------------------------------------------

// nodeFile represents a file stored in the Node.
type nodeFile struct {
	name    string
	content []byte
	modTime time.Time
}

func (f *nodeFile) Stat() (fs.FileInfo, error) { return &nodeFileInfo{file: f}, nil }
func (f *nodeFile) Read([]byte) (int, error)    { return 0, io.EOF }
func (f *nodeFile) Close() error                { return nil }

// nodeFileInfo implements fs.FileInfo for a nodeFile.
type nodeFileInfo struct{ file *nodeFile }

func (i *nodeFileInfo) Name() string       { return path.Base(i.file.name) }
func (i *nodeFileInfo) Size() int64        { return int64(len(i.file.content)) }
func (i *nodeFileInfo) Mode() fs.FileMode  { return 0444 }
func (i *nodeFileInfo) ModTime() time.Time { return i.file.modTime }
func (i *nodeFileInfo) IsDir() bool        { return false }
func (i *nodeFileInfo) Sys() interface{}   { return nil }

// nodeFileReader implements fs.File for reading a nodeFile.
type nodeFileReader struct {
	file   *nodeFile
	reader *bytes.Reader
}

func (r *nodeFileReader) Stat() (fs.FileInfo, error) { return r.file.Stat() }
func (r *nodeFileReader) Read(p []byte) (int, error) {
	if r.reader == nil {
		r.reader = bytes.NewReader(r.file.content)
	}
	return r.reader.Read(p)
}
func (r *nodeFileReader) Close() error { return nil }

// dirInfo implements fs.FileInfo for an implicit directory.
type dirInfo struct {
	name    string
	modTime time.Time
}

func (d *dirInfo) Name() string       { return d.name }
func (d *dirInfo) Size() int64        { return 0 }
func (d *dirInfo) Mode() fs.FileMode  { return fs.ModeDir | 0555 }
func (d *dirInfo) ModTime() time.Time { return d.modTime }
func (d *dirInfo) IsDir() bool        { return true }
func (d *dirInfo) Sys() interface{}   { return nil }

// dirFile implements fs.File for a directory.
type dirFile struct {
	path    string
	modTime time.Time
}

func (d *dirFile) Stat() (fs.FileInfo, error) {
	return &dirInfo{name: path.Base(d.path), modTime: d.modTime}, nil
}
func (d *dirFile) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.path, Err: fs.ErrInvalid}
}
func (d *dirFile) Close() error { return nil }
