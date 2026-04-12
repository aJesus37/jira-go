// internal/api/client.go
package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/aJesus37/jira-go/internal/config"
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
