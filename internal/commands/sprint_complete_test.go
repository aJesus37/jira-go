package commands_test

import (
	"testing"

	"github.com/user/jira-go/internal/commands"
)

func TestSprintCompleteHasMoveToFlag(t *testing.T) {
	cmd, _, err := commands.RootCmd.Find([]string{"sprint", "complete"})
	if err != nil || cmd == nil {
		t.Fatalf("sprint complete not found: %v", err)
	}
	if f := cmd.Flags().Lookup("move-to"); f == nil {
		t.Fatal("sprint complete --move-to flag not registered")
	}
}
