package commands_test

import (
	"testing"

	"github.com/aJesus37/jira-go/internal/commands"
)

func TestSprintIssuesHasFormatFlag(t *testing.T) {
	cmd, _, err := commands.RootCmd.Find([]string{"sprint", "issues"})
	if err != nil || cmd == nil {
		t.Fatalf("sprint issues not found: %v", err)
	}
	if f := cmd.Flags().Lookup("format"); f == nil {
		t.Fatal("sprint issues --format flag not registered")
	}
}

func TestSprintIssuesHasStatusFlag(t *testing.T) {
	cmd, _, _ := commands.RootCmd.Find([]string{"sprint", "issues"})
	if f := cmd.Flags().Lookup("status"); f == nil {
		t.Fatal("sprint issues --status flag not registered")
	}
}

func TestSprintIssuesHasLimitFlag(t *testing.T) {
	cmd, _, _ := commands.RootCmd.Find([]string{"sprint", "issues"})
	if f := cmd.Flags().Lookup("limit"); f == nil {
		t.Fatal("sprint issues --limit flag not registered")
	}
}
