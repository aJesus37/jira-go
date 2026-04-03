// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configFileName = "config.yaml"
)

// Config holds application configuration
type Config struct {
	DefaultProject string             `yaml:"default_project"`
	Auth           AuthConfig         `yaml:"auth"`
	Projects       map[string]Project `yaml:"projects"`
	Cache          CacheConfig        `yaml:"cache"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Email    string `yaml:"email"`
	APIToken string `yaml:"api_token"`
}

// Project holds project-specific configuration
type Project struct {
	JiraURL         string            `yaml:"jira_url"`
	BoardID         int               `yaml:"board_id"`
	MultiOwnerField string            `yaml:"multi_owner_field"`
	SprintField     string            `yaml:"sprint_field"`
	IssueTypes      map[string]string `yaml:"issue_types"`
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	Enabled    bool   `yaml:"enabled"`
	DefaultTTL string `yaml:"default_ttl"`
	Location   string `yaml:"location"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	if envPath := os.Getenv("JIRA_GO_CONFIG"); envPath != "" {
		return envPath
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".config", "jira-go", configFileName)
}

// GetCachePath returns the default cache directory path
func GetCachePath() string {
	if envPath := os.Getenv("JIRA_GO_CACHE"); envPath != "" {
		return envPath
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "jira-go-cache.db")
	}

	return filepath.Join(home, ".cache", "jira-go", "cache.db")
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	configPath := GetConfigPath()

	cfg := &Config{
		Cache: CacheConfig{
			Enabled:    true,
			DefaultTTL: "30m",
			Location:   GetCachePath(),
		},
	}

	// Load from file if exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("reading config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	// Override with environment variables
	cfg.overrideFromEnv()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Auth.Email == "" {
		return fmt.Errorf("email is required (set via config or JIRA_GO_EMAIL env var)")
	}

	if c.Auth.APIToken == "" {
		return fmt.Errorf("API token is required (set via config or JIRA_GO_API_TOKEN env var)")
	}

	if c.DefaultProject == "" {
		return fmt.Errorf("default project is required (set via config or JIRA_GO_DEFAULT_PROJECT env var)")
	}

	if _, err := c.GetProject(c.DefaultProject); err != nil {
		return fmt.Errorf("default project '%s' not found in projects configuration", c.DefaultProject)
	}

	project, _ := c.GetProject(c.DefaultProject)
	if project.JiraURL == "" {
		return fmt.Errorf("Jira URL is required for project %s", c.DefaultProject)
	}

	return nil
}

// overrideFromEnv overrides config with environment variables
func (c *Config) overrideFromEnv() {
	if email := os.Getenv("JIRA_GO_EMAIL"); email != "" {
		c.Auth.Email = email
	}

	if token := os.Getenv("JIRA_GO_API_TOKEN"); token != "" {
		c.Auth.APIToken = token
	}

	if project := os.Getenv("JIRA_GO_DEFAULT_PROJECT"); project != "" {
		c.DefaultProject = project
	}
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := GetConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// GetProject returns project configuration by key
func (c *Config) GetProject(key string) (*Project, error) {
	if project, ok := c.Projects[key]; ok {
		return &project, nil
	}
	return nil, fmt.Errorf("project %s not found in config", key)
}
