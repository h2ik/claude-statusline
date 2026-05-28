package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	dir string
}

func New(dir string) *Cache {
	return &Cache{dir: dir}
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) error {
	path := c.path(key)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	if err := os.WriteFile(path, value, 0644); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

func (c *Cache) Get(key string, ttl time.Duration) ([]byte, error) {
	path := c.path(key)

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}

	age := time.Since(info.ModTime())
	if age > ttl {
		return nil, fmt.Errorf("cache expired (age: %v, ttl: %v)", age, ttl)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return data, nil
}

// Prune removes cache files older than maxAge. Errors on individual
// files are silently ignored so a single permission issue doesn't
// prevent cleanup of the rest.
func (c *Cache) Prune(maxAge time.Duration) error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("read cache dir: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(c.dir, entry.Name()))
		}
	}
	return nil
}

// Clear removes every file in the cache directory and returns the number
// of files removed. A missing cache directory is treated as already empty
// (count 0, no error).
func (c *Cache) Clear() (int, error) {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read cache dir: %w", err)
	}

	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := os.Remove(filepath.Join(c.dir, entry.Name())); err != nil {
			return removed, fmt.Errorf("remove %s: %w", entry.Name(), err)
		}
		removed++
	}
	return removed, nil
}

func (c *Cache) path(key string) string {
	hash := sha256.Sum256([]byte(key))
	filename := hex.EncodeToString(hash[:])
	return filepath.Join(c.dir, filename)
}
