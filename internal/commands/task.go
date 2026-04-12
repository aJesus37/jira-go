// internal/commands/task.go
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/aJesus37/jira-go/internal/api"
	"github.com/aJesus37/jira-go/internal/config"
	"github.com/aJesus37/jira-go/internal/models"
	"github.com/aJesus37/jira-go/internal/tui"
)

func escapeJQL(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage Jira tasks/issues",
	Long:  `Create, view, edit, and delete Jira tasks and issues.`,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks (interactive by default)",
	Long: `List tasks in interactive TUI mode by default.
Use --no-interactive flag for plain text output suitable for automation.

Filter by assignee or owner email:
  jira-go task list --assignee "user@example.com"
  jira-go task list --owner "user@example.com"

Show only active tasks (exclude status category 'Done'):
  jira-go task list --active

Show only backlog tasks (not in sprint):
  jira-go task list --backlog`,
	RunE: runTaskList,
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

var taskStatusCmd = &cobra.Command{
	Use:   "status [key] [status]",
	Short: "Transition issue to a new status",
	Args:  cobra.ExactArgs(2),
	RunE:  runTaskStatus,
}

var taskCommentCmd = &cobra.Command{
	Use:   "comment [key] [message]",
	Short: "Add a comment to an issue",
	Args:  cobra.ExactArgs(2),
	RunE:  runTaskComment,
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskViewCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskStatusCmd)
	taskCmd.AddCommand(taskCommentCmd)
	taskStatusCmd.Flags().String("comment", "", "Optional comment to add with the status change")

	// List flags
	taskListCmd.Flags().String("project", "", "Project key (defaults to config)")
	taskListCmd.Flags().String("assignee", "", "Filter by assignee email")
	taskListCmd.Flags().String("owner", "", "Filter by owner email (multi-owner field)")
	taskListCmd.Flags().String("status", "", "Filter by status")
	taskListCmd.Flags().Bool("active", false, "Show only active tasks (exclude status category 'Done')")
	taskListCmd.Flags().Bool("backlog", false, "Show only backlog tasks (not in any sprint)")
	taskListCmd.Flags().Int("limit", 50, "Maximum results (default 50 for interactive mode)")
	taskListCmd.Flags().String("format", "table", "Output format: table or json")
	taskListCmd.Flags().Bool("age", false, "Show days in current status column")

	// Create flags
	taskCreateCmd.Flags().String("project", "", "Project key (defaults to config)")
	taskCreateCmd.Flags().String("type", "Task", "Issue type")
	taskCreateCmd.Flags().String("summary", "", "Issue summary (required)")
	taskCreateCmd.Flags().String("description", "", "Issue description")
	taskCreateCmd.Flags().String("assignee", "", "Assignee email")
	taskCreateCmd.Flags().String("owners", "", "Comma-separated owner emails")
	taskCreateCmd.Flags().String("status", "", "Set initial status after creation (e.g. 'Done')")

	// Edit flags
	taskEditCmd.Flags().String("summary", "", "New summary")
	taskEditCmd.Flags().String("description", "", "New description")
	taskEditCmd.Flags().String("assignee", "", "New assignee email")
	taskEditCmd.Flags().String("owners", "", "Comma-separated owner emails")
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
	jql := fmt.Sprintf("project = '%s'", escapeJQL(projectKey))

	if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		// Resolve email to account ID for JQL
		user, err := client.ResolveEmail(assignee)
		if err != nil {
			return fmt.Errorf("resolving assignee email: %w", err)
		}
		jql += fmt.Sprintf(" AND assignee = '%s'", escapeJQL(user.AccountID))
	}

	if status, _ := cmd.Flags().GetString("status"); status != "" {
		jql += fmt.Sprintf(" AND status = '%s'", escapeJQL(status))
	}

	// Filter for active tasks only (exclude done status category)
	if active, _ := cmd.Flags().GetBool("active"); active {
		jql += " AND statusCategory != Done"
	}

	// Filter for backlog tasks only (not in any sprint)
	if backlog, _ := cmd.Flags().GetBool("backlog"); backlog {
		jql += " AND sprint is EMPTY"
	}

	// Get project config for owner field
	project, _ := cfg.GetProject(projectKey)
	ownerFieldID := project.MultiOwnerField

	// Handle owner filter - we'll filter locally since JQL for custom fields is unreliable
	var ownerFilterEmail string
	if ownerEmail, _ := cmd.Flags().GetString("owner"); ownerEmail != "" {
		if ownerFieldID == "" {
			return fmt.Errorf("multi_owner_field not configured, cannot filter by owner")
		}
		ownerFilterEmail = strings.ToLower(ownerEmail)
	}

	limit, _ := cmd.Flags().GetInt("limit")
	if limit <= 0 {
		limit = 50 // Default limit
	}

	resp, err := client.SearchIssues(jql, 0, limit, ownerFieldID, project.SprintField)
	if err != nil {
		return fmt.Errorf("searching issues: %w", err)
	}

	// Filter by owner locally if specified
	if ownerFilterEmail != "" {
		var filteredIssues []models.Issue
		for _, issue := range resp.Issues {
			for _, owner := range issue.Owners {
				if strings.ToLower(owner.Email) == ownerFilterEmail {
					filteredIssues = append(filteredIssues, issue)
					break
				}
			}
		}
		resp.Issues = filteredIssues
	}

	// Check if we should run in non-interactive mode
	format, _ := cmd.Flags().GetString("format")
	showAge, _ := cmd.Flags().GetBool("age")
	if format == "" {
		format = "table"
	}
	if noInteractiveFlag || format != "table" || showAge {
		mergeAssigneeOwner := project.MergeAssigneeOwner == nil || *project.MergeAssigneeOwner
		return displayTaskListTable(resp.Issues, resp.Total, mergeAssigneeOwner, format, showAge)
	}

	// Run in interactive TUI mode
	model := tui.NewIssueList(resp.Issues, client, projectKey, ownerFieldID, project.SprintField, project.BoardID)
	return tui.Run(model)
}

