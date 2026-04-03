// internal/commands/sprint.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/config"
)

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Manage Jira sprints",
	Long:  `List, create, and manage Jira sprints.`,
}

var sprintListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sprints",
	RunE:  runSprintList,
}

var sprintIssuesCmd = &cobra.Command{
	Use:   "issues [sprint-id]",
	Short: "List issues in a sprint",
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintIssues,
}

func init() {
	rootCmd.AddCommand(sprintCmd)
	sprintCmd.AddCommand(sprintListCmd)
	sprintCmd.AddCommand(sprintIssuesCmd)
}

func runSprintList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := cfg.DefaultProject
	project, err := cfg.GetProject(projectKey)
	if err != nil {
		return fmt.Errorf("getting project: %w", err)
	}

	fmt.Printf("Fetching sprints for board %d...\n", project.BoardID)
	fmt.Println("Sprint management coming soon!")

	return nil
}

func runSprintIssues(cmd *cobra.Command, args []string) error {
	fmt.Println("Sprint issues not yet implemented")
	return nil
}
