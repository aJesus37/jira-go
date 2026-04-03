// internal/commands/tui.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/config"
	"github.com/user/jira-go/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Long:  `Open the interactive terminal user interface for browsing and managing Jira.`,
}

var tuiListCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues in TUI",
	RunE:  runTUIList,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.AddCommand(tuiListCmd)

	tuiListCmd.Flags().String("project", "", "Project key (defaults to config)")
}

func runTUIList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// Get project config for owner field
	project, _ := cfg.GetProject(projectKey)
	ownerFieldID := project.MultiOwnerField

	// Fetch issues
	jql := fmt.Sprintf("project = %s ORDER BY updated DESC", projectKey)
	resp, err := client.SearchIssues(jql, 0, 50, ownerFieldID)
	if err != nil {
		return fmt.Errorf("searching issues: %w", err)
	}

	// Launch TUI
	model := tui.NewIssueList(resp.Issues, client, projectKey, ownerFieldID)
	return tui.Run(model)
}
