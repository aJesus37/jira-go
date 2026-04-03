// internal/commands/task.go
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/config"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage Jira tasks/issues",
	Long:  `Create, view, edit, and delete Jira tasks and issues.`,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE:  runTaskList,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	RunE:  runTaskCreate,
}

var taskViewCmd = &cobra.Command{
	Use:   "view [key]",
	Short: "View task details",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskView,
}

var taskEditCmd = &cobra.Command{
	Use:   "edit [key]",
	Short: "Edit a task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskEdit,
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskDelete,
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskViewCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.AddCommand(taskDeleteCmd)

	// List flags
	taskListCmd.Flags().String("project", "", "Project key (defaults to config)")
	taskListCmd.Flags().String("assignee", "", "Filter by assignee")
	taskListCmd.Flags().String("status", "", "Filter by status")
	taskListCmd.Flags().Int("limit", 25, "Maximum results")

	// Create flags
	taskCreateCmd.Flags().String("project", "", "Project key (defaults to config)")
	taskCreateCmd.Flags().String("type", "Task", "Issue type")
	taskCreateCmd.Flags().String("summary", "", "Issue summary (required)")
	taskCreateCmd.Flags().String("description", "", "Issue description")
	taskCreateCmd.Flags().String("assignee", "", "Assignee email")
	taskCreateCmd.Flags().String("owners", "", "Comma-separated owner emails")

	// Edit flags
	taskEditCmd.Flags().String("summary", "", "New summary")
	taskEditCmd.Flags().String("description", "", "New description")
	taskEditCmd.Flags().String("assignee", "", "New assignee email")
	taskEditCmd.Flags().String("owners", "", "Comma-separated owner emails")
}

func getProjectKey(cmd *cobra.Command, cfg *config.Config) string {
	// Check global flag first
	if projectFlag != "" {
		return projectFlag
	}
	if project, _ := cmd.Flags().GetString("project"); project != "" {
		return project
	}
	return cfg.DefaultProject
}

func runTaskList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// Build JQL query
	jql := fmt.Sprintf("project = %s", projectKey)

	if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		jql += fmt.Sprintf(" AND assignee = '%s'", assignee)
	}

	if status, _ := cmd.Flags().GetString("status"); status != "" {
		jql += fmt.Sprintf(" AND status = '%s'", status)
	}

	limit, _ := cmd.Flags().GetInt("limit")

	resp, err := client.SearchIssues(jql, 0, limit)
	if err != nil {
		return fmt.Errorf("searching issues: %w", err)
	}

	// Simple table output (will be replaced with TUI later)
	fmt.Printf("%-12s %-10s %-12s %s\n", "KEY", "TYPE", "STATUS", "SUMMARY")
	fmt.Println(strings.Repeat("-", 80))

	for _, issue := range resp.Issues {
		status := issue.Status
		if len(status) > 12 {
			status = status[:9] + "..."
		}

		summary := issue.Summary
		if len(summary) > 40 {
			summary = summary[:37] + "..."
		}

		fmt.Printf("%-12s %-10s %-12s %s\n", issue.Key, issue.Type, status, summary)
	}

	fmt.Printf("\nShowing %d of %d issues\n", len(resp.Issues), resp.Total)

	return nil
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
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
	issueType, _ := cmd.Flags().GetString("type")

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// Create issue
	issue, err := client.CreateIssue(projectKey, summary, description, issueType)
	if err != nil {
		return fmt.Errorf("creating issue: %w", err)
	}

	fmt.Printf("✓ Created %s\n", issue.Key)

	// Handle assignee
	if assigneeEmail, _ := cmd.Flags().GetString("assignee"); assigneeEmail != "" {
		user, err := client.ResolveEmail(assigneeEmail)
		if err != nil {
			return fmt.Errorf("resolving assignee: %w", err)
		}

		if err := client.AssignIssue(issue.Key, user.AccountID); err != nil {
			return fmt.Errorf("assigning issue: %w", err)
		}

		fmt.Printf("✓ Assigned to %s\n", assigneeEmail)
	}

	// Handle multi-owners
	if ownersStr, _ := cmd.Flags().GetString("owners"); ownersStr != "" {
		project, _ := cfg.GetProject(projectKey)
		if project.MultiOwnerField == "" {
			return fmt.Errorf("multi_owner_field not configured for project %s", projectKey)
		}

		emails := strings.Split(ownersStr, ",")
		var accountIDs []string

		for _, email := range emails {
			email = strings.TrimSpace(email)
			user, err := client.ResolveEmail(email)
			if err != nil {
				return fmt.Errorf("resolving owner %s: %w", email, err)
			}
			accountIDs = append(accountIDs, user.AccountID)
		}

		if err := client.UpdateMultiOwnerField(issue.Key, project.MultiOwnerField, accountIDs); err != nil {
			return fmt.Errorf("updating owners: %w", err)
		}

		fmt.Printf("✓ Set owners: %s\n", ownersStr)
	}

	return nil
}

