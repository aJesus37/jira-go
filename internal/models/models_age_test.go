package models_test

import (
	"testing"
	"time"

	"github.com/aJesus37/jira-go/internal/models"
)

func TestDaysInStatusUsesUpdated(t *testing.T) {
	issue := models.Issue{
		Updated: time.Now().Add(-3 * 24 * time.Hour),
	}
	days := issue.DaysInStatus()
	if days < 2 || days > 4 {
		t.Fatalf("expected ~3 days, got %d", days)
	}
}

func TestDaysInStatusZeroWhenNoTime(t *testing.T) {
	issue := models.Issue{}
	if issue.DaysInStatus() != 0 {
		t.Fatal("expected 0 when no time set")
	}
}
