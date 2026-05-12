package main

import "os"

// githubToken returns a token for GitHub API calls (changelog compare, gist).
// GH_TOKEN is set by the GitHub CLI and many CI setups; GITHUB_TOKEN is the Actions default.
func githubToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}
