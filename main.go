package main

import (
	"flag"
	"os"

	"golang.org/x/mod/modfile"
)

func main() {
	goBinary := os.Getenv("GOVERSION")
	if goBinary == "" {
		goBinary = "go"
	}

	var runTests bool
	flag.BoolVar(&runTests, "test", false, "run tests for each dependency upgrade")
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

			if success && runTests {
				err = cmd(goBinary, "test", "./...")
				if err != nil {
					printerr("tests failed, reverting go.mod")
					save(lastMod)
					success = false
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
		println("Problems:")
		for _, m := range badModules {
			println(m)
		}
	}
}
