# Jira CLI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a comprehensive Go CLI for Jira Software with task/sprint/epic management, agile ceremonies, multi-owner support, and rich TUI using Charmbracelet libraries.

**Architecture:** Modular monolithic CLI with clean separation between commands, services (API client, cache, config), and TUI components. SQLite for caching, YAML+env for config, Cobra for CLI structure, Bubble Tea for TUI.

**Tech Stack:** Go 1.21+, Cobra, Charmbracelet (Bubble Tea, Lipgloss, Bubbles, Glow), SQLite, Jira REST API v3

---

## Phase 1: Project Setup & Foundation

### Task 1: Initialize Go Module

**Files:**
- Create: `go.mod`
- Create: `go.sum` (generated)
- Create: `.gitignore`
- Create: `Makefile`

**Step 1: Initialize Go module**

```bash
cd /home/jesus/Projects/jira-go
go mod init github.com/user/jira-go
```

**Step 2: Create .gitignore**

```bash
cat > .gitignore << 'EOF'
# Binaries
jira-go
*.exe
*.dll
*.so
*.dylib

# Test binaries
*.test

# Output of go coverage
*.out

# Dependency directories
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Local config (may contain secrets)
/config.yaml
/.config/

# Cache
/.cache/
*.db

# Environment
.env
.env.local
EOF
```

**Step 3: Create Makefile**

```makefile
.PHONY: build test clean install lint

BINARY_NAME=jira-go
BUILD_DIR=./build

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/jira-go

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

lint:
	golangci-lint run

dev:
	go run ./cmd/jira-go

.DEFAULT_GOAL := build
```

**Step 4: Initial commit**

```bash
git add go.mod .gitignore Makefile
git commit -m "chore: initialize Go module with build tooling"
```

---

### Task 2: Create Project Structure

**Files:**
- Create: `cmd/jira-go/main.go`
- Create: `internal/config/config.go`
- Create: `internal/api/client.go`
- Create: `internal/cache/cache.go`
- Create: `internal/models/models.go`
- Create: `internal/tui/tui.go`
- Create: `pkg/utils/utils.go`
- Create: `docs/README.md`

**Step 1: Create directory structure**

```bash
mkdir -p cmd/jira-go
mkdir -p internal/{config,api,cache,models,tui,commands}
mkdir -p pkg/utils
mkdir -p docs
```

**Step 2: Create main.go**

```go
// cmd/jira-go/main.go
package main

import (
	"fmt"
	"os"

	"github.com/user/jira-go/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Create placeholder files**

```go
// internal/config/config.go
package config

// Config holds application configuration
type Config struct {
	DefaultProject string               `yaml:"default_project"`
	Auth           AuthConfig           `yaml:"auth"`
	Projects       map[string]Project   `yaml:"projects"`
	Cache          CacheConfig          `yaml:"cache"`
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
```

```go
// internal/models/models.go
package models

import "time"

// Issue represents a Jira issue
type Issue struct {
	Key         string    `json:"key"`
	ID          string    `json:"id"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Assignee    *User     `json:"assignee"`
	Owners      []User    `json:"owners"` // Multi-owner support
	Reporter    *User     `json:"reporter"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Labels      []string  `json:"labels"`
}

// User represents a Jira user
type User struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	Email       string `json:"emailAddress"`
	AvatarURL   string `json:"avatarUrls"`
}

// Sprint represents a Jira sprint
type Sprint struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	State        string    `json:"state"` // future, active, closed
	StartDate    time.Time `json:"startDate"`
	EndDate      time.Time `json:"endDate"`
	CompleteDate time.Time `json:"completeDate"`
}
```

```go
// internal/api/client.go
package api

import (
	"net/http"
	"time"

	"github.com/user/jira-go/internal/config"
)

// Client wraps the Jira API client
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Email      string
	APIToken   string
}

// NewClient creates a new Jira API client
func NewClient(cfg *config.Config, projectKey string) (*Client, error) {
	// TODO: Implement client creation
	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}
```

```go
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
```

```go
// internal/tui/tui.go
package tui

// TUI handles terminal user interface
type TUI struct {
	// TODO: Implement Bubble Tea TUI
}

// New creates a new TUI instance
func New() *TUI {
	return &TUI{}
}
```

```go
// internal/commands/root.go
package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jira-go",
	Short: "A CLI tool for managing Jira Software projects",
	Long: `jira-go is a comprehensive CLI for Jira Software that supports
task management, sprint operations, epics, and agile ceremonies.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// TODO: Add subcommands
}
```

**Step 4: Commit structure**

```bash
git add .
git commit -m "chore: create project structure with placeholder packages"
```

---

## Phase 2: Configuration Management

### Task 3: Implement Config Package

**Files:**
- Modify: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write failing test**

```go
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
```

**Step 2: Run test to verify failure**

```bash
go test ./internal/config -v
```

Expected: FAIL - Load function not fully implemented

**Step 3: Implement config loading**

```go
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

// Load loads configuration from file and environment
func Load() (*Config, error) {
	configPath := GetConfigPath()
	
	cfg := &Config{
		Cache: CacheConfig{
			Enabled:    true,
			DefaultTTL: "30m",
			Location:   filepath.Join(os.TempDir(), "jira-go-cache.db"),
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
	
	return cfg, nil
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
```

**Step 4: Add yaml dependency and run tests**

```bash
go get gopkg.in/yaml.v3
go test ./internal/config -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat(config): implement configuration loading with env override"
```

---

### Task 4: Implement Init Command

**Files:**
- Create: `internal/commands/init.go`
- Modify: `internal/commands/root.go`

**Step 1: Create init command**

