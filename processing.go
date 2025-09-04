package main

import (
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
)

func process(original *modfile.File) []Result {
	var results []Result
	proxy := NewGoProxy("")

	okMod := parseMod(config.GoModSrc)
	var newMod *modfile.File

	for _, r := range original.Require {
		if !r.Indirect {
			success := true

			out.BeginPreformatted(config.GoBinary, "get", r.Mod.Path)
			versions, err := proxy.FetchVersions(r.Mod.Path, r.Mod.Version)
			if err != nil {
				out.Error("failed to fetch versions:", err.Error())
				out.EndPreformattedCond(false)
				continue
			}
			if len(versions) == 0 {
				continue
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
					success = false
				} else if newMod == nil || newMod.Go == nil {
					out.Error("parsing error, reverting go.mod")
					saveMod(config.GoModDst, okMod)
					success = false
				} else if strings.TrimSuffix(okMod.Go.Version, ".0") != strings.TrimSuffix(newMod.Go.Version, ".0") {
					out.Error("upgrade changes required Go version", okMod.Go.Version, " => ", newMod.Go.Version, "reverting go.mod")
					saveMod(config.GoModDst, okMod)
					success = false
				} else {
					out.Println("compare", okMod.Go.Version, " => ", newMod.Go.Version)
					success = true
					break
				}
			}
			out.EndPreformattedCond(!success)

			if success {
				for _, c := range config.Commands {
					if c == "" {
						continue
					}

					out.BeginPreformatted(c)
					if err := cmds(c); err != nil {
						out.Error("tests failed, reverting go.mod")
						saveMod(config.GoModDst, okMod)
						success = false
					}
					out.EndPreformattedCond(!success)
				}
			}

			result := Result{
				ModulePath:    r.Mod.Path,
				VersionBefore: r.Mod.Version,
			}

			if success {
				okMod = newMod
				result.Success = true
			} else {
				result.Success = false
			}

			if newMod != nil {
				mi := slices.IndexFunc(newMod.Require, func(re *modfile.Require) bool {
					return re.Mod.Path == r.Mod.Path
				})
				newRequire := newMod.Require[mi]
				result.VersionAfter = newRequire.Mod.Version
			}

			results = append(results, result)
		}
	}

	slices.SortFunc(results, func(a, b Result) int {
		return strings.Compare(a.ModulePath, b.ModulePath)
	})

	return results
}
