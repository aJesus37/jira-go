// internal/commands/project.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/config"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `Switch between projects and view project configuration.`,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured projects",
	RunE:  runProjectList,
}

var projectSwitchCmd = &cobra.Command{
	Use:   "switch [project-key]",
	Short: "Switch default project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectSwitch,
}

var projectConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "View current project configuration",
	RunE:  runProjectConfig,
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectSwitchCmd)
	projectCmd.AddCommand(projectConfigCmd)
}

func runProjectList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Printf("Current default: %s\n\n", cfg.DefaultProject)
	fmt.Println("Configured projects:")
	fmt.Println("-------------------")

	for key, project := range cfg.Projects {
		prefix := "  "
		if key == cfg.DefaultProject {
			prefix = "* "
		}
		fmt.Printf("%s%s\n", prefix, key)
		fmt.Printf("  URL: %s\n", project.JiraURL)
		fmt.Printf("  Board ID: %d\n", project.BoardID)
		if project.MultiOwnerField != "" {
			fmt.Printf("  Multi-owner field: %s\n", project.MultiOwnerField)
		}
		fmt.Println()
	}

	return nil
}

func runProjectSwitch(cmd *cobra.Command, args []string) error {
	newProject := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Verify project exists
	if _, err := cfg.GetProject(newProject); err != nil {
		return fmt.Errorf("project %s not found in config", newProject)
	}

	cfg.DefaultProject = newProject

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("✓ Switched to project %s\n", newProject)
	return nil
}

func runProjectConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	project, err := cfg.GetProject(cfg.DefaultProject)
	if err != nil {
		return err
	}

	fmt.Printf("Project: %s\n", cfg.DefaultProject)
	fmt.Printf("Jira URL: %s\n", project.JiraURL)
	fmt.Printf("Board ID: %d\n", project.BoardID)
	fmt.Printf("Multi-owner field: %s\n", project.MultiOwnerField)

	return nil
}
