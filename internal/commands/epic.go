// internal/commands/epic.go
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/aJesus37/jira-go/internal/api"
	"github.com/aJesus37/jira-go/internal/config"
)

var epicCmd = &cobra.Command{
	Use:   "epic",
	Short: "Manage Jira epics",
	Long:  `Create, view, and manage Jira epics and their child issues.`,
}

var epicListCmd = &cobra.Command{
	Use:   "list",
	Short: "List epics",
	RunE:  runEpicList,
}

var epicCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new epic",
	RunE:  runEpicCreate,
}

var epicViewCmd = &cobra.Command{
	Use:   "view [epic-key]",
	Short: "View epic details with child issues",
	Args:  cobra.ExactArgs(1),
	RunE:  runEpicView,
}

var epicAddCmd = &cobra.Command{
	Use:   "add [epic-key] [issue-keys...]",
	Short: "Add issues to epic",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runEpicAdd,
}

var epicRemoveCmd = &cobra.Command{
	Use:   "remove [issue-keys...]",
	Short: "Remove issues from epic",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runEpicRemove,
}

func init() {
	rootCmd.AddCommand(epicCmd)
	epicCmd.AddCommand(epicListCmd)
	epicCmd.AddCommand(epicCreateCmd)
	epicCmd.AddCommand(epicViewCmd)
	epicCmd.AddCommand(epicAddCmd)
	epicCmd.AddCommand(epicRemoveCmd)

	// Create flags
	epicCreateCmd.Flags().String("summary", "", "Epic summary (required)")
	epicCreateCmd.Flags().String("description", "", "Epic description")
	epicCreateCmd.Flags().String("project", "", "Project key (defaults to config)")
}

func runEpicList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	epics, err := client.GetEpics(projectKey)
	if err != nil {
		return fmt.Errorf("fetching epics: %w", err)
	}

	if len(epics) == 0 {
		fmt.Println("No epics found")
		return nil
	}

	fmt.Printf("%-12s %-12s %-20s %s\n", "KEY", "STATUS", "ASSIGNEE", "SUMMARY")
	fmt.Println(strings.Repeat("-", 100))

	for _, epic := range epics {
		status := epic.Status
		if len(status) > 12 {
			status = status[:9] + "..."
		}

		assignee := "Unassigned"
		if epic.Assignee != nil {
			assignee = epic.Assignee.DisplayName
			if len(assignee) > 18 {
				assignee = assignee[:15] + "..."
			}
		}

		summary := epic.Summary
		if len(summary) > 50 {
			summary = summary[:47] + "..."
		}

		fmt.Printf("%-12s %-12s %-20s %s\n", epic.Key, status, assignee, summary)
	}

	fmt.Printf("\nTotal: %d epics\n", len(epics))
	return nil
}

func runEpicCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)

	summary, _ := cmd.Flags().GetString("summary")
	if summary == "" {
		return fmt.Errorf("--summary is required")
	}

	description, _ := cmd.Flags().GetString("description")

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	epic, err := client.CreateEpic(projectKey, summary, description)
	if err != nil {
		return fmt.Errorf("creating epic: %w", err)
	}

	fmt.Printf("✓ Created epic %s\n", epic.Key)
	return nil
}

func runEpicView(cmd *cobra.Command, args []string) error {
	epicKey := args[0]

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

	// Get epic details
	epic, err := client.GetEpic(epicKey)
	if err != nil {
		return fmt.Errorf("fetching epic: %w", err)
	}

	// Get child issues
	issues, err := client.GetEpicIssues(epicKey, project.MultiOwnerField, project.SprintField)
	if err != nil {
		return fmt.Errorf("fetching epic issues: %w", err)
	}

	// Calculate progress
	total := len(issues)
	done := 0
	for _, issue := range issues {
		if issue.Status == "Done" || issue.Status == "Closed" {
			done++
		}
	}
	progress := 0
	if total > 0 {
		progress = (done * 100) / total
	}

	// Display epic
	fmt.Printf("\n📚 %s: %s\n", epic.Key, epic.Summary)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Status: %s\n", epic.Status)
	if epic.Assignee != nil {
		fmt.Printf("Assignee: %s\n", epic.Assignee.DisplayName)
	} else {
		fmt.Println("Assignee: Unassigned")
	}
	fmt.Printf("Progress: %d%% (%d/%d issues completed)\n", progress, done, total)
	fmt.Println(strings.Repeat("-", 80))

	if epic.Description != "" {
		fmt.Println("Description:")
		fmt.Println(epic.Description)
		fmt.Println()
	}

	// Display child issues
	if len(issues) > 0 {
		fmt.Printf("Child Issues (%d):\n\n", len(issues))
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
	} else {
		fmt.Println("No child issues")
	}

	fmt.Println()
	return nil
}

func runEpicAdd(cmd *cobra.Command, args []string) error {
	epicKey := args[0]
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

	// Link each issue to epic
	for _, issueKey := range issueKeys {
		if err := client.LinkIssueToEpic(issueKey, epicKey); err != nil {
			fmt.Printf("⚠ Failed to add %s: %v\n", issueKey, err)
		} else {
			fmt.Printf("✓ Added %s to %s\n", issueKey, epicKey)
		}
	}

	return nil
}

func runEpicRemove(cmd *cobra.Command, args []string) error {
	issueKeys := args

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// Unlink each issue from epic
	for _, issueKey := range issueKeys {
		if err := client.UnlinkIssueFromEpic(issueKey); err != nil {
			fmt.Printf("⚠ Failed to remove %s: %v\n", issueKey, err)
		} else {
			fmt.Printf("✓ Removed %s from epic\n", issueKey)
		}
	}

	return nil
}
