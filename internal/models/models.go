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
	SprintName  string    `json:"sprint_name"` // Current sprint name (if any)
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
		case map[string]interface{}:
			// ADF (Atlassian Document Format) - extract text content
			issue.Description = extractADFText(d)
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

// DaysInStatus returns how many days the issue has been in its current status.
// Uses Updated as a proxy since Jira changelog requires a separate API call.
func (i Issue) DaysInStatus() int {
	if i.Updated.IsZero() {
		return 0
	}
	return int(time.Since(i.Updated).Hours() / 24)
}

func (i *Issue) GetAllParticipants() []User {
	seen := make(map[string]bool)
	var result []User

	addUnique := func(u User) {
		if u.AccountID != "" && !seen[u.AccountID] {
			seen[u.AccountID] = true
			result = append(result, u)
		}
	}

	for _, o := range i.Owners {
		addUnique(o)
	}
	if i.Assignee != nil {
		addUnique(*i.Assignee)
	}

	return result
}

// ToIssueWithOwners converts a RawIssue to an Issue with owners from custom field
func (r *RawIssue) ToIssueWithOwners(ownerFieldID string, sprintFieldID string) Issue {
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

	// Parse sprint from custom field
	if sprintFieldID != "" {
		if sprints, ok := r.Fields[sprintFieldID].([]interface{}); ok && len(sprints) > 0 {
			// Get the last sprint (most recent)
			if lastSprint, ok := sprints[len(sprints)-1].(map[string]interface{}); ok {
				if name, ok := lastSprint["name"].(string); ok {
					issue.SprintName = name
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

// extractADFText extracts readable text from Atlassian Document Format (ADF)
func extractADFText(adf map[string]interface{}) string {
	var result strings.Builder

	// Get content array
	content, ok := adf["content"].([]interface{})
	if !ok {
		return ""
	}

	for _, item := range content {
		node, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		nodeType, _ := node["type"].(string)

		switch nodeType {
		case "paragraph":
			text := extractADFNodeText(node)
			if text != "" {
				result.WriteString(text)
				result.WriteString("\n\n")
			}
		case "heading":
			level, _ := node["attrs"].(map[string]interface{})["level"].(float64)
			for i := 0; i < int(level); i++ {
				result.WriteString("#")
			}
			result.WriteString(" ")
			result.WriteString(extractADFNodeText(node))
			result.WriteString("\n\n")
		case "bulletList", "orderedList":
			result.WriteString(extractADFListItems(node, ""))
			result.WriteString("\n")
		case "codeBlock":
			result.WriteString("```\n")
			result.WriteString(extractADFNodeText(node))
			result.WriteString("\n```\n\n")
		case "blockquote":
			text := extractADFNodeText(node)
			if text != "" {
				result.WriteString("> ")
				result.WriteString(strings.ReplaceAll(text, "\n", "\n> "))
				result.WriteString("\n\n")
			}
		case "rule":
			result.WriteString("---\n\n")
		}
	}

	return strings.TrimSpace(result.String())
}

// extractADFNodeText extracts text from a single ADF node
func extractADFNodeText(node map[string]interface{}) string {
	var result strings.Builder

	content, ok := node["content"].([]interface{})
	if !ok {
		return ""
	}

	for _, item := range content {
		switch child := item.(type) {
		case map[string]interface{}:
			childType, _ := child["type"].(string)
			switch childType {
			case "text":
				text, _ := child["text"].(string)
				// Handle marks (formatting)
				if marks, ok := child["marks"].([]interface{}); ok {
					for _, mark := range marks {
						if markMap, ok := mark.(map[string]interface{}); ok {
							markType, _ := markMap["type"].(string)
							switch markType {
							case "strong":
								text = "**" + text + "**"
							case "em":
								text = "*" + text + "*"
							case "code":
								text = "`" + text + "`"
							}
						}
					}
				}
				result.WriteString(text)
			case "hardBreak":
				result.WriteString("\n")
			case "inlineCard", "emoji":
				// Try to get text representation
				if attrs, ok := child["attrs"].(map[string]interface{}); ok {
					if text, ok := attrs["text"].(string); ok {
						result.WriteString(text)
					}
				}
			}
		}
	}

	return result.String()
}

// extractADFListItems extracts list items from ADF
func extractADFListItems(node map[string]interface{}, indent string) string {
	var result strings.Builder

	content, ok := node["content"].([]interface{})
	if !ok {
		return ""
	}

	for i, item := range content {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		itemType, _ := itemMap["type"].(string)
		if itemType != "listItem" {
			continue
		}

		// Get the content of this list item
		itemContent, ok := itemMap["content"].([]interface{})
		if !ok {
			continue
		}

		// Write bullet or number
		if node["type"] == "orderedList" {
			result.WriteString(fmt.Sprintf("%s%d. ", indent, i+1))
		} else {
			result.WriteString(indent + "- ")
		}

		// Process the content
		for _, contentItem := range itemContent {
			contentMap, ok := contentItem.(map[string]interface{})
			if !ok {
				continue
			}

			contentType, _ := contentMap["type"].(string)
			switch contentType {
			case "paragraph":
				result.WriteString(extractADFNodeText(contentMap))
			case "bulletList", "orderedList":
				result.WriteString("\n")
				result.WriteString(extractADFListItems(contentMap, indent+"  "))
			}
		}

		result.WriteString("\n")
	}

	return result.String()
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
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	State         string   `json:"state"` // future, active, closed
	Goal          string   `json:"goal"`
	OriginBoardID int      `json:"originBoardId"`
	StartDate     JiraTime `json:"startDate"`
	EndDate       JiraTime `json:"endDate"`
	CompleteDate  JiraTime `json:"completeDate"`
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
