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

func (c *Cache) path(key string) string {
	hash := sha256.Sum256([]byte(key))
	filename := hex.EncodeToString(hash[:])
	return filepath.Join(c.dir, filename)
}
