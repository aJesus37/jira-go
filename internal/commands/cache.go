// internal/commands/cache.go
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aJesus37/jira-go/internal/cache"
	"github.com/aJesus37/jira-go/internal/config"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage local cache",
	Long:  `View cache status, clear cache, and configure caching options.`,
}

var cacheStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cache status",
	RunE:  runCacheStatus,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cached data",
	RunE:  runCacheClear,
}

var cachePathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show cache file path",
	RunE:  runCachePath,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheStatusCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cachePathCmd)
}

func getCachePath() string {
	if envPath := os.Getenv("JIRA_GO_CACHE"); envPath != "" {
		return envPath
	}

	cfg, err := config.Load()
	if err == nil && cfg.Cache.Location != "" {
		return cfg.Cache.Location
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "jira-go", "cache.db")
}

func runCacheStatus(cmd *cobra.Command, args []string) error {
	cachePath := getCachePath()

	fmt.Printf("Cache file: %s\n", cachePath)

	// Check if cache exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		fmt.Println("Status: Not initialized (cache will be created on first use)")
		return nil
	}

	// Open and get stats
	c, err := cache.New(cachePath)
	if err != nil {
		return fmt.Errorf("opening cache: %w", err)
	}
	defer c.Close() //nolint:errcheck

	total, expired, err := c.Stats()
	if err != nil {
		return fmt.Errorf("getting stats: %w", err)
	}

	fmt.Printf("Status: Active\n")
	fmt.Printf("Total entries: %d\n", total)
	fmt.Printf("Expired entries: %d\n", expired)

	return nil
}

func runCacheClear(cmd *cobra.Command, args []string) error {
	cachePath := getCachePath()

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		fmt.Println("Cache is already empty")
		return nil
	}

	c, err := cache.New(cachePath)
	if err != nil {
		return fmt.Errorf("opening cache: %w", err)
	}
	defer c.Close()

	if err := c.Clear(); err != nil {
		return fmt.Errorf("clearing cache: %w", err)
	}

	fmt.Println("✓ Cache cleared successfully")
	return nil
}

func runCachePath(cmd *cobra.Command, args []string) error {
	fmt.Println(getCachePath())
	return nil
}
