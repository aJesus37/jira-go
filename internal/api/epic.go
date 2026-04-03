// internal/api/epic.go
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/user/jira-go/internal/models"
)

// Epic represents a Jira Epic
type Epic struct {
	Key         string       `json:"key"`
	ID          string       `json:"id"`
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	Status      string       `json:"status"`
	Assignee    *models.User `json:"assignee"`
	Reporter    *models.User `json:"reporter"`
	StoryPoints int          `json:"storyPoints"`
	IssueCount  int          `json:"issueCount"`
}

// GetEpics retrieves epics for a project
func (c *Client) GetEpics(projectKey string) ([]Epic, error) {
	jql := fmt.Sprintf("project = %s AND issuetype = Epic", projectKey)

	params := fmt.Sprintf("jql=%s&maxResults=100&fields=summary,status,assignee,description",
		url.QueryEscape(jql))

	resp, err := c.Get("/rest/api/3/search/jql?" + params)
	if err != nil {
		return nil, fmt.Errorf("fetching epics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get epics failed: %s", resp.Status)
	}

	var result IssueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var epics []Epic
	for _, raw := range result.RawIssues {
		issue := raw.ToIssue()
		epic := Epic{
			Key:         issue.Key,
			ID:          issue.ID,
			Summary:     issue.Summary,
			Description: issue.Description,
			Status:      issue.Status,
			Assignee:    issue.Assignee,
			Reporter:    issue.Reporter,
		}
		epics = append(epics, epic)
	}

	return epics, nil
}

// GetEpicIssues retrieves issues linked to an epic
func (c *Client) GetEpicIssues(epicKey string, ownerFieldID string) ([]models.Issue, error) {
	// JQL to find issues in epic
	jql := fmt.Sprintf("'Epic Link' = %s", epicKey)

	fields := "summary,status,assignee,created,updated,issuetype,description,labels"
	if ownerFieldID != "" {
		fields = fields + "," + ownerFieldID
	}

	params := fmt.Sprintf("jql=%s&maxResults=100&fields=%s",
		url.QueryEscape(jql),
		url.QueryEscape(fields))

	resp, err := c.Get("/rest/api/3/search/jql?" + params)
	if err != nil {
		return nil, fmt.Errorf("fetching epic issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get epic issues failed: %s", resp.Status)
	}

	var result IssueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var issues []models.Issue
	for _, raw := range result.RawIssues {
		issue := raw.ToIssueWithOwners(ownerFieldID)
		issues = append(issues, issue)
	}

	return issues, nil
}

// CreateEpic creates a new epic
func (c *Client) CreateEpic(projectKey, summary, description string) (*Epic, error) {
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": projectKey,
			},
			"summary":     summary,
			"description": description,
			"issuetype": map[string]string{
				"name": "Epic",
			},
		},
	}

	resp, err := c.Post("/rest/api/3/issue", payload)
	if err != nil {
		return nil, fmt.Errorf("creating epic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create epic failed: %s - %s", resp.Status, string(body))
	}

	var result CreateIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Fetch the full epic
	return c.GetEpic(result.Key)
}

// GetEpic retrieves an epic by key
func (c *Client) GetEpic(key string) (*Epic, error) {
	resp, err := c.Get(fmt.Sprintf("/rest/api/3/issue/%s", key))
	if err != nil {
		return nil, fmt.Errorf("getting epic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("epic %s not found", key)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get epic failed: %s", resp.Status)
	}

	var rawIssue models.RawIssue
	if err := json.NewDecoder(resp.Body).Decode(&rawIssue); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	issue := rawIssue.ToIssue()
	return &Epic{
		Key:         issue.Key,
		ID:          issue.ID,
		Summary:     issue.Summary,
		Description: issue.Description,
		Status:      issue.Status,
		Assignee:    issue.Assignee,
		Reporter:    issue.Reporter,
	}, nil
}

// UpdateEpic updates an epic
func (c *Client) UpdateEpic(key string, fields map[string]interface{}) error {
	payload := map[string]interface{}{
		"fields": fields,
	}

	resp, err := c.Put(fmt.Sprintf("/rest/api/3/issue/%s", key), payload)
	if err != nil {
		return fmt.Errorf("updating epic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("update epic failed: %s", resp.Status)
	}

	return nil
}

// LinkIssueToEpic links an issue to an epic
func (c *Client) LinkIssueToEpic(issueKey, epicKey string) error {
	// Use the Epic Link field (customfield_10014 is common but varies)
	// For standard Jira, we use the parent field for next-gen projects
	// or Epic Link for classic projects

	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"parent": map[string]string{
				"key": epicKey,
			},
		},
	}

	resp, err := c.Put(fmt.Sprintf("/rest/api/3/issue/%s", issueKey), payload)
	if err != nil {
		return fmt.Errorf("linking issue to epic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("link issue to epic failed: %s", resp.Status)
	}

	return nil
}

// UnlinkIssueFromEpic removes the epic link from an issue
func (c *Client) UnlinkIssueFromEpic(issueKey string) error {
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"parent": nil,
		},
	}

	resp, err := c.Put(fmt.Sprintf("/rest/api/3/issue/%s", issueKey), payload)
	if err != nil {
		return fmt.Errorf("unlinking issue from epic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unlink issue from epic failed: %s", resp.Status)
	}

	return nil
}
