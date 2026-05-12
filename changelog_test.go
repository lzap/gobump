package main

import "testing"

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
