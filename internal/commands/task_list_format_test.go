package commands_test

import (
	"testing"

	"github.com/aJesus37/jira-go/internal/commands"
)

func TestTaskListHasFormatFlag(t *testing.T) {
	listCmd, _, err := commands.RootCmd.Find([]string{"task", "list"})
	if err != nil {
		t.Fatalf("task list not found: %v", err)
	}
	if f := listCmd.Flags().Lookup("format"); f == nil {
		t.Fatal("task list --format flag not registered")
	}
}

func TestTaskListHasAgeFlag(t *testing.T) {
	listCmd, _, err := commands.RootCmd.Find([]string{"task", "list"})
	if err != nil {
		t.Fatalf("task list not found: %v", err)
	}
	if f := listCmd.Flags().Lookup("age"); f == nil {
		t.Fatal("task list --age flag not registered")
	}
}
