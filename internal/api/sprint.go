// internal/api/sprint.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/jira-go/internal/models"
)

// SprintListResponse represents the response from listing sprints
type SprintListResponse struct {
	IsLast     bool            `json:"isLast"`
	MaxResults int             `json:"maxResults"`
	StartAt    int             `json:"startAt"`
	Values     []models.Sprint `json:"values"`
}

// GetSprints retrieves sprints for a board
func (c *Client) GetSprints(boardID int, state string) ([]models.Sprint, error) {
	endpoint := fmt.Sprintf("/rest/agile/1.0/board/%d/sprint", boardID)
	if state != "" {
		endpoint = fmt.Sprintf("%s?state=%s", endpoint, state)
	}

	resp, err := c.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching sprints: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get sprints failed: %s", resp.Status)
	}

	var result SprintListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Values, nil
}

// CreateSprint creates a new sprint
func (c *Client) CreateSprint(boardID int, name string, goal string, startDate, endDate time.Time) (*models.Sprint, error) {
	payload := map[string]interface{}{
		"name":          name,
		"originBoardId": boardID,
	}

	if goal != "" {
		payload["goal"] = goal
	}

	if !startDate.IsZero() {
		payload["startDate"] = startDate.Format("2006-01-02T15:04:05.000Z0700")
	}

	if !endDate.IsZero() {
		payload["endDate"] = endDate.Format("2006-01-02T15:04:05.000Z0700")
	}

	resp, err := c.Post("/rest/agile/1.0/sprint", payload)
	if err != nil {
		return nil, fmt.Errorf("creating sprint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create sprint failed: %s", resp.Status)
	}

	var sprint models.Sprint
	if err := json.NewDecoder(resp.Body).Decode(&sprint); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &sprint, nil
}

// StartSprint starts a sprint
func (c *Client) StartSprint(sprintID int, goal string) error {
	payload := map[string]interface{}{
		"state": "active",
	}

	if goal != "" {
		payload["goal"] = goal
	}

	resp, err := c.Put(fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID), payload)
	if err != nil {
		return fmt.Errorf("starting sprint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("start sprint failed: %s", resp.Status)
	}

	return nil
}

// CompleteSprint completes a sprint
func (c *Client) CompleteSprint(sprintID int) error {
	payload := map[string]interface{}{
		"state": "closed",
	}

	resp, err := c.Put(fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID), payload)
	if err != nil {
		return fmt.Errorf("completing sprint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("complete sprint failed: %s", resp.Status)
	}

	return nil
}

// GetSprintIssues retrieves issues in a sprint
func (c *Client) GetSprintIssues(sprintID int, ownerFieldID string) ([]models.Issue, error) {
	fields := "summary,status,assignee,created,updated,issuetype,description,labels"
	if ownerFieldID != "" {
		fields = fields + "," + ownerFieldID
	}

	endpoint := fmt.Sprintf("/rest/agile/1.0/sprint/%d/issue?fields=%s", sprintID, fields)

	resp, err := c.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching sprint issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get sprint issues failed: %s", resp.Status)
	}

	var result IssueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Convert RawIssues to Issues with owners
	var issues []models.Issue
	for _, raw := range result.RawIssues {
		issue := raw.ToIssueWithOwners(ownerFieldID)
		issues = append(issues, issue)
	}

	return issues, nil
}

// MoveIssuesToSprint moves issues to a sprint
func (c *Client) MoveIssuesToSprint(sprintID int, issueKeys []string) error {
	payload := map[string]interface{}{
		"issues": issueKeys,
	}

	resp, err := c.Post(fmt.Sprintf("/rest/agile/1.0/sprint/%d/issue", sprintID), payload)
	if err != nil {
		return fmt.Errorf("moving issues to sprint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("move issues to sprint failed: %s", resp.Status)
	}

	return nil
}

// GetBoards retrieves boards for a project
func (c *Client) GetBoards(projectKey string) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/rest/agile/1.0/board?projectKeyOrId=%s", projectKey)

	resp, err := c.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching boards: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get boards failed: %s", resp.Status)
	}

	var result struct {
		Values []map[string]interface{} `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Values, nil
}
