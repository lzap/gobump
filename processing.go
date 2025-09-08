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
func upgradeModule(proxy *GoProxy, r *modfile.Require, okMod *modfile.File) (*modfile.File, bool) {
	var success bool
	out.BeginPreformatted(config.GoBinary, "get", r.Mod.Path)
	defer out.EndPreformattedCond(success)

	versions, err := proxy.FetchVersions(r.Mod.Path, r.Mod.Version)
	if err != nil {
		out.Error("failed to fetch versions:", err.Error())
		return okMod, success
	}
	if len(versions) == 0 {
		success = true
		return okMod, success
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
			out.Error(err.Error(), ", reverting go.mod")
			if err := saveMod(config.GoModDst, okMod); err != nil {
				out.Error("failed to revert go.mod:", err.Error())
			}
			continue
		}

		out.Println("compare", okMod.Go.Version, " => ", newMod.Go.Version)
		success = true
		return newMod, success
	}
	return okMod, success
}

// runCommands executes post-upgrade commands.
func runCommands(mod *modfile.File) bool {
	for _, c := range config.Commands {
		if c == "" {
			continue
		}
		out.BeginPreformatted(c)
		if err := cmds(c); err != nil {
			out.Error("tests failed, reverting go.mod")
			if err := saveMod(config.GoModDst, mod); err != nil {
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
	proxy := NewGoProxy("")
	okMod, err := parseMod(config.GoModSrc)
	if err != nil {
		out.Fatal(err.Error(), ERR_PARSE)
	}

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
		for _, ex := range config.Exclude {
			if r.Mod.Path == ex {
				results = append(results, Result{
					ModulePath:    r.Mod.Path,
					VersionBefore: r.Mod.Version,
					VersionAfter:  r.Mod.Version,
					Success:       false,
					Excluded:      true,
				})
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		newMod, upgradeSuccess := upgradeModule(proxy, r, okMod)

		if upgradeSuccess {
			if !runCommands(newMod) {
				upgradeSuccess = false
				if err := saveMod(config.GoModDst, okMod); err != nil {
					out.Error("failed to revert go.mod:", err.Error())
				}
			}
		}

		result := Result{
			ModulePath:    r.Mod.Path,
			VersionBefore: r.Mod.Version,
		}

		if upgradeSuccess {
			okMod = newMod
			result.Success = true
		} else {
			result.Success = false
		}

		if newMod != nil {
			mi := slices.IndexFunc(newMod.Require, func(re *modfile.Require) bool {
				return re.Mod.Path == r.Mod.Path
			})
			if mi != -1 {
				newRequire := newMod.Require[mi]
				result.VersionAfter = newRequire.Mod.Version
			}
		}
		results = append(results, result)
	}

	slices.SortFunc(results, func(a, b Result) int {
		return strings.Compare(a.ModulePath, b.ModulePath)
	})

	return results
}
