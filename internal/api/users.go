// internal/api/users.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aJesus37/jira-go/internal/models"
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

// SearchUsers searches for users by query string (for autocomplete)
func (c *Client) SearchUsers(query string) ([]models.User, error) {
	if query == "" {
		return nil, nil
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("maxResults", "10")

	resp, err := c.Get("/rest/api/3/user/search?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("searching users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user search failed: %s", resp.Status)
	}

	var users []models.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return users, nil
}