func runTaskView(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	issue, err := client.GetIssue(key)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s: %s\n", issue.Key, issue.Summary)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Type: %s\n", issue.Type)
	fmt.Printf("Status: %s\n", issue.Status)
	if issue.Assignee != nil {
		fmt.Printf("Assignee: %s\n", issue.Assignee.DisplayName)
	}
	if len(issue.Owners) > 0 {
		var ownerNames []string
		for _, o := range issue.Owners {
			ownerNames = append(ownerNames, o.DisplayName)
		}
		fmt.Printf("Owners: %s\n", strings.Join(ownerNames, ", "))
	}
	fmt.Printf("Created: %s\n", issue.Created.Format("2006-01-02 15:04"))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("Description:")
	fmt.Println(issue.Description)
	fmt.Println()

	return nil
}

func runTaskEdit(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	fields := make(map[string]interface{})

	if summary, _ := cmd.Flags().GetString("summary"); summary != "" {
		fields["summary"] = summary
	}

	if description, _ := cmd.Flags().GetString("description"); description != "" {
		fields["description"] = description
	}

	if len(fields) > 0 {
		if err := client.UpdateIssue(key, fields); err != nil {
			return fmt.Errorf("updating issue: %w", err)
		}
		fmt.Printf("✓ Updated %s\n", key)
	}

	// Handle assignee
	if assigneeEmail, _ := cmd.Flags().GetString("assignee"); assigneeEmail != "" {
		user, err := client.ResolveEmail(assigneeEmail)
		if err != nil {
			return fmt.Errorf("resolving assignee: %w", err)
		}

		if err := client.AssignIssue(key, user.AccountID); err != nil {
			return fmt.Errorf("assigning issue: %w", err)
		}

		fmt.Printf("✓ Assigned to %s\n", assigneeEmail)
	}

	// Handle multi-owners
	if ownersStr, _ := cmd.Flags().GetString("owners"); ownersStr != "" {
		project, _ := cfg.GetProject(cfg.DefaultProject)
		if project.MultiOwnerField == "" {
			return fmt.Errorf("multi_owner_field not configured")
		}

		emails := strings.Split(ownersStr, ",")
		var accountIDs []string

		for _, email := range emails {
			email = strings.TrimSpace(email)
			user, err := client.ResolveEmail(email)
			if err != nil {
				return fmt.Errorf("resolving owner %s: %w", email, err)
			}
			accountIDs = append(accountIDs, user.AccountID)
		}

		if err := client.UpdateMultiOwnerField(key, project.MultiOwnerField, accountIDs); err != nil {
			return fmt.Errorf("updating owners: %w", err)
		}

		fmt.Printf("✓ Set owners: %s\n", ownersStr)
	}

	return nil
}

func runTaskDelete(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	if err := client.DeleteIssue(key); err != nil {
		return fmt.Errorf("deleting issue: %w", err)
	}

	fmt.Printf("✓ Deleted %s\n", key)
	return nil
}
