// internal/api/sprint.go
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aJesus37/jira-go/internal/models"
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
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get sprints failed: %s", resp.Status)
	}

	var result SprintListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Values, nil
}

// GetOpenSprints gets all open (active + future) sprints for a board
func (c *Client) GetOpenSprints(boardID int) ([]models.Sprint, error) {
	endpoint := fmt.Sprintf("/rest/agile/1.0/board/%d/sprint?state=active,future", boardID)

	resp, err := c.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching open sprints: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get open sprints failed: %s", resp.Status)
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
	defer resp.Body.Close() //nolint:errcheck

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
	// First fetch the sprint to get required fields
	currentSprint, err := c.GetSprint(sprintID)
	if err != nil {
		return fmt.Errorf("fetching current sprint: %w", err)
	}

	payload := map[string]interface{}{
		"originBoardId": currentSprint.OriginBoardID,
		"state":         "active",
		"name":          currentSprint.Name,
	}

	if goal != "" {
		payload["goal"] = goal
	}

	// Include dates if they exist
	if !currentSprint.StartDate.IsZero() {
		payload["startDate"] = currentSprint.StartDate.Time().Format("2006-01-02T15:04:05.000Z0700")
	}
	if !currentSprint.EndDate.IsZero() {
		payload["endDate"] = currentSprint.EndDate.Time().Format("2006-01-02T15:04:05.000Z0700")
	}

	resp, err := c.Put(fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID), payload)
	if err != nil {
		return fmt.Errorf("starting sprint: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("start sprint failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

// CompleteSprint completes a sprint
func (c *Client) CompleteSprint(sprintID int) error {
	// First fetch the sprint to get required fields
	currentSprint, err := c.GetSprint(sprintID)
	if err != nil {
		return fmt.Errorf("fetching current sprint: %w", err)
	}

	payload := map[string]interface{}{
		"originBoardId": currentSprint.OriginBoardID,
		"state":         "closed",
		"name":          currentSprint.Name,
	}

	// Include dates if they exist
	if !currentSprint.StartDate.IsZero() {
		payload["startDate"] = currentSprint.StartDate.Time().Format("2006-01-02T15:04:05.000Z0700")
	}
	if !currentSprint.EndDate.IsZero() {
		payload["endDate"] = currentSprint.EndDate.Time().Format("2006-01-02T15:04:05.000Z0700")
	}

	resp, err := c.Put(fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID), payload)
	if err != nil {
		return fmt.Errorf("completing sprint: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("complete sprint failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

// CompleteSprintSafe completes a sprint, automatically starting it first
// if it is currently in FUTURE state (Jira does not allow FUTURE→CLOSED directly).
func (c *Client) CompleteSprintSafe(sprintID int) error {
	sprint, err := c.GetSprint(sprintID)
	if err != nil {
		return fmt.Errorf("fetching sprint: %w", err)
	}
	if sprint.State == "future" {
		if err := c.StartSprint(sprintID, ""); err != nil {
			return fmt.Errorf("starting future sprint before close: %w", err)
		}
	}
	return c.CompleteSprint(sprintID)
}

// GetSprint retrieves a single sprint by ID
func (c *Client) GetSprint(sprintID int) (*models.Sprint, error) {
	resp, err := c.Get(fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID))
	if err != nil {
		return nil, fmt.Errorf("fetching sprint: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get sprint failed: %s", resp.Status)
	}

	var sprint models.Sprint
	if err := json.NewDecoder(resp.Body).Decode(&sprint); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &sprint, nil
}

// UpdateSprint updates sprint properties (name, goal, dates)
func (c *Client) UpdateSprint(sprintID int, name, goal string, startDate, endDate time.Time) (*models.Sprint, error) {
	// First fetch the sprint to get the originBoardId (required for updates)
	currentSprint, err := c.GetSprint(sprintID)
	if err != nil {
		return nil, fmt.Errorf("fetching current sprint: %w", err)
	}

	payload := map[string]interface{}{
		"originBoardId": currentSprint.OriginBoardID,
		"state":         currentSprint.State,
		"name":          currentSprint.Name,
	}

	if name != "" {
		payload["name"] = name
	}
	if goal != "" {
		payload["goal"] = goal
	}

	// Use provided dates or keep existing ones
	if !startDate.IsZero() {
		payload["startDate"] = startDate.Format("2006-01-02T15:04:05.000Z0700")
	} else if !currentSprint.StartDate.IsZero() {
		payload["startDate"] = currentSprint.StartDate.Time().Format("2006-01-02T15:04:05.000Z0700")
	}

	if !endDate.IsZero() {
		payload["endDate"] = endDate.Format("2006-01-02T15:04:05.000Z0700")
	} else if !currentSprint.EndDate.IsZero() {
		payload["endDate"] = currentSprint.EndDate.Time().Format("2006-01-02T15:04:05.000Z0700")
	}

	resp, err := c.Put(fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID), payload)
	if err != nil {
		return nil, fmt.Errorf("updating sprint: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		// Try to read error body
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update sprint failed: %s - %s", resp.Status, string(body))
	}

	var sprint models.Sprint
	if err := json.NewDecoder(resp.Body).Decode(&sprint); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &sprint, nil
}

// GetSprintIssues retrieves issues in a sprint using JQL search.
// Pass limit=0 to use default (100). Pass status="" for no filter.
func (c *Client) GetSprintIssues(sprintID int, ownerFieldID, sprintFieldID, status string, limit int) ([]models.Issue, int, error) {
	jql := fmt.Sprintf("sprint = %d", sprintID)
	if status != "" {
		jql += fmt.Sprintf(" AND status = %q", status)
	}
	if limit <= 0 {
		limit = 100
	}

	result, err := c.SearchIssues(jql, 0, limit, ownerFieldID, sprintFieldID)
	if err != nil {
		return nil, 0, fmt.Errorf("searching sprint issues: %w", err)
	}

	return result.Issues, result.Total, nil
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
	defer resp.Body.Close() //nolint:errcheck

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
	defer resp.Body.Close() //nolint:errcheck

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