func displayTaskListTable(issues []models.Issue, total int, mergeAssigneeOwner bool, format string, showAge bool) error {
	if format != "table" && format != "json" {
		return fmt.Errorf("unknown format %q: must be table or json", format)
	}

	// JSON output
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(issues)
	}

	// Table header
	if showAge {
		fmt.Printf("%-12s %-10s %-12s %-6s %-20s %s\n", "KEY", "TYPE", "STATUS", "DAYS", "ASSIGNEE", "SUMMARY")
	} else {
		fmt.Printf("%-12s %-10s %-12s %-20s %s\n", "KEY", "TYPE", "STATUS", "ASSIGNEE", "SUMMARY")
	}
	fmt.Println(strings.Repeat("-", 100))

	for _, issue := range issues {
		status := issue.Status
		if len(status) > 12 {
			status = status[:9] + "..."
		}

		displayAssignee := "Unassigned"
		if mergeAssigneeOwner {
			participants := issue.GetAllParticipants()
			if len(participants) > 0 {
				names := []string{}
				for _, p := range participants {
					names = append(names, p.DisplayName)
				}
				displayAssignee = strings.Join(names, ", ")
			}
		} else if issue.Assignee != nil {
			displayAssignee = issue.Assignee.DisplayName
		}

		if len(displayAssignee) > 18 {
			displayAssignee = displayAssignee[:15] + "..."
		}

		summary := issue.Summary
		if len(summary) > 40 {
			summary = summary[:37] + "..."
		}

		if showAge {
			fmt.Printf("%-12s %-10s %-12s %-6d %-20s %s\n",
				issue.Key, issue.Type, status, issue.DaysInStatus(), displayAssignee, summary)
		} else {
			fmt.Printf("%-12s %-10s %-12s %-20s %s\n", issue.Key, issue.Type, status, displayAssignee, summary)
		}
	}

	fmt.Printf("\nShowing %d of %d issues\n", len(issues), total)
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

	if targetStatus, _ := cmd.Flags().GetString("status"); targetStatus != "" {
		transitions, err := client.GetTransitions(issue.Key)
		if err != nil {
			return fmt.Errorf("getting transitions: %w", err)
		}
		var transitionID string
		for _, t := range transitions {
			if strings.EqualFold(t.Name, targetStatus) {
				transitionID = t.ID
				break
			}
		}
		if transitionID == "" {
			var names []string
			for _, t := range transitions {
				names = append(names, t.Name)
			}
			return fmt.Errorf("status %q not available after creation; available: %s",
				targetStatus, strings.Join(names, ", "))
		}
		if err := client.TransitionIssue(issue.Key, transitionID); err != nil {
			return fmt.Errorf("setting initial status: %w", err)
		}
		fmt.Printf("✓ Status set to %s\n", targetStatus)
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

	// Get project config for owner field
	project, _ := cfg.GetProject(cfg.DefaultProject)
	ownerFieldID := project.MultiOwnerField

	issue, err := client.GetIssue(key, ownerFieldID, project.SprintField)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s: %s\n", issue.Key, issue.Summary)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Type: %s\n", issue.Type)
	fmt.Printf("Status: %s\n", issue.Status)
	if issue.Assignee != nil {
		fmt.Printf("Assignee: %s (%s)\n", issue.Assignee.DisplayName, issue.Assignee.Email)
	} else {
		fmt.Println("Assignee: Unassigned")
	}
	if len(issue.Owners) > 0 {
		var ownerNames []string
		for _, o := range issue.Owners {
			ownerNames = append(ownerNames, fmt.Sprintf("%s (%s)", o.DisplayName, o.Email))
		}
		fmt.Printf("Owners: %s\n", strings.Join(ownerNames, ", "))
	}
	fmt.Printf("Created: %s\n", issue.Created.Format("2006-01-02 15:04"))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("Description:")

	// Render description with markdown
	renderer, err := tui.NewMarkdownRenderer(80)
	if err != nil {
		// Fallback to plain text
		fmt.Println(issue.Description)
	} else {
		rendered, err := renderer.Render(issue.Description)
		if err != nil {
			fmt.Println(issue.Description)
		} else {
			fmt.Println(rendered)
		}
	}
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

