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
	found, err = c.Get(key, &retrieved)
	if err != nil {
		t.Errorf("Get after delete returned an unexpected error: %v", err)
	}
	if found {
		t.Error("expected item to be deleted")
	}

	// Test Expiry
	cshort, err := cache.New(m, "/tmp/cache-short", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create short-lived cache: %v", err)
	}
	if err := cshort.Set(key, data); err != nil {
		t.Fatalf("Set for expiry test failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	found, err = cshort.Get(key, &retrieved)
	if err != nil {
		t.Errorf("Get for expired item returned an unexpected error: %v", err)
	}
	if found {
		t.Error("expected item to be expired")
	}

	// Test Clear
	if err := c.Set("key1", data); err != nil {
		t.Fatalf("Set for clear test failed for key1: %v", err)
	}
	if err := c.Set("key2", data); err != nil {
		t.Fatalf("Set for clear test failed for key2: %v", err)
	}
	if err := c.Clear(); err != nil {
		t.Errorf("Clear failed: %v", err)
	}
	found, err = c.Get("key1", &retrieved)
	if err != nil {
		t.Errorf("Get after clear returned an unexpected error: %v", err)
	}
	if found {
		t.Error("expected key1 to be cleared")
	}
}

func TestCacheDefaults(t *testing.T) {
	// Test default Medium (io.Local) and default TTL
	c, err := cache.New(nil, "", 0)
	if err != nil {
		t.Fatalf("failed to create cache with defaults: %v", err)
	}
	if c == nil {
		t.Fatal("expected cache instance")
	}
}
