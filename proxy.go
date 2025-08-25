package main

import (
	"bufio"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

type GoProxy struct {
	baseURL string
}

func NewGoProxy(baseURL string) *GoProxy {
	if baseURL == "" {
		baseURL = "https://proxy.golang.org"
	}

	baseURL = strings.TrimSuffix(baseURL, "/")
	return &GoProxy{baseURL: baseURL}
}

func isPreRelease(version string) bool {
	return strings.Contains(version, "-")
}

// FetchVersions fetches the list of versions for a given module from the Go proxy.
// It returns a slice of module.Version structs sorted in descending order.
// Pre-release versions will return pre-release versions
func (p *GoProxy) FetchVersions(modName string, version string) ([]module.Version, error) {
	versions := []module.Version{}

	modName, err := module.EscapePath(modName)
	if err != nil {
		return nil, fmt.Errorf("failed to escape module path: %w", err)
	}

	resp, err := http.Get(fmt.Sprintf("%s/%s/@v/list", p.baseURL, modName))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch versions: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// skip pre-release versions
		if !isPreRelease(version) && isPreRelease(line) {
			continue
		}

		// skip lower versions
		if semver.Compare(version, line) >= 0 {
			continue
		}

		v := module.Version{
			Path:    modName,
			Version: strings.TrimSpace(line),
		}

		versions = append(versions, v)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	module.Sort(versions)
	slices.Reverse(versions)

	return versions, nil
}
