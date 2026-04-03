// internal/commands/sprint.go
package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/config"
	"github.com/user/jira-go/internal/models"
	"github.com/user/jira-go/internal/tui"
)

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Manage Jira sprints",
	Long:  `List, create, start, complete, and manage Jira sprints.`,
}

var sprintListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sprints",
	RunE:  runSprintList,
}

var sprintCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new sprint",
	RunE:  runSprintCreate,
}

var sprintStartCmd = &cobra.Command{
	Use:   "start [sprint-id]",
	Short: "Start a sprint",
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintStart,
}

var sprintCompleteCmd = &cobra.Command{
	Use:   "complete [sprint-id]",
	Short: "Complete a sprint",
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintComplete,
}

var sprintIssuesCmd = &cobra.Command{
	Use:   "issues [sprint-id]",
	Short: "List issues in a sprint",
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintIssues,
}

var sprintBoardCmd = &cobra.Command{
	Use:   "board [sprint-id]",
	Short: "View sprint board (kanban view)",
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintBoard,
}

var sprintMoveCmd = &cobra.Command{
	Use:   "move [sprint-id] [issue-keys]",
	Short: "Move issues to sprint",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runSprintMove,
}

func init() {
	rootCmd.AddCommand(sprintCmd)
	sprintCmd.AddCommand(sprintListCmd)
	sprintCmd.AddCommand(sprintCreateCmd)
	sprintCmd.AddCommand(sprintStartCmd)
	sprintCmd.AddCommand(sprintCompleteCmd)
	sprintCmd.AddCommand(sprintIssuesCmd)
	sprintCmd.AddCommand(sprintBoardCmd)
	sprintCmd.AddCommand(sprintMoveCmd)

	// List flags
	sprintListCmd.Flags().String("state", "", "Filter by state (active, future, closed)")

	// Create flags
	sprintCreateCmd.Flags().String("name", "", "Sprint name (required)")
	sprintCreateCmd.Flags().String("goal", "", "Sprint goal")
	sprintCreateCmd.Flags().String("start", "", "Start date (YYYY-MM-DD)")
	sprintCreateCmd.Flags().String("end", "", "End date (YYYY-MM-DD)")
}

func runSprintList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)
	project, err := cfg.GetProject(projectKey)
	if err != nil {
		return fmt.Errorf("getting project: %w", err)
	}

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	state, _ := cmd.Flags().GetString("state")

	sprints, err := client.GetSprints(project.BoardID, state)
	if err != nil {
		return fmt.Errorf("fetching sprints: %w", err)
	}

	if len(sprints) == 0 {
		fmt.Println("No sprints found")
		return nil
	}

	fmt.Printf("%-6s %-30s %-10s %-12s %-12s\n", "ID", "NAME", "STATE", "START", "END")
	fmt.Println(strings.Repeat("-", 90))

	for _, sprint := range sprints {
		startDate := "-"
		if !sprint.StartDate.IsZero() {
			startDate = sprint.StartDate.Time().Format("2006-01-02")
		}

		endDate := "-"
		if !sprint.EndDate.IsZero() {
			endDate = sprint.EndDate.Time().Format("2006-01-02")
		}

		name := sprint.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		fmt.Printf("%-6d %-30s %-10s %-12s %-12s\n", sprint.ID, name, sprint.State, startDate, endDate)
	}

	return nil
}

func runSprintCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)
	project, err := cfg.GetProject(projectKey)
	if err != nil {
		return fmt.Errorf("getting project: %w", err)
	}

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return fmt.Errorf("--name is required")
	}

	goal, _ := cmd.Flags().GetString("goal")
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")

	var startDate, endDate time.Time

	if startStr != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return fmt.Errorf("invalid start date format, use YYYY-MM-DD: %w", err)
		}
	}

	if endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return fmt.Errorf("invalid end date format, use YYYY-MM-DD: %w", err)
		}
	}

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	sprint, err := client.CreateSprint(project.BoardID, name, goal, startDate, endDate)
	if err != nil {
		return fmt.Errorf("creating sprint: %w", err)
	}

	fmt.Printf("✓ Created sprint %d: %s\n", sprint.ID, sprint.Name)
	return nil
}

