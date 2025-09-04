package main

import (
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
)

// upgradeModule attempts to upgrade a single module.
// It returns the new modfile.File and a boolean indicating success.
func upgradeModule(proxy *GoProxy, r *modfile.Require, okMod *modfile.File) (*modfile.File, bool) {
	var success bool
	var newMod *modfile.File

	out.BeginPreformatted(config.GoBinary, "get", r.Mod.Path)
	defer func() { out.EndPreformattedCond(!success) }()

	versions, err := proxy.FetchVersions(r.Mod.Path, r.Mod.Version)
	if err != nil {
		out.Error("failed to fetch versions:", err.Error())
		return okMod, false
	}
	if len(versions) == 0 {
		return okMod, true // No new versions, consider it a success
	}

	for vi, version := range versions {
		if vi >= config.Retries {
			out.Error("too many failed attempts, giving up")
			break
		}
		err := cmd(config.GoBinary, "get", r.Mod.Path+"@"+version.Version)
		newMod = parseMod(config.GoModSrc)
		if err != nil {
			out.Error("upgrade unsuccessful, reverting go.mod")
			saveMod(config.GoModDst, okMod)
		} else if newMod == nil || newMod.Go == nil {
			out.Error("parsing error, reverting go.mod")
			saveMod(config.GoModDst, okMod)
		} else if strings.TrimSuffix(okMod.Go.Version, ".0") != strings.TrimSuffix(newMod.Go.Version, ".0") {
			out.Error("upgrade changes required Go version", okMod.Go.Version, " => ", newMod.Go.Version, "reverting go.mod")
			saveMod(config.GoModDst, okMod)
		} else {
			out.Println("compare", okMod.Go.Version, " => ", newMod.Go.Version)
			success = true
			return newMod, true
		}
	}
	return okMod, false
}

// runCommands executes post-upgrade commands.
// Returns true if all commands succeed.
func runCommands(okMod *modfile.File) bool {
	for _, c := range config.Commands {
		if c == "" {
			continue
		}

		var success bool
		out.BeginPreformatted(c)
		if err := cmds(c); err != nil {
			out.Error("tests failed, reverting go.mod")
			saveMod(config.GoModDst, okMod)
			success = false
		}
		out.EndPreformattedCond(!success)
		if !success {
			return false
		}
	}
	return true
}

func process(original *modfile.File) []Result {
	var results []Result
	proxy := NewGoProxy("")
	okMod := parseMod(config.GoModSrc)

	for _, r := range original.Require {
		if r.Indirect {
			continue
		}

		var newMod *modfile.File
		newMod, upgradeSuccess := upgradeModule(proxy, r, okMod)

		if upgradeSuccess {
			if !runCommands(newMod) {
				upgradeSuccess = false
				// Revert to the last known good mod file
				saveMod(config.GoModDst, okMod)
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
