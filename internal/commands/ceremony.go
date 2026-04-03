// internal/commands/ceremony.go
package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/api"
	"github.com/user/jira-go/internal/config"
	"github.com/user/jira-go/internal/tui"
)

var ceremonyCmd = &cobra.Command{
	Use:   "ceremony",
	Short: "Run agile ceremonies",
	Long:  `Interactive tools for sprint planning, retrospectives, and daily standups.`,
}

var ceremonyPlanningCmd = &cobra.Command{
	Use:   "planning",
	Short: "Run sprint planning ceremony",
	RunE:  runCeremonyPlanning,
}

var ceremonyRetroCmd = &cobra.Command{
	Use:   "retro",
	Short: "Run retrospective ceremony",
	RunE:  runCeremonyRetro,
}

var ceremonyDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Run daily standup",
	RunE:  runCeremonyDaily,
}

func init() {
	rootCmd.AddCommand(ceremonyCmd)
	ceremonyCmd.AddCommand(ceremonyPlanningCmd)
	ceremonyCmd.AddCommand(ceremonyRetroCmd)
	ceremonyCmd.AddCommand(ceremonyDailyCmd)
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
	resp, err := client.SearchIssues(jql, 0, 100, project.MultiOwnerField)
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

	// Export results
	export := model.ExportPlanning()
	filename := fmt.Sprintf("sprint-planning-%s.md", time.Now().Format("2006-01-02"))
	if err := os.WriteFile(filename, []byte(export), 0644); err != nil {
		return fmt.Errorf("exporting planning: %w", err)
	}

	fmt.Printf("✓ Planning exported to %s\n", filename)
	return nil
}

func runCeremonyRetro(cmd *cobra.Command, args []string) error {
	if noInteractiveFlag {
		fmt.Println("📝 Retrospective")
		fmt.Println("\nRun without --no-interactive for full TUI")
		return nil
	}

	// Launch retrospective TUI
	model := tui.NewRetrospectiveCeremony()

	if err := tui.Run(model); err != nil {
		return err
	}

	// Export results
	export := model.ExportRetrospective()
	filename := fmt.Sprintf("retrospective-%s.md", time.Now().Format("2006-01-02"))
	if err := os.WriteFile(filename, []byte(export), 0644); err != nil {
		return fmt.Errorf("exporting retrospective: %w", err)
	}

	fmt.Printf("✓ Retrospective exported to %s\n", filename)
	return nil
}

func runCeremonyDaily(cmd *cobra.Command, args []string) error {
	// Team members (could be fetched from config or passed as args)
	members := []string{"Team Member 1", "Team Member 2", "Team Member 3"}

	if noInteractiveFlag {
		fmt.Println("📅 Daily Standup")
		fmt.Printf("\nTeam members: %d\n", len(members))
		fmt.Println("\nRun without --no-interactive for full TUI")
		return nil
	}

	// Launch standup TUI
	model := tui.NewDailyStandupCeremony(members)

	if err := tui.Run(model); err != nil {
		return err
	}

	// Export results
	export := model.ExportDailyStandup()
	filename := fmt.Sprintf("daily-standup-%s.md", time.Now().Format("2006-01-02"))
	if err := os.WriteFile(filename, []byte(export), 0644); err != nil {
		return fmt.Errorf("exporting standup: %w", err)
	}

	fmt.Printf("✓ Standup exported to %s\n", filename)
	return nil
}
