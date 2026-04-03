// internal/models/models.go
package models

import (
	"encoding/json"
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

// RawIssue represents the raw Jira API response
type RawIssue struct {
	Key    string      `json:"key"`
	ID     string      `json:"id"`
	Fields IssueFields `json:"fields"`
}

// ToIssue converts a RawIssue to an Issue
func (r *RawIssue) ToIssue() Issue {
	issue := Issue{
		Key:      r.Key,
		ID:       r.ID,
		Summary:  r.Fields.Summary,
		Type:     r.Fields.IssueType.Name,
		Status:   r.Fields.Status.Name,
		Assignee: r.Fields.Assignee,
		Reporter: r.Fields.Reporter,
		Created:  r.Fields.Created.Time(),
		Updated:  r.Fields.Updated.Time(),
		Labels:   r.Fields.Labels,
	}

	// Extract description (handle both string and ADF)
	switch d := r.Fields.Description.(type) {
	case string:
		issue.Description = d
	default:
		// For ADF format, we'd need more complex parsing
		// For now, just store as empty or implement ADF to markdown conversion
		issue.Description = ""
	}

	return issue
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
