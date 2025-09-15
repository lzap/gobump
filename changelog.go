package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// GitHubCommit represents a single commit in the GitHub API response.
type GitHubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"commit"`
}

// GitHubCompareResponse represents the response from the GitHub compare API.
type GitHubCompareResponse struct {
	Commits []GitHubCommit `json:"commits"`
}

// GistFile represents a file in a Gist.
type GistFile struct {
	Content string `json:"content"`
}

// GistRequest represents the request to create a Gist.
type GistRequest struct {
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
}

// GistResponse represents the response from creating a Gist.
type GistResponse struct {
	HTMLURL string `json:"html_url"`
}

// getChangelog fetches the changelog for a module from GitHub.
func getChangelog(modulePath, fromVersion, toVersion string) (string, error) {
	parts := strings.Split(modulePath, "/")
	if len(parts) < 3 || parts[0] != "github.com" {
		return "", nil
	}
	owner := parts[1]
	repo := parts[2]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/compare/%s...%s", owner, repo, fromVersion, toVersion)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch changelog from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned non-200 status: %s", resp.Status)
	}

	var compareResp GitHubCompareResponse
	if err := json.NewDecoder(resp.Body).Decode(&compareResp); err != nil {
		return "", fmt.Errorf("failed to decode GitHub API response: %w", err)
	}

	var changelog strings.Builder
	for _, commit := range compareResp.Commits {
		firstLine := strings.Split(commit.Commit.Message, "\n")[0]
		changelog.WriteString(fmt.Sprintf("* %s: %s (%s)\n", commit.SHA[:7], firstLine, commit.Commit.Author.Name))
	}

	return changelog.String(), nil
}

// createGist creates a new GitHub Gist with the provided content.
func createGist(token, description, content string) (string, error) {
	gistRequest := GistRequest{
		Description: description,
		Public:      false,
		Files: map[string]GistFile{
			"changelog.md": {
				Content: content,
			},
		},
	}

	requestBody, err := json.Marshal(gistRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Gist request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/gists", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create Gist request: %w", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send Gist request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitHub API returned non-201 status for Gist creation: %s", resp.Status)
	}

	var gistResponse GistResponse
	if err := json.NewDecoder(resp.Body).Decode(&gistResponse); err != nil {
		return "", fmt.Errorf("failed to decode Gist response: %w", err)
	}

	return gistResponse.HTMLURL, nil
}

// PrintChangelogs prints the changelogs for all updated modules.
func PrintChangelogs(results []Result) {

	if config.ChangelogDest == "gist" {
		var fullChangelog strings.Builder
		fullChangelog.WriteString("# GoBump Changelog\n\n")
		for _, result := range results {
			if result.Success && result.VersionBefore != result.VersionAfter {
				fullChangelog.WriteString(fmt.Sprintf("## %s\n\n", result.ModulePath))
				fullChangelog.WriteString(fmt.Sprintf("Updated from `%s` to `%s`\n\n", result.VersionBefore, result.VersionAfter))
				changelog, err := getChangelog(result.ModulePath, result.VersionBefore, result.VersionAfter)
				if err != nil {
					fullChangelog.WriteString(fmt.Sprintf("Failed to get changelog: %s\n\n", err.Error()))
				} else if changelog == "" {
					fullChangelog.WriteString("No commits found between versions.\n\n")
				} else {
					fullChangelog.WriteString(changelog + "\n")
				}
			}
		}
		gistURL, err := createGist(os.Getenv("GITHUB_TOKEN"), "GoBump Dependency Changelog", fullChangelog.String())
		if err != nil {
			out.Error("Failed to create Gist:", err.Error())
		} else {
			out.Println("\nChangelog Gist created:", gistURL)
		}
	} else {
		sb := strings.Builder{}
		for _, result := range results {
			if result.Success && result.VersionBefore != result.VersionAfter {
				sb.WriteString(fmt.Sprintf("\nModule: %s\n", result.ModulePath))
				sb.WriteString(fmt.Sprintf("Updated from %s to %s\n", result.VersionBefore, result.VersionAfter))
				changelog, err := getChangelog(result.ModulePath, result.VersionBefore, result.VersionAfter)
				if err != nil {
					sb.WriteString(fmt.Sprintf("Failed to get changelog: %s\n", err.Error()))
				} else if changelog == "" {
					sb.WriteString("No commits found between versions.\n")
				} else {
					sb.WriteString(changelog)
				}
			}
		}

		if config.ChangelogDest == "stdout" {
			out.Println("\nGit Changelogs:")
			out.Println(sb.String())
		} else if config.ChangelogDest != "" {
			err := os.WriteFile(config.ChangelogDest, []byte(sb.String()), 0644)
			if err != nil {
				out.Error("Failed to write changelog to file:", err.Error())
			}
		}
	}
}
