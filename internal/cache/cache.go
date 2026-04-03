// internal/cache/cache.go
package cache

// Cache handles local data caching
type Cache struct {
	// TODO: Implement SQLite cache
}

// New creates a new cache instance
func New(dbPath string) (*Cache, error) {
	return &Cache{}, nil
}
