// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create temp config dir
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "jira-go")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `
default_project: PROJ
auth:
  email: test@example.com
  api_token: test-token
projects:
  PROJ:
    jira_url: https://test.atlassian.net
    board_id: 1
    multi_owner_field: customfield_10001
`
	os.WriteFile(configPath, []byte(configContent), 0600)

	// Set config path env var
	os.Setenv("JIRA_GO_CONFIG", configPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DefaultProject != "PROJ" {
		t.Errorf("DefaultProject = %v, want PROJ", cfg.DefaultProject)
	}

	if cfg.Auth.Email != "test@example.com" {
		t.Errorf("Auth.Email = %v, want test@example.com", cfg.Auth.Email)
	}
}
