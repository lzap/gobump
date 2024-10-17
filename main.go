package main

import (
	"os"

	"golang.org/x/mod/modfile"
)

func main() {
	goBinary := os.Getenv("GOVERSION")
	if goBinary == "" {
		goBinary = "go"
	}

	original := parse()
	modules := []*modfile.File{original}

	for _, r := range original.Require {
		if !r.Indirect {
			lastMod := modules[len(modules)-1]
			err := cmd(goBinary, "get", r.Mod.Path+"@latest")
			newMod := parse()
			if err != nil {
				printerr("upgrade unsuccessful, reverting go.mod")
				save(lastMod)
			} else if lastMod.Go.Version != newMod.Go.Version {
				printerr("upgrade changes required Go version, reverting go.mod")
				save(lastMod)
			} else {
				modules = append(modules, newMod)
			}
		}
	}
}
