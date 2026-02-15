package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	key := "test-key"
	value := []byte("test value")
	ttl := 1 * time.Hour

	if err := c.Set(key, value, ttl); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := c.Get(key, ttl)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got) != string(value) {
		t.Errorf("expected %s, got %s", value, got)
	}
}

func TestCache_GetExpired(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	key := "expire-test"
	value := []byte("old value")

	// Write cache file with old mtime
	path := c.path(key)
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, value, 0644)

	// Set mtime to 2 hours ago
	oldTime := time.Now().Add(-2 * time.Hour)
	os.Chtimes(path, oldTime, oldTime)

	// Try to get with 1 hour TTL
	_, err := c.Get(key, 1*time.Hour)
	if err == nil {
		t.Fatal("expected error for expired cache, got nil")
	}
}