```go
// internal/commands/init.go
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/config"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize jira-go configuration",
	Long:  `Interactive setup to configure jira-go with your Jira instance.`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("🚀 Welcome to jira-go!")
	fmt.Println("Let's set up your configuration.\n")
	
	// Jira URL
	fmt.Print("Jira URL (e.g., https://your-domain.atlassian.net): ")
	jiraURL, _ := reader.ReadString('\n')
	jiraURL = strings.TrimSpace(jiraURL)
	
	// Email
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	
	// API Token (hidden input)
	fmt.Print("API Token (input hidden): ")
	apiTokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("reading API token: %w", err)
	}
	apiToken := string(apiTokenBytes)
	fmt.Println()
	
	// Default Project
	fmt.Print("Default Project Key (e.g., PROJ): ")
	projectKey, _ := reader.ReadString('\n')
	projectKey = strings.TrimSpace(projectKey)
	
	// Multi-owner field (optional)
	fmt.Print("Multi-owner Custom Field ID (optional, e.g., customfield_10001): ")
	multiOwnerField, _ := reader.ReadString('\n')
	multiOwnerField = strings.TrimSpace(multiOwnerField)
	
	// Create config
	cfg := &config.Config{
		DefaultProject: projectKey,
		Auth: config.AuthConfig{
			Email:    email,
			APIToken: apiToken,
		},
		Projects: map[string]config.Project{
			projectKey: {
				JiraURL:         jiraURL,
				MultiOwnerField: multiOwnerField,
			},
		},
		Cache: config.CacheConfig{
			Enabled:    true,
			DefaultTTL: "30m",
		},
	}
	
	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	
	fmt.Printf("\n✓ Configuration saved to %s\n", config.GetConfigPath())
	fmt.Println("✓ You're ready to use jira-go!")
	fmt.Println("\nTry: jira-go --help")
	
	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
```

**Step 2: Add term dependency**

```bash
go get golang.org/x/term
```

**Step 3: Update root.go to ensure init is registered**

The init() function in init.go should already register the command. Verify imports.

**Step 4: Build and test**

```bash
go build -o build/jira-go ./cmd/jira-go
./build/jira-go init
```

Expected: Interactive prompts, config saved

**Step 5: Commit**

```bash
git add internal/commands/
git commit -m "feat(init): add interactive init command for configuration setup"
```

---

## Phase 3: Jira API Client

### Task 5: Implement Basic API Client

**Files:**
- Modify: `internal/api/client.go`
- Create: `internal/api/client_test.go`

**Step 1: Write test**

```go
// internal/api/client_test.go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/jira-go/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Email:    "test@example.com",
			APIToken: "test-token",
		},
		Projects: map[string]config.Project{
			"PROJ": {
				JiraURL: "https://test.atlassian.net",
			},
		},
	}
	
	client, err := NewClient(cfg, "PROJ")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	
	if client.BaseURL != "https://test.atlassian.net" {
		t.Errorf("BaseURL = %v, want https://test.atlassian.net", client.BaseURL)
	}
	
	if client.Email != "test@example.com" {
		t.Errorf("Email = %v, want test@example.com", client.Email)
	}
}

func TestDoRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("Expected Authorization header")
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key": "PROJ-1"}`))
	}))
	defer server.Close()
	
	client := &Client{
		HTTPClient: &http.Client{},
		BaseURL:    server.URL,
		Email:      "test@example.com",
		APIToken:   "test-token",
	}
	
	req, _ := http.NewRequest("GET", server.URL+"/rest/api/3/issue/PROJ-1", nil)
	resp, err := client.DoRequest(req)
	if err != nil {
		t.Fatalf("DoRequest() error = %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}
```

**Step 2: Run test to verify failure**

```bash
go test ./internal/api -v
```

**Step 3: Implement client**

```go
// internal/api/client.go
package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/user/jira-go/internal/config"
)

// Client wraps the Jira API client
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Email      string
	APIToken   string
}

// NewClient creates a new Jira API client
func NewClient(cfg *config.Config, projectKey string) (*Client, error) {
	project, err := cfg.GetProject(projectKey)
	if err != nil {
		return nil, err
	}
	
	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    project.JiraURL,
		Email:      cfg.Auth.Email,
		APIToken:   cfg.Auth.APIToken,
	}, nil
}

// DoRequest performs an HTTP request with authentication
func (c *Client) DoRequest(req *http.Request) (*http.Response, error) {
	// Add auth header
	auth := base64.StdEncoding.EncodeToString([]byte(c.Email + ":" + c.APIToken))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	return c.HTTPClient.Do(req)
}

// Get performs a GET request
func (c *Client) Get(path string) (*http.Response, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	return c.DoRequest(req)
}

// Post performs a POST request
func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	url := c.BaseURL + path
	
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}
	
	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, err
	}
	
	return c.DoRequest(req)
}

// Put performs a PUT request
func (c *Client) Put(path string, body interface{}) (*http.Response, error) {
	url := c.BaseURL + path
	
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	
	return c.DoRequest(req)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string) (*http.Response, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	
	return c.DoRequest(req)
}
```

**Step 4: Fix imports and run tests**

```go
// Add to imports in client.go
import "bytes"
```

```bash
go test ./internal/api -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/api/
git commit -m "feat(api): implement basic HTTP client with authentication"
```

---

### Task 6: Implement User Resolution (Email ↔ AccountID)

**Files:**
- Create: `internal/api/users.go`
- Create: `internal/api/users_test.go`

**Step 1: Create user resolution test**

```go
// internal/api/users_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveEmail(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/user/search" {
			t.Errorf("Expected path /rest/api/3/user/search, got %s", r.URL.Path)
		}
		
		query := r.URL.Query().Get("query")
		if query != "test@example.com" {
			t.Errorf("Expected query test@example.com, got %s", query)
		}
		
		users := []map[string]interface{}{
			{
				"accountId":   "12345",
				"displayName": "Test User",
				"emailAddress": "test@example.com",
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()
	
	client := &Client{
		HTTPClient: &http.Client{},
		BaseURL:    server.URL,
		Email:      "admin@example.com",
		APIToken:   "token",
	}
	
	user, err := client.ResolveEmail("test@example.com")
	if err != nil {
		t.Fatalf("ResolveEmail() error = %v", err)
	}
	
	if user.AccountID != "12345" {
		t.Errorf("AccountID = %v, want 12345", user.AccountID)
	}
}
```

**Step 2: Implement user resolution**

```go
// internal/api/users.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/user/jira-go/internal/models"
)

