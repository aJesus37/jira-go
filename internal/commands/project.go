// internal/commands/project.go
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
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

var projectSetBoardCmd = &cobra.Command{
	Use:   "set-board [board-id]",
	Short: "Set board ID for current project",
	Long:  `Update the board ID for the current project. Get the board ID from your Jira board URL (e.g., https://your-domain.atlassian.net/jira/software/c/projects/PROJ/boards/1)`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectSetBoard,
}

var projectDetectBoardsCmd = &cobra.Command{
	Use:   "detect-boards",
	Short: "Auto-detect boards for current project",
	Long:  `Fetch and list all available boards for the current project from Jira.`,
	RunE:  runProjectDetectBoards,
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectSwitchCmd)
	projectCmd.AddCommand(projectConfigCmd)
	projectCmd.AddCommand(projectSetBoardCmd)
	projectCmd.AddCommand(projectDetectBoardsCmd)
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

func runProjectSetBoard(cmd *cobra.Command, args []string) error {
	boardID := 0
	if _, err := fmt.Sscanf(args[0], "%d", &boardID); err != nil {
		return fmt.Errorf("invalid board ID: %s", args[0])
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)
	project, exists := cfg.Projects[projectKey]
	if !exists {
		return fmt.Errorf("project %s not found in config", projectKey)
	}

	project.BoardID = boardID
	cfg.Projects[projectKey] = project

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("✓ Set board ID to %d for project %s\n", boardID, projectKey)
	fmt.Println("\nYou can now use sprint commands:")
	fmt.Println("  jira-go sprint list")
	fmt.Println("  jira-go sprint board")

	return nil
}

func runProjectDetectBoards(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectKey := getProjectKey(cmd, cfg)

	client, err := api.NewClient(cfg, projectKey)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	boards, err := client.GetBoards(projectKey)
	if err != nil {
		return fmt.Errorf("fetching boards: %w", err)
	}

	if len(boards) == 0 {
		fmt.Println("No boards found for this project")
		fmt.Println("\nMake sure:")
		fmt.Println("1. You have admin access to the project")
		fmt.Println("2. The project has at least one board created")
		return nil
	}

	fmt.Printf("Found %d board(s) for project %s:\n\n", len(boards), projectKey)
	fmt.Printf("%-6s %-30s %-15s\n", "ID", "NAME", "TYPE")
	fmt.Println(strings.Repeat("-", 60))

	for _, board := range boards {
		id, _ := board["id"].(float64)
		name, _ := board["name"].(string)
		boardType, _ := board["type"].(string)

		if len(name) > 28 {
			name = name[:25] + "..."
		}

		fmt.Printf("%-6d %-30s %-15s\n", int(id), name, boardType)
	}

	fmt.Println("\nUse the ID to set your board:")
	fmt.Printf("  jira-go project set-board <ID>\n")

	return nil
}
