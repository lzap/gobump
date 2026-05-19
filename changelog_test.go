package main

import (
	"strings"
	"testing"
)

func TestFormatModuleChangelogNonGitHub(t *testing.T) {
	got := formatModuleChangelog("golang.org/x/mod", "v0.22.0", "v0.24.0")
	if !strings.Contains(got, "golang.org/x/mod") {
		t.Fatalf("missing module path: %q", got)
	}
	if !strings.Contains(got, "v0.22.0") || !strings.Contains(got, "v0.24.0") {
		t.Fatalf("missing versions: %q", got)
	}
	if !strings.Contains(got, "No commits found between versions") {
		t.Fatalf("expected no commits message for non-GitHub module: %q", got)
	}
}

func TestShortCommitSHA(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"abcdefg", "abcdefg"},
		{"abcdefgh", "abcdefg"},
	}
	for _, tt := range tests {
		if got := shortCommitSHA(tt.in); got != tt.want {
			t.Errorf("shortCommitSHA(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
