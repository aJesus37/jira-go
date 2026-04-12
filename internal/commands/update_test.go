package commands

import (
	"testing"
)

func TestDetermineAssetName(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{"linux", "amd64", "jira_linux_amd64.tar.gz"},
		{"darwin", "amd64", "jira_darwin_amd64.tar.gz"},
		{"darwin", "arm64", "jira_darwin_arm64.tar.gz"},
		{"windows", "amd64", "jira_windows_amd64.zip"},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"/"+tt.goarch, func(t *testing.T) {
			got := determineAssetName(tt.goos, tt.goarch)
			if got != tt.want {
				t.Errorf("determineAssetName(%q, %q) = %q, want %q",
					tt.goos, tt.goarch, got, tt.want)
			}
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0.2.0", "v0.2.0"},
		{"v0.2.0", "v0.2.0"},
		{"1.0.0", "v1.0.0"},
		{"v1.0.0", "v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeVersion(tt.input)
			if got != tt.want {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
