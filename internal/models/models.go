// internal/models/models.go
package models

import "time"

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

// User represents a Jira user
type User struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	Email       string `json:"emailAddress"`
	AvatarURL   string `json:"avatarUrls"`
}

// Sprint represents a Jira sprint
type Sprint struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	State        string    `json:"state"` // future, active, closed
	StartDate    time.Time `json:"startDate"`
	EndDate      time.Time `json:"endDate"`
	CompleteDate time.Time `json:"completeDate"`
}
