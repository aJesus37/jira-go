// internal/api/issues.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	Total      int               `json:"total"`
	StartAt    int               `json:"startAt"`
	MaxResults int               `json:"maxResults"`
	Issues     []models.Issue    `json:"-"` // Populated manually from RawIssues
	RawIssues  []models.RawIssue `json:"issues"`
}

// CreateIssueResponse represents the response from creating an issue
type CreateIssueResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
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

	var result CreateIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Fetch the full issue to return complete data
	return c.GetIssue(result.Key)
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

	var rawIssue models.RawIssue
	if err := json.NewDecoder(resp.Body).Decode(&rawIssue); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	issue := rawIssue.ToIssue()
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

	// Convert RawIssues to Issues
	for _, raw := range result.RawIssues {
		issue := raw.ToIssue()
		result.Issues = append(result.Issues, issue)
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
