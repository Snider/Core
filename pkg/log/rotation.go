package log

import (
	"fmt"
	"io"
	"sync"
	"time"

	coreio "github.com/host-uk/core/pkg/io"
)

// RotatingWriter implements io.WriteCloser and provides log rotation.
type RotatingWriter struct {
	opts   RotationOptions
	medium coreio.Medium
	mu     sync.Mutex
	file   io.WriteCloser
	size   int64
}

// NewRotatingWriter creates a new RotatingWriter with the given options and medium.
func NewRotatingWriter(opts RotationOptions, m coreio.Medium) *RotatingWriter {
	if m == nil {
		m = coreio.Local
	}
	if opts.MaxSize <= 0 {
		opts.MaxSize = 100 // 100 MB
	}
	if opts.MaxBackups <= 0 {
		opts.MaxBackups = 5
	}
	if opts.MaxAge == 0 {
		opts.MaxAge = 28 // 28 days
	} else if opts.MaxAge < 0 {
		opts.MaxAge = 0 // disabled
	}

	return &RotatingWriter{
		opts:   opts,
		medium: m,
	}
}

// Write writes data to the current log file, rotating it if necessary.
func (w *RotatingWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.openExistingOrNew(); err != nil {
			return 0, err
		}
	}

	if w.size+int64(len(p)) > int64(w.opts.MaxSize)*1024*1024 {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	if err == nil {
		w.size += int64(n)
	}
	return n, err
}

// Close closes the current log file.
func (w *RotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.close()
}

func (w *RotatingWriter) close() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *RotatingWriter) openExistingOrNew() error {
	info, err := w.medium.Stat(w.opts.Filename)
	if err == nil {
		w.size = info.Size()
		f, err := w.medium.Append(w.opts.Filename)
		if err != nil {
			return err
		}
		w.file = f
		return nil
	}

	f, err := w.medium.Create(w.opts.Filename)
	if err != nil {
		return err
	}
	w.file = f
	w.size = 0
	return nil
}

func (w *RotatingWriter) rotate() error {
	if err := w.close(); err != nil {
		return err
	}

	if err := w.rotateFiles(); err != nil {
		// Try to reopen current file even if rotation failed
		_ = w.openExistingOrNew()
		return err
	}

	if err := w.openExistingOrNew(); err != nil {
		return err
	}

	w.cleanup()

	return nil
}

func (w *RotatingWriter) rotateFiles() error {
	// Rotate existing backups: log.N -> log.N+1
	for i := w.opts.MaxBackups; i >= 1; i-- {
		oldPath := w.backupPath(i)
		newPath := w.backupPath(i + 1)

		if w.medium.Exists(oldPath) {
			if i+1 > w.opts.MaxBackups {
				_ = w.medium.Delete(oldPath)
			} else {
				_ = w.medium.Rename(oldPath, newPath)
			}
		}
	}

	// log -> log.1
	return w.medium.Rename(w.opts.Filename, w.backupPath(1))
}

func (w *RotatingWriter) backupPath(n int) string {
	return fmt.Sprintf("%s.%d", w.opts.Filename, n)
}

func (w *RotatingWriter) cleanup() {
	// 1. Remove backups beyond MaxBackups
	// This is already partially handled by rotateFiles but we can be thorough
	for i := w.opts.MaxBackups + 1; ; i++ {
		path := w.backupPath(i)
		if !w.medium.Exists(path) {
			break
		}
		_ = w.medium.Delete(path)
	}

	// 2. Remove backups older than MaxAge
	if w.opts.MaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -w.opts.MaxAge)
		for i := 1; i <= w.opts.MaxBackups; i++ {
			path := w.backupPath(i)
			info, err := w.medium.Stat(path)
			if err == nil && info.ModTime().Before(cutoff) {
				_ = w.medium.Delete(path)
			}
		}
	}
}
