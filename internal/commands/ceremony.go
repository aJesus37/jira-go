// internal/commands/ceremony.go
package commands

import (
	"fmt"
	"time"

	"github.com/aJesus37/jira-go/internal/api"
	"github.com/aJesus37/jira-go/internal/config"
	"github.com/aJesus37/jira-go/internal/models"
	"github.com/aJesus37/jira-go/internal/tui"
	"github.com/spf13/cobra"
)

var ceremonyCmd = &cobra.Command{
	Use:   "ceremony",
	Short: "Run agile ceremonies",
	Long:  `Interactive tools for sprint planning and daily standups.`,
}

var ceremonyPlanningCmd = &cobra.Command{
	Use:   "planning",
	Short: "Run sprint planning ceremony",
	RunE:  runCeremonyPlanning,
}

var ceremonyDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Run daily standup",
	RunE:  runCeremonyDaily,
}

var dailyTimerDuration time.Duration

func init() {
	rootCmd.AddCommand(ceremonyCmd)
	ceremonyCmd.AddCommand(ceremonyPlanningCmd)
	ceremonyCmd.AddCommand(ceremonyDailyCmd)

	ceremonyDailyCmd.Flags().DurationVar(&dailyTimerDuration, "timer", 2*time.Minute, "Timer duration per person (e.g., 2m, 5m, 1h)")
}

func runCeremonyPlanning(cmd *cobra.Command, args []string) error {
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

	// Fetch backlog issues
	jql := fmt.Sprintf("project = %s AND sprint is EMPTY", projectKey)
	resp, err := client.SearchIssues(jql, 0, 100, project.MultiOwnerField, project.SprintField)
	if err != nil {
		return fmt.Errorf("fetching backlog: %w", err)
	}

	if noInteractiveFlag {
		// Show planning info in text mode
		fmt.Println("🎯 Sprint Planning")
		fmt.Printf("\nBacklog issues: %d\n", len(resp.Issues))
		fmt.Println("\nRun without --no-interactive for full TUI")
		return nil
	}

	// Launch planning TUI
	model := tui.NewPlanningCeremony(resp.Issues)

	// Run the TUI
	if err := tui.Run(model); err != nil {
		return err
	}

	return nil
}

func runCeremonyDaily(cmd *cobra.Command, args []string) error {
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

	// Fetch issues from current sprint
	jql := fmt.Sprintf("project = %s AND sprint in openSprints()", projectKey)
	resp, err := client.SearchIssues(jql, 0, 100, project.MultiOwnerField, project.SprintField)
	if err != nil {
		return fmt.Errorf("fetching sprint issues: %w", err)
	}

	// Extract unique assignees and group their tasks by status
	// Uses multi-owner field (Owners) if available, falls back to single assignee (Assignee)
	assigneeTasks := make(map[string]map[string][]models.Issue)
	var members []string
	memberSet := make(map[string]bool)

	for _, issue := range resp.Issues {
		// Get assignees from multi-owner field or fall back to single assignee
		var assigneeNames []string
		if len(issue.Owners) > 0 {
			// Use multi-owner custom field
			for _, owner := range issue.Owners {
				if owner.DisplayName != "" {
					assigneeNames = append(assigneeNames, owner.DisplayName)
				}
			}
		}

		// Fall back to single assignee if no owners
		if len(assigneeNames) == 0 {
			if issue.Assignee != nil && issue.Assignee.DisplayName != "" {
				assigneeNames = append(assigneeNames, issue.Assignee.DisplayName)
			} else {
				assigneeNames = append(assigneeNames, "Unassigned")
			}
		}

		// Group by status
		status := issue.Status
		if status == "" {
			status = "To Do"
		}

		// Add task to each assignee's task list
		for _, assigneeName := range assigneeNames {
			if !memberSet[assigneeName] {
				memberSet[assigneeName] = true
				members = append(members, assigneeName)
				assigneeTasks[assigneeName] = make(map[string][]models.Issue)
			}
			assigneeTasks[assigneeName][status] = append(assigneeTasks[assigneeName][status], issue)
		}
	}

	if len(members) == 0 {
		members = []string{"No assignees found"}
	}

	if noInteractiveFlag {
		fmt.Println("📅 Daily Standup")
		fmt.Printf("\nTeam members with tasks: %d\n", len(members))
		for _, member := range members {
			fmt.Printf("\n👤 %s\n", member)
			for status, tasks := range assigneeTasks[member] {
				fmt.Printf("  [%s] %d tasks\n", status, len(tasks))
				for _, task := range tasks {
					fmt.Printf("    - %s: %s\n", task.Key, task.Summary)
				}
			}
		}
		fmt.Println("\nRun without --no-interactive for full TUI")
		return nil
	}

	// Launch standup TUI with real data
	model := tui.NewDailyStandupCeremony(members, assigneeTasks, dailyTimerDuration)

	if err := tui.Run(model); err != nil {
		return err
	}

	return nil
}
