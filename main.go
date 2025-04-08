package main

import (
	"flag"
	"fmt"
	"os"

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
	var goodModules, badModules []string

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

			if success {
				modules = append(modules, newMod)
				goodModules = append(goodModules, r.Mod.Path)
			} else {
				badModules = append(badModules, r.Mod.Path)
			}
		}
	}

	println()
	if len(badModules) == 0 {
		println("All modules are up to date")
	} else {
		println("Up to date:")
		for _, m := range goodModules {
			println(m)
		}
		println()
		println("Unable to bump version:")
		for _, m := range badModules {
			println(m)
		}
	}

	if dryRun {
		save(original)
	}
}
