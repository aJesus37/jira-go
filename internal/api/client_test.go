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
