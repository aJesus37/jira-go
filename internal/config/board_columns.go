package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ColumnConfig struct {
	Visible bool `json:"visible"`
	Width   int  `json:"width"`
	Order   int  `json:"order"`
}

type BoardColumnPrefs map[string]ColumnConfig

func GetBoardColumnsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "jira-go", "board-columns.json")
}

func LoadBoardColumns(projectKey string) (BoardColumnPrefs, error) {
	path := GetBoardColumnsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(BoardColumnPrefs), nil
		}
		return nil, err
	}

	var allPrefs map[string]BoardColumnPrefs
	if err := json.Unmarshal(data, &allPrefs); err != nil {
		return make(BoardColumnPrefs), nil
	}

	if prefs, ok := allPrefs[projectKey]; ok {
		return prefs, nil
	}
	return make(BoardColumnPrefs), nil
}

func SaveBoardColumns(projectKey string, prefs BoardColumnPrefs) error {
	path := GetBoardColumnsPath()

	// Load existing
	var allPrefs map[string]BoardColumnPrefs
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &allPrefs)
	} else {
		allPrefs = make(map[string]BoardColumnPrefs)
	}

	allPrefs[projectKey] = prefs

	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0755)

	out, _ := json.MarshalIndent(allPrefs, "", "  ")
	return os.WriteFile(path, out, 0600)
}
