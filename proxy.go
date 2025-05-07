package main

import (
	"bufio"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"golang.org/x/mod/module"
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

// FetchVersions fetches the list of versions for a given module from the Go proxy.
// It returns a slice of module.Version structs sorted in descending order.
func (p *GoProxy) FetchVersions(modName string) ([]module.Version, error) {
	versions := []module.Version{}

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