// ResolveEmail looks up a user by email and returns their details
func (c *Client) ResolveEmail(email string) (*models.User, error) {
	endpoint := "/rest/api/3/user/search"
	params := url.Values{}
	params.Set("query", email)
	
	resp, err := c.Get(endpoint + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("searching user: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user search failed: %s", resp.Status)
	}
	
	var users []models.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	
	for _, user := range users {
		if user.Email == email {
			return &user, nil
		}
	}
	
	return nil, fmt.Errorf("user with email %s not found", email)
}

// ResolveEmails resolves multiple emails to users
func (c *Client) ResolveEmails(emails []string) ([]models.User, error) {
	var users []models.User
	
	for _, email := range emails {
		user, err := c.ResolveEmail(email)
		if err != nil {
			return nil, fmt.Errorf("resolving %s: %w", email, err)
		}
		users = append(users, *user)
	}
	
	return users, nil
}

// GetCurrentUser returns the currently authenticated user
func (c *Client) GetCurrentUser() (*models.User, error) {
	resp, err := c.Get("/rest/api/3/myself")
	if err != nil {
		return nil, fmt.Errorf("getting current user: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get current user: %s", resp.Status)
	}
	
	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	
	return &user, nil
}
```

**Step 3: Run tests**

```bash
go test ./internal/api -v -run TestResolveEmail
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/api/users.go internal/api/users_test.go
git commit -m "feat(api): implement email to accountId resolution"
```

---

### Task 7: Implement Issue Operations

**Files:**
- Create: `internal/api/issues.go`
- Create: `internal/api/issues_test.go`

**Step 1: Create issue operations**

```go
// internal/api/issues.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/user/jira-go/internal/models"
)

// IssueSearchRequest represents a JQL search request
type IssueSearchRequest struct {
	JQL        string   `json:"jql"`
	StartAt    int      `json:"startAt,omitempty"`
	MaxResults int      `json:"maxResults,omitempty"`
	Fields     []string `json:"fields,omitempty"`
}

// IssueSearchResponse represents search results
type IssueSearchResponse struct {
	Total      int             `json:"total"`
	StartAt    int             `json:"startAt"`
	MaxResults int             `json:"maxResults"`
	Issues     []models.Issue  `json:"issues"`
}

// CreateIssue creates a new issue
func (c *Client) CreateIssue(projectKey, summary, description, issueType string) (*models.Issue, error) {
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": projectKey,
			},
			"summary":     summary,
			"description": description,
			"issuetype": map[string]string{
				"name": issueType,
			},
		},
	}
	
	resp, err := c.Post("/rest/api/3/issue", payload)
	if err != nil {
		return nil, fmt.Errorf("creating issue: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create issue failed: %s", resp.Status)
	}
	
	var result models.Issue
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	
	return &result, nil
}

