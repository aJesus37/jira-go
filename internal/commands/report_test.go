package commands_test

import (
	"github.com/aJesus37/jira-go/internal/commands"
	"testing"
)

func TestReportCommandRegistered(t *testing.T) {
	reportCmd, _, err := commands.RootCmd.Find([]string{"report"})
	if err != nil || reportCmd == nil {
		t.Fatalf("report command not registered: %v", err)
	}
}

func TestReportHasFormatFlag(t *testing.T) {
	reportCmd, _, _ := commands.RootCmd.Find([]string{"report"})
	if f := reportCmd.Flags().Lookup("format"); f == nil {
		t.Fatal("report --format flag not registered")
	}
}

func TestReportHasSprintFlag(t *testing.T) {
	reportCmd, _, _ := commands.RootCmd.Find([]string{"report"})
	if f := reportCmd.Flags().Lookup("sprint"); f == nil {
		t.Fatal("report --sprint flag not registered")
	}
}
