package cache_test

import (
	"testing"
	"time"

	"github.com/host-uk/core/pkg/cache"
	"github.com/host-uk/core/pkg/io"
)

func TestCache(t *testing.T) {
	m := io.NewMockMedium()
	// Use a path that MockMedium will understand
	baseDir := "/tmp/cache"
	c, err := cache.New(m, baseDir, 1*time.Minute)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	key := "test-key"
	data := map[string]string{"foo": "bar"}

	// Test Set
	if err := c.Set(key, data); err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// Test Get
	var retrieved map[string]string
	found, err := c.Get(key, &retrieved)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if !found {
		t.Error("expected to find cached item")
	}
	if retrieved["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", retrieved["foo"])
	}

	// Test Age
	age := c.Age(key)
	if age < 0 {
		t.Error("expected age >= 0")
	}

	// Test Delete
	if err := c.Delete(key); err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	found, _ = c.Get(key, &retrieved)
	if found {
		t.Error("expected item to be deleted")
	}

	// Test Expiry
	cshort, _ := cache.New(m, "/tmp/cache-short", 10*time.Millisecond)
	_ = cshort.Set(key, data)
	time.Sleep(50 * time.Millisecond)
	found, _ = cshort.Get(key, &retrieved)
	if found {
		t.Error("expected item to be expired")
	}

	// Test Clear
	_ = c.Set("key1", data)
	_ = c.Set("key2", data)
	if err := c.Clear(); err != nil {
		t.Errorf("Clear failed: %v", err)
	}
	found, _ = c.Get("key1", &retrieved)
	if found {
		t.Error("expected key1 to be cleared")
	}
}

func TestCacheDefaults(t *testing.T) {
	// Test default Medium (io.Local) and default TTL
	// We won't actually write to disk if we can avoid it, but let's check constructor
	c, err := cache.New(nil, "", 0)
	if err != nil {
		t.Fatalf("failed to create cache with defaults: %v", err)
	}
	if c == nil {
		t.Fatal("expected cache instance")
	}
	// We can't easily check c.medium since it's unexported, but we can check behavior if needed
}
