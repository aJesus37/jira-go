// internal/cache/cache_test.go
package cache

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cache, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cache.Close() //nolint:errcheck

	// Test Set and Get
	t.Run("Set and Get", func(t *testing.T) {
		data := []byte(`{"key": "value"}`)
		if err := cache.Set("test-key", data, 5*time.Minute); err != nil {
			t.Errorf("Set() error = %v", err)
		}

		got, err := cache.Get("test-key")
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}

		if string(got) != string(data) {
			t.Errorf("Get() = %v, want %v", string(got), string(data))
		}
	})

	// Test expiration
	t.Run("Expiration", func(t *testing.T) {
		data := []byte(`{"temp": true}`)
		if err := cache.Set("expire-key", data, 1*time.Millisecond); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		_, err := cache.Get("expire-key")
		if err == nil {
			t.Error("Expected error for expired key")
		}
	})
}
