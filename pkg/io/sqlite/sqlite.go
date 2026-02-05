// Package sqlite provides a SQLite-backed implementation of the io.Medium interface.
package sqlite

import (
	"bytes"
	"database/sql"
	goio "io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	coreerr "github.com/host-uk/core/pkg/framework/core"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Medium is a SQLite-backed storage backend implementing the io.Medium interface.
type Medium struct {
	db    *sql.DB
	table string
}

// Option configures a Medium.
type Option func(*Medium)

// WithTable sets the table name (default: "files").
func WithTable(table string) Option {
	return func(m *Medium) {
		m.table = table
	}
}

// New creates a new SQLite Medium at the given database path.
// Use ":memory:" for an in-memory database.
func New(dbPath string, opts ...Option) (*Medium, error) {
	if dbPath == "" {
		return nil, coreerr.E("sqlite.New", "database path is required", nil)
	}

	m := &Medium{table: "files"}
	for _, opt := range opts {
		opt(m)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, coreerr.E("sqlite.New", "failed to open database", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, coreerr.E("sqlite.New", "failed to set WAL mode", err)
	}

	// Create the schema
	createSQL := `CREATE TABLE IF NOT EXISTS ` + m.table + ` (
		path    TEXT PRIMARY KEY,
		content BLOB NOT NULL,
		mode    INTEGER DEFAULT 420,
		is_dir  BOOLEAN DEFAULT FALSE,
		mtime   DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := db.Exec(createSQL); err != nil {
		db.Close()
		return nil, coreerr.E("sqlite.New", "failed to create table", err)
	}

	m.db = db
	return m, nil
}

// Close closes the underlying database connection.
func (m *Medium) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// cleanPath normalizes a path for consistent storage.
// Uses a leading "/" before Clean to sandbox traversal attempts.
func cleanPath(p string) string {
	clean := path.Clean("/" + p)
	if clean == "/" {
		return ""
	}
	return strings.TrimPrefix(clean, "/")
}

// Read retrieves the content of a file as a string.
func (m *Medium) Read(p string) (string, error) {
	key := cleanPath(p)
	if key == "" {
		return "", coreerr.E("sqlite.Read", "path is required", os.ErrInvalid)
	}

	var content []byte
	var isDir bool
	err := m.db.QueryRow(
		`SELECT content, is_dir FROM `+m.table+` WHERE path = ?`, key,
	).Scan(&content, &isDir)
	if err == sql.ErrNoRows {
		return "", coreerr.E("sqlite.Read", "file not found: "+key, os.ErrNotExist)
	}
	if err != nil {
		return "", coreerr.E("sqlite.Read", "query failed: "+key, err)
	}
	if isDir {
		return "", coreerr.E("sqlite.Read", "path is a directory: "+key, os.ErrInvalid)
	}
	return string(content), nil
}

// Write saves the given content to a file, overwriting it if it exists.
func (m *Medium) Write(p, content string) error {
	key := cleanPath(p)
	if key == "" {
		return coreerr.E("sqlite.Write", "path is required", os.ErrInvalid)
	}

	_, err := m.db.Exec(
		`INSERT INTO `+m.table+` (path, content, mode, is_dir, mtime) VALUES (?, ?, 420, FALSE, ?)
		 ON CONFLICT(path) DO UPDATE SET content = excluded.content, is_dir = FALSE, mtime = excluded.mtime`,
		key, []byte(content), time.Now().UTC(),
	)
	if err != nil {
		return coreerr.E("sqlite.Write", "insert failed: "+key, err)
	}
	return nil
}

// EnsureDir makes sure a directory exists, creating it if necessary.
func (m *Medium) EnsureDir(p string) error {
	key := cleanPath(p)
	if key == "" {
		// Root always "exists"
		return nil
	}

	_, err := m.db.Exec(
		`INSERT INTO `+m.table+` (path, content, mode, is_dir, mtime) VALUES (?, '', 493, TRUE, ?)
		 ON CONFLICT(path) DO NOTHING`,
		key, time.Now().UTC(),
	)
	if err != nil {
		return coreerr.E("sqlite.EnsureDir", "insert failed: "+key, err)
	}
	return nil
}

// IsFile checks if a path exists and is a regular file.
func (m *Medium) IsFile(p string) bool {
	key := cleanPath(p)
	if key == "" {
		return false
	}

	var isDir bool
	err := m.db.QueryRow(
		`SELECT is_dir FROM `+m.table+` WHERE path = ?`, key,
	).Scan(&isDir)
	if err != nil {
		return false
	}
	return !isDir
}

// FileGet is a convenience function that reads a file from the medium.
func (m *Medium) FileGet(p string) (string, error) {
	return m.Read(p)
}

// FileSet is a convenience function that writes a file to the medium.
func (m *Medium) FileSet(p, content string) error {
	return m.Write(p, content)
}

// Delete removes a file or empty directory.
func (m *Medium) Delete(p string) error {
	key := cleanPath(p)
	if key == "" {
		return coreerr.E("sqlite.Delete", "path is required", os.ErrInvalid)
	}

	// Check if it's a directory with children
	var isDir bool
	err := m.db.QueryRow(
		`SELECT is_dir FROM `+m.table+` WHERE path = ?`, key,
	).Scan(&isDir)
	if err == sql.ErrNoRows {
		return coreerr.E("sqlite.Delete", "path not found: "+key, os.ErrNotExist)
	}
	if err != nil {
		return coreerr.E("sqlite.Delete", "query failed: "+key, err)
	}

	if isDir {
		// Check for children
		prefix := key + "/"
		var count int
		err := m.db.QueryRow(
			`SELECT COUNT(*) FROM `+m.table+` WHERE path LIKE ? AND path != ?`, prefix+"%", key,
		).Scan(&count)
		if err != nil {
			return coreerr.E("sqlite.Delete", "count failed: "+key, err)
		}
		if count > 0 {
			return coreerr.E("sqlite.Delete", "directory not empty: "+key, os.ErrExist)
		}
	}

	res, err := m.db.Exec(`DELETE FROM `+m.table+` WHERE path = ?`, key)
	if err != nil {
		return coreerr.E("sqlite.Delete", "delete failed: "+key, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return coreerr.E("sqlite.Delete", "path not found: "+key, os.ErrNotExist)
	}
	return nil
}

// DeleteAll removes a file or directory and all its contents recursively.
func (m *Medium) DeleteAll(p string) error {
	key := cleanPath(p)
	if key == "" {
		return coreerr.E("sqlite.DeleteAll", "path is required", os.ErrInvalid)
	}

	prefix := key + "/"

	// Delete the exact path and all children
	res, err := m.db.Exec(
		`DELETE FROM `+m.table+` WHERE path = ? OR path LIKE ?`,
		key, prefix+"%",
	)
	if err != nil {
		return coreerr.E("sqlite.DeleteAll", "delete failed: "+key, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return coreerr.E("sqlite.DeleteAll", "path not found: "+key, os.ErrNotExist)
	}
	return nil
}

// Rename moves a file or directory from oldPath to newPath.
func (m *Medium) Rename(oldPath, newPath string) error {
	oldKey := cleanPath(oldPath)
	newKey := cleanPath(newPath)
	if oldKey == "" || newKey == "" {
		return coreerr.E("sqlite.Rename", "both old and new paths are required", os.ErrInvalid)
	}

	tx, err := m.db.Begin()
	if err != nil {
		return coreerr.E("sqlite.Rename", "begin tx failed", err)
	}
	defer tx.Rollback()

	// Check if source exists
	var content []byte
	var mode int
	var isDir bool
	var mtime time.Time
	err = tx.QueryRow(
		`SELECT content, mode, is_dir, mtime FROM `+m.table+` WHERE path = ?`, oldKey,
	).Scan(&content, &mode, &isDir, &mtime)
	if err == sql.ErrNoRows {
		return coreerr.E("sqlite.Rename", "source not found: "+oldKey, os.ErrNotExist)
	}
	if err != nil {
		return coreerr.E("sqlite.Rename", "query failed: "+oldKey, err)
	}

	// Insert or replace at new path
	_, err = tx.Exec(
		`INSERT INTO `+m.table+` (path, content, mode, is_dir, mtime) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(path) DO UPDATE SET content = excluded.content, mode = excluded.mode, is_dir = excluded.is_dir, mtime = excluded.mtime`,
		newKey, content, mode, isDir, mtime,
	)
	if err != nil {
		return coreerr.E("sqlite.Rename", "insert at new path failed: "+newKey, err)
	}

	// Delete old path
	_, err = tx.Exec(`DELETE FROM `+m.table+` WHERE path = ?`, oldKey)
	if err != nil {
		return coreerr.E("sqlite.Rename", "delete old path failed: "+oldKey, err)
	}

	// If it's a directory, move all children
	if isDir {
		oldPrefix := oldKey + "/"
		newPrefix := newKey + "/"

		rows, err := tx.Query(
			`SELECT path, content, mode, is_dir, mtime FROM `+m.table+` WHERE path LIKE ?`,
			oldPrefix+"%",
		)
		if err != nil {
			return coreerr.E("sqlite.Rename", "query children failed", err)
		}

		type child struct {
			path    string
			content []byte
			mode    int
			isDir   bool
			mtime   time.Time
		}
		var children []child
		for rows.Next() {
			var c child
			if err := rows.Scan(&c.path, &c.content, &c.mode, &c.isDir, &c.mtime); err != nil {
				rows.Close()
				return coreerr.E("sqlite.Rename", "scan child failed", err)
			}
			children = append(children, c)
		}
		rows.Close()

		for _, c := range children {
			newChildPath := newPrefix + strings.TrimPrefix(c.path, oldPrefix)
			_, err = tx.Exec(
				`INSERT INTO `+m.table+` (path, content, mode, is_dir, mtime) VALUES (?, ?, ?, ?, ?)
				 ON CONFLICT(path) DO UPDATE SET content = excluded.content, mode = excluded.mode, is_dir = excluded.is_dir, mtime = excluded.mtime`,
				newChildPath, c.content, c.mode, c.isDir, c.mtime,
			)
			if err != nil {
				return coreerr.E("sqlite.Rename", "insert child failed", err)
			}
		}

		// Delete old children
		_, err = tx.Exec(`DELETE FROM `+m.table+` WHERE path LIKE ?`, oldPrefix+"%")
		if err != nil {
			return coreerr.E("sqlite.Rename", "delete old children failed", err)
		}
	}

	return tx.Commit()
}

// List returns the directory entries for the given path.
func (m *Medium) List(p string) ([]fs.DirEntry, error) {
	prefix := cleanPath(p)
	if prefix != "" {
		prefix += "/"
	}

	// Query all paths under the prefix
	rows, err := m.db.Query(
		`SELECT path, content, mode, is_dir, mtime FROM `+m.table+` WHERE path LIKE ? OR path LIKE ?`,
		prefix+"%", prefix+"%",
	)
	if err != nil {
		return nil, coreerr.E("sqlite.List", "query failed", err)
	}
	defer rows.Close()

	seen := make(map[string]bool)
	var entries []fs.DirEntry

	for rows.Next() {
		var rowPath string
		var content []byte
		var mode int
		var isDir bool
		var mtime time.Time
		if err := rows.Scan(&rowPath, &content, &mode, &isDir, &mtime); err != nil {
			return nil, coreerr.E("sqlite.List", "scan failed", err)
		}

		rest := strings.TrimPrefix(rowPath, prefix)
		if rest == "" {
			continue
		}

		// Check if this is a direct child or nested
		if idx := strings.Index(rest, "/"); idx >= 0 {
			// Nested - register as a directory
			dirName := rest[:idx]
			if !seen[dirName] {
				seen[dirName] = true
				entries = append(entries, &dirEntry{
					name:  dirName,
					isDir: true,
					mode:  fs.ModeDir | 0755,
					info: &fileInfo{
						name:  dirName,
						isDir: true,
						mode:  fs.ModeDir | 0755,
					},
				})
			}
		} else {
			// Direct child
			if !seen[rest] {
				seen[rest] = true
				entries = append(entries, &dirEntry{
					name:  rest,
					isDir: isDir,
					mode:  fs.FileMode(mode),
					info: &fileInfo{
						name:    rest,
						size:    int64(len(content)),
						mode:    fs.FileMode(mode),
						modTime: mtime,
						isDir:   isDir,
					},
				})
			}
		}
	}

	return entries, rows.Err()
}

// Stat returns file information for the given path.
func (m *Medium) Stat(p string) (fs.FileInfo, error) {
	key := cleanPath(p)
	if key == "" {
		return nil, coreerr.E("sqlite.Stat", "path is required", os.ErrInvalid)
	}

	var content []byte
	var mode int
	var isDir bool
	var mtime time.Time
	err := m.db.QueryRow(
		`SELECT content, mode, is_dir, mtime FROM `+m.table+` WHERE path = ?`, key,
	).Scan(&content, &mode, &isDir, &mtime)
	if err == sql.ErrNoRows {
		return nil, coreerr.E("sqlite.Stat", "path not found: "+key, os.ErrNotExist)
	}
	if err != nil {
		return nil, coreerr.E("sqlite.Stat", "query failed: "+key, err)
	}

	name := path.Base(key)
	return &fileInfo{
		name:    name,
		size:    int64(len(content)),
		mode:    fs.FileMode(mode),
		modTime: mtime,
		isDir:   isDir,
	}, nil
}

// Open opens the named file for reading.
func (m *Medium) Open(p string) (fs.File, error) {
	key := cleanPath(p)
	if key == "" {
		return nil, coreerr.E("sqlite.Open", "path is required", os.ErrInvalid)
	}

	var content []byte
	var mode int
	var isDir bool
	var mtime time.Time
	err := m.db.QueryRow(
		`SELECT content, mode, is_dir, mtime FROM `+m.table+` WHERE path = ?`, key,
	).Scan(&content, &mode, &isDir, &mtime)
	if err == sql.ErrNoRows {
		return nil, coreerr.E("sqlite.Open", "file not found: "+key, os.ErrNotExist)
	}
	if err != nil {
		return nil, coreerr.E("sqlite.Open", "query failed: "+key, err)
	}
	if isDir {
		return nil, coreerr.E("sqlite.Open", "path is a directory: "+key, os.ErrInvalid)
	}

	return &sqliteFile{
		name:    path.Base(key),
		content: content,
		mode:    fs.FileMode(mode),
		modTime: mtime,
	}, nil
}

// Create creates or truncates the named file.
func (m *Medium) Create(p string) (goio.WriteCloser, error) {
	key := cleanPath(p)
	if key == "" {
		return nil, coreerr.E("sqlite.Create", "path is required", os.ErrInvalid)
	}
	return &sqliteWriteCloser{
		medium: m,
		path:   key,
	}, nil
}

// Append opens the named file for appending, creating it if it doesn't exist.
func (m *Medium) Append(p string) (goio.WriteCloser, error) {
	key := cleanPath(p)
	if key == "" {
		return nil, coreerr.E("sqlite.Append", "path is required", os.ErrInvalid)
	}

	var existing []byte
	err := m.db.QueryRow(
		`SELECT content FROM `+m.table+` WHERE path = ? AND is_dir = FALSE`, key,
	).Scan(&existing)
	if err != nil && err != sql.ErrNoRows {
		return nil, coreerr.E("sqlite.Append", "query failed: "+key, err)
	}

	return &sqliteWriteCloser{
		medium: m,
		path:   key,
		data:   existing,
	}, nil
}

// ReadStream returns a reader for the file content.
func (m *Medium) ReadStream(p string) (goio.ReadCloser, error) {
	key := cleanPath(p)
	if key == "" {
		return nil, coreerr.E("sqlite.ReadStream", "path is required", os.ErrInvalid)
	}

	var content []byte
	var isDir bool
	err := m.db.QueryRow(
		`SELECT content, is_dir FROM `+m.table+` WHERE path = ?`, key,
	).Scan(&content, &isDir)
	if err == sql.ErrNoRows {
		return nil, coreerr.E("sqlite.ReadStream", "file not found: "+key, os.ErrNotExist)
	}
	if err != nil {
		return nil, coreerr.E("sqlite.ReadStream", "query failed: "+key, err)
	}
	if isDir {
		return nil, coreerr.E("sqlite.ReadStream", "path is a directory: "+key, os.ErrInvalid)
	}

	return goio.NopCloser(bytes.NewReader(content)), nil
}

// WriteStream returns a writer for the file content. Content is stored on Close.
func (m *Medium) WriteStream(p string) (goio.WriteCloser, error) {
	return m.Create(p)
}

// Exists checks if a path exists (file or directory).
func (m *Medium) Exists(p string) bool {
	key := cleanPath(p)
	if key == "" {
		// Root always exists
		return true
	}

	var count int
	err := m.db.QueryRow(
		`SELECT COUNT(*) FROM `+m.table+` WHERE path = ?`, key,
	).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// IsDir checks if a path exists and is a directory.
func (m *Medium) IsDir(p string) bool {
	key := cleanPath(p)
	if key == "" {
		return false
	}

	var isDir bool
	err := m.db.QueryRow(
		`SELECT is_dir FROM `+m.table+` WHERE path = ?`, key,
	).Scan(&isDir)
	if err != nil {
		return false
	}
	return isDir
}

// --- Internal types ---

// fileInfo implements fs.FileInfo for SQLite entries.
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

// dirEntry implements fs.DirEntry for SQLite listings.
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

// sqliteFile implements fs.File for SQLite entries.
type sqliteFile struct {
	name    string
	content []byte
	offset  int64
	mode    fs.FileMode
	modTime time.Time
}

func (f *sqliteFile) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		name:    f.name,
		size:    int64(len(f.content)),
		mode:    f.mode,
		modTime: f.modTime,
	}, nil
}

func (f *sqliteFile) Read(b []byte) (int, error) {
	if f.offset >= int64(len(f.content)) {
		return 0, goio.EOF
	}
	n := copy(b, f.content[f.offset:])
	f.offset += int64(n)
	return n, nil
}

func (f *sqliteFile) Close() error {
	return nil
}

// sqliteWriteCloser buffers writes and stores to SQLite on Close.
type sqliteWriteCloser struct {
	medium *Medium
	path   string
	data   []byte
}

func (w *sqliteWriteCloser) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *sqliteWriteCloser) Close() error {
	_, err := w.medium.db.Exec(
		`INSERT INTO `+w.medium.table+` (path, content, mode, is_dir, mtime) VALUES (?, ?, 420, FALSE, ?)
		 ON CONFLICT(path) DO UPDATE SET content = excluded.content, is_dir = FALSE, mtime = excluded.mtime`,
		w.path, w.data, time.Now().UTC(),
	)
	if err != nil {
		return coreerr.E("sqlite.WriteCloser.Close", "store failed: "+w.path, err)
	}
	return nil
}
