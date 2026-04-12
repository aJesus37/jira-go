// internal/cache/cache.go
package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Cache handles local data caching with SQLite
type Cache struct {
	db *sql.DB
}

// New creates a new cache instance
func New(dbPath string) (*Cache, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	cache := &Cache{db: db}

	if err := cache.createTables(); err != nil {
		db.Close() //nolint:errcheck
		return nil, fmt.Errorf("creating tables: %w", err)
	}

	return cache, nil
}

// createTables creates the necessary database tables
func (c *Cache) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS cache (
		key TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_expires ON cache(expires_at);

	CREATE TABLE IF NOT EXISTS user_mappings (
		email TEXT PRIMARY KEY,
		account_id TEXT NOT NULL,
		expires_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_user_expires ON user_mappings(expires_at);
	`

	_, err := c.db.Exec(query)
	return err
}

// Set stores data in the cache with a TTL
func (c *Cache) Set(key string, data []byte, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)

	query := `
	INSERT INTO cache (key, data, expires_at)
	VALUES (?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET
		data = excluded.data,
		expires_at = excluded.expires_at
	`

	_, err := c.db.Exec(query, key, data, expiresAt)
	return err
}

// Get retrieves data from the cache
func (c *Cache) Get(key string) ([]byte, error) {
	// Clean expired entries first (ignore cleanup errors; best-effort)
	_ = c.Cleanup()

	var data []byte
	query := `SELECT data FROM cache WHERE key = ? AND expires_at > ?`
	err := c.db.QueryRow(query, key, time.Now()).Scan(&data)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cache miss for key: %s", key)
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Delete removes an entry from the cache
func (c *Cache) Delete(key string) error {
	_, err := c.db.Exec("DELETE FROM cache WHERE key = ?", key)
	return err
}

// Cleanup removes expired entries
func (c *Cache) Cleanup() error {
	if _, err := c.db.Exec("DELETE FROM cache WHERE expires_at <= ?", time.Now()); err != nil {
		return err
	}
	_, err := c.db.Exec("DELETE FROM user_mappings WHERE expires_at <= ?", time.Now())
	return err
}

// SetUserMapping caches email to accountId mapping
func (c *Cache) SetUserMapping(email, accountID string, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)

	query := `
	INSERT INTO user_mappings (email, account_id, expires_at)
	VALUES (?, ?, ?)
	ON CONFLICT(email) DO UPDATE SET
		account_id = excluded.account_id,
		expires_at = excluded.expires_at
	`

	_, err := c.db.Exec(query, email, accountID, expiresAt)
	return err
}

// GetUserMapping retrieves accountId from cache
func (c *Cache) GetUserMapping(email string) (string, error) {
	var accountID string
	query := `SELECT account_id FROM user_mappings WHERE email = ? AND expires_at > ?`
	err := c.db.QueryRow(query, email, time.Now()).Scan(&accountID)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("cache miss for email: %s", email)
	}
	if err != nil {
		return "", err
	}

	return accountID, nil
}

// Close closes the database connection
func (c *Cache) Close() error {
	return c.db.Close()
}

// Clear removes all cached data
func (c *Cache) Clear() error {
	_, err := c.db.Exec("DELETE FROM cache")
	if err != nil {
		return err
	}
	_, err = c.db.Exec("DELETE FROM user_mappings")
	return err
}

// Stats returns cache statistics
func (c *Cache) Stats() (totalEntries int, expiredEntries int, err error) {
	err = c.db.QueryRow("SELECT COUNT(*) FROM cache").Scan(&totalEntries)
	if err != nil {
		return 0, 0, err
	}

	err = c.db.QueryRow("SELECT COUNT(*) FROM cache WHERE expires_at <= ?", time.Now()).Scan(&expiredEntries)
	if err != nil {
		return 0, 0, err
	}

	return totalEntries, expiredEntries, nil
}
