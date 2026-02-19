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

func TestCache_Prune_RemovesOldFiles(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	// Create a cache entry and backdate it to 31 days ago
	_ = c.Set("old-key", []byte("stale"), 0)
	oldPath := c.path("old-key")
	oldTime := time.Now().Add(-31 * 24 * time.Hour)
	_ = os.Chtimes(oldPath, oldTime, oldTime)

	// Create a fresh cache entry
	_ = c.Set("new-key", []byte("fresh"), 0)

	if err := c.Prune(30 * 24 * time.Hour); err != nil {
		t.Fatalf("Prune failed: %v", err)
	}

	// Old file should be gone
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("expected old cache file to be removed")
	}

	// Fresh file should still exist
	newPath := c.path("new-key")
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("expected new cache file to still exist, got: %v", err)
	}
}

func TestCache_Prune_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	if err := c.Prune(30 * 24 * time.Hour); err != nil {
		t.Fatalf("Prune on empty dir failed: %v", err)
	}
}

func TestCache_Prune_MissingDir(t *testing.T) {
	c := New("/nonexistent/cache/dir")

	err := c.Prune(30 * 24 * time.Hour)
	if err == nil {
		t.Fatal("expected error for missing dir, got nil")
	}
}

func TestCache_GetExpired(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	key := "expire-test"
	value := []byte("old value")

	// Write cache file with old mtime
	path := c.path(key)
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	_ = os.WriteFile(path, value, 0644)

	// Set mtime to 2 hours ago
	oldTime := time.Now().Add(-2 * time.Hour)
	_ = os.Chtimes(path, oldTime, oldTime)

	// Try to get with 1 hour TTL
	_, err := c.Get(key, 1*time.Hour)
	if err == nil {
		t.Fatal("expected error for expired cache, got nil")
	}
}