func runTaskStatus(cmd *cobra.Command, args []string) error {
	key, targetStatus := args[0], args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	transitions, err := client.GetTransitions(key)
	if err != nil {
		return fmt.Errorf("getting transitions: %w", err)
	}

	var transitionID string
	for _, t := range transitions {
		if strings.EqualFold(t.Name, targetStatus) {
			transitionID = t.ID
			break
		}
	}
	if transitionID == "" {
		var names []string
		for _, t := range transitions {
			names = append(names, t.Name)
		}
		return fmt.Errorf("status %q not found; available: %s", targetStatus, strings.Join(names, ", "))
	}

	if err := client.TransitionIssue(key, transitionID); err != nil {
		return fmt.Errorf("transitioning issue: %w", err)
	}
	fmt.Printf("✓ %s → %s\n", key, targetStatus)

	if comment, _ := cmd.Flags().GetString("comment"); comment != "" {
		if err := client.AddComment(key, comment); err != nil {
			return fmt.Errorf("adding comment: %w", err)
		}
		fmt.Printf("✓ Comment added\n")
	}

	return nil
}

func runTaskComment(cmd *cobra.Command, args []string) error {
	key, message := args[0], args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := api.NewClient(cfg, cfg.DefaultProject)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	if err := client.AddComment(key, message); err != nil {
		return fmt.Errorf("adding comment: %w", err)
	}

	fmt.Printf("✓ Comment added to %s\n", key)
	return nil
}