// GetIssue retrieves an issue by key
func (c *Client) GetIssue(key string) (*models.Issue, error) {
	resp, err := c.Get(fmt.Sprintf("/rest/api/3/issue/%s", key))
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("issue %s not found", key)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get issue failed: %s", resp.Status)
	}
	
	var issue models.Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	
	return &issue, nil
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(jql string, startAt, maxResults int) (*IssueSearchResponse, error) {
	payload := IssueSearchRequest{
		JQL:        jql,
		StartAt:    startAt,
		MaxResults: maxResults,
		Fields:     []string{"summary", "status", "assignee", "created", "updated", "issuetype", "description", "labels"},
	}
	
	resp, err := c.Post("/rest/api/3/search", payload)
	if err != nil {
		return nil, fmt.Errorf("searching issues: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}
	
	var result IssueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	
	return &result, nil
}

// UpdateIssue updates an existing issue
func (c *Client) UpdateIssue(key string, fields map[string]interface{}) error {
	payload := map[string]interface{}{
		"fields": fields,
	}
	
	resp, err := c.Put(fmt.Sprintf("/rest/api/3/issue/%s", key), payload)
	if err != nil {
		return fmt.Errorf("updating issue: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("update issue failed: %s", resp.Status)
	}
	
	return nil
}

// DeleteIssue deletes an issue
func (c *Client) DeleteIssue(key string) error {
	resp, err := c.Delete(fmt.Sprintf("/rest/api/3/issue/%s", key))
	if err != nil {
		return fmt.Errorf("deleting issue: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete issue failed: %s", resp.Status)
	}
	
	return nil
}

// AssignIssue assigns an issue to a user
func (c *Client) AssignIssue(key, accountID string) error {
	payload := map[string]string{
		"accountId": accountID,
	}
	
	resp, err := c.Put(fmt.Sprintf("/rest/api/3/issue/%s/assignee", key), payload)
	if err != nil {
		return fmt.Errorf("assigning issue: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("assign issue failed: %s", resp.Status)
	}
	
	return nil
}

// UpdateMultiOwnerField updates the custom multi-owner field
func (c *Client) UpdateMultiOwnerField(key, fieldID string, accountIDs []string) error {
	// Build the multi-user picker value
	var users []map[string]string
	for _, id := range accountIDs {
		users = append(users, map[string]string{
			"accountId": id,
		})
	}
	
	fields := map[string]interface{}{
		fieldID: users,
	}
	
	return c.UpdateIssue(key, fields)
}
```

**Step 2: Run tests**

```bash
go build ./...
```

Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/api/issues.go
git commit -m "feat(api): implement CRUD operations for issues"
```

---

## Phase 4: Cache Implementation

### Task 8: Implement SQLite Cache

**Files:**
- Modify: `internal/cache/cache.go`
- Create: `internal/cache/cache_test.go`

**Step 1: Write cache test**

```go
// internal/cache/cache_test.go
package cache

import (
	"os"
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
	defer cache.Close()
	
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
		cache.Set("expire-key", data, 1*time.Millisecond)
		
		time.Sleep(10 * time.Millisecond)
		
		_, err := cache.Get("expire-key")
		if err == nil {
			t.Error("Expected error for expired key")
		}
	})
}
```

**Step 2: Implement cache with SQLite**

```go
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

// cacheEntry represents a cached item
type cacheEntry struct {
	Key       string
	Data      []byte
	ExpiresAt time.Time
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
		db.Close()
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
	// Clean expired entries first
	c.Cleanup()
	
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
	_, err := c.db.Exec("DELETE FROM cache WHERE expires_at <= ?", time.Now())
	_, err = c.db.Exec("DELETE FROM user_mappings WHERE expires_at <= ?", time.Now())
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
```

**Step 3: Add SQLite dependency and run tests**

```bash
go get github.com/mattn/go-sqlite3
go test ./internal/cache -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/cache/
git commit -m "feat(cache): implement SQLite-based caching with TTL support"
```

---

## Phase 5: Core Commands

### Task 9: Implement Task Commands

**Files:**
- Create: `internal/commands/task.go`

**Step 1: Create task command**

```go
// internal/commands/task.go
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/cache"
	"github.com/user/jira-go/internal/config"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage Jira tasks/issues",
	Long:  `Create, view, edit, and delete Jira tasks and issues.`,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE:  runTaskList,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	RunE:  runTaskCreate,
}

var taskViewCmd = &cobra.Command{
	Use:   "view [key]",
	Short: "View task details",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskView,
}

var taskEditCmd = &cobra.Command{
	Use:   "edit [key]",
	Short: "Edit a task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskEdit,
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskDelete,
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskViewCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	
	// List flags
	taskListCmd.Flags().String("project", "", "Project key (defaults to config)")
	taskListCmd.Flags().String("assignee", "", "Filter by assignee")
	taskListCmd.Flags().String("status", "", "Filter by status")
	taskListCmd.Flags().Int("limit", 25, "Maximum results")
	
	// Create flags
	taskCreateCmd.Flags().String("project", "", "Project key (defaults to config)")
	taskCreateCmd.Flags().String("type", "Task", "Issue type")
	taskCreateCmd.Flags().String("summary", "", "Issue summary (required)")
	taskCreateCmd.Flags().String("description", "", "Issue description")
	taskCreateCmd.Flags().String("assignee", "", "Assignee email")
	taskCreateCmd.Flags().String("owners", "", "Comma-separated owner emails")
	
	// Edit flags
	taskEditCmd.Flags().String("summary", "", "New summary")
	taskEditCmd.Flags().String("description", "", "New description")
	taskEditCmd.Flags().String("assignee", "", "New assignee email")
	taskEditCmd.Flags().String("owners", "", "Comma-separated owner emails")
}

func getProjectKey(cmd *cobra.Command, cfg *config.Config) string {
	if project, _ := cmd.Flags().GetString("project"); project != "" {
		return project
	}
	return cfg.DefaultProject
}

func runTaskList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	projectKey := getProjectKey(cmd, cfg)
	
	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	
	// Build JQL query
	jql := fmt.Sprintf("project = %s", projectKey)
	
	if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		jql += fmt.Sprintf(" AND assignee = '%s'", assignee)
	}
	
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		jql += fmt.Sprintf(" AND status = '%s'", status)
	}
	
	limit, _ := cmd.Flags().GetInt("limit")
	
	resp, err := client.SearchIssues(jql, 0, limit)
	if err != nil {
		return fmt.Errorf("searching issues: %w", err)
	}
	
	// Simple table output (will be replaced with TUI later)
	fmt.Printf("%-12s %-10s %-12s %s\n", "KEY", "TYPE", "STATUS", "SUMMARY")
	fmt.Println(strings.Repeat("-", 80))
	
	for _, issue := range resp.Issues {
		status := issue.Status
		if len(status) > 12 {
			status = status[:9] + "..."
		}
		
		summary := issue.Summary
		if len(summary) > 40 {
			summary = summary[:37] + "..."
		}
		
		fmt.Printf("%-12s %-10s %-12s %s\n", issue.Key, issue.Type, status, summary)
	}
	
	fmt.Printf("\nShowing %d of %d issues\n", len(resp.Issues), resp.Total)
	
	return nil
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	projectKey := getProjectKey(cmd, cfg)
	
	summary, _ := cmd.Flags().GetString("summary")
	if summary == "" {
		return fmt.Errorf("--summary is required")
	}
	
	description, _ := cmd.Flags().GetString("description")
	issueType, _ := cmd.Flags().GetString("type")
	
	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	
	// Create issue
	issue, err := client.CreateIssue(projectKey, summary, description, issueType)
	if err != nil {
		return fmt.Errorf("creating issue: %w", err)
	}
	
	fmt.Printf("✓ Created %s\n", issue.Key)
	
	// Handle assignee
	if assigneeEmail, _ := cmd.Flags().GetString("assignee"); assigneeEmail != "" {
		user, err := client.ResolveEmail(assigneeEmail)
		if err != nil {
			return fmt.Errorf("resolving assignee: %w", err)
		}
		
		if err := client.AssignIssue(issue.Key, user.AccountID); err != nil {
			return fmt.Errorf("assigning issue: %w", err)
		}
		
		fmt.Printf("✓ Assigned to %s\n", assigneeEmail)
	}
	
	// Handle multi-owners
	if ownersStr, _ := cmd.Flags().GetString("owners"); ownersStr != "" {
		project, _ := cfg.GetProject(projectKey)
		if project.MultiOwnerField == "" {
			return fmt.Errorf("multi_owner_field not configured for project %s", projectKey)
		}
		
		emails := strings.Split(ownersStr, ",")
		var accountIDs []string
		
		for _, email := range emails {
			email = strings.TrimSpace(email)
			user, err := client.ResolveEmail(email)
			if err != nil {
				return fmt.Errorf("resolving owner %s: %w", email, err)
			}
			accountIDs = append(accountIDs, user.AccountID)
		}
		
		if err := client.UpdateMultiOwnerField(issue.Key, project.MultiOwnerField, accountIDs); err != nil {
			return fmt.Errorf("updating owners: %w", err)
		}
		
		fmt.Printf("✓ Set owners: %s\n", ownersStr)
	}
	
	return nil
}

func runTaskView(cmd *cobra.Command, args []string) error {
	key := args[0]
	
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	
	issue, err := client.GetIssue(key)
	if err != nil {
		return err
	}
	
	fmt.Printf("\n%s: %s\n", issue.Key, issue.Summary)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Type: %s\n", issue.Type)
	fmt.Printf("Status: %s\n", issue.Status)
	if issue.Assignee != nil {
		fmt.Printf("Assignee: %s\n", issue.Assignee.DisplayName)
	}
	if len(issue.Owners) > 0 {
		var ownerNames []string
		for _, o := range issue.Owners {
			ownerNames = append(ownerNames, o.DisplayName)
		}
		fmt.Printf("Owners: %s\n", strings.Join(ownerNames, ", "))
	}
	fmt.Printf("Created: %s\n", issue.Created.Format("2006-01-02 15:04"))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("Description:")
	fmt.Println(issue.Description)
	fmt.Println()
	
	return nil
}

func runTaskEdit(cmd *cobra.Command, args []string) error {
	key := args[0]
	
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	
	fields := make(map[string]interface{})
	
	if summary, _ := cmd.Flags().GetString("summary"); summary != "" {
		fields["summary"] = summary
	}
	
	if description, _ := cmd.Flags().GetString("description"); description != "" {
		fields["description"] = description
	}
	
	if len(fields) > 0 {
		if err := client.UpdateIssue(key, fields); err != nil {
			return fmt.Errorf("updating issue: %w", err)
		}
		fmt.Printf("✓ Updated %s\n", key)
	}
	
	// Handle assignee
	if assigneeEmail, _ := cmd.Flags().GetString("assignee"); assigneeEmail != "" {
		user, err := client.ResolveEmail(assigneeEmail)
		if err != nil {
			return fmt.Errorf("resolving assignee: %w", err)
		}
		
		if err := client.AssignIssue(key, user.AccountID); err != nil {
			return fmt.Errorf("assigning issue: %w", err)
		}
		
		fmt.Printf("✓ Assigned to %s\n", assigneeEmail)
	}
	
	// Handle multi-owners
	if ownersStr, _ := cmd.Flags().GetString("owners"); ownersStr != "" {
		project, _ := cfg.GetProject(cfg.DefaultProject)
		if project.MultiOwnerField == "" {
			return fmt.Errorf("multi_owner_field not configured")
		}
		
		emails := strings.Split(ownersStr, ",")
		var accountIDs []string
		
		for _, email := range emails {
			email = strings.TrimSpace(email)
			user, err := client.ResolveEmail(email)
			if err != nil {
				return fmt.Errorf("resolving owner %s: %w", email, err)
			}
			accountIDs = append(accountIDs, user.AccountID)
		}
		
		if err := client.UpdateMultiOwnerField(key, project.MultiOwnerField, accountIDs); err != nil {
			return fmt.Errorf("updating owners: %w", err)
		}
		
		fmt.Printf("✓ Set owners: %s\n", ownersStr)
	}
	
	return nil
}

func runTaskDelete(cmd *cobra.Command, args []string) error {
	key := args[0]
	
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	
	if err := client.DeleteIssue(key); err != nil {
		return fmt.Errorf("deleting issue: %w", err)
	}
	
	fmt.Printf("✓ Deleted %s\n", key)
	return nil
}
```

**Step 2: Build and verify**

```bash
go build ./...
```

Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/commands/task.go
git commit -m "feat(commands): implement task CRUD operations with multi-owner support"
```

---

### Task 10: Implement Cache Management Commands

**Files:**
- Create: `internal/commands/cache.go`

**Step 1: Create cache command**

```go
// internal/commands/cache.go
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/cache"
	"github.com/user/jira-go/internal/config"
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
	defer c.Close()
	
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
```

**Step 2: Build and commit**

```bash
go build ./...
git add internal/commands/cache.go
git commit -m "feat(commands): add cache management commands (status, clear, path)"
```

---

## Phase 6: TUI Foundation (Charmbracelet)

### Task 11: Setup Bubble Tea Framework

**Files:**
- Modify: `internal/tui/tui.go`
- Create: `internal/tui/issue_list.go`

**Step 1: Add Bubble Tea dependencies**

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

**Step 2: Create base TUI structure**

```go
// internal/tui/tui.go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the base interface for all TUI models
type Model interface {
	tea.Model
	GetTitle() string
}

// Run starts a TUI program
func Run(model tea.Model) error {
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
```

**Step 3: Create issue list TUI**

```go
// internal/tui/issue_list.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/jira-go/internal/models"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)
	
	statusStyle = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)
	
	todoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	inProgressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00A8E8"))
	doneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00C851"))
	blockedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
)

// IssueItem represents a list item for an issue
type IssueItem struct {
	Issue models.Issue
}

func (i IssueItem) Title() string {
	return fmt.Sprintf("%s: %s", i.Issue.Key, i.Issue.Summary)
}

func (i IssueItem) Description() string {
	status := i.Issue.Status
	if i.Issue.Assignee != nil {
		status += " • " + i.Issue.Assignee.DisplayName
	}
	return status
}

func (i IssueItem) FilterValue() string {
	return i.Issue.Key + " " + i.Issue.Summary
}

// IssueListModel is the TUI model for listing issues
type IssueListModel struct {
	list   list.Model
	issues []models.Issue
	err    error
}

// NewIssueList creates a new issue list TUI
func NewIssueList(issues []models.Issue) IssueListModel {
	var items []list.Item
	for _, issue := range issues {
		items = append(items, IssueItem{Issue: issue})
	}
	
	l := list.New(items, list.NewDefaultDelegate(), 80, 20)
	l.Title = "Jira Issues"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	
	return IssueListModel{
		list:   l,
		issues: issues,
	}
}

func (m IssueListModel) Init() tea.Cmd {
	return nil
}

func (m IssueListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m IssueListModel) View() string {
	return m.list.View()
}

func (m IssueListModel) GetTitle() string {
	return "Issue List"
}

// GetSelected returns the currently selected issue
func (m IssueListModel) GetSelected() *models.Issue {
	if item, ok := m.list.SelectedItem().(IssueItem); ok {
		return &item.Issue
	}
	return nil
}
```

**Step 4: Build and commit**

```bash
go build ./...
git add internal/tui/
git commit -m "feat(tui): add Bubble Tea framework with issue list component"
```

---

### Task 12: Implement TUI Task List Command

**Files:**
- Create: `internal/commands/tui.go`

**Step 1: Create TUI command**

```go
// internal/commands/tui.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/config"
	"github.com/user/jira-go/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Long:  `Open the interactive terminal user interface for browsing and managing Jira.`,
}

var tuiListCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues in TUI",
	RunE:  runTUIList,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.AddCommand(tuiListCmd)
	
	tuiListCmd.Flags().String("project", "", "Project key (defaults to config)")
}

func runTUIList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	projectKey := getProjectKey(cmd, cfg)
	
	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	
	// Fetch issues
	jql := fmt.Sprintf("project = %s ORDER BY updated DESC", projectKey)
	resp, err := client.SearchIssues(jql, 0, 50)
	if err != nil {
		return fmt.Errorf("searching issues: %w", err)
	}
	
	// Launch TUI
	model := tui.NewIssueList(resp.Issues)
	return tui.Run(model)
}
```

**Step 2: Build and commit**

```bash
go build ./...
git add internal/commands/tui.go
git commit -m "feat(tui): add tui list command with Bubble Tea interface"
```

---

## Phase 7: Sprint and Project Commands

### Task 13: Implement Sprint Commands

**Files:**
- Create: `internal/commands/sprint.go`

**Step 1: Create sprint command with basic operations**

```go
// internal/commands/sprint.go
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/config"
)

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Manage Jira sprints",
	Long:  `List, create, and manage Jira sprints.`,
}

var sprintListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sprints",
	RunE:  runSprintList,
}

var sprintIssuesCmd = &cobra.Command{
	Use:   "issues [sprint-id]",
	Short: "List issues in a sprint",
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintIssues,
}

func init() {
	rootCmd.AddCommand(sprintCmd)
	sprintCmd.AddCommand(sprintListCmd)
	sprintCmd.AddCommand(sprintIssuesCmd)
}

func runSprintList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	projectKey := cfg.DefaultProject
	project, err := cfg.GetProject(projectKey)
	if err != nil {
		return fmt.Errorf("getting project: %w", err)
	}
	
	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	
	// Fetch sprints from board
	resp, err := client.Get(fmt.Sprintf("/rest/agile/1.0/board/%d/sprint", project.BoardID))
	if err != nil {
		return fmt.Errorf("fetching sprints: %w", err)
	}
	defer resp.Body.Close()
	
	// For now, just show raw response
	// TODO: Parse and display properly
	fmt.Printf("Fetching sprints for board %d...\n", project.BoardID)
	
	return nil
}

func runSprintIssues(cmd *cobra.Command, args []string) error {
	// TODO: Implement
	fmt.Println("Sprint issues not yet implemented")
	return nil
}
```

**Step 2: Commit**

```bash
git add internal/commands/sprint.go
git commit -m "feat(commands): add sprint command structure"
```

---

### Task 14: Implement Project Commands

**Files:**
- Create: `internal/commands/project.go`

**Step 1: Create project command**

```go
// internal/commands/project.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/config"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `Switch between projects and view project configuration.`,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured projects",
	RunE:  runProjectList,
}

