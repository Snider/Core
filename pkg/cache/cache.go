// Package cache provides a file-based cache for GitHub API responses.
package cache

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/host-uk/core/pkg/io"
)

// DefaultTTL is the default cache expiry time.
const DefaultTTL = 1 * time.Hour

// Cache represents a file-based cache.
type Cache struct {
	medium  io.Medium
	baseDir string
	ttl     time.Duration
}

// Entry represents a cached item with metadata.
type Entry struct {
	Data      json.RawMessage `json:"data"`
	CachedAt  time.Time       `json:"cached_at"`
	ExpiresAt time.Time       `json:"expires_at"`
}

// New creates a new cache instance.
// If baseDir is empty, uses .core/cache in current directory.
// If m is nil, uses io.Local.
func New(m io.Medium, baseDir string, ttl time.Duration) (*Cache, error) {
	if m == nil {
		m = io.Local
	}

	if baseDir == "" {
		// Use .core/cache in current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		baseDir = filepath.Join(cwd, ".core", "cache")
	}

	if ttl == 0 {
		ttl = DefaultTTL
	}

	// Convert to absolute path for consistency
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, err
	}

	// Ensure cache directory exists
	if err := m.EnsureDir(absBaseDir); err != nil {
		return nil, err
	}

	baseDir = absBaseDir

	return &Cache{
		medium:  m,
		baseDir: baseDir,
		ttl:     ttl,
	}, nil
}

// Path returns the full path for a cache key.
func (c *Cache) Path(key string) string {
	return filepath.Join(c.baseDir, key+".json")
}

// Get retrieves a cached item if it exists and hasn't expired.
func (c *Cache) Get(key string, dest interface{}) (bool, error) {
	path := c.Path(key)

	content, err := c.medium.Read(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	data := []byte(content)

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		// Invalid cache file, treat as miss
		return false, nil
	}

	// Check expiry
	if time.Now().After(entry.ExpiresAt) {
		return false, nil
	}

	// Unmarshal the actual data
	if err := json.Unmarshal(entry.Data, dest); err != nil {
		return false, err
	}

	return true, nil
}

// Set stores an item in the cache.
func (c *Cache) Set(key string, data interface{}) error {
	path := c.Path(key)

	// Marshal the data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	entry := Entry{
		Data:      dataBytes,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
	}

	entryBytes, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	// medium.Write creates parent directories automatically
	return c.medium.Write(path, string(entryBytes))
}

// Delete removes an item from the cache.
func (c *Cache) Delete(key string) error {
	path := c.Path(key)
	err := c.medium.Delete(path)
	if errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err) {
		return nil
	}
	return err
}

// Clear removes all cached items.
func (c *Cache) Clear() error {
	return c.medium.DeleteAll(c.baseDir)
}

// Age returns how old a cached item is, or -1 if not cached.
func (c *Cache) Age(key string) time.Duration {
	path := c.Path(key)

	content, err := c.medium.Read(path)
	if err != nil {
		return -1
	}
	data := []byte(content)

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return -1
	}

	return time.Since(entry.CachedAt)
}

// GitHub-specific cache keys

// GitHubReposKey returns the cache key for an org's repo list.
func GitHubReposKey(org string) string {
	return filepath.Join("github", org, "repos")
}

// GitHubRepoKey returns the cache key for a specific repo's metadata.
func GitHubRepoKey(org, repo string) string {
	return filepath.Join("github", org, repo, "meta")
}
