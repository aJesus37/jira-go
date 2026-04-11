// internal/commands/report.go
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/config"
)

// AssigneeSummary aggregates issue data per person.
type AssigneeSummary struct {
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	Total      int            `json:"total"`
	ByStatus   map[string]int `json:"by_status"`
	AvgAgeDays int            `json:"avg_age_days"`
	OldestDays int            `json:"oldest_days"`
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Project status report",
	Long:  `Summarize active tasks per assignee with age-in-status. Use --format json for machine-readable output.`,
	RunE:  runReport,
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().String("format", "table", "Output format: table or json")
	reportCmd.Flags().String("sprint", "", "Filter: active, future, closed, or sprint ID")
	reportCmd.Flags().String("assignee", "", "Filter by assignee email")
	reportCmd.Flags().String("project", "", "Project key (defaults to config)")
}

func runReport(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)
	project, _ := cfg.GetProject(projectKey)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	jql := fmt.Sprintf("project = %s AND statusCategory != Done", projectKey)

	if sprint, _ := cmd.Flags().GetString("sprint"); sprint != "" {
		switch sprint {
		case "active":
			jql += " AND sprint in openSprints()"
		case "future":
			jql += " AND sprint in futureSprints()"
		case "closed":
			jql += " AND sprint in closedSprints()"
		default:
			jql += fmt.Sprintf(" AND sprint = %s", sprint)
		}
	}

	if assigneeEmail, _ := cmd.Flags().GetString("assignee"); assigneeEmail != "" {
		user, err := client.ResolveEmail(assigneeEmail)
		if err != nil {
			return fmt.Errorf("resolving assignee: %w", err)
		}
		jql += fmt.Sprintf(" AND assignee = '%s'", user.AccountID)
	}

	var ownerField, sprintField string
	if project != nil {
		ownerField = project.MultiOwnerField
		sprintField = project.SprintField
	}

	resp, err := client.SearchIssues(jql, 0, 200, ownerField, sprintField)
	if err != nil {
		return fmt.Errorf("searching issues: %w", err)
	}

	// Aggregate by assignee
	byAssignee := map[string]*AssigneeSummary{}
	for _, issue := range resp.Issues {
		var email, name string
		if issue.Assignee != nil {
			email = issue.Assignee.Email
			name = issue.Assignee.DisplayName
		} else {
			email = "unassigned"
			name = "Unassigned"
		}

		if _, ok := byAssignee[email]; !ok {
			byAssignee[email] = &AssigneeSummary{
				Name:     name,
				Email:    email,
				ByStatus: map[string]int{},
			}
		}

		s := byAssignee[email]
		s.Total++
		s.ByStatus[issue.Status]++
		age := issue.DaysInStatus()
		s.AvgAgeDays += age
		if age > s.OldestDays {
			s.OldestDays = age
		}
	}

	// Compute averages and sort
	summaries := make([]AssigneeSummary, 0, len(byAssignee))
	for _, s := range byAssignee {
		if s.Total > 0 {
			s.AvgAgeDays = s.AvgAgeDays / s.Total
		}
		summaries = append(summaries, *s)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Total > summaries[j].Total
	})

	format, _ := cmd.Flags().GetString("format")

	if format != "table" && format != "json" {
		return fmt.Errorf("unknown format %q: must be table or json", format)
	}

	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summaries)
	}

	// Table output
	fmt.Printf("\nProject: %s — Active Tasks\n", projectKey)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("%-25s %-6s %-10s %-10s\n", "ASSIGNEE", "TOTAL", "AVG DAYS", "OLDEST")
	fmt.Println(strings.Repeat("-", 55))
	for _, s := range summaries {
		name := s.Name
		if len(name) > 23 {
			name = name[:20] + "..."
		}
		fmt.Printf("%-25s %-6d %-10d %-10d\n", name, s.Total, s.AvgAgeDays, s.OldestDays)
		// Status breakdown (indented), sorted alphabetically
		statuses := make([]string, 0, len(s.ByStatus))
		for st := range s.ByStatus {
			statuses = append(statuses, st)
		}
		sort.Strings(statuses)
		for _, st := range statuses {
			fmt.Printf("  %-23s %d\n", st, s.ByStatus[st])
		}
	}
	fmt.Printf("\nTotal active: %d\n", len(resp.Issues))
	return nil
}
