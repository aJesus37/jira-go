// internal/models/models.go
package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JiraTime is a custom time type that handles Jira's timestamp format
type JiraTime time.Time

// UnmarshalJSON handles Jira's timestamp format (e.g., "2026-04-03T10:53:45.694-0300")
func (j *JiraTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	str := strings.Trim(string(data), `"`)
	if str == "" {
		return nil
	}

	// Try different Jira timestamp formats
	formats := []string{
		time.RFC3339,                   // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05.000-0700", // With milliseconds and offset
		"2006-01-02T15:04:05.000Z",     // With milliseconds
		"2006-01-02T15:04:05-0700",     // With offset
		"2006-01-02T15:04:05.000Z0700", // Alternative offset format
		"2006-01-02T15:04:05Z0700",     // Without milliseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			*j = JiraTime(t)
			return nil
		}
	}

	// If all formats fail, return the time as-is with zero value
	*j = JiraTime(time.Time{})
	return nil
}

// Time returns the underlying time.Time
func (j JiraTime) Time() time.Time {
	return time.Time(j)
}

// IsZero returns true if the time is zero
func (j JiraTime) IsZero() bool {
	return time.Time(j).IsZero()
}

// Issue represents a Jira issue
type Issue struct {
	Key         string    `json:"key"`
	ID          string    `json:"id"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Assignee    *User     `json:"assignee"`
	Owners      []User    `json:"owners"` // Multi-owner support
	Reporter    *User     `json:"reporter"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Labels      []string  `json:"labels"`
}

// IssueFields represents the fields of a Jira issue (for API parsing)
type IssueFields struct {
	Summary     string      `json:"summary"`
	Description interface{} `json:"description"` // Can be string or ADF (Atlassian Document Format)
	IssueType   IssueType   `json:"issuetype"`
	Status      Status      `json:"status"`
	Assignee    *User       `json:"assignee"`
	Reporter    *User       `json:"reporter"`
	Created     JiraTime    `json:"created"`
	Updated     JiraTime    `json:"updated"`
	Labels      []string    `json:"labels"`
}

// IssueType represents an issue type
type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Status represents an issue status
type Status struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RawIssue represents the raw Jira API response with custom fields
type RawIssue struct {
	Key    string                 `json:"key"`
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

// ToIssue converts a RawIssue to an Issue
func (r *RawIssue) ToIssue() Issue {
	issue := Issue{
		Key: r.Key,
		ID:  r.ID,
	}

	// Extract standard fields
	if summary, ok := r.Fields["summary"].(string); ok {
		issue.Summary = summary
	}

	if desc, ok := r.Fields["description"]; ok {
		switch d := desc.(type) {
		case string:
			issue.Description = d
		default:
			issue.Description = ""
		}
	}

	if issuetype, ok := r.Fields["issuetype"].(map[string]interface{}); ok {
		if name, ok := issuetype["name"].(string); ok {
			issue.Type = name
		}
	}

	if status, ok := r.Fields["status"].(map[string]interface{}); ok {
		if name, ok := status["name"].(string); ok {
			issue.Status = name
		}
	}

	// Parse times
	if created, ok := r.Fields["created"].(string); ok && created != "" {
		if t, err := parseJiraTime(created); err == nil {
			issue.Created = t
		}
	}

	if updated, ok := r.Fields["updated"].(string); ok && updated != "" {
		if t, err := parseJiraTime(updated); err == nil {
			issue.Updated = t
		}
	}

	// Parse labels
	if labels, ok := r.Fields["labels"].([]interface{}); ok {
		for _, l := range labels {
			if label, ok := l.(string); ok {
				issue.Labels = append(issue.Labels, label)
			}
		}
	}

	// Parse assignee
	if assignee, ok := r.Fields["assignee"].(map[string]interface{}); ok && assignee != nil {
		issue.Assignee = parseUser(assignee)
	}

	// Parse reporter
	if reporter, ok := r.Fields["reporter"].(map[string]interface{}); ok && reporter != nil {
		issue.Reporter = parseUser(reporter)
	}

	return issue
}

// ToIssueWithOwners converts a RawIssue to an Issue with owners from custom field
func (r *RawIssue) ToIssueWithOwners(ownerFieldID string) Issue {
	issue := r.ToIssue()

	// Parse owners from custom field
	if ownerFieldID != "" {
		if owners, ok := r.Fields[ownerFieldID].([]interface{}); ok {
			for _, o := range owners {
				if ownerMap, ok := o.(map[string]interface{}); ok {
					if user := parseUser(ownerMap); user != nil {
						issue.Owners = append(issue.Owners, *user)
					}
				}
			}
		}
	}

	return issue
}

// parseUser extracts user data from a map
func parseUser(data map[string]interface{}) *User {
	if data == nil {
		return nil
	}

	user := &User{}

	if accountID, ok := data["accountId"].(string); ok {
		user.AccountID = accountID
	}

	if displayName, ok := data["displayName"].(string); ok {
		user.DisplayName = displayName
	}

	if email, ok := data["emailAddress"].(string); ok {
		user.Email = email
	}

	// Parse avatar URLs if present
	if avatars, ok := data["avatarUrls"].(map[string]interface{}); ok {
		user.AvatarURLs = make(map[string]string)
		for k, v := range avatars {
			if url, ok := v.(string); ok {
				user.AvatarURLs[k] = url
			}
		}
	}

	return user
}

// parseJiraTime tries to parse Jira timestamp
func parseJiraTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05.000Z0700",
		"2006-01-02T15:04:05Z0700",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// User represents a Jira user
type User struct {
	AccountID   string            `json:"accountId"`
	DisplayName string            `json:"displayName"`
	Email       string            `json:"emailAddress"`
	AvatarURLs  map[string]string `json:"avatarUrls"`
}

// Sprint represents a Jira sprint
type Sprint struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	State        string   `json:"state"` // future, active, closed
	StartDate    JiraTime `json:"startDate"`
	EndDate      JiraTime `json:"endDate"`
	CompleteDate JiraTime `json:"completeDate"`
}

// Ensure Sprint implements json.Unmarshaler for the time fields
func (s *Sprint) UnmarshalJSON(data []byte) error {
	type Alias Sprint
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal(data, &aux)
}
