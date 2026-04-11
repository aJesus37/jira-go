package models

import "testing"

func TestIssueGetAllParticipants(t *testing.T) {
	issue := Issue{
		Owners: []User{
			{AccountID: "acc1", DisplayName: "Alice"},
		},
		Assignee: &User{AccountID: "acc2", DisplayName: "Bob"},
	}
	participants := issue.GetAllParticipants()
	if len(participants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(participants))
	}
}

func TestIssueGetAllParticipantsDeduplicates(t *testing.T) {
	issue := Issue{
		Owners: []User{
			{AccountID: "acc1", DisplayName: "Alice"},
		},
		Assignee: &User{AccountID: "acc1", DisplayName: "Alice Clone"}, // same ID
	}
	participants := issue.GetAllParticipants()
	if len(participants) != 1 {
		t.Errorf("expected 1 (deduplicated), got %d", len(participants))
	}
}
