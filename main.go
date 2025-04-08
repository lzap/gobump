package main

import (
	"flag"
	"fmt"
	"os"
	"slices"

	"golang.org/x/mod/modfile"
)

type stringSlice []string

func (i *stringSlice) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	dryRun  bool
	verbose bool
)

func main() {
	goBinary := os.Getenv("GOVERSION")
	if goBinary == "" {
		goBinary = "go"
	}

	var commands stringSlice
	flag.BoolVar(&dryRun, "dry-run", false, "revert to original go.mod after running")
	flag.BoolVar(&verbose, "verbose", false, "print more information including stderr of executed commands")
	flag.Var(&commands, "exec", "exec command for each individual bump, can be used multiple times")
	flag.Parse()

	original := parse()
	modules := []*modfile.File{original}
	var results []Result

	for _, r := range original.Require {
		if !r.Indirect {
			success := true
			lastMod := modules[len(modules)-1]
			err := cmd(goBinary, "get", r.Mod.Path+"@latest")
			newMod := parse()
			if err != nil {
				printerr("upgrade unsuccessful, reverting go.mod")
				save(lastMod)
				success = false
			} else if lastMod.Go.Version != newMod.Go.Version {
				printerr("upgrade changes required Go version, reverting go.mod")
				save(lastMod)
				success = false
			}

			if success {
				for _, c := range commands {
					if err := cmds(c); err != nil {
						printerr("tests failed, reverting go.mod")
						save(lastMod)
						success = false
					}
				}
			}

			mi := slices.IndexFunc(newMod.Require, func(re *modfile.Require) bool {
				return re.Mod.Path == r.Mod.Path
			})
			newRequire := newMod.Require[mi]

			result := Result{
				ModulePath:    r.Mod.Path,
				MinGoVersion:  newMod.Go.Version,
				VersionBefore: r.Mod.Version,
				VersionAfter:  newRequire.Mod.Version,
			}
			if success {
				modules = append(modules, newMod)
				result.Success = true
			} else {
				result.Success = false
			}
			results = append(results, result)
		}
	}

	println()
	println("Summary:")
	for _, r := range results {
		action := "skipped"
		if r.Success {
			if r.VersionAfter == r.VersionBefore {
				action = "no action"
			} else {
				action = "upgraded"
			}
		}
		if r.VersionAfter != "" && r.VersionAfter != r.VersionBefore && action != "skipped" {
			println(r.ModulePath, action, fmt.Sprintf("mingo:%s", r.MinGoVersion), r.VersionBefore, "->", r.VersionAfter)
		} else {
			println(r.ModulePath, action, fmt.Sprintf("mingo:%s", r.MinGoVersion))

		}
	}

	if dryRun {
		save(original)
	}
}