var projectSwitchCmd = &cobra.Command{
	Use:   "switch [project-key]",
	Short: "Switch default project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectSwitch,
}

var projectConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "View current project configuration",
	RunE:  runProjectConfig,
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectSwitchCmd)
	projectCmd.AddCommand(projectConfigCmd)
}

func runProjectList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	fmt.Printf("Current default: %s\n\n", cfg.DefaultProject)
	fmt.Println("Configured projects:")
	fmt.Println("-------------------")
	
	for key, project := range cfg.Projects {
		prefix := "  "
		if key == cfg.DefaultProject {
			prefix = "* "
		}
		fmt.Printf("%s%s\n", prefix, key)
		fmt.Printf("  URL: %s\n", project.JiraURL)
		fmt.Printf("  Board ID: %d\n", project.BoardID)
		if project.MultiOwnerField != "" {
			fmt.Printf("  Multi-owner field: %s\n", project.MultiOwnerField)
		}
		fmt.Println()
	}
	
	return nil
}

func runProjectSwitch(cmd *cobra.Command, args []string) error {
	newProject := args[0]
	
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	// Verify project exists
	if _, err := cfg.GetProject(newProject); err != nil {
		return fmt.Errorf("project %s not found in config", newProject)
	}
	
	cfg.DefaultProject = newProject
	
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	
	fmt.Printf("✓ Switched to project %s\n", newProject)
	return nil
}

func runProjectConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	
	project, err := cfg.GetProject(cfg.DefaultProject)
	if err != nil {
		return err
	}
	
	fmt.Printf("Project: %s\n", cfg.DefaultProject)
	fmt.Printf("Jira URL: %s\n", project.JiraURL)
	fmt.Printf("Board ID: %d\n", project.BoardID)
	fmt.Printf("Multi-owner field: %s\n", project.MultiOwnerField)
	
	return nil
}
```

**Step 2: Build and commit**

```bash
go build ./...
git add internal/commands/project.go
git commit -m "feat(commands): add project management commands (list, switch, config)"
```

---

## Phase 8: Ceremony Commands (Foundation)

### Task 15: Implement Ceremony Command Structure

**Files:**
- Create: `internal/commands/ceremony.go`

**Step 1: Create ceremony command with subcommands**

```go
// internal/commands/ceremony.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ceremonyCmd = &cobra.Command{
	Use:   "ceremony",
	Short: "Run agile ceremonies",
	Long:  `Interactive tools for sprint planning, retrospectives, and daily standups.`,
}

var ceremonyPlanningCmd = &cobra.Command{
	Use:   "planning",
	Short: "Run sprint planning ceremony",
	RunE:  runCeremonyPlanning,
}

var ceremonyRetroCmd = &cobra.Command{
	Use:   "retro",
	Short: "Run retrospective ceremony",
	RunE:  runCeremonyRetro,
}

var ceremonyDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Run daily standup",
	RunE:  runCeremonyDaily,
}

func init() {
	rootCmd.AddCommand(ceremonyCmd)
	ceremonyCmd.AddCommand(ceremonyPlanningCmd)
	ceremonyCmd.AddCommand(ceremonyRetroCmd)
	ceremonyCmd.AddCommand(ceremonyDailyCmd)
}

func runCeremonyPlanning(cmd *cobra.Command, args []string) error {
	fmt.Println("🎯 Sprint Planning")
	fmt.Println("This will launch an interactive planning session.")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  • View and sort backlog")
	fmt.Println("  • Assign issues to sprint")
	fmt.Println("  • Story point estimation")
	fmt.Println("  • Export planning notes")
	fmt.Println()
	fmt.Println("Not yet implemented - coming soon!")
	return nil
}

func runCeremonyRetro(cmd *cobra.Command, args []string) error {
	fmt.Println("📝 Retrospective")
	fmt.Println("This will launch an interactive retrospective session.")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  • Anonymous card submission")
	fmt.Println("  • Voting and grouping")
	fmt.Println("  • Export action items")
	fmt.Println()
	fmt.Println("Not yet implemented - coming soon!")
	return nil
}

func runCeremonyDaily(cmd *cobra.Command, args []string) error {
	fmt.Println("📅 Daily Standup")
	fmt.Println("This will launch an interactive daily standup.")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  • Team member checklist")
	fmt.Println("  • Blocker highlighting")
	fmt.Println("  • Timer for standup")
	fmt.Println("  • Export summary")
	fmt.Println()
	fmt.Println("Not yet implemented - coming soon!")
	return nil
}
```

**Step 2: Commit**

```bash
git add internal/commands/ceremony.go
git commit -m "feat(commands): add ceremony command structure (planning, retro, daily)"
```

---

## Phase 9: Integration & Polish

### Task 16: Add Global Flags

**Files:**
- Modify: `internal/commands/root.go`

**Step 1: Add global flags**

```go
// internal/commands/root.go
package commands

import (
	"github.com/spf13/cobra"
)

var (
	// Global flags
	projectFlag   string
	noCacheFlag   bool
	cacheTTLFlag  string
	verboseFlag   bool
)

