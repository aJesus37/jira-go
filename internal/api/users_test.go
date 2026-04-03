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
				"accountId":    "12345",
				"displayName":  "Test User",
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
