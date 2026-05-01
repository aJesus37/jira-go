package skills

import "embed"

//go:embed jira-go jira-go-tasks jira-go-epics jira-go-sprints jira-go-reports
var FS embed.FS
