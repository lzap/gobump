package main

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
)

// attemptUpgrade tries to upgrade a module to a specific version.
func attemptUpgrade(modulePath, version string) (*modfile.File, error) {
	err := cmd(config.GoBinary, "get", modulePath+"@"+version)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %w", err)
	}
	return parseMod(config.GoModSrc)
}

// validateUpgrade checks if the upgrade is valid.
func validateUpgrade(originalMod, newMod *modfile.File) error {
	if newMod == nil || newMod.Go == nil {
		return fmt.Errorf("parsing error")
	}
	if strings.TrimSuffix(originalMod.Go.Version, ".0") != strings.TrimSuffix(newMod.Go.Version, ".0") {
		return fmt.Errorf("upgrade changes required Go version %s => %s", originalMod.Go.Version, newMod.Go.Version)
	}
	return nil
}

// upgradeModule attempts to upgrade a single module.
// The bool is success; the third return is true when the proxy listed no newer versions (no go get run).
func upgradeModule(proxy *GoProxy, r *modfile.Require, okMod *modfile.File) (*modfile.File, bool, bool) {
	var success bool
	var noProxyVersions bool
	out.BeginPreformatted(config.GoBinary, "get", r.Mod.Path)
	defer out.EndPreformattedCond(!success)

	versions, err := proxy.FetchVersions(r.Mod.Path, r.Mod.Version)
	if err != nil {
		out.Error("failed to fetch versions:", err.Error())
		return okMod, success, false
	}
	if len(versions) == 0 {
		success = true
		noProxyVersions = true
		return okMod, success, noProxyVersions
	}

	for vi, version := range versions {
		if vi >= config.Retries {
			out.Error("too many failed attempts, giving up")
			break
		}

		newMod, err := attemptUpgrade(r.Mod.Path, version.Version)
		if err != nil {
			out.Error("upgrade unsuccessful, reverting go.mod")
			if err := saveMod(config.GoModDst, okMod); err != nil {
				out.Error("failed to revert go.mod:", err.Error())
			}
			continue
		}

		if err := validateUpgrade(okMod, newMod); err != nil {
			out.Error(fmt.Sprintf("%s; reverting go.mod", err.Error()))
			if err := saveMod(config.GoModDst, okMod); err != nil {
				out.Error("failed to revert go.mod:", err.Error())
			}
			continue
		}

		if config.Verbose {
			out.Println("compare", okMod.Go.Version, " => ", newMod.Go.Version)
		}

		if !runCommands(okMod) {
			continue
		}

		success = true
		return newMod, success, false
	}
	return okMod, success, false
}

// runCommands executes post-upgrade commands against the current go.mod on disk
// (expected to match a successful upgrade). On failure it restores revertTo.
func runCommands(revertTo *modfile.File) bool {
	for _, c := range config.Commands {
		if c == "" {
			continue
		}
		out.BeginPreformatted(c)
		if err := cmds(c); err != nil {
			out.Error("tests failed, reverting go.mod")
			if err := saveMod(config.GoModDst, revertTo); err != nil {
				out.Error("failed to revert go.mod:", err.Error())
			}
			out.EndPreformattedCond(false)
			return false
		}
		out.EndPreformattedCond(true)
	}
	return true
}

func process(original *modfile.File) []Result {
	var results []Result
	proxy := NewGoProxy(config.ModuleProxy)
	okMod, err := parseMod(config.GoModSrc)
	if err != nil {
		out.Fatal(err.Error(), ERR_PARSE)
	}

	perDepGit := perDependencyGitEnabled()

	dependencies := original.Require
	if len(config.Dependencies) > 0 {
		dependencies = []*modfile.Require{}
		for _, r := range original.Require {
			for _, d := range config.Dependencies {
				if r.Mod.Path == d {
					dependencies = append(dependencies, r)
				}
			}
		}
	}

	for _, r := range dependencies {
		if r.Indirect {
			continue
		}

		excluded := false
		if slices.Contains(config.Exclude, r.Mod.Path) {
			results = append(results, Result{
				ModulePath:    r.Mod.Path,
				VersionBefore: r.Mod.Version,
				VersionAfter:  r.Mod.Version,
				Success:       false,
				Excluded:      true,
			})
			excluded = true
		}
		if excluded {
			continue
		}

		newMod, upgradeSuccess, noProxyVersions := upgradeModule(proxy, r, okMod)

		versionAfter := r.Mod.Version
		if newMod != nil {
			mi := slices.IndexFunc(newMod.Require, func(re *modfile.Require) bool {
				return re.Mod.Path == r.Mod.Path
			})
			if mi != -1 {
				versionAfter = newMod.Require[mi].Mod.Version
			}
		}

		if perDepGit {
			if !upgradeSuccess {
				if err := gitResetHardHEAD(); err != nil {
					out.Error("git reset --hard HEAD failed:", err.Error())
				}
			} else if versionAfter != r.Mod.Version && gitWorktreeDiffersFromHEAD() {
				if err := gitCommitDependencyBump(r.Mod.Path, versionAfter); err != nil {
					out.Error("git commit failed:", err.Error())
				}
			}
		}

		result := Result{
			ModulePath:      r.Mod.Path,
			VersionBefore:   r.Mod.Version,
			VersionAfter:    versionAfter,
			NoProxyVersions: noProxyVersions,
		}

		if upgradeSuccess {
			okMod = newMod
			result.Success = true
		} else {
			result.Success = false
		}

		results = append(results, result)
	}

	slices.SortFunc(results, func(a, b Result) int {
		return strings.Compare(a.ModulePath, b.ModulePath)
	})

	return results
}
