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