func runSprintStart(cmd *cobra.Command, args []string) error {
	sprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sprint ID: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	if err := client.StartSprint(sprintID, ""); err != nil {
		return fmt.Errorf("starting sprint: %w", err)
	}

	fmt.Printf("✓ Started sprint %d\n", sprintID)
	return nil
}

func runSprintComplete(cmd *cobra.Command, args []string) error {
	sprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sprint ID: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	if err := client.CompleteSprint(sprintID); err != nil {
		return fmt.Errorf("completing sprint: %w", err)
	}

	fmt.Printf("✓ Completed sprint %d\n", sprintID)
	return nil
}

func runSprintIssues(cmd *cobra.Command, args []string) error {
	sprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sprint ID: %w", err)
	}

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

	issues, err := client.GetSprintIssues(sprintID, project.MultiOwnerField)
	if err != nil {
		return fmt.Errorf("fetching sprint issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No issues in sprint")
		return nil
	}

	fmt.Printf("Issues in sprint %d:\n\n", sprintID)
	fmt.Printf("%-12s %-10s %-12s %-20s %s\n", "KEY", "TYPE", "STATUS", "ASSIGNEE", "SUMMARY")
	fmt.Println(strings.Repeat("-", 100))

	for _, issue := range issues {
		status := issue.Status
		if len(status) > 12 {
			status = status[:9] + "..."
		}

		assignee := "Unassigned"
		if issue.Assignee != nil {
			assignee = issue.Assignee.DisplayName
			if len(assignee) > 18 {
				assignee = assignee[:15] + "..."
			}
		}

		summary := issue.Summary
		if len(summary) > 40 {
			summary = summary[:37] + "..."
		}

		fmt.Printf("%-12s %-10s %-12s %-20s %s\n", issue.Key, issue.Type, status, assignee, summary)
	}

	fmt.Printf("\nTotal: %d issues\n", len(issues))
	return nil
}

func runSprintBoard(cmd *cobra.Command, args []string) error {
	sprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sprint ID: %w", err)
	}

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

	issues, err := client.GetSprintIssues(sprintID, project.MultiOwnerField)
	if err != nil {
		return fmt.Errorf("fetching sprint issues: %w", err)
	}

	if noInteractiveFlag {
		// Show kanban-style text output
		return displayKanbanBoard(issues)
	}

	// Launch kanban TUI
	model := tui.NewKanbanBoard(issues, sprintID, client)
	return tui.Run(model)
}

func displayKanbanBoard(issues []models.Issue) error {
	// Group issues by status
	columns := make(map[string][]models.Issue)

	for _, issue := range issues {
		columns[issue.Status] = append(columns[issue.Status], issue)
	}

	// Print each column
	for status, issues := range columns {
		fmt.Printf("\n📋 %s (%d)\n", strings.ToUpper(status), len(issues))
		fmt.Println(strings.Repeat("─", 60))

		for _, issue := range issues {
			assignee := "Unassigned"
			if issue.Assignee != nil {
				assignee = issue.Assignee.DisplayName
			}
			fmt.Printf("  %-12s %-30s 👤 %s\n", issue.Key, issue.Summary, assignee)
		}
	}

	fmt.Println()
	return nil
}

func runSprintMove(cmd *cobra.Command, args []string) error {
	sprintID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sprint ID: %w", err)
	}

	issueKeys := args[1:]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	if err := client.MoveIssuesToSprint(sprintID, issueKeys); err != nil {
		return fmt.Errorf("moving issues: %w", err)
	}

	fmt.Printf("✓ Moved %d issue(s) to sprint %d\n", len(issueKeys), sprintID)
	return nil
}
