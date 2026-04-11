package config

import (
	"os"
	"testing"
)

func TestLoadBoardColumns(t *testing.T) {
	prefs, err := LoadBoardColumns("SWCSIRT")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if prefs == nil {
		t.Errorf("expected empty map, got nil")
	}
}

func TestSaveBoardColumns(t *testing.T) {
	prefs := BoardColumnPrefs{
		"REVISAR": {Visible: true, Width: 25},
	}
	err := SaveBoardColumns("SWCSIRT", prefs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	loaded, err := LoadBoardColumns("SWCSIRT")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if loaded["REVISAR"].Width != 25 {
		t.Errorf("expected width 25, got %d", loaded["REVISAR"].Width)
	}

	// Cleanup
	os.Remove(GetBoardColumnsPath())
}
