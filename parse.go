package main

import (
	"fmt"
	"os"

	"golang.org/x/mod/modfile"
)

func parseMod(file string) (*modfile.File, error) {
	buf, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading go.mod: %w", err)
	}

	mod, err := modfile.Parse(file, buf, nil)
	if err != nil {
		return nil, fmt.Errorf("error parsing go.mod: %w", err)
	}

	if config.Verbose {
		out.Println("parsed go.mod:", mod.Go.Version)
	}
	return mod, nil
}

func saveMod(file string, mod *modfile.File) error {
	buf, err := mod.Format()
	if err != nil {
		return fmt.Errorf("error formatting go.mod: %w", err)
	}

	err = os.WriteFile(file, buf, 0644)
	if err != nil {
		return fmt.Errorf("error writing go.mod: %w", err)
	}
	return nil
}
