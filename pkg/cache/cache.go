package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

// Cache handles the caching of artifacts.
type Cache struct {
	baseDir string
}

// New creates a new Cache instance.
// The baseDir must be provided and should already have its default handled by the caller (e.g., config package).
func New(baseDir string) (*Cache, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("cache: base directory cannot be empty")
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("cache: error creating cache directory %s: %w", baseDir, err)
	}

	return &Cache{baseDir: baseDir}, nil
}

// GetPath returns the full path for a given artifact name in the cache.
func (c *Cache) GetPath(artifactName string) string {
	return filepath.Join(c.baseDir, artifactName)
}

// Has checks if an artifact exists in the cache.
// For directories (like git repos), it checks for the existence of a specific file within them
// to confirm validity (e.g., a ".git" directory). If innerPath is empty, it just checks the artifact path itself.
func (c *Cache) Has(artifactName string, innerPath string) bool {
	path := c.GetPath(artifactName)
	if innerPath != "" {
		path = filepath.Join(path, innerPath)
	}
	_, err := os.Stat(path)
	return err == nil
}