var rootCmd = &cobra.Command{
	Use:   "jira-go",
	Short: "A CLI tool for managing Jira Software projects",
	Long: `jira-go is a comprehensive CLI for Jira Software that supports
task management, sprint operations, epics, and agile ceremonies.

Use "jira-go init" to get started with the initial configuration.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip for init command
		if cmd.Name() == "init" || cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}
		
		// TODO: Initialize cache based on flags
		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&projectFlag, "project", "p", "", "Project key (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&noCacheFlag, "no-cache", false, "Disable cache for this command")
	rootCmd.PersistentFlags().StringVar(&cacheTTLFlag, "cache-ttl", "", "Cache TTL (e.g., 5m, 1h)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
}
```

**Step 2: Update task command to use global project flag**

Modify `runTaskList`, `runTaskCreate` and other functions to check the global flag first:

```go
func getProjectKey(cmd *cobra.Command, cfg *config.Config) string {
	// Check global flag first
	if projectFlag != "" {
		return projectFlag
	}
	
	if project, _ := cmd.Flags().GetString("project"); project != "" {
		return project
	}
	return cfg.DefaultProject
}
```

**Step 3: Commit**

```bash
git add internal/commands/root.go internal/commands/task.go
git commit -m "feat(commands): add global flags for project, cache control, and verbosity"
```

---

### Task 17: Add Version Command

**Files:**
- Create: `internal/commands/version.go`

**Step 1: Create version command**

```go
// internal/commands/version.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// These are set during build
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("jira-go version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  build date: %s\n", buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
```

**Step 2: Update Makefile for version injection**

```makefile
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X github.com/user/jira-go/internal/commands.version=$(VERSION) \
           -X github.com/user/jira-go/internal/commands.commit=$(COMMIT) \
           -X github.com/user/jira-go/internal/commands.buildDate=$(BUILD_DATE)

build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/jira-go
```

**Step 3: Test and commit**

```bash
make build
./build/jira-go version
git add internal/commands/version.go Makefile
git commit -m "feat(commands): add version command with build info injection"
```

---

## Phase 10: Documentation

### Task 18: Create README

**Files:**
- Create: `README.md`

**Step 1: Write comprehensive README**

```markdown
# jira-go

A comprehensive CLI tool for managing Jira Software projects with support for task/sprint/epic management, agile ceremonies, and a beautiful TUI powered by Charmbracelet.

## Features

- 🎯 **Task Management**: Create, edit, delete, and list Jira issues
- 🏃 **Sprint Operations**: Manage sprints and view sprint boards
- 📊 **Epic Management**: Organize work with epics
- 👥 **Multi-Owner Support**: Assign multiple owners to tasks via custom fields
- 📧 **Email-Based Interaction**: Use emails instead of account IDs
- 🖥️ **Rich TUI**: Beautiful terminal interface using Bubble Tea
- ⚡ **Local Caching**: SQLite-based cache for offline access
- 🎪 **Agile Ceremonies**: Interactive planning, retrospectives, and dailies

## Installation

### From Source

```bash
git clone https://github.com/user/jira-go
cd jira-go
make build
make install
```

### Download Binary

Download the latest release from the releases page.

## Quick Start

### 1. Initialize Configuration

```bash
jira-go init
```

This interactive wizard will guide you through:
- Jira URL (e.g., https://your-domain.atlassian.net)
- Email address
- API Token ([create one here](https://id.atlassian.com/manage-profile/security/api-tokens))
- Default project key
- Multi-owner custom field ID (optional)

### 2. Verify Connection

```bash
jira-go task list
```

## Usage

### Task Commands

```bash
# List tasks
jira-go task list
jira-go task list --assignee user@example.com
jira-go task list --status "In Progress"

# Create a task
jira-go task create --summary "Fix login bug" --type Bug --assignee dev@example.com

# Create with multiple owners
jira-go task create --summary "Refactor API" --owners "dev1@example.com,dev2@example.com"

# View task details
jira-go task view PROJ-123

# Edit a task
jira-go task edit PROJ-123 --summary "Updated summary"

# Delete a task
jira-go task delete PROJ-123
```

### TUI Mode

```bash
# Interactive issue browser
jira-go tui list
```

### Project Management

```bash
# List configured projects
jira-go project list

# Switch default project
jira-go project switch OTHER

# View project config
jira-go project config
```

### Cache Management

```bash
# View cache status
jira-go cache status

# Clear cache
jira-go cache clear

# Show cache path
jira-go cache path
```

### Agile Ceremonies

```bash
# Sprint planning
jira-go ceremony planning

# Retrospective
jira-go ceremony retro

# Daily standup
jira-go ceremony daily
```

## Configuration

Configuration is stored in `~/.config/jira-go/config.yaml`:

```yaml
default_project: PROJ
auth:
  email: user@example.com
  api_token: ${JIRA_GO_API_TOKEN}  # Can reference env vars

projects:
  PROJ:
    jira_url: https://company.atlassian.net
    board_id: 1
    multi_owner_field: customfield_10001

cache:
  enabled: true
  default_ttl: 30m
  location: ~/.cache/jira-go/cache.db
```

### Environment Variables

- `JIRA_GO_CONFIG` - Path to config file
- `JIRA_GO_API_TOKEN` - API token (overrides config)
- `JIRA_GO_EMAIL` - Email (overrides config)
- `JIRA_GO_DEFAULT_PROJECT` - Default project (overrides config)
- `JIRA_GO_CACHE` - Cache file path

## Multi-Owner Configuration

Jira only supports a single assignee per issue. To enable multiple owners:

1. Go to Jira Project Settings → Issue Types → Fields
2. Create a custom field: "Additional Owners" (Multi User Picker type)
3. Note the field ID (e.g., `customfield_10001`)
4. Run `jira-go init` again and enter the field ID
5. Use `--owners` flag when creating/editing tasks

## Global Flags

All commands support these global flags:

- `-p, --project` - Override default project
- `--no-cache` - Disable cache for this command
- `--cache-ttl` - Custom cache TTL (e.g., `5m`, `1h`)
- `-v, --verbose` - Enable verbose output

## Development

```bash
# Run tests
make test

# Build for development
go run ./cmd/jira-go

# Lint
make lint
```

## License

MIT License
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add comprehensive README with usage examples"
```

---

## Summary

This implementation plan creates a fully functional Jira CLI with:

1. ✅ **Configuration Management**: YAML + env var support, interactive init
2. ✅ **Jira API Client**: Full CRUD for issues, user resolution, authentication
3. ✅ **Multi-Owner Support**: Custom field integration with email resolution
4. ✅ **Caching**: SQLite-based with TTL and user mapping cache
5. ✅ **Task Commands**: List, create, view, edit, delete with email-based interactions
6. ✅ **TUI Foundation**: Bubble Tea setup with issue list component
7. ✅ **Project Management**: List, switch, config commands
8. ✅ **Sprint Structure**: Basic sprint commands (extend as needed)
9. ✅ **Ceremony Foundation**: Planning, retro, daily placeholders
10. ✅ **Cache Management**: Status, clear, path commands
11. ✅ **Global Flags**: Project, cache control, verbosity
12. ✅ **Documentation**: Comprehensive README

### Next Steps (Beyond This Plan)

1. **Complete Sprint API**: Implement full sprint operations with Jira Agile API
2. **Epic Management**: CRUD operations and epic-issue relationships
3. **Full TUI Ceremonies**: Planning, retro, and daily interfaces
4. **Glow Integration**: Add markdown rendering for issue descriptions
5. **Advanced TUI**: Kanban board, detailed issue view with markdown
6. **Export Features**: Markdown/Confluence export for ceremonies
7. **Webhook Support**: Real-time updates
8. **CI/CD Commands**: Pipeline integration helpers

**Plan complete and saved to `docs/plans/2025-01-22-jira-cli-implementation-plan.md`.**

Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach would you prefer?
