package main

type Result struct {
	ModulePath      string
	Success         bool
	VersionBefore   string
	VersionAfter    string
	Excluded        bool
	NoProxyVersions bool // proxy returned no semver newer than current (no go get attempted)
}

// resultsHaveErrors reports whether any module that was considered for update
// ended in a failed state (excluded modules are ignored).
func resultsHaveErrors(results []Result) bool {
	for _, r := range results {
		if r.Excluded {
			continue
		}
		if !r.Success {
			return true
		}
	}
	return false
}
