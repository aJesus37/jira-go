// internal/api/issues.go
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/aJesus37/jira-go/internal/models"
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
	fields := map[string]interface{}{
		"project": map[string]string{
			"key": projectKey,
		},
		"summary": summary,
		"issuetype": map[string]string{
			"name": issueType,
		},
	}

	// Only add description if provided (Jira requires ADF format, not plain string)
	if description != "" {
		fields["description"] = map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": description,
						},
					},
				},
			},
		}
	}

	payload := map[string]interface{}{
		"fields": fields,
	}

	resp, err := c.Post("/rest/api/3/issue", payload)
	if err != nil {
		return nil, fmt.Errorf("creating issue: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create issue failed: %s - %s", resp.Status, string(body))
	}

	var result CreateIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Fetch the full issue to return complete data
	return c.GetIssue(result.Key, "", "")
}

// GetIssue retrieves an issue by key
func (c *Client) GetIssue(key string, ownerFieldID string, sprintFieldID string) (*models.Issue, error) {
	resp, err := c.Get(fmt.Sprintf("/rest/api/3/issue/%s", key))
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

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

	issue := rawIssue.ToIssueWithOwners(ownerFieldID, sprintFieldID)
	return &issue, nil
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(jql string, startAt, maxResults int, ownerFieldID string, sprintFieldID string) (*IssueSearchResponse, error) {
	// Build fields list
	fields := "summary,status,assignee,created,updated,issuetype,description,labels"
	if ownerFieldID != "" {
		fields = fields + "," + ownerFieldID
	}
	if sprintFieldID != "" {
		fields = fields + "," + sprintFieldID
	}

	// Build query parameters
	params := fmt.Sprintf("jql=%s&startAt=%d&maxResults=%d&fields=%s",
		url.QueryEscape(jql),
		startAt,
		maxResults,
		url.QueryEscape(fields))

	resp, err := c.Get("/rest/api/3/search/jql?" + params)
	if err != nil {
		return nil, fmt.Errorf("searching issues: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s - %s", resp.Status, string(body))
	}

	var result IssueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Convert RawIssues to Issues with owners and sprint
	for _, raw := range result.RawIssues {
		issue := raw.ToIssueWithOwners(ownerFieldID, sprintFieldID)
		result.Issues = append(result.Issues, issue)
	}

	return &result, nil
}

// SearchIssuesAll searches for all issues matching JQL, automatically fetching all pages
func (c *Client) SearchIssuesAll(jql string, maxResults int, ownerFieldID string, sprintFieldID string) ([]models.Issue, int, error) {
	if maxResults <= 0 {
		maxResults = 100
	}

	var allIssues []models.Issue
	startAt := 0

	for {
		result, err := c.SearchIssues(jql, startAt, maxResults, ownerFieldID, sprintFieldID)
		if err != nil {
			return nil, 0, err
		}

		allIssues = append(allIssues, result.Issues...)

		if len(allIssues) >= result.Total || len(result.Issues) == 0 {
			break
		}

		startAt += len(result.Issues)
	}

	return allIssues, len(allIssues), nil
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
	defer resp.Body.Close() //nolint:errcheck

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
	defer resp.Body.Close() //nolint:errcheck

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
	defer resp.Body.Close() //nolint:errcheck

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

// AddOwnerToIssue adds a single owner to an issue's multi-owner field (by email)
func (c *Client) AddOwnerToIssue(issueKey, fieldID, ownerEmail string) error {
	user, err := c.ResolveEmail(ownerEmail)
	if err != nil {
		return fmt.Errorf("resolving owner email: %w", err)
	}

	return c.AddOwnerByAccountID(issueKey, fieldID, user.AccountID)
}

// AddOwnerByAccountID adds a single owner to an issue's multi-owner field (by account ID)
func (c *Client) AddOwnerByAccountID(issueKey, fieldID, accountID string) error {
	issue, err := c.GetIssue(issueKey, fieldID, "")
	if err != nil {
		return fmt.Errorf("getting issue: %w", err)
	}

	for _, owner := range issue.Owners {
		if owner.AccountID == accountID {
			return nil
		}
	}

	currentOwners := make([]string, 0)
	for _, owner := range issue.Owners {
		if owner.AccountID != "" {
			currentOwners = append(currentOwners, owner.AccountID)
		}
	}

	currentOwners = append(currentOwners, accountID)

	return c.UpdateMultiOwnerField(issueKey, fieldID, currentOwners)
}

// RemoveOwnerFromIssue removes a single owner from an issue's multi-owner field (by email)
func (c *Client) RemoveOwnerFromIssue(issueKey, fieldID, ownerEmail string) error {
	user, err := c.ResolveEmail(ownerEmail)
	if err != nil {
		return fmt.Errorf("resolving owner email: %w", err)
	}

	return c.RemoveOwnerByAccountID(issueKey, fieldID, user.AccountID)
}

// RemoveOwnerByAccountID removes a single owner from an issue's multi-owner field (by account ID)
func (c *Client) RemoveOwnerByAccountID(issueKey, fieldID, accountID string) error {
	issue, err := c.GetIssue(issueKey, fieldID, "")
	if err != nil {
		return fmt.Errorf("getting issue: %w", err)
	}

	currentOwners := make([]string, 0)
	for _, owner := range issue.Owners {
		if owner.AccountID != "" && owner.AccountID != accountID {
			currentOwners = append(currentOwners, owner.AccountID)
		}
	}

	return c.UpdateMultiOwnerField(issueKey, fieldID, currentOwners)
}

// AddComment adds a comment to an issue
func (c *Client) AddComment(key, body string) error {
	payload := map[string]interface{}{
		"body": map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": body,
						},
					},
				},
			},
		},
	}

	resp, err := c.Post(fmt.Sprintf("/rest/api/3/issue/%s/comment", key), payload)
	if err != nil {
		return fmt.Errorf("adding comment: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("add comment failed: %s", resp.Status)
	}

	return nil
}

// Transition represents a status transition
type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetTransitions gets available status transitions for an issue
func (c *Client) GetTransitions(key string) ([]Transition, error) {
	resp, err := c.Get(fmt.Sprintf("/rest/api/3/issue/%s/transitions", key))
	if err != nil {
		return nil, fmt.Errorf("getting transitions: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get transitions failed: %s", resp.Status)
	}

	var result struct {
		Transitions []Transition `json:"transitions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding transitions: %w", err)
	}

	return result.Transitions, nil
}

// TransitionIssue transitions an issue to a new status
func (c *Client) TransitionIssue(key, transitionID string) error {
	payload := map[string]interface{}{
		"transition": map[string]string{
			"id": transitionID,
		},
	}

	resp, err := c.Post(fmt.Sprintf("/rest/api/3/issue/%s/transitions", key), payload)
	if err != nil {
		return fmt.Errorf("transitioning issue: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("transition issue failed: %s", resp.Status)
	}

	return nil
}
