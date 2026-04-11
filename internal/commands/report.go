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
	"github.com/user/jira-go/internal/models"
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

// SprintReport groups assignee summaries under a sprint.
type SprintReport struct {
	Sprint    string            `json:"sprint"`
	Total     int               `json:"total"`
	Assignees []AssigneeSummary `json:"assignees"`
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
	reportCmd.Flags().Int("limit", 200, "Maximum issues to fetch (default 200)")
}

func runReport(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	format, _ := cmd.Flags().GetString("format")
	if format != "table" && format != "json" {
		return fmt.Errorf("unknown format %q: must be table or json", format)
	}

	projectKey := getProjectKey(cmd, cfg)
	project, _ := cfg.GetProject(projectKey)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	jql := fmt.Sprintf("project = %s AND statusCategory != Done", projectKey)

	sprintFilter, _ := cmd.Flags().GetString("sprint")
	if sprintFilter != "" {
		switch sprintFilter {
		case "active":
			jql += " AND sprint in openSprints()"
		case "future":
			jql += " AND sprint in futureSprints()"
		case "closed":
			jql += " AND sprint in closedSprints()"
		default:
			jql += fmt.Sprintf(` AND sprint = "%s"`, sprintFilter)
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

	limit, _ := cmd.Flags().GetInt("limit")
	resp, err := client.SearchIssues(jql, 0, limit, ownerField, sprintField)
	if err != nil {
		return fmt.Errorf("searching issues: %w", err)
	}

	if sprintFilter != "" {
		// Sprint filter provided: flat assignee breakdown
		summaries := buildAssigneeSummaries(resp.Issues)
		if format == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(summaries)
		}
		fmt.Printf("\nProject: %s — Active Tasks\n", projectKey)
		fmt.Println(strings.Repeat("=", 60))
		printAssigneeTable(summaries)
		fmt.Printf("\nTotal active: %d\n", len(resp.Issues))
		return nil
	}

	// No sprint filter: group by sprint, then show assignee breakdown per sprint
	sprintReports := buildSprintReports(resp.Issues)

	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(sprintReports)
	}

	fmt.Printf("\nProject: %s — Active Tasks by Sprint\n", projectKey)
	for _, sr := range sprintReports {
		fmt.Printf("\n%s (%d tickets)\n", sr.Sprint, sr.Total)
		fmt.Println(strings.Repeat("─", 55))
		printAssigneeTable(sr.Assignees)
	}
	fmt.Printf("\nTotal active: %d\n", len(resp.Issues))
	return nil
}

// buildAssigneeSummaries aggregates a slice of issues by assignee.
func buildAssigneeSummaries(issues []models.Issue) []AssigneeSummary {
	byAssignee := map[string]*AssigneeSummary{}
	for _, issue := range issues {
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
	return summaries
}

// buildSprintReports groups issues by sprint name, then aggregates by assignee within each sprint.
func buildSprintReports(issues []models.Issue) []SprintReport {
	bySprint := map[string][]models.Issue{}
	sprintOrder := []string{}

	for _, issue := range issues {
		name := issue.SprintName
		if name == "" {
			name = "Backlog"
		}
		if _, ok := bySprint[name]; !ok {
			sprintOrder = append(sprintOrder, name)
		}
		bySprint[name] = append(bySprint[name], issue)
	}

	// Named sprints alphabetically, Backlog last
	sort.Slice(sprintOrder, func(i, j int) bool {
		a, b := sprintOrder[i], sprintOrder[j]
		if a == "Backlog" {
			return false
		}
		if b == "Backlog" {
			return true
		}
		return a < b
	})

	reports := make([]SprintReport, 0, len(sprintOrder))
	for _, sName := range sprintOrder {
		summaries := buildAssigneeSummaries(bySprint[sName])
		total := 0
		for _, s := range summaries {
			total += s.Total
		}
		reports = append(reports, SprintReport{
			Sprint:    sName,
			Total:     total,
			Assignees: summaries,
		})
	}
	return reports
}

// printAssigneeTable renders the assignee summary table to stdout.
func printAssigneeTable(summaries []AssigneeSummary) {
	fmt.Printf("  %-23s %-6s %-10s %-10s\n", "ASSIGNEE", "TOTAL", "AVG DAYS", "MAX DAYS")
	fmt.Printf("  %s\n", strings.Repeat("-", 53))
	for _, s := range summaries {
		name := s.Name
		if len(name) > 21 {
			name = name[:18] + "..."
		}
		fmt.Printf("  %-23s %-6d %-10d %-10d\n", name, s.Total, s.AvgAgeDays, s.OldestDays)
		statuses := make([]string, 0, len(s.ByStatus))
		for st := range s.ByStatus {
			statuses = append(statuses, st)
		}
		sort.Strings(statuses)
		for _, st := range statuses {
			fmt.Printf("    %-21s %d\n", st, s.ByStatus[st])
		}
	}
}
