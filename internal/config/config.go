// internal/config/config.go
package config

// Config holds application configuration
type Config struct {
	DefaultProject string             `yaml:"default_project"`
	Auth           AuthConfig         `yaml:"auth"`
	Projects       map[string]Project `yaml:"projects"`
	Cache          CacheConfig        `yaml:"cache"`
}

type AuthConfig struct {
	Email    string `yaml:"email"`
	APIToken string `yaml:"api_token"`
}

type Project struct {
	JiraURL         string            `yaml:"jira_url"`
	BoardID         int               `yaml:"board_id"`
	MultiOwnerField string            `yaml:"multi_owner_field"`
	IssueTypes      map[string]string `yaml:"issue_types"`
}

type CacheConfig struct {
	Enabled    bool   `yaml:"enabled"`
	DefaultTTL string `yaml:"default_ttl"`
	Location   string `yaml:"location"`
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	// TODO: Implement config loading
	return &Config{}, nil
}
